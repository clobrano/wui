# wui - Warrior UI

A modern, fast Terminal User Interface (TUI) for [Taskwarrior](https://taskwarrior.org), built with Go and [bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- **Intuitive keyboard-driven interface** - Navigate and manage tasks efficiently without touching the mouse
- **Multiple views** - Next, Waiting, Projects, Tags, and custom filtered views
- **Detailed task sidebar** - View all task metadata, annotations, dependencies, and UDAs
- **Quick task modifications** - Add tags, change due dates, annotate, and more with simple commands
- **Grouped views** - Browse tasks by project or tag with task counts
- **Rich task metadata** - Full support for priorities, dates, dependencies, recurrence, and custom UDAs
- **Configurable** - Customize keybindings, colors, columns, and fully customize tabs/sections
- **Respects Taskwarrior config** - Reads your `.taskrc` for UDAs, contexts, and settings

## Installation

### From Source

Requirements:
- Go 1.21 or higher
- Taskwarrior 3.x

```bash
go install github.com/clobrano/wui@latest
```

Or clone and build:

```bash
git clone https://github.com/clobrano/wui.git
cd wui
make build
sudo make install
```

## Quick Start

Simply run:

```bash
wui
```

Press `?` for help and keybinding reference.

## Keyboard Shortcuts

### Task Navigation
- `j` / `↓` - Move down
- `k` / `↑` - Move up
- `g` - Jump to first task
- `G` - Jump to last task
- `1-9` - Quick jump to visible task

### Section Navigation
- `Tab` / `l` / `→` - Next section
- `Shift+Tab` / `h` / `←` - Previous section
- `1-5` - Jump to section by number

### Task Actions
- `d` - Mark task done
- `s` - Start/Stop task
- `x` - Delete task (with confirmation)
- `e` - Edit task in $EDITOR
- `n` - Create new task
- `m` - Quick modify (e.g., `due:tomorrow +urgent` or `due:2025-10-20T14:30`)
- `a` - Add annotation
- `u` - Undo last operation

**Note:** Dates with time can be set using `due:YYYY-MM-DDTHH:MM` or `scheduled:YYYY-MM-DDTHH:MM` format. Times are displayed only when not midnight.

### View Controls
- `Enter` - Toggle task details sidebar / Drill into group
- `Esc` - Close sidebar / Back to group list
- `/` - Filter tasks (Taskwarrior syntax)
- `r` - Refresh task list
- `?` - Toggle help screen
- `q` - Quit

### Sidebar Scrolling (when sidebar is open)
- `Ctrl+d` - Scroll down half page
- `Ctrl+u` - Scroll up half page
- `Ctrl+f` - Scroll down full page
- `Ctrl+b` - Scroll up full page

## Configuration

wui reads from `~/.config/wui/config.yaml` (created automatically on first run).

### Example Configuration

```yaml
# Task binary location
task_bin: /usr/local/bin/task

# Taskwarrior config file
taskrc_path: ~/.taskrc

# TUI-specific settings
tui:
  # Sidebar width (percentage of terminal width, 1-100)
  sidebar_width: 33

  # Display columns (max 6, case-insensitive)
  # Available: id, project, priority, due, tags, description
  # Default: id, project, priority, due, description
  columns:
    - id
    - project
    - priority
    - due
    - description

  # Tabs/sections - fully customizable!
  # You can reorder, remove defaults, or add your own
  tabs:
    - name: "Next"
      filter: "( status:pending or status:active ) -WAITING"
    - name: "Today"
      filter: "due:today"
    - name: "Urgent"
      filter: "+urgent"
    - name: "Work"
      filter: "+work -someday"
    - name: "All"
      filter: "status:pending or status:active"

  # Theme (dark or light)
  theme:
    name: dark
```

## CLI Flags

```
  --config string      Config file path (default: ~/.config/wui/config.yaml)
  --taskrc string      Taskrc file path (default: ~/.taskrc)
  --task-bin string    Task binary path (default: /usr/local/bin/task)
  --log-level string   Log level: debug, info, warn, error (default: error)
  --log-format string  Log format: text, json (default: text)
```

Example:
```bash
wui --taskrc ~/work/.taskrc --log-level debug
```

Logs are written to `/tmp/wui.log` by default (or `$WUI_LOG_FILE`).

## Filtering

wui supports Taskwarrior's powerful filter syntax:

```
+work -someday              # Tasks with +work tag, without +someday
project:Home due:tomorrow   # Home project tasks due tomorrow
status:pending priority:H   # High priority pending tasks
(+urgent or +important)     # Tasks with urgent OR important tag
```

Press `/` to enter filter mode, then type your filter and press Enter.

## Projects and Tags Views

- **Projects**: Press `Tab` to navigate to Projects section, shows tasks grouped by project
- **Tags**: Navigate to Tags section to see tasks grouped by tag
- Press `Enter` on a group to drill into it and see the tasks
- Press `Esc` to go back to the group list

## Sections

Default tabs (fully customizable via config):
- **Next** - `( status:pending or status:active ) -WAITING` (tasks ready to work on)
- **Waiting** - `status:waiting` (blocked or scheduled for later)
- **Projects** - Grouped view by project
- **Tags** - Grouped view by tag
- **All** - All pending and active tasks

Customize tabs in your config file - reorder, remove, or add your own!

## Development

See [CLAUDE.md](CLAUDE.md) for development guide and architecture overview.

## Requirements

- Taskwarrior 3.x (must be installed and in PATH)
- Terminal with color support (recommended: 256 colors or truecolor)

## Comparison to taskwarrior-tui

wui is inspired by [taskwarrior-tui](https://github.com/kdheepak/taskwarrior-tui) but built with different goals:

- Modern bubbletea framework vs termbox
- Written in Go vs Rust
- Focus on speed and simplicity
- Better UDA support
- Grouped views (Projects/Tags)
- Fully customizable tabs

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please open an issue or pull request on GitHub.

## Credits

- Built with [bubbletea](https://github.com/charmbracelet/bubbletea) by Charm
- Inspired by [gh-dash](https://github.com/dlvhdr/gh-dash) and [taskwarrior-tui](https://github.com/kdheepak/taskwarrior-tui)
- Taskwarrior by [Göteborg Bit Factory](https://taskwarrior.org)

---

Made with ❤️ for productive task management
