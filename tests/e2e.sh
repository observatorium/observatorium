#!/bin/bash

set -e
set -o pipefail

KUBECTL="${KUBECTL:-./kubectl}"

kind() {
    curl -LO https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/linux/amd64/kubectl
    curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-linux-amd64
    chmod +x kind kubectl
    ./kind create cluster
}

deploy() {
    $KUBECTL apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
    $KUBECTL apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml
    $KUBECTL create ns minio || true
    $KUBECTL create ns observatorium || true
    $KUBECTL apply -f environments/dev/manifests/
}

run_test() {
    $KUBECTL wait --for=condition=available --timeout=10m -n observatorium deploy/minio || ($KUBECTL get pods --all-namespaces && exit 1)
    $KUBECTL wait --for=condition=available --timeout=10m -n observatorium deploy/observatorium-xyz-thanos-query || ($KUBECTL get pods --all-namespaces && exit 1)

    $KUBECTL apply -f tests/manifests/observatorium-up.yaml

    sleep 5

    # This should wait for ~2min for the job to finish.
    $KUBECTL wait --for=condition=complete --timeout=5m -n default job/observatorium-up || ($KUBECTL get pods --all-namespaces && exit 1)
}

case $1 in
kind)
    kind
    ;;

deploy)
    deploy
    ;;

test)
    run_test
    ;;

*)
    echo "usage: $(basename "$0") { kind | deploy | test }"
    ;;
esac
