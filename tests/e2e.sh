#!/bin/bash

set -e
set -o pipefail

ARTIFACT_DIR="${ARTIFACT_DIR:-/tmp/artifacts}"
KUBECTL="${KUBECTL:-./kubectl}"
OS_TYPE=$(echo `uname -s` | tr '[:upper:]' '[:lower:]')

kind() {
    curl -LO https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/$OS_TYPE/amd64/kubectl
    curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.8.1/kind-$OS_TYPE-amd64
    chmod +x kind kubectl
    ./kind create cluster
}

dex() {
    $KUBECTL create ns dex || true
    $KUBECTL apply -f environments/dev/manifests/dex-secret.yaml
    $KUBECTL apply -f environments/dev/manifests/dex-pvc.yaml
    $KUBECTL apply -f environments/dev/manifests/dex-deployment.yaml
    $KUBECTL apply -f environments/dev/manifests/dex-service.yaml
    # Observatorium needs the Dex API to be ready for authentication to work and thus for the tests to pass.
    $KUBECTL wait --for=condition=available --timeout=10m -n dex deploy/dex || (must_gather "$ARTIFACT_DIR" && exit 1)
}

deploy() {
    $KUBECTL apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
    $KUBECTL apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml
    $KUBECTL create ns observatorium-minio || true
    $KUBECTL create ns observatorium || true
    dex
    $KUBECTL apply -f environments/dev/manifests/
}

wait_for_cr() {
    observatorium_cr_status=""
    target_status="Finished"
    timeout=$true
    interval=0
    intervals=600
    while [ $interval -ne $intervals ]; do
      echo "Waiting for" $1 "currentStatus="$observatorium_cr_status
      observatorium_cr_status=$($KUBECTL -n observatorium get observatoria.core.observatorium.io $1 -o=jsonpath='{.status.conditions[*].currentStatus}')
      if [ "$observatorium_cr_status" = "$target_status" ]; then
        echo $1 CR status is now: $observatorium_cr_status
	    timeout=$false
	    break
	  fi
	  sleep 5
	  interval=$((interval+5))
    done

    if [ $timeout ]; then
      echo "Timeout waiting for" $1 "CR status to be " $target_status
      exit 1
    fi
}

run_test() {
    local suffix
    while [ $# -gt 0 ]; do
        case $1 in
            --tls)
                suffix=-tls
                ;;
        esac
        shift
    done

    $KUBECTL wait --for=condition=available --timeout=10m -n observatorium-minio deploy/minio || (must_gather "$ARTIFACT_DIR" && exit 1)
    $KUBECTL wait --for=condition=available --timeout=10m -n observatorium deploy/observatorium-xyz-thanos-query-frontend || (must_gather "$ARTIFACT_DIR" && exit 1)
    $KUBECTL wait --for=condition=available --timeout=10m -n observatorium deploy/observatorium-xyz-loki-query-frontend || (must_gather "$ARTIFACT_DIR" && exit 1)
    $KUBECTL apply -f tests/manifests/observatorium-xyz-tls-configmap.yaml
    $KUBECTL apply -f tests/manifests/observatorium-up-metrics"$suffix".yaml

    sleep 5

    # This should wait for ~2min for the job to finish.
    $KUBECTL wait --for=condition=complete --timeout=5m -n default job/observatorium-up-metrics"$suffix" || (must_gather "$ARTIFACT_DIR" && exit 1)
    $KUBECTL apply -f tests/manifests/observatorium-up-logs"$suffix".yaml

    sleep 5

    # This should wait for ~2min for the job to finish.
    $KUBECTL wait --for=condition=complete --timeout=5m -n default job/observatorium-up-logs"$suffix" || (must_gather "$ARTIFACT_DIR" && exit 1)
}

must_gather() {
    local artifact_dir="$1"

    for namespace in default dex observatorium observatorium-minio; do
        mkdir -p "$artifact_dir/$namespace"

        for name in $($KUBECTL get pods -n "$namespace" -o jsonpath='{.items[*].metadata.name}') ; do
            $KUBECTL -n "$namespace" describe pod "$name" > "$artifact_dir/$namespace/$name.describe"
            $KUBECTL -n "$namespace" get pod "$name" -o yaml > "$artifact_dir/$namespace/$name.yaml"

            for initContainer in $($KUBECTL -n "$namespace" get po "$name" -o jsonpath='{.spec.initContainers[*].name}') ; do
                $KUBECTL -n "$namespace" logs "$name" -c "$initContainer" > "$artifact_dir/$namespace/$name-$initContainer.logs"
            done

            for container in $($KUBECTL -n "$namespace" get po "$name" -o jsonpath='{.spec.containers[*].name}') ; do
                $KUBECTL -n "$namespace" logs "$name" -c "$container" > "$artifact_dir/$namespace/$name-$container.logs"
            done
        done
    done

    $KUBECTL describe nodes > "$artifact_dir/nodes"
    $KUBECTL get pods --all-namespaces > "$artifact_dir/pods"
    $KUBECTL get deploy --all-namespaces > "$artifact_dir/deployments"
    $KUBECTL get statefulset --all-namespaces > "$artifact_dir/statefulsets"
    $KUBECTL get services --all-namespaces > "$artifact_dir/services"
    $KUBECTL get endpoints --all-namespaces > "$artifact_dir/endpoints"
}

case $1 in
kind)
    kind
    ;;

deploy)
    deploy
    ;;

test)
    shift
    run_test "$@"
    ;;

deploy-operator)
    deploy_operator
    ;;

delete-cr)
    delete_cr
    ;;

*)
    echo "usage: $(basename "$0") { kind | deploy | test | deploy-operator | delete-cr }"
    ;;
esac
