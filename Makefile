CONFIGURATION_DIR ?= ./configuration
WEBSITE_DIR ?= website
WEBSITE_BASE_URL ?= https://observatorium.io

all: generate validate

.PHONY: fmt
fmt:
	@$(MAKE) -C $(CONFIGURATION_DIR) fmt

.PHONY: lint
lint:
	@$(MAKE) -C $(CONFIGURATION_DIR) lint

.PHONY: generate
generate:
	@$(MAKE) -C $(CONFIGURATION_DIR) generate

$(CONFIGURATION_DIR)/tests/manifests:
	@$(MAKE) -C $(CONFIGURATION_DIR) tests/manifests

.PHONY: validate
validate:
	@$(MAKE) -C $(CONFIGURATION_DIR) validate

.PHONY: vendor
vendor:
	@$(MAKE) -C $(CONFIGURATION_DIR) vendor

.PHONY: web-theme
web-theme:
	cd $(WEBSITE_DIR)/themes/doks/ && \
	npm install && \
	rm -rf content

.PHONY: web
web:
	cd $(WEBSITE_DIR) && \
	hugo -b $(WEBSITE_BASE_URL)

.PHONY: web-serve
web-serve:
	cd $(WEBSITE_DIR) && \
	hugo serve
