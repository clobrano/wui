.PHONY: build test install clean help image image-push run service-install service-uninstall service-status service-logs service-restart service-linger

WUI_CONFIG ?= $(HOME)/.config/wui/config.yaml

# Build variables
BINARY_NAME=wui
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Container image variables
CONTAINER_ENGINE ?= podman
IMAGE_REPO ?= quay.io/clobrano/wui
IMAGE_TAG ?= $(VERSION)
IMAGE ?= $(IMAGE_REPO):$(IMAGE_TAG)

# Host address:port to publish the container's web GUI on. Defaults to loopback
# (the GUI has no auth). To reach it over Tailscale, publish on the Tailscale
# IP, e.g.: make run HOST_ADDR=$(tailscale ip -4)
HOST_ADDR ?= 127.0.0.1
HOST_PORT ?= 7008
LDFLAGS=-ldflags "-X github.com/clobrano/wui/internal/version.Version=$(VERSION) \
                   -X github.com/clobrano/wui/internal/version.Commit=$(COMMIT) \
                   -X github.com/clobrano/wui/internal/version.BuildDate=$(BUILD_DATE)"

# Default target
all: build

## build: Build the wui binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Tests complete"

## coverage: Run tests with coverage report
coverage: test
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

## install: Install the wui binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) .
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.txt coverage.html
	@rm -rf dist/
	@echo "Clean complete"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

## mod-tidy: Tidy Go modules
mod-tidy:
	@echo "Tidying modules..."
	@go mod tidy

## image: Build the container image ($(IMAGE))
image:
	@echo "Building container image $(IMAGE)..."
	@$(CONTAINER_ENGINE) build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE) \
		-f Containerfile .
	@echo "Image built: $(IMAGE)"

## image-push: Push the container image to the registry
image-push: image
	@echo "Pushing $(IMAGE)..."
	@$(CONTAINER_ENGINE) push $(IMAGE)

## run: Run the container image, publishing the web GUI on $(HOST_ADDR):$(HOST_PORT) (mounts your Taskwarrior data)
run:
	@$(CONTAINER_ENGINE) run --rm \
		-p $(HOST_ADDR):$(HOST_PORT):7008 \
		-v "$(HOME)/.task:/home/wui/.task:z" \
		-v "$(HOME)/.taskrc:/home/wui/.taskrc:ro,z" \
		$(IMAGE)

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## serve: Run the wui gui
serve:
	wui gui --config $(WUI_CONFIG)	

## rebuild-and-serve
rebuild-and-serve: build install serve

# ---------------------------------------------------------------------------
# wui serve as a systemd --user daemon (Linux)
# ---------------------------------------------------------------------------
# Defaults to where `make install` (go install) puts the binary
WUI_BIN         ?= $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)/$(BINARY_NAME)
WUI_ADDR        ?= localhost:7007
WUI_LOG_LEVEL   ?= info
WUI_SERVE_FLAGS ?=
SYSTEMD_USER_DIR ?= $(HOME)/.config/systemd/user
SERVICE_NAME     = wui-serve.service
SERVICE_TEMPLATE = contrib/systemd/$(SERVICE_NAME).in

## service-install: Install wui and enable+start it as a systemd user service (WUI_ADDR, WUI_LOG_LEVEL, WUI_SERVE_FLAGS)
service-install: install
	@command -v systemctl >/dev/null || { echo "systemctl not found: systemd is required"; exit 1; }
	@test -x "$(WUI_BIN)" || { echo "wui binary not found at $(WUI_BIN); set WUI_BIN=/path/to/wui"; exit 1; }
	@mkdir -p $(SYSTEMD_USER_DIR)
	@sed -e 's|@WUI_BIN@|$(WUI_BIN)|g' \
	     -e 's|@WUI_ADDR@|$(WUI_ADDR)|g' \
	     -e 's|@WUI_LOG_LEVEL@|$(WUI_LOG_LEVEL)|g' \
	     -e 's|@WUI_SERVE_FLAGS@|$(WUI_SERVE_FLAGS)|g' \
	     $(SERVICE_TEMPLATE) > $(SYSTEMD_USER_DIR)/$(SERVICE_NAME)
	@echo "Installed $(SYSTEMD_USER_DIR)/$(SERVICE_NAME)"
	systemctl --user daemon-reload
	systemctl --user enable $(SERVICE_NAME)
	systemctl --user restart $(SERVICE_NAME)
	@echo "wui serve is running on $(WUI_ADDR). To keep it running while logged out: make service-linger"

## service-uninstall: Stop, disable and remove the wui systemd user service
service-uninstall:
	-systemctl --user disable --now $(SERVICE_NAME)
	rm -f $(SYSTEMD_USER_DIR)/$(SERVICE_NAME)
	systemctl --user daemon-reload

## service-restart: Restart the wui systemd user service
service-restart:
	systemctl --user restart $(SERVICE_NAME)

## service-status: Show the status of the wui systemd user service
service-status:
	-systemctl --user status --no-pager $(SERVICE_NAME)

## service-logs: Follow the wui systemd user service logs
service-logs:
	journalctl --user -u $(SERVICE_NAME) -f

## service-linger: Let user services run at boot / without an active login session
service-linger:
	loginctl enable-linger $(USER)
