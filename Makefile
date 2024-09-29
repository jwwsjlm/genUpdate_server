BINARY_NAME := ./build/Update
CMD_PATH := ./cmd/main

# Go commands
GO = go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOMOD := $(GO) mod
GOGET := $(GO) get

# Build flags
LDFLAGS := -s -w
#TAGS := jsoniter

# OS-specific settings
GOOS_WINDOWS := windows
GOARCH_WINDOWS := amd64
GOOS_LINUX := linux
GOARCH_LINUX := amd64

.PHONY: all build-windows build-linux install clean help get-u

all: build-windows build-linux

build-windows:
	@echo Building application for Windows...
	set GOOS=$(GOOS_WINDOWS)& set GOARCH=$(GOARCH_WINDOWS)& $(GOBUILD) -tags=$(TAGS) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-windows.exe $(CMD_PATH)

build-linux:
	@echo Building application for Linux...
	set GOOS=$(GOOS_LINUX)& set GOARCH=$(GOARCH_LINUX)& $(GOBUILD) -tags=$(TAGS) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux $(CMD_PATH)

install:
	@echo Installing dependencies...
	$(GOMOD) tidy
upx: build-windows build-linux
	@echo "Compressing with UPX..."
	upx --best --lzma $(BINARY_NAME)-windows.exe
	upx --best --lzma $(BINARY_NAME)-linux

clean:
	@echo Cleaning build artifacts...
	$(GOCLEAN)
	@if exist $(BINARY_NAME)-windows.exe del /F /Q $(BINARY_NAME)-windows.exe
	@if exist $(BINARY_NAME)-linux del /F /Q $(BINARY_NAME)-linux

get-u:
	@echo Updating dependencies...
	$(GOGET) -u ./...

help:
	@echo Usage:
	@echo   make build-windows - Build application for Windows
	@echo   make build-linux   - Build application for Linux
	@echo   make install       - Install dependencies
	@echo   make clean         - Clean build artifacts
	@echo   make get-u         - Update dependencies
	@echo   make help          - Display this help message
