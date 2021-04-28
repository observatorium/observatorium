TMP_DIR := $(shell pwd)/tmp
BIN_DIR ?= $(TMP_DIR)/bin
GOBIN ?= $(BIN_DIR)
include .bingo/Variables.mk

SHELL=/usr/bin/env bash -o pipefail
CERT_DIR ?= $(TMP_DIR)/certs
CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
GENERATE_TLS_CERT ?= $(BIN_DIR)/generate-tls-cert

DEPLOYMENTS ?= deployments
JSONNET_SRC = $(shell find . -type f -not -path './*vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))

all: generate validate

vendor: $(JB)
	cd $(DEPLOYMENTS) && $(JB) install

.PHONY: fmt
fmt: $(JSONNETFMT) $(JSONNET_SRC)
	$(JSONNETFMT) -n 2 --max-blank-lines 2 --string-style s --comment-style s -i $(JSONNET_SRC)

.PHONY: lint
lint: $(JSONNET_LINT) vendor
	echo ${JSONNET_SRC} | xargs -n 1 -- $(JSONNET_LINT) -J $(DEPLOYMENTS)/vendor

.PHONY: generate
generate: $(DEPLOYMENTS)/environments/base/manifests $(DEPLOYMENTS)/environments/dev/manifests $(DEPLOYMENTS)/environments/local/manifests

.PHONY: validate
validate: $(KUBEVAL) generate
	$(KUBEVAL) --ignore-missing-schemas $(DEPLOYMENTS)/environments/base/manifests/*.yaml $(DEPLOYMENTS)/environments/dev/manifests/*.yaml $(DEPLOYMENTS)/environments/local/manifests/*.yaml $(DEPLOYMENTS)/tests/manifests/*.yaml

$(DEPLOYMENTS)/environments/base/manifests: $(DEPLOYMENTS)/environments/base/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf $(DEPLOYMENTS)/environments/base/manifests
	-mkdir $(DEPLOYMENTS)/environments/base/manifests
	cd $(DEPLOYMENTS) && $(JSONNET) -J vendor -m environments/base/manifests environments/base/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find $(DEPLOYMENTS)/environments/base/manifests -type f ! -name '*.yaml' -delete

$(DEPLOYMENTS)/environments/dev/manifests: $(DEPLOYMENTS)/environments/dev/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf $(DEPLOYMENTS)/environments/dev/manifests
	-mkdir $(DEPLOYMENTS)/environments/dev/manifests
	cd $(DEPLOYMENTS) && $(JSONNET) -J vendor -m environments/dev/manifests environments/dev/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find $(DEPLOYMENTS)/environments/dev/manifests -type f ! -name '*.yaml' -delete

$(DEPLOYMENTS)/environments/local/manifests: $(DEPLOYMENTS)/environments/local/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf $(DEPLOYMENTS)/environments/local/manifests
	-mkdir $(DEPLOYMENTS)/environments/local/manifests
	cd $(DEPLOYMENTS) && $(JSONNET) -J vendor -m environments/local/manifests environments/local/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find $(DEPLOYMENTS)/environments/local/manifests -type f ! -name '*.yaml' -delete

$(DEPLOYMENTS)/tests/manifests: $(DEPLOYMENTS)/tests/main.jsonnet vendor generate-cert $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf $(DEPLOYMENTS)/tests/manifests
	-mkdir $(DEPLOYMENTS)/tests/manifests
	cd $(DEPLOYMENTS) && $(JSONNET) -J vendor -m tests/manifests tests/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find $(DEPLOYMENTS)/tests/manifests -type f ! -name '*.yaml' -delete

.PHONY: generate-cert
# Generate TLS certificates for local development.
generate-cert: $(GENERATE_TLS_CERT) | $(CERT_DIR)
	cd $(CERT_DIR) && $(GENERATE_TLS_CERT) -server-common-name=observatorium-xyz-observatorium-api.observatorium.svc.cluster.local -server-sans localhost,127.0.0.1,dex.dex.svc.cluster.local,observatorium-xyz-observatorium-api.observatorium.svc.cluster.local

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(CERT_DIR):
	mkdir -p $(CERT_DIR)

# Not managed by Bingo directly, as it requires the -tags tools flag
# TODO(bwplotka): Fix with https://github.com/bwplotka/bingo/issues/46.
$(GENERATE_TLS_CERT): $(BINGO_DIR)/api.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/generate-tls-cert"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=api.mod -tags=tools -o=$(BIN_DIR)/generate-tls-cert "github.com/observatorium/observatorium/test/tls"

