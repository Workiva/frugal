.DEFAULT_GOAL := help
SHELL=/bin/bash -o pipefail

# Cite: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Display this help page
	@grep -E '^[a-zA-Z/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: gen
gen: ## Regenerates dart source-gen files
	dart run build_runner build lib --delete-conflicting-outputs --build-filter=**.g.dart

.PHONY: init
init: install

.PHONY: clean
clean: ## Remove build assets
	$(RM) -r .packages
	$(RM) -r .dart_tool
	$(RM) - r pubspec.lock

.PHONY: install
install: ## Install the dependencies for your repositories
	dart pub get

.PHONY: validate
validate: ## Validates dart dependencies
	dart run dependency_validator

.PHONY: format
format: ## Formats dart files
	@echo Formatting Dart Files...
	@dart run dart_dev format

.PHONY: test
test: ## Runs unit tests
	@echo "Running Unit Tests..."
	@dart run dart_dev test --xunit

.PHONY: analyze
analyze: ## Analyze dart project.
	@echo "Analyzing project..."
	@dart run dart_dev analyze
