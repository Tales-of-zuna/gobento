# Development tasks for shop. Run `make` to list them.
#
# Every target runs plain go or docker commands. Without make (for example
# on Windows), run the commands from the Development section of README.md.

.DEFAULT_GOAL := help

GO               ?= go
BINARY           := shop$(if $(filter Windows_NT,$(OS)),.exe)
IMAGE            ?= shop
GOLANGCI_LINT    ?= $(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
DOCKER_RUN_FLAGS ?=

# Keep the SQLite database of the container in a named volume.
DOCKER_RUN_FLAGS += --volume $(IMAGE)-data:/data

.PHONY: help build run test cover vet lint fmt tidy check docker-build docker-run clean

help: ## Show this list of targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-z0-9-]+:.*## / {printf "  %-13s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary into bin/
	$(GO) build -o bin/$(BINARY) ./cmd/shop

run: ## Run the server
	$(GO) run ./cmd/shop

test: ## Run the tests
	$(GO) test ./...

cover: ## Run the tests with coverage and print a summary
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

vet: ## Report suspicious constructs
	$(GO) vet ./...

lint: ## Lint and check formatting with golangci-lint
	$(GOLANGCI_LINT) run

fmt: ## Format the code
	$(GOLANGCI_LINT) fmt

tidy: ## Tidy go.mod and go.sum
	$(GO) mod tidy

check: vet lint test ## Run the checks CI runs: vet, lint, and test

.PHONY: generate
generate: ## Regenerate generated code
	$(GO) run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate

docker-build: ## Build the container image
	docker build --tag $(IMAGE) .

docker-run: ## Run the container image on port 8080
	docker run --rm --publish 8080:8080 $(DOCKER_RUN_FLAGS) $(if $(DATABASE_URL),--env 'DATABASE_URL=$(DATABASE_URL)') $(IMAGE)

clean: ## Remove build and coverage output
	$(RM) -r bin coverage.out
