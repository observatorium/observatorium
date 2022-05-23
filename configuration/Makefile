TMP_DIR := $(shell pwd)/tmp
BIN_DIR ?= $(TMP_DIR)/bin
GOBIN ?= $(BIN_DIR)
include ../.bingo/Variables.mk

SHELL=/usr/bin/env bash -o pipefail
CERT_DIR ?= $(TMP_DIR)/certs
CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
GENERATE_TLS_CERT ?= $(BIN_DIR)/generate-tls-cert
RNDR ?= ../../rndr/.bin/rndr # DEBUG

JSONNET_SRC = $(shell find . -type f -not -path './*vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))

UNAME := $(shell uname)
ifeq ($(UNAME), Darwin)
    # macOS/BSD xargs not compatible with this Makefile (because of -S replsize default of 256).
    XARGS ?= gxargs
    ifeq ($(shell command -v "$(XARGS)"), "")
        $(error "Cannot find or execute gxargs binary, you can install it with `brew install findutils`")
    endif
else
    XARGS ?= xargs
endif

all: generate validate

render:
	@mkdir -p tests/.gen/kubernetes
	@$(RNDR) --template="rndr.yaml" kubernetes manifests \
 		 --values-file="tests/values.yaml" \
 		 -o "tests/.gen/kubernetes"

vendor: $(JB)
	$(JB) install

jsonnet-update: $(JB)
	$(JB) update

.PHONY: fmt
fmt: $(JSONNETFMT) $(JSONNET_SRC)
	$(JSONNETFMT) -n 2 --max-blank-lines 2 --string-style s --comment-style s -i $(JSONNET_SRC)

.PHONY: lint
lint: $(JSONNET_LINT) vendor
	echo ${JSONNET_SRC} | $(XARGS) -n 1 -- $(JSONNET_LINT) -J vendor

.PHONY: generate
generate: examples/base/manifests examples/dev/manifests examples/local/manifests

.PHONY: validate
validate: $(KUBEVAL) generate
	$(KUBEVAL) --ignore-missing-schemas -d examples/base/manifests/,examples/dev/manifests/,examples/local/manifests/,tests/manifests/

# Use Jsonnet to generate JSON versions of K8s manifests without .json extensions,
# then send all the JSON through xargs to be converted into YAML and given .yaml extension.
examples/base/manifests: examples/base/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	echo ecs "$OS"
	-make fmt
	-rm -rf examples/base/manifests
	-mkdir examples/base/manifests
	$(JSONNET) -J vendor -c -m examples/base/manifests examples/base/main.jsonnet | $(XARGS) -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find examples/base/manifests -type f ! -name '*.yaml' -delete

examples/dev/manifests: examples/dev/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf examples/dev/manifests
	-mkdir examples/dev/manifests
	$(JSONNET) -J vendor -c -m examples/dev/manifests examples/dev/main.jsonnet | $(XARGS) -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find examples/dev/manifests -type f ! -name '*.yaml' -delete

examples/local/manifests: examples/local/main.jsonnet vendor $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf examples/local/manifests
	-mkdir examples/local/manifests
	$(JSONNET) -J vendor -c -m examples/local/manifests examples/local/main.jsonnet | $(XARGS) -I{} $(XARGS_FLAGS) sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
	find examples/local/manifests -type f ! -name '*.yaml' -delete

tests/manifests: tests/main.jsonnet vendor generate-cert $(JSONNET_SRC) $(JSONNET) $(GOJSONTOYAML)
	-make fmt
	-rm -rf tests/manifests
	-mkdir tests/manifests
	$(JSONNET) -J vendor -c -m tests/manifests tests/main.jsonnet | $(XARGS) -I{} sh -c 'cat {} | $(GOJSONTOYAML) > {}.yaml' -- {}
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
# TODO(bwplotka): Fix with https://github.com/bwplotka/bingo/issues/46.
$(GENERATE_TLS_CERT): $(BINGO_DIR)/api.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/generate-tls-cert"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=api.mod -tags=tools -o=$(BIN_DIR)/generate-tls-cert "github.com/observatorium/api/test/tls"
