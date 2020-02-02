# /usr/bin/env bash

set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

curl -q https://raw.githubusercontent.com/openshift/cluster-monitoring-operator/master/manifests/0000_50_cluster_monitoring_operator_04-config.yaml | gojsontoyaml -yamltojson | jq -r '.data["metrics.yaml"]' | gojsontoyaml -yamltojson | jq -r '.matches' > environments/openshift/metrics.json
