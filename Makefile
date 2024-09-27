BINARY_NAME := Update
CMD_PATH := ./cmd/main

# Go commands
GO_BUILD:=go build
GO_CLEAN:=go clean
GO_TIDY:=go mod tidy
GO_GET_U:=go get -u ./...
GOARCH_Win:=amd64
GOOS_Win:=windows
GOARCH_Linux:=amd64
GOOS_Linux:=linux
LDFLAGS:="-s -w"
TAG:=jsoniter
get-u:
	@echo "Updating dependencies..."
	$(GO_GET_U)
# Build binary for Windows
build-windows:
	@echo "Building application for Windows..."
	set GOOS=$(GOOS_Win)&& set GOARCH=$(GOARCH_Win)&& $(GO_BUILD) -tags=$(TAG) -ldflags=$(LDFLAGS) -o $(BINARY_NAME)-windows.exe $(CMD_PATH)
# Build binary for Linux
build-linux:
	@echo "Building application for Linux..."
	set GOOS=$(GOOS_Linux)&& set GOARCH=$(GOARCH_Linux)&& $(GO_BUILD) -tags=$(TAG) -ldflags=$(LDFLAGS) -o $(BINARY_NAME)-linux $(CMD_PATH)

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GO_TIDY)

# Clean generated fileutils
clean:
	@echo "Cleaning build artifacts..."
	$(GO_CLEAN)
	del /F /Q $(BINARY_NAME)-windows.exe
	del /F /Q $(BINARY_NAME)-linux

# Default target
all: build-windows

# Help information
help:
	@echo "Usage:"
	@echo "  make build-windows - Build application for Windows"
	@echo "  make install       - Install dependencies"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make help          - Display this help message"

.PHONY: build-windows install clean proto help all
