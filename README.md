<h1 align="center">wui</h1>
<p align="center"><strong>Warrior UI &mdash; A terminal interface for Taskwarrior that gets out of your way</strong></p>

<p align="center">
  <a href="#installation">Install</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#features-at-a-glance">Features</a> &bull;
  <a href="#configuration">Configuration</a> &bull;
  <a href="#google-calendar-sync">Calendar Sync</a>
</p>

---

If you already live in the terminal and rely on [Taskwarrior](https://taskwarrior.org) to manage your tasks, **wui** gives you a fast, keyboard-driven interface to do it all without leaving the command line.

Built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea), wui is designed for speed, simplicity, and respect for the Taskwarrior workflow you already know.

```
wui
```

That's it. Press `?` for help.

## Why wui?

- **Instant startup** &mdash; no loading screens, no web servers, no Electron
- **Vim-style navigation** &mdash; `h` `j` `k` `l`, `g` `G`, `Ctrl+d` `Ctrl+u` &mdash; feels like home
- **Taskwarrior-native filtering** &mdash; use the same filter syntax you already know (`+work due:tomorrow project:Home`)
- **Grouped views** &mdash; browse tasks organized by project or tag, drill in with `Enter`, back out with `Esc`
- **Batch operations** &mdash; select multiple tasks with `Space`, then act on all of them at once
- **Custom commands** &mdash; wire up any key to run a shell command with task data (`xdg-open {{.url}}`, copy to clipboard, git clone, etc.)
- **Google Calendar sync** &mdash; push tasks to your calendar with `wui sync`
- **Fully configurable** &mdash; tabs, columns, keybindings, themes, sort order &mdash; your workflow, your rules
- **Respects your `.taskrc`** &mdash; reads your existing Taskwarrior contexts and settings

## Features at a Glance

| Feature | Description |
|---|---|
| **Customizable tabs** | Define any number of tabs with their own Taskwarrior filters and sort order |
| **Projects & Tags views** | Special grouped views with task counts &mdash; drill into any group |
| **Search tab** | Persistent search across all tasks (pending, completed, deleted) using Taskwarrior filters |
| **Task detail sidebar** | Full metadata, annotations, dependencies &mdash; scrollable with Ctrl+d/u |
| **Multi-select** | Select tasks with `Space`, then batch-apply done, modify, annotate, delete |
| **Quick modify** | Press `m` to modify tasks inline (`due:tomorrow +urgent priority:H`) |
| **Markdown export** | Press `M` to copy tasks to clipboard as markdown (`* [ ] Description (uuid)`) |
| **Annotation links** | Press `o` to open URLs and file paths found in annotations |
| **Custom commands** | Map any key to a shell command with `{{.field}}` templates |
| **Flexible sorting** | Per-tab sorting: alphabetic, due, scheduled, created, modified &mdash; with reverse option |
| **Short view** | Compact multi-line layout for narrow terminals (auto or forced) |
| **Theme engine** | Dark and light base themes with full ANSI 256-color customization |
| **Custom keybindings** | Remap every action to your preferred keys |
| **Google Calendar sync** | One-way sync to Google Calendar with priority color-coding |
| **CLI integration** | `--search` flag to open with a pre-applied filter (great for scripts and aliases) |

## Installation

### Requirements

- [Taskwarrior 3.x](https://taskwarrior.org) installed and in your PATH
- Go 1.24+ (for building from source)
- Terminal with 256-color support (recommended)

### Install with Go

```bash
go install github.com/clobrano/wui@latest
```

### Build from source

```bash
git clone https://github.com/clobrano/wui.git
cd wui
make build
sudo make install
```

## Quick Start

```bash
# Launch the TUI
wui

# Open straight into a search
wui --search "project:Home +urgent"

# Sync tasks to Google Calendar
wui sync

# Show version
wui version
```

## Keyboard Shortcuts

wui uses vim-style keybindings by default. Every binding is [customizable](#keybindings).

### Navigation

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Jump to first task |
| `G` | Jump to last task |
| `Tab` / `l` / `→` | Next tab |
| `Shift+Tab` / `h` / `←` | Previous tab |
| `1`–`9` | Quick-jump to task or tab |

### Task Actions

| Key | Action |
|---|---|
| `d` | Mark task(s) done |
| `s` | Start / Stop task(s) |
| `x` | Delete task(s) (with confirmation) |
| `e` | Edit task in `$EDITOR` |
| `n` | Create new task |
| `m` | Quick modify task(s) &mdash; e.g. `due:tomorrow +urgent` |
| `M` | Export task(s) as markdown to clipboard |
| `a` | Add annotation to task(s) |
| `o` | Open URL or file path from annotation |
| `u` | Undo last operation |
| `Space` | Toggle multi-select on current task |
| `Esc` | Clear selection |

### Views & Filtering

| Key | Action |
|---|---|
| `Enter` | Toggle sidebar / Drill into group |
| `Esc` | Close sidebar / Back to group list |
| `/` | Enter filter mode (Taskwarrior syntax) |
| `r` | Refresh task list |
| `?` | Toggle help screen |
| `q` | Quit |

### Sidebar Scrolling

| Key | Action |
|---|---|
| `Ctrl+d` / `Ctrl+f` | Scroll down (half / full page) |
| `Ctrl+u` / `Ctrl+b` | Scroll up (half / full page) |

> **Tip:** Dates with time are supported: `due:2026-03-15T14:30` or `scheduled:2026-03-15T09:00`. Times are displayed only when they are not midnight.

## Configuration

wui reads from `~/.config/wui/config.yaml` (created automatically on first run).

### Tabs

Tabs are the heart of wui. Each tab defines a Taskwarrior filter and optional sort order. You can add, remove, and reorder tabs freely.

```yaml
tui:
  tabs:
    - name: "Next"
      filter: "( status:pending or status:active ) -WAITING"

    - name: "Today"
      filter: "due:today"
      sort: "due"

    - name: "Urgent"
      filter: "+urgent"

    - name: "Work"
      filter: "+work -someday"
      sort: "alphabetic"

    - name: "Projects"                           # Special name → grouped view
      filter: "status:pending or status:active"

    - name: "Tags"                               # Special name → grouped view
      filter: "status:pending or status:active"

    - name: "Recent"
      filter: "status:pending"
      sort: "modified"
      reverse: true                              # Most recently modified first
```

**Special tab names:**

- **Search** &mdash; always auto-prepended as the first tab (⌕). Searches across all statuses by default. Cannot be removed or reordered.
- **Projects** &mdash; shows tasks grouped by project with counts. Press `Enter` to drill in.
- **Tags** &mdash; shows tasks grouped by tag with counts. Press `Enter` to drill in.

> Renaming "Projects" or "Tags" to anything else turns them into regular flat-list tabs.

### Sorting

Each tab supports per-tab sorting:

| Sort value | Description |
|---|---|
| `alphabetic` (or `alpha`, `description`) | Sort by description, case-insensitive |
| `due` | Sort by due date (no-date tasks last) |
| `scheduled` | Sort by scheduled date (no-date tasks last) |
| `created` (or `entry`) | Sort by creation date |
| `modified` | Sort by modification date (no-date tasks last) |

Add `reverse: true` to invert the order. Completed tasks always sort to the bottom.

### Columns

Choose up to 6 columns for the task list (case-insensitive):

```yaml
tui:
  columns:
    - id
    - project
    - priority
    - due
    - description
```

Available columns: `id`, `project`, `priority`, `due`, `tags`, `description`.

### Sidebar

```yaml
tui:
  sidebar_width: 33  # Percentage of terminal width (1–100)
```

### Short View (Narrow Terminals)

When the terminal is less than 80 columns wide, wui switches to a compact layout:

```
▶ Fix login page crash
  DUE:  2026-02-20
  TAGS: +bug, +frontend
```

Configure up to 3 fields below each task description:

```yaml
tui:
  narrow_view_fields:
    - name: due
      label: "DUE"
    - name: tags
      label: "TAGS"
    - name: project
      label: "PROJECT"

  # Force narrow view at any terminal width:
  # force_small_screen: true
```

### Keybindings

Remap any action:

```yaml
tui:
  keybindings:
    quit: q
    help: "?"
    up: k
    down: j
    first: g
    last: G
    page_up: ctrl+u
    page_down: ctrl+d
    done: d
    delete: x
    edit: e
    modify: m
    annotate: a
    new: n
    undo: u
    filter: "/"
    refresh: r
```

### Custom Commands

Map any key to a shell command using `{{.fieldname}}` templates. All Taskwarrior fields and custom UDAs are available.

```yaml
tui:
  custom_commands:
    O:
      name: "Open URL"
      command: "xdg-open {{.url}}"
      description: "Opens the task's URL in default browser"

    "1":
      name: "Copy Description"
      command: "echo {{.description}} | xclip -selection clipboard"
      description: "Copy task description to clipboard"

    c:
      name: "Git Clone"
      command: "git clone {{.url}} ~/projects/{{.project}}"
      description: "Clone repository to project folder"
```

Custom commands appear in the help screen (`?`). Platform-specific examples:

| Platform | Command |
|---|---|
| Linux | `xdg-open {{.url}}` |
| macOS | `open {{.url}}` |
| Termux (Android) | `termux-open-url {{.url}}` |
| Windows | `cmd /c start {{.url}}` |

> If a custom command key conflicts with a built-in shortcut, wui warns on exit. Add `silence_shortcut_override_warnings: true` to suppress this.

For the full reference, see [`docs/custom-commands.md`](docs/custom-commands.md).

### Themes

wui ships with **dark** and **light** base themes. Override any color with ANSI 256-color codes:

```yaml
tui:
  theme:
    name: dark  # "dark" or "light" (any other name uses dark as base)

    # Priority colors
    # priority_high: "9"
    # priority_medium: "11"
    # priority_low: "12"

    # Due date colors
    # due_overdue: "9"
    # due_today: "11"
    # due_soon: "11"

    # Status colors
    # status_active: "15"
    # status_waiting: "8"
    # status_completed: "8"

    # UI elements
    # header_fg: "12"
    # footer_fg: "246"
    # separator_fg: "8"
    # selection_bg: "12"
    # selection_fg: "0"
    # sidebar_border: "8"
    # sidebar_title: "12"
    # label_fg: "12"
    # value_fg: "15"
    # dim_fg: "8"
    # error_fg: "9"
    # success_fg: "10"
    # tag_fg: "14"

    # Tab colors
    # section_active_fg: "15"
    # section_active_bg: "63"
    # section_inactive_fg: "246"
```

Color values are [ANSI 8-bit codes](https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit) (0–255). Standard colors 0–15 work everywhere; 16–255 require 256-color terminal support.

## Filtering

wui uses Taskwarrior's native filter syntax. Press `/` to filter:

```
+work -someday                 # Tag include/exclude
project:Home due:tomorrow      # Field matches
status:pending priority:H      # Status and priority
(+urgent or +important)        # Boolean logic
```

The **Search tab** (⌕) searches across all statuses by default. Other tabs filter within their own scope.

## Google Calendar Sync

Sync your tasks to Google Calendar as color-coded all-day events.

### Setup

1. Create a project in [Google Cloud Console](https://console.cloud.google.com) and enable the **Google Calendar API**
2. Create OAuth 2.0 credentials (Desktop app) with `http://localhost:8080` as redirect URI
3. Download `credentials.json` to `~/.config/wui/`
4. Configure in `config.yaml`:

```yaml
calendar_sync:
  enabled: true
  calendar_name: "Tasks"
  task_filter: "status:pending or status:completed"
  credentials_path: ~/.config/wui/credentials.json
  token_path: ~/.config/wui/token.json
```

### Usage

```bash
# Sync using config settings
wui sync

# Override calendar or filter on the fly
wui sync --calendar "Work" --filter "+work due.before:eow"
wui sync --calendar "Urgent" --filter "+urgent priority:H"
```

On first run, you'll authorize via browser. The token is saved to `~/.config/wui/token.json`.

### How it works

- Tasks become all-day events based on due date (or scheduled date as fallback)
- Events include UUID, project, tags, and status in the description
- Completed tasks show a **✓** checkmark in the title
- Events are color-coded by priority (red = high, yellow = medium)
- Existing events are updated when tasks change
- Sync is **one-way**: Taskwarrior → Google Calendar

## CLI Reference

```
wui                              Launch the TUI
wui version                      Print version info
wui sync                         Sync tasks to Google Calendar

Flags:
  --config string                Config file path (default: ~/.config/wui/config.yaml)
  --taskrc string                Taskrc file path (default: ~/.taskrc)
  --task-bin string              Task binary path (default: /usr/local/bin/task)
  --search string                Open in Search tab with a pre-applied filter
  --log-level string             Log level: debug, info, warn, error (default: error)
  --log-format string            Log format: text, json (default: text)
```

### Logging

Log level priority: CLI flag > `WUI_LOG_LEVEL` env var > config file.

Logs go to `/tmp/wui.log` by default. Override with `WUI_LOG_FILE` environment variable.

```bash
# Debug a specific session
wui --log-level debug

# Or via environment
export WUI_LOG_LEVEL=info
wui
```

## Comparison to taskwarrior-tui

wui is inspired by [taskwarrior-tui](https://github.com/kdheepak/taskwarrior-tui) and builds on the idea with a different set of priorities:

| | wui | taskwarrior-tui |
|---|---|---|
| Language | Go | Rust |
| Framework | [Bubbletea](https://github.com/charmbracelet/bubbletea) | termbox |
| Grouped views | Projects & Tags with drill-down | &mdash; |
| Custom tabs | Unlimited, with per-tab filters & sorting | &mdash; |
| Custom commands | `{{.field}}` template shell commands | &mdash; |
| Calendar sync | Google Calendar | &mdash; |
| Multi-select | Batch operations on selected tasks | &mdash; |
| Markdown export | Copy tasks to clipboard | &mdash; |
| Short view | Auto-adapts to narrow terminals | &mdash; |

## Development

See [CLAUDE.md](CLAUDE.md) for the development guide and architecture overview.

```bash
make build       # Build binary
make test        # Run tests with coverage
make lint        # Run linter
make fmt         # Format code
make clean       # Clean build artifacts
```

## License

MIT License &mdash; see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please open an issue or pull request on [GitHub](https://github.com/clobrano/wui).

## Credits

- Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) by [Charm](https://charm.sh)
- Inspired by [gh-dash](https://github.com/dlvhdr/gh-dash) and [taskwarrior-tui](https://github.com/kdheepak/taskwarrior-tui)
- [Taskwarrior](https://taskwarrior.org) by Göteborg Bit Factory
