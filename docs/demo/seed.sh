#!/bin/sh
# Create a throwaway Taskwarrior setup with sample tasks for the README demo.
#
# Usage: docs/demo/seed.sh <dir>
#
# <dir> becomes a fake $HOME holding .taskrc, .task/ and .config/wui/, so the
# demo never touches your real tasks. Run wui against it with HOME=<dir>.
set -eu

dir=${1:?usage: $0 <dir>}
rm -rf "$dir"
mkdir -p "$dir/.task"
dir=$(cd "$dir" && pwd)

cat > "$dir/.taskrc" <<EOF
data.location=$dir/.task
confirmation=off
news.version=99.0.0
EOF

export HOME="$dir" TASKRC="$dir/.taskrc"
t() { task rc.hooks=off rc.verbose=nothing "$@" >/dev/null; }

t add "Review pull request for the API server" project:wui priority:H due:today +review
t add "Write release notes for v1.4" project:wui due:tomorrow +docs
t add "Fix flaky calendar sync test" project:wui priority:M +bug
t add "Record a demo video for the README" project:wui due:friday +docs
t add "Book flights for the conference" project:travel priority:H due:monday
t add "Renew passport" project:travel due:eom +errand
t add "Buy groceries" project:home +errand due:today
t add "Call the plumber about the kitchen sink" project:home priority:M
t add "Plan weekend hike" project:home +outdoors scheduled:saturday
t add "Read \"Designing Data-Intensive Applications\"" project:learning +book
t add "Finish Go concurrency course" project:learning due:2w
t add "Reply to the landlord" wait:3d +errand

# A bit of extra detail for the sidebar.
t 1 annotate "https://github.com/clobrano/wui/pulls"
t 1 annotate "Check the TLS flags and the new endpoints"
t 2 modify depends:1
t 3 start
