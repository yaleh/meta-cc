# Makefile for meta-cc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X github.com/yale/meta-cc/cmd.Version=$(VERSION) \
                     -X github.com/yale/meta-cc/cmd.Commit=$(COMMIT) \
                     -X github.com/yale/meta-cc/cmd.BuildTime=$(BUILD_TIME)"

GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOMOD := $(GOCMD) mod
BUILD_DIR := build
BINARY_NAME := meta-cc
MCP_BINARY_NAME := meta-cc-mcp
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build build-cli build-mcp test clean install cross-compile lint fmt vet help

all: lint test build

build: build-cli build-mcp

build-cli:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

build-mcp:
	@echo "Building $(MCP_BINARY_NAME) $(VERSION)..."
	$(GOBUILD) -o $(MCP_BINARY_NAME) ./cmd/mcp-server

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(MCP_BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install:
	@echo "Installing..."
	$(GOCMD) install $(LDFLAGS) .

cross-compile:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} .; \
		if [ "$${platform%/*}" = "windows" ]; then \
			mv $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} $(BUILD_DIR)/$(BINARY_NAME)-$${platform%/*}-$${platform#*/}.exe; \
		fi; \
	done
	@echo "Cross-compilation complete. Binaries in $(BUILD_DIR)/"

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

lint: fmt vet
	@echo "Running static analysis..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo "Skipping lint checks..."; \
	fi

fmt:
	@echo "Formatting code..."
	@gofmt -l -w .

vet:
	@echo "Running go vet..."
	@$(GOCMD) vet ./...

help:
	@echo "Available targets:"
	@echo "  make build           - Build both meta-cc and meta-cc-mcp"
	@echo "  make build-cli       - Build meta-cc CLI only"
	@echo "  make build-mcp       - Build meta-cc-mcp MCP server only"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make lint            - Run static analysis (fmt + vet + golangci-lint)"
	@echo "  make fmt             - Format code with gofmt"
	@echo "  make vet             - Run go vet"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make install         - Install to GOPATH/bin"
	@echo "  make cross-compile   - Build for all platforms"
	@echo "  make deps            - Download and tidy dependencies"
	@echo "  make help            - Show this help message"
