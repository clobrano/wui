#!/bin/sh
# Entrypoint for the wui container.
#
# Ensures Taskwarrior has a data directory and a taskrc before starting, so the
# image runs out of the box even when no existing Taskwarrior data is mounted.
# Bind-mounting a real ~/.taskrc or ~/.task simply overlays these defaults.
set -eu

: "${HOME:=/home/wui}"
: "${TASKDATA:=$HOME/.task}"
export TASKDATA

mkdir -p "$TASKDATA"

# Create a minimal taskrc if none is present. Taskwarrior would otherwise prompt
# interactively on first run, which fails in a non-interactive container.
if [ ! -f "$HOME/.taskrc" ]; then
    printf 'data.location=%s\n' "$TASKDATA" > "$HOME/.taskrc"
fi

exec wui "$@"
