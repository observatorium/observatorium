TMP_DIR := $(shell pwd)/tmp
BIN_DIR ?= $(TMP_DIR)/bin
GOBIN ?= $(BIN_DIR)
include .bingo/Variables.mk

SHELL=/usr/bin/env bash -o pipefail
CERT_DIR ?= $(TMP_DIR)/certs
CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
GENERATE_TLS_CERT ?= $(BIN_DIR)/generate-tls-cert

JSONNET_SRC = $(shell find . -type f -not -path './*vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))

all: generate validate

vendor: $(JB)
	$(JB) install

.PHONY: fmt
fmt: $(JSONNETFMT) $(JSONNET_SRC)
	$(JSONNETFMT) -n 2 --max-blank-lines 2 --string-style s --comment-style s -i $(JSONNET_SRC)

.PHONY: lint
lint: $(JSONNET_LINT) vendor
	echo ${JSONNET_SRC} | xargs -n 1 -- $(JSONNET_LINT) -J vendor

.PHONY: generate
generate: environments/base/manifests environments/dev/manifests

.PHONY: validate
validate: $(KUBEVAL) generate
	$(KUBEVAL) --ignore-missing-schemas environments/base/manifests/*.yaml environments/dev/manifests/*.yaml tests/manifests/*.yaml

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

tests/manifests: tests/main.jsonnet vendor generate-cert $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf tests/manifests
	-mkdir tests/manifests
	$(JSONNET) -J vendor -m tests/manifests tests/main.jsonnet | xargs -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find tests/manifests -type f ! -name '*.yaml' -delete

.PHONY: generate-cert
# Generate TLS certificates for local development.
generate-cert: $(GENERATE_TLS_CERT) | $(CERT_DIR)
	cd $(CERT_DIR) && $(GENERATE_TLS_CERT) -server-common-name=observatorium-xyz-observatorium-api.observatorium.svc.cluster.local -server-sans localhost,127.0.0.1,dex.dex.svc.cluster.local,observatorium-xyz-observatorium-api.observatorium.svc.cluster.local

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(CERT_DIR):
	mkdir -p $(CERT_DIR)

# Not managed by Bingo directly, as it requires the -tags tools flag
$(GENERATE_TLS_CERT):
	@echo "(re)installing $(GOBIN)/generate-tls-cert"
	@cd .bingo && $(GO) build -modfile=generate-tls-cert.mod -tags tools -o=$(GOBIN)/generate-tls-cert "github.com/observatorium/observatorium/test/tls"
