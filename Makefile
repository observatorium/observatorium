SHELL=/usr/bin/env bash -o pipefail

VERSION := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE := $(shell date -u +"%Y-%m-%d")
BUILD_TIMESTAMP := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_BRANCH := $(strip $(shell git rev-parse --abbrev-ref HEAD))
VCS_REF := $(strip $(shell [ -d .git ] && git rev-parse --short HEAD))
DOCKER_REPO ?= quay.io/observatorium/observatorium-operator

TMP_DIR := $(shell pwd)/tmp
BIN_DIR ?= $(TMP_DIR)/bin
CERT_DIR ?= $(TMP_DIR)/certs

CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
JB ?= $(BIN_DIR)/jb
GENERATE_TLS_CERT ?= $(BIN_DIR)/generate-tls-cert

.PHONY: generate-cert
# Generate TLS certificates for local development.
generate-cert: $(GENERATE_TLS_CERT) | $(CERT_DIR)
	cd $(CERT_DIR) && $(GENERATE_TLS_CERT) -server-common-name=observatorium-xyz-observatorium-api.observatorium.svc.cluster.local -server-hosts=observatorium-xyz-observatorium-api.observatorium.svc.cluster.local

$(GENERATE_TLS_CERT): | $(BIN_DIR)
	# A thin wrapper around github.com/cloudflare/cfssl
	cd operator;  GO111MODULE="on" go build -tags tools -o $@ github.com/observatorium/observatorium/test/tls

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	cd operator; $(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=manifests/crds

# Run go fmt against code
fmt:
	cd operator; go fmt ./...

# Run go vet against code
vet:
	cd operator; go vet ./...

# Generate code
generate: $(CONTROLLER_GEN) environments/base/manifests environments/dev/manifests example/manifests tests/manifests
	cd operator; $(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

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

vendor-jsonnet: $(JB)
	cd operator/jsonnet; $(JB) install

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(CERT_DIR):
	mkdir -p $(CERT_DIR)

$(CONTROLLER_GEN): $(BIN_DIR)
	cd operator; GO111MODULE="on" go build -o $@ sigs.k8s.io/controller-tools/cmd/controller-gen

$(JB): $(BIN_DIR)
	cd operator; GO111MODULE="on" go build -o $@ github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb

JSONNET_SRC = $(shell find . -type f -not -path './*vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))

.PHONY: jsonnetfmt
jsonnetfmt: $(JSONNET_SRC)
	jsonnetfmt -n 2 --max-blank-lines 2 --string-style s --comment-style s -i $(JSONNET_SRC)

environments/base/manifests: environments/base/main.jsonnet $(JSONNET_SRC)
	-make jsonnetfmt
	-rm -rf environments/base/manifests
	-mkdir environments/base/manifests
	jsonnet -J operator/jsonnet/vendor -m environments/base/manifests environments/base/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find environments/base/manifests -type f ! -name '*.yaml' -delete

environments/dev/manifests: environments/dev/main.jsonnet $(JSONNET_SRC) vendor-jsonnet
	-make jsonnetfmt
	-rm -rf environments/dev/manifests
	-mkdir environments/dev/manifests
	jsonnet -J operator/jsonnet/vendor -m environments/dev/manifests environments/dev/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find environments/dev/manifests -type f ! -name '*.yaml' -delete

example/manifests: example/main.jsonnet $(JSONNET_SRC) vendor-jsonnet
	-make jsonnetfmt
	-rm -rf example/manifests
	-mkdir example/manifests
	jsonnet -J operator/jsonnet/vendor example/main.jsonnet | gojsontoyaml > example/manifests/observatorium.yaml
	find example/manifests -type f ! -name '*.yaml' -delete

tests/manifests: tests/main.jsonnet $(JSONNET_SRC) vendor-jsonnet
	-make jsonnetfmt
	-rm -rf tests/manifests
	-mkdir tests/manifests
	jsonnet -J operator/jsonnet/vendor -m tests/manifests tests/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find tests/manifests -type f ! -name '*.yaml' -delete
