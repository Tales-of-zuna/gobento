# Development tasks for gobento. Run `make` to list them.
# Each target wraps one or two plain commands; CONTRIBUTING.md lists them
# for systems without make (for example, Windows).

.DEFAULT_GOAL := help

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
PNPM          ?= pnpm
BIN           := bin/gobento
WEB           := website

.PHONY: help build test test-all lint fmt tidy cover clean \
	web-dev web-check web-build web-e2e

help: ## Show this list of targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-z0-9-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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

clean: ## Remove build and coverage output
	$(RM) -r bin dist coverage.out

web-dev: ## Start the website dev server (run pnpm install in website/ first)
	cd $(WEB) && $(PNPM) dev

web-check: ## Lint, type-check, format-check and unit-test the website
	cd $(WEB) && $(PNPM) lint && $(PNPM) typecheck && $(PNPM) format:check && $(PNPM) test

web-build: ## Build the website's static export into website/out/
	cd $(WEB) && $(PNPM) build

web-e2e: ## Run the website end-to-end and accessibility tests (Playwright)
	cd $(WEB) && $(PNPM) e2e
