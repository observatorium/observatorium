JSONNET_FILES=$(shell find . -type f -not -path './vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))

.DEFAULT_GOAL := all

.PHONY: check
check:
	@echo Checking...
	@echo $(JSONNET_FILES) | xargs -L1 jsonnetfmt -i
	@git diff -s --exit-code . || (echo "Build failed: one of more files aren't staged to be committed." && git diff && exit 1)

.PHONY: clean
clean:
	@echo Cleaning...
	@rm -rf environments/kubernetes/manifests
	@rm -rf environments/openshift/manifests
	@rm -rf environments/sre/servicemonitors
	@rm -rf environments/sre/prometheusrules
	@rm -rf environments/sre/grafana

.PHONY: build
build:
	@echo Building...
	@./build.sh

.PHONY: vendor
vendor:
	@echo Bringing dependencies...
	@jb install

.PHONY: install-tools
install-tools:
	@## this target is missing the installation of jsonnetfmt, which doesn't seem to have a binary available (!!)
	@go get \
		github.com/google/go-jsonnet/cmd/jsonnet \
		github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb \
		github.com/brancz/gojsontoyaml

.PHONY: all
all: check vendor build

.PHONY: ci
ci: all
