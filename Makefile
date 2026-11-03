# Development tasks for gobento. Run `make` to list them.
# Each target wraps one or two plain commands; CONTRIBUTING.md lists them
# for systems without make (for example, Windows).

.DEFAULT_GOAL := help

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
BIN           := bin/gobento

.PHONY: help build test test-all lint fmt tidy cover catalog catalog-check clean

help: ## Show this list of targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-13s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the gobento binary into bin/
	@if [ -d cmd/gobento ]; then $(GO) build -o $(BIN) ./cmd/gobento; \
	else echo "cmd/gobento does not exist yet; nothing to build"; fi

test: ## Run unit tests (fast; skips integration tests)
	$(GO) test -short -race ./...

test-all: ## Run all tests, including integration tests (needs network)
	$(GO) test -race ./...

lint: ## Run linters
	$(GOLANGCI_LINT) run

fmt: ## Format code
	$(GOLANGCI_LINT) fmt

tidy: ## Tidy go.mod and go.sum
	$(GO) mod tidy

cover: ## Run unit tests with coverage and print a summary
	$(GO) test -short -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

catalog: ## Regenerate the website's catalog.json (and its CLI cross-check commands)
	$(GO) run ./internal/tools/catalogexport -o website/src/generated/catalog.json
	@if [ -d website/node_modules ]; then cd website && pnpm exec vitest run lib/command/cli-commands.test.ts --update; \
	else echo "website/node_modules is missing; run pnpm install and pnpm test -u in website/ to refresh lib/command/supported-commands.txt"; fi

catalog-check: ## Fail if catalog.json is stale or the website's commands fail in the CLI
	$(GO) run ./internal/tools/catalogexport | diff -u website/src/generated/catalog.json -
	$(GO) test -count=1 -run 'TestCommittedCatalogIsCurrent|TestWebsiteCommands' ./internal/tools/catalogexport ./internal/cli

clean: ## Remove build and coverage output
	$(RM) -r bin dist coverage.out
