.PHONY: build test install clean help image image-push run service-install service-uninstall service-status service-logs service-restart service-linger container-install container-uninstall tailscale-serve tailscale-serve-off tailscale-serve-status

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
# wui as a systemd --user daemon (Linux)
# ---------------------------------------------------------------------------
# Defaults to where `make install` (go install) puts the binary
WUI_BIN         ?= $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)/$(BINARY_NAME)
# Web GUI port. `wui gui` listens on all interfaces (reachable from the LAN)
# and starts the REST API itself on localhost:7007.
WUI_PORT        ?= 7008
# What the service runs. For the REST API only, e.g.: WUI_ARGS="serve --addr :7007"
WUI_ARGS        ?= gui --port $(WUI_PORT) --no-browser
WUI_LOG_LEVEL   ?= info
# 1 = start the service at boot, without a login session (loginctl enable-linger)
WUI_LINGER      ?= 1
# 1 = also serve the web GUI over HTTPS on your tailnet (see tailscale-serve)
WUI_TAILSCALE   ?= 0
SYSTEMD_USER_DIR ?= $(HOME)/.config/systemd/user
SERVICE_NAME     = wui-serve.service
SERVICE_TEMPLATE = deploy/systemd/$(SERVICE_NAME).in

## service-install: Install wui and enable+start it as a systemd user service, started at boot (WUI_PORT, WUI_ARGS, WUI_LOG_LEVEL, WUI_LINGER)
service-install: install
	@command -v systemctl >/dev/null || { echo "systemctl not found: systemd is required"; exit 1; }
	@test -x "$(WUI_BIN)" || { echo "wui binary not found at $(WUI_BIN); set WUI_BIN=/path/to/wui"; exit 1; }
	@mkdir -p $(SYSTEMD_USER_DIR)
	@sed -e 's|@WUI_BIN@|$(WUI_BIN)|g' \
	     -e 's|@WUI_ARGS@|$(WUI_ARGS)|g' \
	     -e 's|@WUI_LOG_LEVEL@|$(WUI_LOG_LEVEL)|g' \
	     $(SERVICE_TEMPLATE) > $(SYSTEMD_USER_DIR)/$(SERVICE_NAME)
	@echo "Installed $(SYSTEMD_USER_DIR)/$(SERVICE_NAME)"
	systemctl --user daemon-reload
	systemctl --user enable $(SERVICE_NAME)
	systemctl --user restart $(SERVICE_NAME)
ifeq ($(WUI_LINGER),1)
	loginctl enable-linger "$$(id -un)"
	@echo "wui is running ($(WUI_ARGS)) and will start at boot"
else
	@echo "wui is running ($(WUI_ARGS)); starts at login only, run 'make service-linger' to start it at boot"
endif
ifeq ($(WUI_TAILSCALE),1)
	@$(MAKE) --no-print-directory tailscale-serve TAILSCALE_PORT=$(WUI_PORT)
endif

## service-uninstall: Stop, disable and remove the wui systemd user service (and its Tailscale HTTPS proxy if WUI_TAILSCALE=1)
service-uninstall:
	-systemctl --user disable --now $(SERVICE_NAME)
	rm -f $(SYSTEMD_USER_DIR)/$(SERVICE_NAME)
	systemctl --user daemon-reload
ifeq ($(WUI_TAILSCALE),1)
	-@$(MAKE) --no-print-directory tailscale-serve-off
endif

## service-restart: Restart the wui systemd user service
service-restart:
	systemctl --user restart $(SERVICE_NAME)

## service-status: Show the status of the wui systemd user service
service-status:
	-systemctl --user status --no-pager $(SERVICE_NAME)

## service-logs: Follow the wui systemd user service logs
service-logs:
	journalctl --user -u $(SERVICE_NAME) -f

## service-linger: Let user services run at boot / without an active login session (done by service-install unless WUI_LINGER=0)
service-linger:
	loginctl enable-linger "$$(id -un)"

# ---------------------------------------------------------------------------
# wui container as a systemd --user service (Podman Quadlet)
# ---------------------------------------------------------------------------
QUADLET_DIR      ?= $(HOME)/.config/containers/systemd
QUADLET_UNIT      = deploy/systemd/wui.container

## container-install: Install the Podman Quadlet unit and start the wui container, started at boot (WUI_LINGER, WUI_TAILSCALE)
container-install:
	@command -v podman >/dev/null || { echo "podman not found: Podman 4.4+ is required"; exit 1; }
	@mkdir -p $(HOME)/.task $(HOME)/.config/wui $(QUADLET_DIR)
	@test -e $(HOME)/.taskrc || touch $(HOME)/.taskrc
	cp $(QUADLET_UNIT) $(QUADLET_DIR)/wui.container
	systemctl --user daemon-reload
	systemctl --user restart wui.service
ifeq ($(WUI_LINGER),1)
	loginctl enable-linger "$$(id -un)"
endif
ifeq ($(WUI_TAILSCALE),1)
	@$(MAKE) --no-print-directory tailscale-serve TAILSCALE_PORT=$(HOST_PORT)
endif

## container-uninstall: Stop the wui container and remove its Quadlet unit (and its Tailscale HTTPS proxy if WUI_TAILSCALE=1)
container-uninstall:
	-systemctl --user stop wui.service
	rm -f $(QUADLET_DIR)/wui.container
	systemctl --user daemon-reload
ifeq ($(WUI_TAILSCALE),1)
	-@$(MAKE) --no-print-directory tailscale-serve-off
endif

# ---------------------------------------------------------------------------
# HTTPS over Tailscale (needed to install the web GUI as an app on other devices)
# ---------------------------------------------------------------------------
# `tailscale serve` gets and renews a certificate for <machine>.<tailnet>.ts.net
# and proxies https://<machine>.<tailnet>.ts.net to the local web GUI. Tailscale
# keeps this setting across reboots, so it only has to be run once. Enable
# "HTTPS Certificates" at https://login.tailscale.com/admin/dns first.
TAILSCALE_PORT  ?= $(WUI_PORT)
# Changing tailscale serve needs root unless you ran `sudo tailscale set --operator=$$USER`;
# in that case pass TAILSCALE_SUDO=
TAILSCALE_SUDO  ?= sudo

## tailscale-serve: Serve the web GUI over HTTPS on your tailnet (TAILSCALE_PORT, TAILSCALE_SUDO)
tailscale-serve:
	@command -v tailscale >/dev/null || { echo "tailscale not found: install it from https://tailscale.com/download"; exit 1; }
	$(TAILSCALE_SUDO) tailscale serve --bg --https=443 http://127.0.0.1:$(TAILSCALE_PORT)
	@tailscale serve status

## tailscale-serve-off: Stop serving the web GUI over HTTPS on your tailnet
tailscale-serve-off:
	$(TAILSCALE_SUDO) tailscale serve --https=443 off

## tailscale-serve-status: Show what Tailscale serves over HTTPS
tailscale-serve-status:
	tailscale serve status
