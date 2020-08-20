JSONNET_SRC = $(shell find . -type f -not -path './*vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \))

.PHONY: jsonnetfmt
jsonnetfmt: $(JSONNET_SRC)
	jsonnetfmt -n 2 --max-blank-lines 2 --string-style s --comment-style s -i $(JSONNET_SRC)

environments/base/manifests: environments/base/main.jsonnet $(JSONNET_SRC)
	-make jsonnetfmt
	-rm -rf environments/base/manifests
	-mkdir environments/base/manifests
	jsonnet -J vendor -m environments/base/manifests environments/base/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find environments/base/manifests -type f ! -name '*.yaml' -delete

environments/dev/manifests: environments/dev/main.jsonnet $(JSONNET_SRC) vendor-jsonnet
	-make jsonnetfmt
	-rm -rf environments/dev/manifests
	-mkdir environments/dev/manifests
	jsonnet -J vendor -m environments/dev/manifests environments/dev/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find environments/dev/manifests -type f ! -name '*.yaml' -delete

example/manifests: example/main.jsonnet $(JSONNET_SRC) vendor-jsonnet
	-make jsonnetfmt
	-rm -rf example/manifests
	-mkdir example/manifests
	jsonnet -J vendor example/main.jsonnet | gojsontoyaml > example/manifests/observatorium.yaml
	find example/manifests -type f ! -name '*.yaml' -delete

tests/manifests: tests/main.jsonnet $(JSONNET_SRC) vendor-jsonnet
	-make jsonnetfmt
	-rm -rf tests/manifests
	-mkdir tests/manifests
	jsonnet -J vendor -m tests/manifests tests/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find tests/manifests -type f ! -name '*.yaml' -delete
