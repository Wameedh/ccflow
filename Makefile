.PHONY: build test lint clean install run help

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
	@echo "  deps          - Download and tidy dependencies"
	@echo "  verify        - Verify dependencies"
	@echo "  snapshot      - Build snapshot with goreleaser"
