# Running wui in a container

This directory contains everything needed to build a container image of the
`wui gui` web interface and run it as a rootless **systemd user service** via
Podman.

- [`../Containerfile`](../Containerfile) — multi-stage build (Go builder →
  Alpine runtime with Taskwarrior installed).
- [`container-entrypoint.sh`](container-entrypoint.sh) — bootstraps a
  Taskwarrior data dir and taskrc so the image runs out of the box.
- [`systemd/wui.container`](systemd/wui.container) — Podman Quadlet unit.

The container runs `wui gui`, which serves the **web UI on port 7008** and
starts the REST API as an internal child process on 7007. The GUI proxies
`/api/v1/` to that API, so **only 7008 needs to be published** — open
`http://<host>:7008` in a browser. It has **no authentication**, so the unit
binds it to loopback only; use [Tailscale](../README.md#secure-access-with-tailscale)
or a reverse proxy for remote access.

> To run the raw REST API instead (e.g. for the
> [wui-android](../README.md#using-from-the-wui-android-flutter-app-android-linux-web)
> client), override the command to `serve --addr 0.0.0.0:7007` and publish
> 7007 — see the notes at the end.

## Build the image

```bash
# With Make (injects version metadata):
make image

# Or directly:
podman build -t quay.io/clobrano/wui:latest .
```

Override the engine, repo, or tag:

```bash
make image CONTAINER_ENGINE=docker IMAGE_REPO=localhost/wui IMAGE_TAG=dev
```

## Run once (foreground, for testing)

```bash
make run
# equivalent to:
podman run --rm -p 127.0.0.1:7008:7008 \
    -v "$HOME/.task:/home/wui/.task:z" \
    -v "$HOME/.taskrc:/home/wui/.taskrc:ro,z" \
    quay.io/clobrano/wui:latest

# then open http://localhost:7008 in a browser, or:
curl http://localhost:7008/api/v1/version   # proxied to the internal API
```

`make run` publishes on `127.0.0.1:7008` by default, so it is reachable only
from the host. Note that binding the GUI to `0.0.0.0` *inside* the container
(the image default) does not expose it to your network on its own — what
matters is the host address Podman publishes to. Override `HOST_ADDR` (and
`HOST_PORT`) to change that:

```bash
# Reach it over Tailscale — publish on the Tailscale IP only:
make run HOST_ADDR=$(tailscale ip -4)

# Or expose on all host interfaces (LAN + internet — no auth, avoid this):
make run HOST_ADDR=0.0.0.0
```

Publishing on the Tailscale IP keeps the GUI off the LAN and public internet
while making it reachable from your other Tailscale devices — just open
`http://<tailscale-ip>:7008` there. See
[Secure access with Tailscale](../README.md#secure-access-with-tailscale) for
the full walkthrough.

## Run as a systemd user service (Podman Quadlet)

Requires Podman 4.4+ (Quadlet).

```bash
# 1. Make sure the host paths the unit mounts exist:
mkdir -p ~/.task ~/.config/wui
touch ~/.taskrc          # or symlink/copy your real ~/.taskrc

# 2. Install the unit:
mkdir -p ~/.config/containers/systemd
cp deploy/systemd/wui.container ~/.config/containers/systemd/

# 3. Reload and start:
systemctl --user daemon-reload
systemctl --user start wui
systemctl --user status wui

# 4. (Optional) keep it running after logout, without an active session:
loginctl enable-linger "$USER"
```

Logs:

```bash
journalctl --user -u wui -f
```

Update to a newer image:

```bash
podman pull quay.io/clobrano/wui:latest
systemctl --user restart wui
# or enable automatic updates (the unit sets AutoUpdate=registry):
systemctl --user enable --now podman-auto-update.timer
```

### Notes

- **Ownership / SELinux.** The unit uses `UserNS=keep-id:uid=1000,gid=1000` so
  the host user maps to the image's `wui` user (uid 1000), keeping bind-mounted
  files readable/writable. The `:z` volume flag relabels for SELinux hosts;
  drop it on non-SELinux systems if you prefer.
- **Existing Taskwarrior data.** Point the `%h/.task` and `%h/.taskrc` volumes
  at your real files to drive your existing tasks. If `~/.taskrc` does not
  exist when the container starts, the entrypoint writes a minimal one.
- **Changing flags / API-only mode.** Edit the `Exec=` line in the unit (e.g.
  add `--log-level info`). To serve the raw REST API instead of the web GUI,
  set `Exec=serve --addr 0.0.0.0:7007` and change `PublishPort` to `7007`.
