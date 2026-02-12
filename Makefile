.PHONY: build install clean test run fmt vet build-gui build-daemon build-all-bins run-gui run-daemon

# Binary names
BINARY=darkstorage
GUI_BINARY=darkstorage-gui
DAEMON_BINARY=darkstorage-daemon

# Version from git or default
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Build the CLI binary
build:
	go build $(LDFLAGS) -o $(BINARY)

# Build the GUI application
build-gui:
	@echo "Building GUI application..."
	go build $(LDFLAGS) -o $(GUI_BINARY) cmd/gui/main.go

# Build the daemon
build-daemon:
	@echo "Building daemon..."
	go build $(LDFLAGS) -o $(DAEMON_BINARY) cmd/daemon/*.go

# Build all binaries (CLI, GUI, daemon)
build-all-bins: build build-gui build-daemon
	@echo "✓ All binaries built successfully"

# Install CLI to /usr/local/bin
install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

# Install all binaries
install-all: build-all-bins
	@echo "Installing all binaries..."
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)
	install -m 755 $(GUI_BINARY) /usr/local/bin/$(GUI_BINARY)
	install -m 755 $(DAEMON_BINARY) /usr/local/bin/$(DAEMON_BINARY)
	@echo "✓ All binaries installed to /usr/local/bin"

# Clean build artifacts
clean:
	rm -f $(BINARY) $(GUI_BINARY) $(DAEMON_BINARY)
	rm -f $(BINARY)-* $(GUI_BINARY)-* $(DAEMON_BINARY)-*
	go clean

# Run tests
test:
	go test -v ./...

# Run the CLI binary
run: build
	./$(BINARY)

# Run the GUI application
run-gui: build-gui
	./$(GUI_BINARY)

# Run the daemon
run-daemon: build-daemon
	./$(DAEMON_BINARY)

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-darwin-arm64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-windows-amd64.exe

# Help target
help:
	@echo "Available targets:"
	@echo "  build          - Build the CLI binary"
	@echo "  build-gui      - Build the GUI application"
	@echo "  build-daemon   - Build the daemon"
	@echo "  build-all-bins - Build all binaries (CLI, GUI, daemon)"
	@echo "  install        - Install CLI to /usr/local/bin"
	@echo "  install-all    - Install all binaries to /usr/local/bin"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  run            - Build and run the CLI"
	@echo "  run-gui        - Build and run the GUI"
	@echo "  run-daemon     - Build and run the daemon"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  build-all      - Build for multiple platforms"
