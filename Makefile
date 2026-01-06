.PHONY: all build test lint bench clean fmt vet coverage

all: lint test build

build:
	go build ./...

test:
	go test -race -v ./...

test-short:
	go test -short ./...

lint:
	golangci-lint run ./...

bench:
	go test -bench=. -benchmem ./...

clean:
	go clean ./...
	rm -f coverage.out coverage.html

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fuzz:
	go test -fuzz=Fuzz -fuzztime=30s ./parse/...

# Install development tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Run all checks before commit
pre-commit: fmt vet lint test
