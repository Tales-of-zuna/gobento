# Development tasks for gobento. Run `make` to list them.
# Each target wraps one or two plain commands; CONTRIBUTING.md lists them
# for systems without make (for example, Windows).

.DEFAULT_GOAL := help

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
BIN           := bin/gobento

.PHONY: help build test test-all lint fmt tidy cover clean

help: ## Show this list of targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-z-]+:.*## / {printf "  %-9s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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
