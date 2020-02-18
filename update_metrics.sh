#!/bin/bash

# Copy current file, as in-place doesn't work...
cp environments/openshift/metrics.json /tmp/metrics.json 

# Generate the metrics.json from the ConfigMap in YAML that the ClusterMonitoringOperator has.
cat ~/src/github.com/openshift/cluster-monitoring-operator/manifests/0000_50_cluster_monitoring_operator_04-config.yaml | gojsontoyaml -yamltojson | jq -r '.data."metrics.yaml"' | gojsontoyaml -yamltojson | jq .matches > /tmp/matches-new.json 

# Merge both files and create a unique array which also sorts
jq -s 'add | unique' /tmp/matches-new.json /tmp/metrics.json > environments/openshift/metrics.json

