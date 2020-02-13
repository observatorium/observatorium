DEFAULT_GOAL := help

# Ensure everything works even if GOPATH is not set, which is often the case.
# The `go env GOPATH` will work for all cases for Go 1.8+.
GOPATH            ?= $(shell go env GOPATH)

TMP_GOPATH        ?= /tmp/thanos-go
GOBIN             ?= $(firstword $(subst :, ,${GOPATH}))/bin
GOPROXY           ?= https://proxy.golang.org
GIT               ?= $(shell which git)

# Version from 12.12.2019
GOJSONTOYAML_VERSION    ?= bf2969bbd742d117a9524b859fb417fefb67565d
GOJSONTOYAML            ?= $(GOBIN)/gojsontoyaml-$(GOJSONTOYAML_VERSION)

# Version from 9.02.2020. We use https://github.com/google/go-jsonnet for ease of building.
GO_JSONNET_VERSION         ?= 31b9ace0f65a5a5717ed9521b57d8d9e18f24070
GO_JSONNET                 ?= $(GOBIN)/jsonnet-$(GO_JSONNET_VERSION)
# v0.2.0
JSONNET_BUNDLER_VERSION ?= 184841238bb5df1b886ccaecd3e179bc55c17687
JSONNET_BUNDLER         ?= $(GOBIN)/jb-$(JSONNET_BUNDLER_VERSION)

# fetch_go_bin_version downloads (go gets) the binary from specific version and installs it in $(GOBIN)/<bin>-<version>
# arguments:
# $(1): Install path. (e.g github.com/campoy/embedmd)
# $(2): Tag or revision for checkout.
# TODO(bwplotka): Move to just using modules, however make sure to not use or edit Thanos go.mod file!
define fetch_go_bin_version
	@mkdir -p $(GOBIN)
	@mkdir -p $(TMP_GOPATH)

	@echo ">> fetching $(1)@$(2) revision/version"
	@if [ ! -d '$(TMP_GOPATH)/src/$(1)' ]; then \
    GOPATH='$(TMP_GOPATH)' GO111MODULE='off' go get -d -u '$(1)/...'; \
  else \
    CDPATH='' cd -- '$(TMP_GOPATH)/src/$(1)' && git fetch; \
  fi
	@CDPATH='' cd -- '$(TMP_GOPATH)/src/$(1)' && git checkout -f -q '$(2)'
	@echo ">> installing $(1)@$(2)"
	@GOBIN='$(TMP_GOPATH)/bin' GOPATH='$(TMP_GOPATH)' GO111MODULE='off' go install '$(1)'
	@mv -- '$(TMP_GOPATH)/bin/$(shell basename $(1))' '$(GOBIN)/$(shell basename $(1))-$(2)'
	@echo ">> produced $(GOBIN)/$(shell basename $(1))-$(2)"

endef

help: ## Displays this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# deps ensures jsonnet deps are installed.
.PHONY: deps
deps: ## Install libsonnet dependencies using jsonnet-bundler.
deps: $(JSONNET_BUNDLER) jsonnetfile.json jsonnetfile.lock.json
	@echo ">> install jsonnet deps via jb"
	$(JSONNET_BUNDLER) install

JSONNET_FMT := jsonnetfmt -n 2 --max-blank-lines 2 --string-style s --comment-style s -i

.PHONY: format
format: ## Auto-formats jsonnet files.
	@which jsonnetfmt 2>/dev/null || ( \
		echo "Cannot find jsonnetfmt command, please install from https://github.com/google/jsonnet/releases. If your C++ does not support GLIBCXX_3.4.20, please use xxx-in-container target like format-in-container." && exit 1 \
	)

	@echo ">> formatting jsonnet"
	@find . -type f -not -path './vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \) | xargs -n 1 -- $(JSONNET_FMT)

.PHONY: generate
generate: ## Generates manifests for all environments from jsonnet.
generate: deps
	@ENV="dev" $(MAKE) generate-env
	@ENV="base" $(MAKE) generate-env
	@ENV="openshift" $(MAKE) generate-env

.PHONY: generate-env
generate-env: ## Generates manifests for environment given in ENV=... environment variable from jsonnet.
generate-env: $(GO_JSONNET) $(GOJSONTOYAML)
	@echo ">> making sure to start with a clean 'manifests' dir for $(ENV)"
	@rm -rf environments/$(ENV)/manifests
	@mkdir environments/$(ENV)/manifests
	@echo ">> building manifests for 'main*.jsonnet' files in $(ENV)"
ifeq ($(ENV),openshift)
	@$(GO_JSONNET) -J vendor environments/openshift/main-observatorium-template.jsonnet | $(GOJSONTOYAML) >environments/openshift/manifests/observatorium-template.yaml
	@$(GO_JSONNET) -J vendor environments/openshift/main-jaeger-template.jsonnet | $(GOJSONTOYAML) >environments/openshift/manifests/jaeger-template.yaml
	@$(GO_JSONNET) -J vendor environments/openshift/main-observatorium-api-template.jsonnet | $(GOJSONTOYAML) >environments/openshift/manifests/observatorium-api-template.yaml
else
	@$(GO_JSONNET) -J vendor -m environments/$(ENV)/manifests environments/$(ENV)/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
endif
	@find environments/$(ENV)/manifests -type f ! -name '*.yaml' -delete

# Check https://github.com/coreos/prometheus-operator/blob/master/scripts/jsonnet/Dockerfile for the image.
JSONNET_CONTAINER_CMD:=docker run --rm \
		-u="$(shell id -u):$(shell id -g)" \
		-v "$(shell go env GOCACHE):/.cache/go-build" \
		-v "$(PWD):/configuration:Z" \
		-w "/configuration" \
		-e USER=deadbeef \
		-e GO111MODULE=on \
		quay.io/coreos/jsonnet-ci:release-0.35

.PHONY: format-in-container
format-in-container: ## Auto-formats jsonnet files in docker contaier.
	$(JSONNET_CONTAINER_CMD) $(MAKE) $(MFLAGS) format

.PHONY: generate-in-container
generate-in-container: ## Generates manifests for all environments from jsonnet in docker container.
	@echo ">> Compiling and generating thanos-mixin"
	$(JSONNET_CONTAINER_CMD) $(MAKE) $(MFLAGS) JSONNET_BUNDLER='/go/bin/jb' deps
	$(JSONNET_CONTAINER_CMD) $(MAKE) $(MFLAGS) \
		JSONNET='/go/bin/jsonnet' \
		JSONNET_BUNDLER='/go/bin/jb' \
		GOJSONTOYAML='/go/bin/gojsontoyaml' \
		generate

$(GO_JSONNET):
	$(call fetch_go_bin_version,github.com/google/go-jsonnet/cmd/jsonnet,$(GO_JSONNET_VERSION))

$(GOJSONTOYAML):
	$(call fetch_go_bin_version,github.com/brancz/gojsontoyaml,$(GOJSONTOYAML_VERSION))

$(JSONNET_BUNDLER):
	$(call fetch_go_bin_version,github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb,$(JSONNET_BUNDLER_VERSION))
