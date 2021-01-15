# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.3.0. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Bellow generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for api variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(API)
#	@echo "Running api"
#	@$(API) <flags/args..>
#
API := $(GOBIN)/api-v0.1.1
$(API): $(BINGO_DIR)/api.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/api-v0.1.1"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=api.mod -o=$(GOBIN)/api-v0.1.1 "github.com/observatorium/observatorium"

BINGO := $(GOBIN)/bingo-v0.3.0
$(BINGO): $(BINGO_DIR)/bingo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/bingo-v0.3.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=bingo.mod -o=$(GOBIN)/bingo-v0.3.0 "github.com/bwplotka/bingo"

GOJSONTOYAML := $(GOBIN)/gojsontoyaml-v0.0.0-20200602132005-3697ded27e8c
$(GOJSONTOYAML): $(BINGO_DIR)/gojsontoyaml.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/gojsontoyaml-v0.0.0-20200602132005-3697ded27e8c"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=gojsontoyaml.mod -o=$(GOBIN)/gojsontoyaml-v0.0.0-20200602132005-3697ded27e8c "github.com/brancz/gojsontoyaml"

JB := $(GOBIN)/jb-v0.4.0
$(JB): $(BINGO_DIR)/jb.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/jb-v0.4.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=jb.mod -o=$(GOBIN)/jb-v0.4.0 "github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb"

JSONNET_LINT := $(GOBIN)/jsonnet-lint-v0.17.0
$(JSONNET_LINT): $(BINGO_DIR)/jsonnet-lint.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/jsonnet-lint-v0.17.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=jsonnet-lint.mod -o=$(GOBIN)/jsonnet-lint-v0.17.0 "github.com/google/go-jsonnet/cmd/jsonnet-lint"

JSONNET := $(GOBIN)/jsonnet-v0.17.0
$(JSONNET): $(BINGO_DIR)/jsonnet.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/jsonnet-v0.17.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=jsonnet.mod -o=$(GOBIN)/jsonnet-v0.17.0 "github.com/google/go-jsonnet/cmd/jsonnet"

JSONNETFMT := $(GOBIN)/jsonnetfmt-v0.17.0
$(JSONNETFMT): $(BINGO_DIR)/jsonnetfmt.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/jsonnetfmt-v0.17.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=jsonnetfmt.mod -o=$(GOBIN)/jsonnetfmt-v0.17.0 "github.com/google/go-jsonnet/cmd/jsonnetfmt"

KUBEVAL := $(GOBIN)/kubeval-v0.0.0-20201005082916-38668c6c5b23
$(KUBEVAL): $(BINGO_DIR)/kubeval.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/kubeval-v0.0.0-20201005082916-38668c6c5b23"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=kubeval.mod -o=$(GOBIN)/kubeval-v0.0.0-20201005082916-38668c6c5b23 "github.com/instrumenta/kubeval"

