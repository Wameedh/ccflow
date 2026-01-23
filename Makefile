.PHONY: build test lint clean install run help dev dev-install clean-dev

# Build variables
BINARY_NAME=ccflow
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-X github.com/wameedh/ccflow/internal/config.Version=$(VERSION) -X github.com/wameedh/ccflow/internal/config.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./...

# Run the CLI (for development)
run:
	go run $(LDFLAGS) ./main.go $(ARGS)

# Development: build and run in one command
# Usage: make dev ARGS="run go-cli-dev"
dev: build
	./$(BINARY_NAME) $(ARGS)

# Install to local ./bin directory (avoids PATH conflicts with Homebrew)
dev-install:
	@mkdir -p ./bin
	go build $(LDFLAGS) -o ./bin/$(BINARY_NAME) ./main.go
	@echo "Installed to ./bin/$(BINARY_NAME)"
	@echo "Run with: ./bin/ccflow <command>"

# Clean up development artifacts (keeps released Homebrew version intact)
clean-dev:
	rm -rf ./bin
	@echo "Removed ./bin/ - released version at /opt/homebrew/bin/ccflow is unchanged"

# Download dependencies
deps:
	go mod download
	go mod tidy

# Verify dependencies
verify:
	go mod verify

# Build for all platforms (local snapshot)
snapshot:
	goreleaser build --snapshot --clean

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  run ARGS=...  - Run the CLI with arguments"
	@echo "  dev ARGS=...  - Build and run in one command"
	@echo "  dev-install   - Install to ./bin/ for local testing"
	@echo "  clean-dev     - Remove ./bin/ development artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  verify        - Verify dependencies"
	@echo "  snapshot      - Build snapshot with goreleaser"
