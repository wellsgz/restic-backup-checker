# Restic Backup Checker Makefile

# Application name
APP_NAME := restic-backup-checker

# Version
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build directory
BUILD_DIR := build

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Cross-compilation targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	@echo "Building $(APP_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/main.go

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building $(APP_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		output_name=$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then \
			output_name=$$output_name.exe; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		env GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$$output_name ./cmd/main.go; \
	done

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install locally
.PHONY: install
install: build
	@echo "Installing $(APP_NAME)..."
	cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

# Run the application
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

# Run setup
.PHONY: setup
setup: build
	@echo "Running setup..."
	./$(BUILD_DIR)/$(APP_NAME) setup

# Run manual check
.PHONY: check
check: build
	@echo "Running manual check..."
	./$(BUILD_DIR)/$(APP_NAME) check

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Development workflow
.PHONY: dev
dev: fmt vet test build

# Release workflow
.PHONY: release
release: clean fmt vet test build-all

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  install      - Install locally"
	@echo "  run          - Run the application"
	@echo "  setup        - Run setup wizard"
	@echo "  check        - Run manual check"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  lint         - Run linter"
	@echo "  dev          - Development workflow (fmt, vet, test, build)"
	@echo "  release      - Release workflow (clean, fmt, vet, test, build-all)"
	@echo "  help         - Show this help message" 