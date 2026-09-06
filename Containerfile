# Containerfile for wui — builds a container image that runs the `wui gui`
# web interface (backed by Taskwarrior).
#
# Build (with Make): make image
# Build (directly):  podman build -t quay.io/clobrano/wui:latest .
#
# Runtime notes:
#   * `wui gui` serves the web UI on 0.0.0.0:7008 inside the container and
#     starts the REST API as an internal child process on localhost:7007. The
#     GUI proxies /api/v1/ to it, so only 7008 needs to be published; browsers
#     (local or remote over Tailscale) talk only to the GUI port.
#   * To run the raw REST API instead, override the command with: serve
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

EXPOSE 7008

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["gui", "--port", "7008"]
