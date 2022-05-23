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
  echo "-------------------------------------------"
  echo "- Creating KIND cluster...  -"
  echo "-------------------------------------------"
  $KIND create cluster
}

deploy() {
  echo "-------------------------------------------"
  echo "- Creating namespaces...  -"
  echo "-------------------------------------------"
  $KUBECTL create namespace hydra --dry-run=client -o yaml | $KUBECTL apply -f -
  $KUBECTL create namespace observatorium --dry-run=client -o yaml | $KUBECTL apply -f -
  $KUBECTL create namespace observability --dry-run=client -o yaml | $KUBECTL apply -f -
  $KUBECTL create namespace observatorium-minio --dry-run=client -o yaml | $KUBECTL apply -f -

  echo "-------------------------------------------"
  echo "- Deploying MinIO...  -"
  echo "-------------------------------------------"
  $KUBECTL apply -f ./manifests/minio
  $KUBECTL wait --for=condition=available --timeout=5m -n observatorium-minio deploy/minio

  echo "-------------------------------------------"
  echo "- Deploying Hydra...  -"
  echo "-------------------------------------------"
  $KUBECTL apply -f ./manifests/hydra

  echo "-------------------------------------------"
  echo "- Deploying kube-prometheus...  -"
  echo "-------------------------------------------"
  $KUBECTL apply --server-side -f ./manifests/kube-prometheus/setup
  until $KUBECTL get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done
  $KUBECTL apply --server-side -f ./manifests/kube-prometheus/

  echo "-------------------------------------------"
  echo "- Deploying Jaeger Operator...  -"
  echo "-------------------------------------------"
  $KUBECTL apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml
  $KUBECTL wait --for=condition=available --timeout=5m -n cert-manager deploy/cert-manager
  $KUBECTL wait --for=condition=available --timeout=5m -n cert-manager deploy/cert-manager-cainjector
  $KUBECTL wait --for=condition=available --timeout=5m -n cert-manager deploy/cert-manager-webhook
  $KUBECTL apply -f https://github.com/jaegertracing/jaeger-operator/releases/download/v1.32.0/jaeger-operator.yaml -n observability

  echo "-------------------------------------------"
  echo "- Deploying OpenTelemetry Operator...  -"
  echo "-------------------------------------------"
  $KUBECTL apply -f https://github.com/open-telemetry/opentelemetry-operator/releases/latest/download/opentelemetry-operator.yaml

  echo "-------------------------------------------"
  echo "- Deploying Observatorium...  -"
  echo "-------------------------------------------"
  $KUBECTL wait --for=condition=available --timeout=5m -n observability deploy/jaeger-operator
  $KUBECTL wait --for=condition=available --timeout=5m -n opentelemetry-operator-system deploy/opentelemetry-operator-controller-manager
  $KUBECTL wait --for=condition=available --timeout=10m -n hydra deploy/hydra
  $KUBECTL apply -f ./manifests/token-refresher
  $KUBECTL apply -f ./manifests/observatorium

  echo "-------------------------------------------"
  echo "- Waiting for Observatorium to come up...  -"
  echo "-------------------------------------------"
  $KUBECTL wait --for=condition=available --timeout=5m -n observatorium deploy/observatorium-xyz-thanos-query-frontend
  $KUBECTL wait --for=condition=available --timeout=5m -n observatorium deploy/observatorium-xyz-observatorium-api
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
