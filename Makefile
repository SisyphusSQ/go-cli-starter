BINARY_NAME ?= go-cli-starter
VARS_PKG ?= github.com/SisyphusSQ/go-cli-starter/vars
BUILD_DIR ?= bin
GO ?= go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

BUILD_FLAGS  = -X '$(VARS_PKG).AppName=$(BINARY_NAME)'
BUILD_FLAGS += -X '$(VARS_PKG).AppVersion=$(VERSION)'
BUILD_FLAGS += -X '$(VARS_PKG).GoVersion=$(shell $(GO) version)'
BUILD_FLAGS += -X '$(VARS_PKG).BuildTime=$(shell date +"%Y-%m-%d %H:%M:%S")'
BUILD_FLAGS += -X '$(VARS_PKG).GitCommit=$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)'
BUILD_FLAGS += -X '$(VARS_PKG).GitRemote=$(shell git config --get remote.origin.url 2>/dev/null || echo unknown)'

LDFLAGS = -ldflags="$(BUILD_FLAGS)"

.PHONY: help all build release run test lint fmt tidy clean install

help:
	@echo "Available targets:"
	@echo "  all      - Run fmt, test, and build"
	@echo "  build    - Build local binary"
	@echo "  release  - Build release binary with trimpath"
	@echo "  run      - Run starter locally"
	@echo "  test     - Run tests with race detector"
	@echo "  lint     - Run golangci-lint"
	@echo "  fmt      - Run go fmt"
	@echo "  tidy     - Run go mod tidy"
	@echo "  install  - Install binary into GOPATH/bin"
	@echo "  clean    - Remove build artifacts"

all: fmt test build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

release:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go

run:
	$(GO) run ./main.go

test:
	$(GO) test -race ./...

lint:
	golangci-lint run

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

install:
	$(GO) install $(LDFLAGS) .

clean:
	rm -rf $(BUILD_DIR)
