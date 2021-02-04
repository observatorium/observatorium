#!/bin/bash

set -e
set -o pipefail

trap 'kill 0' SIGTERM

KUBECTL="${KUBECTL:-kubectl}"
KIND="${KIND:-kind}"

if [ ! $(command -v "$KIND") ]; then
  echo "Cannot find or execute KIND binary $KIND, you can override it by setting the KIND env variable"
  exit 1
fi

if [ ! $(command -v "$KUBECTL") ]; then
  echo "Cannot find or execute Kubectl binary $KUBECTL, you can override it by setting the KUBECTL env variable"
  exit 1
fi

setup() {
  mkdir -p tmp/bin
  echo "-------------------------------------------"
  echo "- Downloading ORY Hydra...  -"
  echo "-------------------------------------------"
  curl -L "https://github.com/ory/hydra/releases/download/v1.9.1/hydra_1.9.1-sqlite_linux_64bit.tar.gz" | tar -xzf - -C tmp/bin hydra

  echo "-------------------------------------------"
  echo "- Cloning observatorium/token-refresher and building...  -"
  echo "-------------------------------------------"
  git clone https://github.com/observatorium/token-refresher tmp/token-refresher
  cd tmp/token-refresher
  make build
  mv ./token-refresher ../bin/
  cd -

  echo "-------------------------------------------"
  echo "- Downloading Prometheus...  -"
  echo "-------------------------------------------"
  curl -L "https://github.com/prometheus/prometheus/releases/download/v2.24.1/prometheus-2.24.1.linux-amd64.tar.gz" | tar -xzf - -C tmp prometheus-2.24.1.linux-amd64/prometheus
  mv ./tmp/prometheus-2.24.1.linux-amd64/prometheus ./tmp/bin/

  echo "-------------------------------------------"
  echo "- Pulling docker image for Grafana...  -"
  echo "-------------------------------------------"
  docker pull grafana/grafana:7.3.7

  echo "-------------------------------------------"
  echo "- Creating KIND cluster...  -"
  echo "-------------------------------------------"
  $KIND create cluster
}

deploy() {
  # Hydra
  (DSN=memory ./tmp/bin/hydra serve all --dangerous-force-http --config ./configs/hydra.yaml &> /dev/null) &
  echo "-------------------------------------------"
  echo "- Waiting for Hydra to come up...  -"
  echo "-------------------------------------------"
  until curl --output /dev/null --silent --fail --insecure http://127.0.0.1:4444/.well-known/openid-configuration; do
    printf '.'
    sleep 1
  done
  echo ""

  curl \
    --output /dev/null --silent \
    --header "Content-Type: application/json" \
    --request POST \
    --data '{"audience": ["observatorium"], "client_id": "user", "client_secret": "secret", "grant_types": ["client_credentials"], "token_endpoint_auth_method": "client_secret_basic"}' \
    http://127.0.0.1:4445/clients

  # MinIO
  echo "-------------------------------------------"
  echo "- Deploying MinIO...  -"
  echo "-------------------------------------------"
  $KUBECTL create namespace observatorium-minio --dry-run=client -o yaml | $KUBECTL apply -f -

  $KUBECTL apply -f ./manifests/minio-pvc.yaml
  $KUBECTL apply -f ./manifests/minio-deployment.yaml
  $KUBECTL apply -f ./manifests/minio-service.yaml

  echo "-------------------------------------------"
  echo "- Waiting for MinIO to come up...  -"
  echo "-------------------------------------------"
  $KUBECTL wait --for=condition=available --timeout=5m -n observatorium-minio deploy/minio

  # Observatorium
  echo "-------------------------------------------"
  echo "- Deploying Observatorium...  -"
  echo "-------------------------------------------"
  $KUBECTL create namespace observatorium --dry-run=client -o yaml | $KUBECTL apply -f -
  $KUBECTL apply -f ./manifests/

  echo "-------------------------------------------"
  echo "- Waiting for Observatorium to come up...  -"
  echo "-------------------------------------------"
  $KUBECTL wait --for=condition=available --timeout=5m -n observatorium deploy/observatorium-xyz-thanos-query-frontend
  $KUBECTL wait --for=condition=available --timeout=5m -n observatorium deploy/observatorium-xyz-observatorium-api
  ($KUBECTL port-forward -n observatorium svc/observatorium-xyz-observatorium-api 8443:8080 &> /dev/null) &

  # Token Refresher
  echo "-------------------------------------------"
  echo "- Starting Token Refresher proxy...  -"
  echo "-------------------------------------------"
  (./tmp/bin/token-refresher \
    --oidc.issuer-url=http://172.17.0.1:4444/ \
    --oidc.client-id=user \
    --oidc.client-secret=secret \
    --oidc.audience=observatorium \
    --url=http://127.0.0.1:8443 &> /dev/null) &
  sleep 1

  # Prometheus
  echo "-------------------------------------------"
  echo "- Starting Prometheus...  -"
  echo "-------------------------------------------"
  (./tmp/bin/prometheus --config.file=./configs/prom.yaml --storage.tsdb.path=tmp/data/ &> /dev/null) &

  # Grafana
  echo "-------------------------------------------"
  echo "- Starting Grafana using docker...  -"
  echo "-------------------------------------------"
  mkdir -p tmp/grafana
  (docker run -p 3000:3000 --user $(id -u) --volume "$PWD/tmp/grafana:/var/lib/grafana" grafana/grafana:7.3.7 &> /dev/null) &
  echo "Open http://localhost:3000 in your browser. Add Prometheus datasource with endpoint http://172.17.0.1:8080/api/metrics/v1/test-oidc."
}

case $1 in
setup)
    setup
    ;;

deploy)
    deploy
    ;;

help)
    echo "usage: $(basename "$0") { setup | deploy }"
    ;;

*)
    setup
    deploy
    ;;
esac

wait
