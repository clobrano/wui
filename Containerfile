# Containerfile for wui — builds a container image that runs the
# `wui serve` REST API server backed by Taskwarrior.
#
# Build (with Make): make image
# Build (directly):  podman build -t quay.io/clobrano/wui:latest .
#
# Runtime notes:
#   * The server listens on 0.0.0.0:7007 inside the container.
#   * Taskwarrior data lives under $TASKDATA (default /home/wui/.task) and the
#     taskrc under /home/wui/.taskrc — bind-mount your own to use existing data.
#   * Runs as an unprivileged user (uid/gid 1000), suitable for rootless podman.

# ---- build stage ---------------------------------------------------------
FROM golang:1.24-alpine AS build

WORKDIR /src

# Cache module downloads separately from the source for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Version metadata is injected at build time (see Makefile).
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 go build \
        -ldflags "-s -w \
            -X github.com/clobrano/wui/internal/version.Version=${VERSION} \
            -X github.com/clobrano/wui/internal/version.Commit=${COMMIT} \
            -X github.com/clobrano/wui/internal/version.BuildDate=${BUILD_DATE}" \
        -o /out/wui .

# ---- runtime stage -------------------------------------------------------
FROM alpine:3.20

# Taskwarrior provides the `task` binary that wui shells out to. tzdata keeps
# due-date handling correct; ca-certificates is needed for outbound sync.
RUN apk add --no-cache task tzdata ca-certificates

# Unprivileged user. uid/gid 1000 pairs with `UserNS=keep-id:uid=1000,gid=1000`
# in the systemd unit so bind-mounted host files line up in ownership.
RUN addgroup -g 1000 wui \
    && adduser -D -u 1000 -G wui wui

COPY --from=build /out/wui /usr/local/bin/wui
COPY deploy/container-entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER wui
ENV HOME=/home/wui \
    TASKDATA=/home/wui/.task

# Data and config live on volumes so they survive container recreation.
VOLUME ["/home/wui/.task", "/home/wui/.config/wui"]

EXPOSE 7007

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["serve", "--addr", "0.0.0.0:7007"]
