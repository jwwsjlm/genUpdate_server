BINARY_NAME := ./build/Update
CMD_PATH := ./cmd/main
GO ?= go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOMOD := $(GO) mod
GOGET := $(GO) get
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)
GOOS_WINDOWS := windows
GOARCH_WINDOWS := amd64
GOOS_LINUX := linux
GOARCH_LINUX := amd64

.PHONY: all build-windows build-linux install clean help get upx

all: build-windows build-linux

build-windows:
	@echo "Building application for Windows..."
	GOOS=$(GOOS_WINDOWS) GOARCH=$(GOARCH_WINDOWS) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-windows.exe $(CMD_PATH)

build-linux:
	@echo "Building application for Linux..."
	GOOS=$(GOOS_LINUX) GOARCH=$(GOARCH_LINUX) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux $(CMD_PATH)

install:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

upx: build-windows build-linux
	@echo "Compressing with UPX..."
	upx --best --lzma -o $(BINARY_NAME)-windows-upx.exe $(BINARY_NAME)-windows.exe || { echo "Windows UPX compression failed"; exit 1; }
	upx --best --lzma -o $(BINARY_NAME)-linux-upx $(BINARY_NAME)-linux || { echo "Linux UPX compression failed"; exit 1; }

clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)-windows.exe $(BINARY_NAME)-linux $(BINARY_NAME)-windows-upx.exe $(BINARY_NAME)-linux-upx

get:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...

help:
	@echo "Usage:"
	@echo "  make build-windows - Build application for Windows"
	@echo "  make build-linux   - Build application for Linux"
	@echo "  make install       - Tidy dependencies"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make get           - Update dependencies"
	@echo "  make help          - Display this help message"
