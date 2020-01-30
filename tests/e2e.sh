#!/bin/bash

set -e
set -o pipefail

kind() {
    curl -LO https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/linux/amd64/kubectl
    curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.6.1/kind-linux-amd64
    chmod +x kind kubectl
    ./kind create cluster --image kindest/node:v1.15.6
}

deploy() {
    ./kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
    ./kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml
    ./kubectl create ns minio || true
    ./kubectl create ns observatorium || true
    ./kubectl apply -f environments/testing/
    ./kubectl apply -f environments/kubernetes/manifests/

    ./kubectl scale statefulset -n observatorium telemeter-server --replicas 0  # Remove once we move Telemeter out of this repo
    ./kubectl scale deployment -n observatorium jaeger-all-in-one --replicas 0  # Don't need it in CI
    ./kubectl scale deployment -n observatorium thanos-querier --replicas 1     # One is enough for CI
}

run_test() {
    ./kubectl wait --for=condition=available --timeout=10m -n minio deploy/minio || (./kubectl get pods --all-namespaces && exit 1)
    ./kubectl wait --for=condition=available --timeout=10m -n observatorium deploy/thanos-querier || (./kubectl get pods --all-namespaces && exit 1)

    ./kubectl apply -f tests/manifests/observatorium-up.yaml

    sleep 5

    # This should wait for ~2min for the job to finish.
    ./kubectl wait --for=condition=complete --timeout=5m -n default job/observatorium-up || (./kubectl get pods --all-namespaces && exit 1)
}

case $1 in
    kind) 
        kind;;

    deploy)
        deploy;;

    test)
        run_test;;

    *)
        echo "usage: $(basename "$0") { kind | deploy | test }";;
esac
