# Copyright 2026 The LicenseOps Authors
# SPDX-License-Identifier: Apache-2.0

BINARY := lops
CMD := ./cmd/lops
MODULE := github.com/licenseops/licenseops
VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build run test lint lint-fix vet fmt tidy docker clean check fix help

## Build

build: ## Build the binary
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

run: build ## Build and run (check mode on current directory)
	./$(BINARY) check -v .

install: ## Install to GOPATH/bin
	go install $(LDFLAGS) $(CMD)

## Quality

test: ## Run all tests
	go test ./... -v -race -count=1

lint: ## Run golangci-lint
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofmt -s -w .

fmt-check: ## Check formatting (CI-friendly)
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

tidy: ## Tidy go.mod
	go mod tidy

## LicenseOps (self-check)

check: build ## Check license headers in this repo
	./$(BINARY) check

fix: build ## Fix license headers in this repo
	./$(BINARY) fix

## Docker

docker: ## Build Docker image
	docker build -t $(BINARY):latest .

docker-run: docker ## Run lops in Docker (mount current dir)
	docker run --rm -v "$$(pwd)":/src -w /src $(BINARY):latest check -v .

## Release

snapshot: ## Build release snapshot (no publish)
	goreleaser release --snapshot --clean

## Cleanup

clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -rf dist/

## Help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
