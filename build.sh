#!/usr/bin/env bash

# This script uses arg $1 (name of *.jsonnet file to use) to generate the manifests/*.yaml files.

set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

./build_dev.sh

# Make sure to start with a clean 'manifests' dir
rm -rf jsonnet/environments/base/manifests
mkdir jsonnet/environments/base/manifests

jsonnet -J jsonnet/vendor -m jsonnet/environments/base/manifests jsonnet/environments/base/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
find jsonnet/environments/base/manifests -type f ! -name '*.yaml' -delete
