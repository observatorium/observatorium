# Observatorium Operator
![Observatorium](https://avatars0.githubusercontent.com/u/51818702?s=100&v=4sanitize=true)

## Prelude

Check the following resources for more information about Observatorium:
* [Documentation repository](https://github.com/observatorium/docs/)
* [Observatorium API](https://github.com/observatorium/observatorium/)
* [Observatorium configuration repository](https://github.com/observatorium/configuration)
* [Locutus - The framework which the operator is based on](https://github.com/brancz/locutus)


## How to deploy - Kubernetes and OpenShift

In order to ease the installation of Observatorium, an operator is available.

### Prerequisites

#### Create Namespaces
```shell script
kubectl create namespace minio
kubectl create namespace observatorium
```


#### S3 storage endpoint and secret
For **testing purposes** you may use [minio](https://github.com/minio/minio) as describe below.
```shell script
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/environments/dev/manifests/minio-secret.yaml
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/environments/dev/manifests/minio-pvc.yaml
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/environments/dev/manifests/minio-deployment.yaml
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/environments/dev/manifests/minio-service.yaml
```

### Deployment

#### Prometheus CRDs - Kubernetes Only
You may skip this step if you are using OpenShift, in which the CRD is already available as a part of the monitoring stack.
```shell script
kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml
```

#### RBAC Configuration
```shell script
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/deploy/cluster_role.yaml
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/deploy/cluster_role_binding.yaml
```

#### Deploy Observatorium CRD and Operator
* In case you need to force a new image download (e.g. development environment), please refer to the [development section](#Development).
```shell script
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/deploy/crds/core.observatorium.io_observatoria.yaml
kubectl apply -f https://raw.githubusercontent.com/observatorium/configuration/master/deploy/operator.yaml
```

## Deploy an example CR
```shell script
kubectl apply -n observatorium -f https://raw.githubusercontent.com/observatorium/configuration/master/example/manifests/observatorium.yaml
```
Monitor the CR status and wait for status --> Finished
```shell script
kubectl -n observatorium get observatoria.core.observatorium.io observatorium-xyz -o=jsonpath='{.status.conditions[*].currentStatus}'

Finished
```

### Expected Result
```shell script
kubectl -n observatorium get all

NAME                                                                  READY   STATUS    RESTARTS   AGE
pod/observatorium-xyz-cortex-query-frontend-7bc9b64c65-5ph64          1/1     Running   0          59s
pod/observatorium-xyz-observatorium-api-gateway-56c7fbcc89-xzcp2      1/1     Running   0          62s
pod/observatorium-xyz-observatorium-api-gateway-thanos-query-5lqrxw   1/1     Running   0          60s
pod/observatorium-xyz-thanos-compact-0                                1/1     Running   0          58s
pod/observatorium-xyz-thanos-query-599859d6cb-ztxxg                   1/1     Running   0          57s
pod/observatorium-xyz-thanos-receive-controller-f89fdf5c7-x5pvc       1/1     Running   0          53s
pod/observatorium-xyz-thanos-receive-default-0                        1/1     Running   0          53s
pod/observatorium-xyz-thanos-receive-default-1                        1/1     Running   0          24s
pod/observatorium-xyz-thanos-receive-default-2                        1/1     Running   0          15s
pod/observatorium-xyz-thanos-rule-0                                   1/1     Running   0          52s
pod/observatorium-xyz-thanos-rule-1                                   1/1     Running   0          28s
pod/observatorium-xyz-thanos-store-shard-0-0                          1/1     Running   0          51s

NAME                                                               TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                         AGE
service/observatorium-xyz-cortex-query-frontend                    ClusterIP   172.30.113.249   <none>        9090/TCP                        59s
service/observatorium-xyz-observatorium-api-gateway                ClusterIP   172.30.240.202   <none>        8080/TCP                        62s
service/observatorium-xyz-observatorium-api-gateway-thanos-query   ClusterIP   172.30.210.14    <none>        10901/TCP,9090/TCP              60s
service/observatorium-xyz-thanos-compact                           ClusterIP   172.30.168.154   <none>        10902/TCP                       58s
service/observatorium-xyz-thanos-query                             ClusterIP   172.30.169.145   <none>        10901/TCP,9090/TCP              57s
service/observatorium-xyz-thanos-receive                           ClusterIP   172.30.124.6     <none>        10901/TCP,10902/TCP,19291/TCP   53s
service/observatorium-xyz-thanos-receive-controller                ClusterIP   172.30.194.82    <none>        8080/TCP                        54s
service/observatorium-xyz-thanos-receive-default                   ClusterIP   None             <none>        10901/TCP,10902/TCP,19291/TCP   54s
service/observatorium-xyz-thanos-rule                              ClusterIP   None             <none>        10901/TCP,10902/TCP             52s
service/observatorium-xyz-thanos-store-shard-0                     ClusterIP   None             <none>        10901/TCP,10902/TCP             51s

NAME                                                                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/observatorium-xyz-cortex-query-frontend                    1/1     1            1           59s
deployment.apps/observatorium-xyz-observatorium-api-gateway                1/1     1            1           62s
deployment.apps/observatorium-xyz-observatorium-api-gateway-thanos-query   1/1     1            1           60s
deployment.apps/observatorium-xyz-thanos-query                             1/1     1            1           57s
deployment.apps/observatorium-xyz-thanos-receive-controller                1/1     1            1           56s

NAME                                                                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/observatorium-xyz-cortex-query-frontend-7bc9b64c65                    1         1         1       59s
replicaset.apps/observatorium-xyz-observatorium-api-gateway-56c7fbcc89                1         1         1       62s
replicaset.apps/observatorium-xyz-observatorium-api-gateway-thanos-query-55999ff96c   1         1         1       60s
replicaset.apps/observatorium-xyz-thanos-query-599859d6cb                             1         1         1       57s
replicaset.apps/observatorium-xyz-thanos-receive-controller-f89fdf5c7                 1         1         1       56s

NAME                                                        READY   AGE
statefulset.apps/observatorium-xyz-thanos-compact           1/1     58s
statefulset.apps/observatorium-xyz-thanos-receive-default   3/3     53s
statefulset.apps/observatorium-xyz-thanos-rule              2/2     52s
statefulset.apps/observatorium-xyz-thanos-store-shard-0     1/1     51s
```

## Test

### Expose observatorium api gateway for external traffic
* In Kubernetes
```shell script
kubectl -n observatorium patch svc observatorium-xyz-observatorium-api-gateway --type='json' -p '[{"op":"replace","path":"/spec/type","value":"NodePort"}]'
```
* In OpenShift
```shell script
oc -n observatorium expose svc observatorium-xyz-observatorium-api-gateway
```

### (Option A) Transmit Metrics via Remote Write Client
```shell script
kubectl -n default apply -f https://raw.githubusercontent.com/observatorium/configuration/master/tests/manifests/observatorium-up.yaml
kubectl wait --for=condition=complete --timeout=5m -n default job/observatorium-up
````
Result
```shell script
job.batch/observatorium-up condition met
```

### (Option B) Configure Prometheus Remote Write
* Example taken from CRC (Openshift), Prometheus deployed as a part of the monitoring operator.
* Note: If this is applied to a separate cluster, the url should be dns resolvable.
```shell script
cat << EOF | kubectl -n openshift-monitoring apply -f -
apiVersion: v1
data:
  config.yaml: |
    prometheusK8s:
      remoteWrite:
      - url: http://observatorium-api-gateway-observatorium.apps-crc.testing/api/metrics/v1/write
      externalLabels:
        demo_spoke_cluster: observatorium_demo
kind: ConfigMap
metadata:
  name: cluster-monitoring-config
  namespace: openshift-monitoring
EOF
```

### Grafana
Data source is set to query the observatorium-api gateway, which proxies the request to thanos-query.
Thus, the data source URL: `http://observatorium-xyz-observatorium-api-gateway.observatorium.svc.cluster.local:8080/api/metrics/v1`

#### Deploy Grafana
```shell script
kubectl -n observatorium apply -f https://raw.githubusercontent.com/observatorium/configuration/master/doc/operator/grafana/grafana.yaml
kubectl -n observatorium apply -f https://raw.githubusercontent.com/observatorium/configuration/master/doc/operator/grafana/grafana-cm.yaml
kubectl -n observatorium apply -f https://raw.githubusercontent.com/observatorium/configuration/master/doc/operator/grafana/grafana-svc.yaml
```

#### Expose Grafana for external traffic
* In Kubernetes
```shell script
kubectl -n observatorium patch svc grafana --type='json' -p '[{"op":"replace","path":"/spec/type","value":"NodePort"}]'
```
* In OpenShift
```shell script
oc -n observatorium expose svc grafana
```

#### Browse Grafana
You should now be able to see the 'foo' metric generated by the up client you invoked beforehand.
![Multi Cluster Architecture](./grafana-observatorium.png)

## Development
* Create the operator while forcing Kubernetes / OpenShift to download the image.  
```shell script
curl https://raw.githubusercontent.com/observatorium/configuration/master/deploy/operator.yaml | \
sed 's/imagePullPolicy\: IfNotPresent/imagePullPolicy\: Always/g' > observatorium-operator.yaml && \
kubectl -n default create -f observatorium-operator.yaml && rm -f observatorium-operator.yaml
```