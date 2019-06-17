#!/usr/bin/env bash

# This script uses arg $1 (name of *.jsonnet file to use) to generate the manifests/*.yaml files.

set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

# Make sure to start with a clean 'manifests' dir
rm -rf environments/kubernetes/manifests
mkdir environments/kubernetes/manifests

jsonnet -J vendor -m environments/kubernetes/manifests environments/kubernetes/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml; rm -f {}' -- {}

# Make sure to start with a clean 'manifests' dir
rm -rf environments/openshift/manifests
mkdir environments/openshift/manifests

jsonnet -J vendor environments/openshift/main.jsonnet | gojsontoyaml > environments/openshift/manifests/thanos-template.yaml
