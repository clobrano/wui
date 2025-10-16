.PHONY: build test install clean help

# Build variables
BINARY_NAME=wui
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/clobrano/wui/internal/version.Version=$(VERSION) \
                   -X github.com/clobrano/wui/internal/version.Commit=$(COMMIT) \
                   -X github.com/clobrano/wui/internal/version.BuildDate=$(BUILD_DATE)"

# Default target
all: build

## build: Build the wui binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) cmd/wui/main.go
	@echo "Build complete: ./$(BINARY_NAME)"

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests complete"

## coverage: Run tests with coverage report
coverage: test
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

## install: Install the wui binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) cmd/wui/main.go
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.txt coverage.html
	@rm -rf dist/
	@echo "Clean complete"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

## mod-tidy: Tidy Go modules
mod-tidy:
	@echo "Tidying modules..."
	@go mod tidy

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
