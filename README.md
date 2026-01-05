# wui - Warrior UI

A modern, fast Terminal User Interface (TUI) for [Taskwarrior](https://taskwarrior.org), built with Go and [bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- **Intuitive keyboard-driven interface** - Navigate and manage tasks efficiently without touching the mouse
- **Multiple views** - Next, Waiting, Projects, Tags, and custom filtered views
- **Detailed task sidebar** - View all task metadata, annotations, and dependencies
- **Quick task modifications** - Add tags, change due dates, annotate, and more with simple commands
- **Grouped views** - Browse tasks by project or tag with task counts
- **Rich task metadata** - Full support for priorities, dates, dependencies, and recurrence
- **Google Calendar sync** - Synchronize tasks to Google Calendar with customizable filters
- **Configurable** - Customize keybindings, colors, columns, and fully customize tabs/sections
- **Respects Taskwarrior config** - Reads your `.taskrc` for contexts and settings

## Installation

### From Source

Requirements:
- Go 1.24 or higher
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

### Multi-Select
- `Space` - Toggle task selection
- `Esc` - Clear all selections

**Note:** You can select multiple tasks with `Space`, then apply actions (mark done, modify, annotate, delete, etc.) to all selected tasks at once.

### Task Actions
- `d` - Mark task(s) done
- `s` - Start/Stop task(s)
- `x` - Delete task(s) (with confirmation)
- `e` - Edit task in $EDITOR
- `n` - Create new task
- `m` - Quick modify task(s) (e.g., `due:tomorrow +urgent` or `due:2025-10-20T14:30`)
- `M` - Export task(s) as markdown to clipboard (format: `* [ ] Description (uuid)`)
- `a` - Add annotation to task(s)
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
  #
  # SPECIAL TAB NAMES:
  # - "Search" - Reserved, always auto-prepended as first tab (⌕)
  # - "Projects" - Shows grouped view by project (not a flat task list)
  # - "Tags" - Shows grouped view by tag (not a flat task list)
  #
  # Note: Renaming "Projects" or "Tags" will change their behavior to show
  # a regular flat task list instead of the grouped view.
  tabs:
    - name: "Next"
      filter: "( status:pending or status:active ) -WAITING"
    - name: "Today"
      filter: "due:today"
    - name: "Urgent"
      filter: "+urgent"
    - name: "Work"
      filter: "+work -someday"
    - name: "Projects"
      filter: "status:pending or status:active"
    - name: "Tags"
      filter: "status:pending or status:active"
    - name: "All"
      filter: "status:pending or status:active"

  # Keybindings - customize keyboard shortcuts
  # All keybindings are optional; omitted keys use defaults
  # Available actions: quit, help, up, down, first, last, page_up, page_down,
  #                    done, delete, edit, modify, annotate, new, undo, filter, refresh
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

  # Theme (dark or light)
  theme:
    name: dark
```

## CLI Flags

```
  --config string      Config file path (default: ~/.config/wui/config.yaml)
  --taskrc string      Taskrc file path (default: ~/.taskrc)
  --task-bin string    Task binary path (default: /usr/local/bin/task)
  --search string      Open in Search tab with the specified filter
  --log-level string   Log level: debug, info, warn, error (default: error)
  --log-format string  Log format: text, json (default: text)
```

Examples:
```bash
# Open with custom taskrc and debug logging
wui --taskrc ~/work/.taskrc --log-level debug

# Open in Search tab with a pre-applied filter (useful in pipelines)
wui --search "project:Home +urgent"

# Search for tasks due today
wui --search "due:today"

# Search for completed tasks
wui --search "status:completed"
```

Logs are written to `/tmp/wui.log` by default (or `$WUI_LOG_FILE`).

## Google Calendar Sync

wui can synchronize your Taskwarrior tasks to Google Calendar. This feature allows you to visualize your tasks in a calendar view and integrate with your existing workflow.

### Setup

1. **Create a Google Cloud Project**:
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project or select an existing one
   - Enable the Google Calendar API

2. **Download Credentials**:
   - In Google Cloud Console, go to APIs & Services > Credentials
   - Create OAuth 2.0 credentials (Desktop app)
   - Add `http://localhost:8080` as an authorized redirect URI
   - Download the credentials file as `credentials.json`
   - Place it in `~/.config/wui/credentials.json`

3. **Configure sync** in `~/.config/wui/config.yaml`:
```yaml
calendar_sync:
  enabled: true
  calendar_name: "Tasks"                      # Name of your Google Calendar (required)
  task_filter: "status:pending or status:completed"  # Taskwarrior filter (required)
  credentials_path: ~/.config/wui/credentials.json
  token_path: ~/.config/wui/token.json
```

**Note**: Including `status:completed` in the filter allows completed tasks to be updated in the calendar (they'll show a ✓ checkmark). You can customize this filter as needed.

### Usage

Once configured in `config.yaml`, simply run:
```bash
# Sync using config.yaml settings
wui sync
```

You can also override config settings with command-line flags:
```bash
# Override calendar name from config
wui sync --calendar "Important Tasks"

# Override both calendar and filter
wui sync --calendar "Work" --filter "+work due.before:eow"

# Sync high-priority tasks to a different calendar
wui sync --calendar "Urgent" --filter "+urgent priority:H"
```

**Note**: On first run, you'll be prompted to authorize the app in your browser. The authorization token will be saved to `~/.config/wui/token.json`.

### How it works

- Tasks are synced as all-day events in Google Calendar
- Each event includes the task UUID, project, tags, and status in its description
- Event dates are based on the task's due date, or scheduled date if no due date is set
- Completed tasks are marked with a ✓ checkmark in the title
- Events are color-coded based on priority (red for high, yellow for medium)
- Existing events are updated if the task changes
- The sync is one-way: Taskwarrior → Google Calendar

## Filtering

wui supports Taskwarrior's powerful filter syntax:

```
+work -someday              # Tasks with +work tag, without +someday
project:Home due:tomorrow   # Home project tasks due tomorrow
status:pending priority:H   # High priority pending tasks
(+urgent or +important)     # Tasks with urgent OR important tag
```

Press `/` to enter filter mode, then type your filter and press Enter.

## Special Tabs

### Search Tab (⌕)
The Search tab is a special, non-configurable tab that:
- Always appears first (cannot be removed or reordered)
- Shows nothing initially - press `/` to enter a search filter
- Searches across **all tasks** in your database (pending, completed, deleted, etc.) by default
- Remembers your search filter for the session (persists when switching tabs)
- Uses `status.any:` automatically unless you specify a status filter

Examples:
```
bug                    # Search for 'bug' in all tasks
project:home           # Home project tasks (all statuses)
status:completed       # Only completed tasks
+urgent due.before:eom # Urgent tasks due before end of month
```

### Projects and Tags Views
**Projects** and **Tags** are special tab names that trigger grouped views:
- **Projects tab**: Shows tasks grouped by project (with task counts)
- **Tags tab**: Shows tasks grouped by tag (with task counts)
- Press `Enter` on a group to drill into it and see the tasks
- Press `Esc` to go back to the group list

**Important**: If you rename these tabs (e.g., "Projects" → "My Projects"), they will behave like regular tabs and show a flat task list instead of the grouped view. The exact names "Projects" and "Tags" are required for the special grouping behavior.

## Default Tabs

wui comes with these default tabs (all customizable except Search):
- **⌕ Search** - Special search tab (auto-prepended, non-configurable)
- **Next** - `( status:pending or status:active ) -WAITING` (tasks ready to work on)
- **Waiting** - `status:waiting` (blocked or scheduled for later)
- **Projects** - Grouped view by project (special name - see above)
- **Tags** - Grouped view by tag (special name - see above)
- **All** - All pending and active tasks

Customize tabs in your config file - reorder, remove, or add your own (except Search)!

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
