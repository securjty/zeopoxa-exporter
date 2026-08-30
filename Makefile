# Variables
APP_NAME ?= zeopoxa_exporter
BUILD_DIR ?= build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Platforms
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all clean test build build-all

all: clean test build-all

clean:
	rm -rf $(BUILD_DIR)

build:
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/cli

# Cross-compile loop for all platforms
build-all:
	$(foreach plat,$(PLATFORMS), \
		$(eval GOOS := $(word 1,$(subst /, ,$(plat)))) \
		$(eval GOARCH := $(word 2,$(subst /, ,$(plat)))) \
		$(eval BIN := $(APP_NAME)-$(GOOS)-$(GOARCH)) \
		$(if $(filter windows,$(GOOS)),$(eval BIN := $(BIN).exe)) \
		echo "Building for $(GOOS)/$(GOARCH)..."; \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-ldflags="-s -w -X main.Version=$(VERSION)" \
			-trimpath \
			-o $(BUILD_DIR)/$(BIN) ./cmd/cli; \
	)