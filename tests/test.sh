#!/bin/bash

set -e
set -x
set -o pipefail

./kubectl wait --for=condition=available --timeout=10m -n minio deploy/minio || (./kubectl get pods --all-namespaces && exit 1)
./kubectl wait --for=condition=available --timeout=10m -n observatorium deploy/thanos-querier || (./kubectl get pods --all-namespaces && exit 1)

./kubectl apply -f tests/manifests/observatorium-up.yaml

sleep 5

# This should wait for ~2min for the job to finish.
./kubectl wait --for=condition=complete --timeout=5m -n default job/observatorium-up || (./kubectl get pods --all-namespaces && exit 1)
