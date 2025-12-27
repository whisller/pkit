.PHONY: build test install clean lint coverage fmt help

# Build variables
BINARY_NAME=pkit
BIN_DIR=bin
CMD_DIR=cmd/pkit

# Default target
.DEFAULT_GOAL := help

## build: Build the pkit binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "✓ Built $(BIN_DIR)/$(BINARY_NAME)"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v -cover ./...

## install: Install pkit to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(CMD_DIR)/main.go
	@echo "✓ Installed $(BINARY_NAME)"

## clean: Clean build artifacts and test caches
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)/
	@go clean -testcache
	@echo "✓ Cleaned"

## lint: Run linters
lint:
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...
	@echo "✓ Linters passed"

## coverage: Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
