#!/usr/bin/env bash

# This script uses arg $1 (name of *.jsonnet file to use) to generate the manifests/*.yaml files.

set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

./build_dev.sh

# Make sure to start with a clean 'manifests' dir
rm -rf environments/base/manifests
mkdir environments/base/manifests

jsonnet -J operator/jsonnet/vendor -m environments/base/manifests environments/base/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
find environments/base/manifests -type f ! -name '*.yaml' -delete
