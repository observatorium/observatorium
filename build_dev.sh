#!/usr/bin/env bash

# This script uses arg $1 (name of *.jsonnet file to use) to generate the manifests/*.yaml files.

set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

# Make sure to start with a clean 'manifests' dir
rm -rf environments/dev/manifests
mkdir environments/dev/manifests

jsonnet -J operator/jsonnet/vendor -m environments/dev/manifests environments/dev/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
find environments/dev/manifests -type f ! -name '*.yaml' -delete

rm -rf tests/manifests
mkdir tests/manifests

jsonnet -J operator/jsonnet/vendor -m tests/manifests tests/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
find tests/manifests -type f ! -name '*.yaml' -delete

rm -rf example/manifests
mkdir example/manifests

jsonnet -J operator/jsonnet/vendor example/main.jsonnet | gojsontoyaml >example/manifests/observatorium.yaml
find example/manifests -type f ! -name '*.yaml' -delete
