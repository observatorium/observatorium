# Observatorium - Getting Started

This tutorial will help you get started with Observatorium. We will be running the whole Observatorium stack in a local Kubernetes cluster. Observatorium uses OIDC (OpenID Connect) for authentication, so we will deploy our own OIDC provider. After we have Observatorium up and running, we will push some metrics as a tenant from outside the cluster using remote write.

If you just want to run Observatroium locally and get started quickly, take a look at the bash script `quickstart.sh`.

## Prerequisites

- Clone this repo
  ```bash
  git clone https://github.com/observatorium/deployments.git
  cd deployments/environments/local
  ```

- Create a temporary folder
  ```bash
  mkdir -p tmp/bin
  ```

### Local K8s cluster - KIND

We need to run a local Kubernetes cluster to run our stack in. We are going to use KIND (Kubernetes in Docker) for that. You can follow the [Getting Started](https://kind.sigs.k8s.io/docs/user/quick-start/) guide to install it.

Let's create a new cluster in kind.
```bash
kind create cluster
```

### OIDC Provider - Hydra

We need to run our own OIDC provider to handle authentication. We are going to use ORY Hydra for that. First download and extract the binary release from [here](https://github.com/ory/hydra/releases/download/v1.9.1/hydra_1.9.1-sqlite_linux_64bit.tar.gz).

```bash
curl -L "https://github.com/ory/hydra/releases/download/v1.9.1/hydra_1.9.1-sqlite_linux_64bit.tar.gz" | tar -xzf - -C tmp/bin hydra
```

The configuration file for `hydra` is present in `configs/hydra.yaml`.

```yaml
strategies:
  access_token: jwt
urls:
  self:
    issuer: http://172.17.0.1:4444/ 
```


We will be running `hydra` outside the cluster to simulate an external tenant, but we need to access `hydra` from inside the cluster. As our K8s cluster is running inside Docker containers, we can use the ip address of `docker0` interface to access host from inside the containers. In most cases it will be `172.17.0.1` but you can find yours using

```bash
ip -o -4 addr list docker0 | awk '{print $4}' | cut -d/ -f1
```

If this value is not `172.17.0.1`, you need to update the `issuer` URL in the config file for hydra.

Next step is to run `hydra`.

```bash
DSN=memory ./tmp/bin/hydra serve all --dangerous-force-http --config ./configs/hydra.yaml
```

Now that we have our OIDC provider running we need create a client to authenticate as. To create a new client in `hydra` run this

```bash
curl \
    --header "Content-Type: application/json" \
    --request POST \
    --data '{"audience": ["observatorium"], "client_id": "user", "client_secret": "secret", "grant_types": ["client_credentials"], "token_endpoint_auth_method": "client_secret_basic"}' \
    http://127.0.0.1:4445/clients
```

## Deploying Observatorium

We will deploy the Observatorium using Kubernetes manifests generated from jsonnet. If the IP of `docker0` interface was different then the default in the above steps, you will need to update the `tenant.issuerURL` with correct IP address and run `make generate` to recreate the manifests.

Let's deploy `minio` first, as Thanos and Loki have a dependency on it.

```bash
kubectl create ns observatorium-minio
kubectl apply -f ./manifests/minio-pvc.yaml
kubectl apply -f ./manifests/minio-deployment.yaml
kubectl apply -f ./manifests/minio-service.yaml
```

Wait for `minio` to come up.

```bash
kubectl wait --for=condition=available -n observatorium-minio deploy/minio
```

Afetr `minio` starts running, deploy everything else.

```bash
kubectl create ns observatorium
kubectl apply -f ./manifests/
```

Wait for everything in `observatorium` namespace to come up before moving forward.

## Writing data from outside

The Observatorium API allows you to push timeseries data using Prometheus remote write protocol. The write endpoint is protected, so we need to authenticate as a tenant. The authentication is handled by the OIDC Provider we started earlier. If you take a look at `manifests/api-secret.yaml`, you can see that we have configured a single tenant `test-oidc`. The remote write endpoint is `/api/metrics/v1/<tenant-id>/api/v1/receive`.

### Token refresher

The token issued by the OIDC providers often have a small validity, but it can be automatically refreshed using a refresh token. As Prometheus doesn't natively support refresh token flow, we use a proxy in-between to do exactly that for us. Take a look at the [GitHub repo](https://github.com/observatorium/token-refresher) to read more about it.

- Clone the repo locally and build the binary using `make build`.
  ```bash
  git clone https://github.com/observatorium/token-refresher tmp/token-refresher
  cd tmp/token-refresher
  make build
  mv ./token-refresher ../bin/
  cd -
  ```
- We need to put up a proxy in front of the Observatorium API so we first need to expose it first. We will use `kubectl port-forward` for that.
  ```bash
  kubectl port-forward -n observatorium svc/observatorium-xyz-observatorium-api 8443:8080
  ```
- Now that Observatorium API is listening on `localhost:8443`, we will run the token refresher to forward traffic to it. You may have to replace the `--oidc.issuer-url` with appropriate value if the IP of the `docker0` interface is different.
  ```bash
  ./tmp/bin/token-refresher --oidc.issuer-url=http://172.17.0.1:4444/ --oidc.client-id=user --oidc.client-secret=secret --oidc.audience=observatorium --url=http://127.0.0.1:8443
  ```

### Push some metrics, shall we?

Take a look at the [Prometheus first steps](https://prometheus.io/docs/introduction/first_steps/) to get a quick overview of how to get started. By default Prometheus stores the data locally in form of TSDB blocks. We will configure it to remote write this data to Observatorium. As we know that the write endpoint is protected, we will write to the token-refresher proxy, and the proxy in turn will forward this data to the Observatorium API with proper tokens.

Download the Prometheus binary from [here](https://github.com/prometheus/prometheus/releases/download/v2.24.1/prometheus-2.24.1.linux-amd64.tar.gz).

```bash
curl -L "https://github.com/prometheus/prometheus/releases/download/v2.24.1/prometheus-2.24.1.linux-amd64.tar.gz" | tar -xzf - -C tmp prometheus-2.24.1.linux-amd64/prometheus
mv ./tmp/prometheus-2.24.1.linux-amd64/prometheus ./tmp/bin/
```

The Prometheus config file present in `configs/prom.yaml` looks like this:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: prom
  static_configs:
  - targets:
    - localhost:9090

remote_write:
- url: http://localhost:8080/api/metrics/v1/test-oidc/api/v1/receive
```

Notice that the remote_write url points to the token refresher, not the Observatorium API directly.

Next, let's run Prometheus with this config.

```bash
./tmp/bin/prometheus --config.file=./configs/prom.yaml --storage.tsdb.path=tmp/data/
```

## Querying data from Observatorium

We are going to use Grafana to query the data we wrote into Observatorium. To start Grafana in a docker container, run

```bash
docker run -p 3000:3000 grafana/grafana:7.3.7
```

- Now open your web browser and go to `http://localhost:3000`. the default username:password is `admin:admin`.
- Add a new Prometheus data source with url `http://172.17.0.1:8080/api/metrics/v1/test-oidc`. We are using `172.17.0.1` as the host because we are trying to access the `token-refresher` running on the host from Grafana, which is running inside a docker container.

You can now go the the `Explore` tab to run queries against the Observatorium API.
