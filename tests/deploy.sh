#!/bin/bash

set -e
set -x
set -o pipefail

./kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0servicemonitorCustomResourceDefinition.yaml
./kubectl apply -f https://raw.githubusercontent.com/coreos/kube-prometheus/master/manifests/setup/prometheus-operator-0prometheusruleCustomResourceDefinition.yaml
./kubectl create ns minio || true
./kubectl create ns observatorium || true
./kubectl apply -f environments/testing/
./kubectl apply -f environments/kubernetes/manifests/

./kubectl scale statefulset -n observatorium telemeter-server --replicas 0  # Remove once we move Telemeter out of this repo
./kubectl scale deployment -n observatorium jaeger-all-in-one --replicas 0  # Don't need it in CI
./kubectl scale deployment -n observatorium thanos-querier --replicas 1     # One is enough for CI
