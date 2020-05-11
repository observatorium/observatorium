SHELL=/usr/bin/env bash -o pipefail

VERSION := $(strip $(shell [ -d .git ] && git describe --always --tags --dirty))
BUILD_DATE := $(shell date -u +"%Y-%m-%d")
BUILD_TIMESTAMP := $(shell date -u +"%Y-%m-%dT%H:%M:%S%Z")
VCS_BRANCH := $(strip $(shell git rev-parse --abbrev-ref HEAD))
VCS_REF := $(strip $(shell [ -d .git ] && git rev-parse --short HEAD))
DOCKER_REPO ?= quay.io/observatorium/observatorium-operator
CACHE_DIR="_cache"
FULL_OPERATOR_IMAGE="quay.io/observatorium/observatorium-operator:latest"
REGISTRY_NAMESPACE ?= "observatorium"
TOOLS_DIR="$(CACHE_DIR)/tools"
OPERATOR_SDK_VERSION="v0.17.0"
OPERATOR_SDK_PLATFORM ?= "x86_64-linux-gnu"
OPERATOR_SDK_BIN="operator-sdk-$(OPERATOR_SDK_VERSION)-$(OPERATOR_SDK_PLATFORM)"
OPERATOR_SDK="$(TOOLS_DIR)/$(OPERATOR_SDK_BIN)"

BIN_DIR ?= $(shell pwd)/tmp/bin

CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen
JB ?= $(BIN_DIR)/jb


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

operator-sdk:
	@if [ ! -x "$(OPERATOR_SDK)" ]; then\
		echo "Downloading operator-sdk $(OPERATOR_SDK_VERSION)";\
		mkdir -p $(TOOLS_DIR);\
		curl -JL https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/$(OPERATOR_SDK_BIN) -o $(OPERATOR_SDK);\
		chmod +x $(OPERATOR_SDK);\
	else\
		echo "Using operator-sdk cached at $(OPERATOR_SDK)";\
	fi

generate-csv: operator-sdk dist-csv-generator
	@if [ -z "$(REGISTRY_NAMESPACE)" ]; then\
		echo "REGISTRY_NAMESPACE env-var must be set to your $(IMAGE_REGISTRY) namespace";\
		exit 1;\
	fi
	OPERATOR_SDK=$(OPERATOR_SDK) FULL_OPERATOR_IMAGE=$(FULL_OPERATOR_IMAGE) operator/hack/csv-generate.sh

dist-csv-generator:
	@if [ ! -x build/_output/bin/csv-generator ]; then\
		echo "Building csv-generator tool";\
		mkdir -p build/_output/bin;\
		go build -i -ldflags="-s -w" -mod=vendor -o build/_output/bin/csv-generator ./tools/csv-generator;\
	else \
		echo "Using pre-built csv-generator tool";\
	fi
