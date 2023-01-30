GOBIN ?= $(BIN_DIR)
include ../.bingo/Variables.mk

.PHONY: validate
validate: $(KUBEVAL) generate
	$(KUBEVAL) --ignore-missing-schemas -d manifests/

.PHONY: generate
generate:
	@rm -rf manifests
	@go run main.go generate --output=manifests
