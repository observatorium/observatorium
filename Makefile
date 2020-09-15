TMP_DIR := $(shell pwd)/tmp
BIN_DIR ?= $(TMP_DIR)/bin
GOBIN ?= $(BIN_DIR)
include .bingo/Variables.mk

SHELL=/usr/bin/env bash -o pipefail

VERSION := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE := $(shell date -u +"%Y-%m-%d")
BUILD_TIMESTAMP := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_BRANCH := $(strip $(shell git rev-parse --abbrev-ref HEAD))
VCS_REF := $(strip $(shell [ -d .git ] && git rev-parse --short HEAD))
DOCKER_REPO ?= quay.io/observatorium/observatorium-operator

CERT_DIR ?= $(TMP_DIR)/certs

CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
GENERATE_TLS_CERT ?= $(BIN_DIR)/generate-tls-cert

.PHONY: generate-cert
# Generate TLS certificates for local development.
generate-cert: $(GENERATE_TLS_CERT) | $(CERT_DIR)
	cd $(CERT_DIR) && $(GENERATE_TLS_CERT) -server-common-name=observatorium-xyz-observatorium-api.observatorium.svc.cluster.local

# Not managed by Bingo directly, as it requires the -tags tools flag
GENERATE_TLS_CERT := $(GOBIN)/generate-tls-cert
$(GENERATE_TLS_CERT): .bingo/generate-tls-cert.mod
	@echo "(re)installing $(GOBIN)/generate-tls-cert"
	@cd .bingo && $(GO) build -modfile=generate-tls-cert.mod -tags tools -o=$(GOBIN)/generate-tls-cert "github.com/observatorium/observatorium/test/tls"

# Build the docker image
container-build:
	docker build --build-arg BUILD_DATE="$(BUILD_TIMESTAMP)" \
		--build-arg VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(VCS_REF)" \
		--build-arg VCS_BRANCH="$(VCS_BRANCH)" \
		--build-arg DOCKERFILE_PATH="/Dockerfile" \
		-t $(DOCKER_REPO):$(VCS_BRANCH)-$(BUILD_DATE)-$(VERSION) .

# Push the image
container-push: container-build
	docker tag $(DOCKER_REPO):$(VCS_BRANCH)-$(BUILD_DATE)-$(VERSION) $(DOCKER_REPO):latest
	docker push $(DOCKER_REPO):$(VCS_BRANCH)-$(BUILD_DATE)-$(VERSION)
	docker push $(DOCKER_REPO):latest

vendor: $(JB)
	$(JB) install

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(CERT_DIR):
	mkdir -p $(CERT_DIR)

JSONNET_SRC = $(shell find . -type f -not -path './*vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))

.PHONY: fmt
fmt: $(JSONNETFMT) $(JSONNET_SRC)
	$(JSONNETFMT) -n 2 --max-blank-lines 2 --string-style s --comment-style s -i $(JSONNET_SRC)

environments/base/manifests: environments/base/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf environments/base/manifests
	-mkdir environments/base/manifests
	$(JSONNET) -J vendor -m environments/base/manifests environments/base/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find environments/base/manifests -type f ! -name '*.yaml' -delete

environments/dev/manifests: environments/dev/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf environments/dev/manifests
	-mkdir environments/dev/manifests
	$(JSONNET) -J vendor -m environments/dev/manifests environments/dev/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find environments/dev/manifests -type f ! -name '*.yaml' -delete

example/manifests: example/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf example/manifests
	-mkdir example/manifests
	$(JSONNET) -J vendor example/main.jsonnet | $(GOJSONTOYAML) > example/manifests/observatorium.yaml
	find example/manifests -type f ! -name '*.yaml' -delete

tests/manifests: tests/main.jsonnet vendor generate-cert $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf tests/manifests
	-mkdir tests/manifests
	$(JSONNET) -J vendor -m tests/manifests tests/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find tests/manifests -type f ! -name '*.yaml' -delete
