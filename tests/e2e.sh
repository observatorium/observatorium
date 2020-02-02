#!/bin/bash

set -e
set -o pipefail

kind() {
    curl -LO https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/linux/amd64/kubectl
    curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-linux-amd64
    chmod +x kind kubectl
    ./kind create cluster
}

deploy() {
    ./kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
    ./kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml
    ./kubectl create ns minio || true
    ./kubectl create ns observatorium || true
    ./kubectl apply -f environments/dev/manifests/
}

run_test() {
    ./kubectl wait --for=condition=available --timeout=10m -n observatorium deploy/minio || (./kubectl get pods --all-namespaces && exit 1)
    ./kubectl wait --for=condition=available --timeout=10m -n observatorium deploy/thanos-query || (./kubectl get pods --all-namespaces && exit 1)

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
