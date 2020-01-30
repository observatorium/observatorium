#!/bin/bash

set -e
set -x
set -o pipefail

curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.6.1/kind-linux-amd64
chmod +x kind kubectl
./kind create cluster --image kindest/node:v1.15.6
