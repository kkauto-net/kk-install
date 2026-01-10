.PHONY: build build-all test clean install release lint fmt deps test-coverage uninstall

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X github.com/kkauto-net/kk-install/cmd.Version=$(VERSION)"
BINARY := kk
BUILD_DIR := build

# Build for current platform
build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .

# Build for all platforms (Linux only)
build-all: clean
	mkdir -p dist/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist/
	rm -f coverage.out coverage.html

# Install locally
install: build
	sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/

# Uninstall
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY)

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Download dependencies
deps:
	go mod download
	go mod tidy
