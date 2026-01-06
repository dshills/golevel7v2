.PHONY: all build test lint bench clean fmt vet coverage help

all: lint test build

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all packages
	go build ./...

test: ## Run tests with race detection
	go test -race -v ./...

test-short: ## Run tests in short mode
	go test -short ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

clean: ## Clean build artifacts
	go clean ./...
	rm -f coverage.out coverage.html

fmt: ## Format code with gofmt and goimports
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	go vet ./...

coverage: ## Generate test coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fuzz: ## Run fuzz tests for 30 seconds
	go test -fuzz=Fuzz -fuzztime=30s ./parse/...

tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

pre-commit: fmt vet lint test ## Run all checks before commit
