SHELL=/usr/bin/env bash -o pipefail

# Image URL to use all building/pushing image targets
IMG ?= quay.io/observatorium/observatorium-operator:latest
BIN_DIR ?= $(shell pwd)/tmp/bin

CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen

# Install CRDs into a cluster
install: manifests
	kubectl apply -f deploy/crds

# Uninstall CRDs from a cluster
uninstall: manifests
	kubectl delete -f  deploy/crds

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=deploy/crds

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(CONTROLLER_GEN): vendor $(BIN_DIR)
	go build -mod=vendor -o $@ sigs.k8s.io/controller-tools/cmd/controller-gen

