.PHONY: build build-windows build-linux build-all run test lint clean deps help

BINARY_NAME = logi-sim-leds
BUILD_DIR   = build
MAIN_PATH   = ./cmd/logi-sim-leds
GCC        ?= gcc
CGO_ENV     = CGO_ENABLED=1 CC=$(GCC)

UNAME := $(shell uname -s)

all: test build

## build: Build for the current OS
ifeq ($(findstring MINGW,$(UNAME))$(findstring MSYS,$(UNAME))$(findstring NT,$(UNAME)),)
build: build-linux
else
build: build-windows
endif

## build-windows: Build for Windows
build-windows:
	@mkdir -p $(BUILD_DIR)
	@$(CGO_ENV) go build -ldflags '-s -w -H windowsgui' -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME).exe"

## build-linux: Build for Linux
build-linux:
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for all platforms
build-all: build-windows build-linux

## run: Build and run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME).exe

## test: Run tests
test:
	@go test -v ./...

## lint: Run linter
lint:
	@golangci-lint run

## clean: Remove build artifacts
clean:
	@go clean
	@rm -rf $(BUILD_DIR)

## deps: Download dependencies
deps:
	@go get -v ./...
	@go mod tidy

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' | sed -e 's/^/ /'
