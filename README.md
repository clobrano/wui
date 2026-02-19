# wui - Warrior UI

A modern, fast Terminal User Interface (TUI) for [Taskwarrior](https://taskwarrior.org), built with Go and [bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- **Intuitive keyboard-driven interface** - Navigate and manage tasks efficiently without touching the mouse
- **Multiple views** - Next, Waiting, Projects, Tags, and custom filtered views
- **Flexible sorting** - Sort tasks by due date, creation date, alphabetically, and more with per-tab configuration
- **Detailed task sidebar** - View all task metadata, annotations, and dependencies
- **Quick task modifications** - Add tags, change due dates, annotate, and more with simple commands
- **Grouped views** - Browse tasks by project or tag with task counts
- **Rich task metadata** - Full support for priorities, dates, dependencies, and recurrence
- **Custom commands** - Execute system commands with task data using flexible templates (e.g., open URLs, copy to clipboard)
- **Google Calendar sync** - Synchronize tasks to Google Calendar with customizable filters
- **Short view** - Compact single-column layout for narrow terminals, showing configurable fields below each task description
- **Highly configurable** - Customize keybindings, colors, columns, tabs/sections, sorting, and custom commands
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
- `o` - Open link from annotation (when sidebar is visible)
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

# Logging level (debug, info, warn, error)
# Can be overridden by WUI_LOG_LEVEL env variable or --log-level flag
# Priority: CLI flag > env variable > config file
log_level: error

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

  # Short view - fields shown below each task description when terminal width < 80
  # (or when force_small_screen is true). Max 3 fields. Defaults to due and tags.
  # Supports the same field names as columns (id, project, priority, due, tags, etc.)
  narrow_view_fields:
    - name: due
      label: "DUE"
    - name: tags
      label: "TAGS"

  # Force narrow/short view regardless of terminal width
  # force_small_screen: false

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
  #
  # SORTING:
  # Each tab can have custom sorting using the 'sort' and 'reverse' fields:
  # - sort: "alphabetic" (or "alpha", "description") - Sort by description (case-insensitive)
  # - sort: "due" - Sort by due date (tasks without due date appear last)
  # - sort: "scheduled" - Sort by scheduled date (tasks without date appear last)
  # - sort: "created" (or "entry") - Sort by creation date
  # - sort: "modified" - Sort by modification date (tasks without date appear last)
  # - reverse: true - Reverse the sort order (newest/latest first)
  #
  # Note: Completed tasks always appear last, regardless of sorting
  tabs:
    - name: "Next"
      filter: "( status:pending or status:active ) -WAITING"
    - name: "Today"
      filter: "due:today"
      sort: "due"  # Sort by due date, earliest first
    - name: "Urgent"
      filter: "+urgent"
    - name: "Work"
      filter: "+work -someday"
      sort: "alphabetic"  # Sort alphabetically by description
    - name: "Projects"
      filter: "status:pending or status:active"
    - name: "Tags"
      filter: "status:pending or status:active"
    - name: "All"
      filter: "status:pending or status:active"
      sort: "modified"  # Sort by last modified
      reverse: true     # Most recently modified first

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

  # Custom commands - execute system commands with task data
  # Use {{.fieldname}} to insert any task field into the command
  # Supports standard fields (id, description, project, priority, tags, due, etc.)
  # and custom UDAs (url, github, contact, etc.)
  custom_commands:
    O:  # Press 'O' (uppercase) to trigger this command
      name: "Open URL"
      command: "xdg-open {{.url}}"
      description: "Opens the task's URL in default browser"

    # Platform-specific examples:
    # Linux/macOS: "xdg-open {{.url}}" or "open {{.url}}"
    # Termux: "termux-open-url {{.url}}"
    # Windows: "cmd /c start {{.url}}"
    #
    # More examples:
    # "1":
    #   name: "Copy Description"
    #   command: "echo {{.description}} | xclip -selection clipboard"
    #   description: "Copy task description to clipboard"
    #
    # c:
    #   name: "Git Clone"
    #   command: "git clone {{.url}} ~/projects/{{.project}}"
    #   description: "Clone repository to projects folder"

  # Theme customization
  theme:
    # Theme name determines which base color palette to start from:
    # - "dark" (or omitted): Uses dark theme defaults (cyan/white on dark background)
    # - "light": Uses light theme defaults (dark blue/black on light background)
    # - Any other name: Uses "dark" as base (e.g., "custom", "myTheme")
    #
    # The name ONLY selects the starting palette. To customize colors,
    # specify color fields below - they override the base theme's defaults.
    name: dark

    # All color fields are optional - only specify what you want to customize
    # Colors use ANSI color codes (0-255) or standard names (e.g., "9" for red)
    # If omitted, the predefined theme's defaults are used

    # Priority colors (for tasks marked H, M, L)
    # priority_high: "9"      # High priority tasks (default: red)
    # priority_medium: "11"   # Medium priority tasks (default: yellow)
    # priority_low: "12"      # Low priority tasks (default: blue)

    # Due date colors (based on urgency)
    # due_overdue: "9"        # Overdue tasks (default: red)
    # due_today: "11"         # Tasks due today (default: yellow)
    # due_soon: "11"          # Tasks due soon (default: yellow)

    # Status colors
    # status_active: "15"     # Active/pending tasks (default: white/black)
    # status_waiting: "8"     # Waiting tasks (default: gray)
    # status_completed: "8"   # Completed tasks (default: gray, with strikethrough)

    # UI element colors
    # header_fg: "12"         # Header text (default: cyan)
    # footer_fg: "246"        # Footer text (default: light/dark gray)
    # separator_fg: "8"       # Separators between columns (default: gray)
    # selection_bg: "12"      # Background of selected task (default: cyan)
    # selection_fg: "0"       # Foreground of selected task (default: black)
    # sidebar_border: "8"     # Sidebar border (default: gray)
    # sidebar_title: "12"     # Sidebar title (default: cyan)
    # label_fg: "12"          # Field labels in sidebar (default: cyan)
    # value_fg: "15"          # Field values in sidebar (default: white/black)
    # dim_fg: "8"             # Dimmed text (default: gray)
    # error_fg: "9"           # Error messages (default: red)
    # success_fg: "10"        # Success messages (default: green)
    # tag_fg: "14"            # Task tags (default: cyan)

    # Section/tab colors
    # section_active_fg: "15"   # Active tab foreground (default: white)
    # section_active_bg: "63"   # Active tab background (default: purple/blue)
    # section_inactive_fg: "246" # Inactive tab foreground (default: gray)
```

### Theme Customization Examples

**Use the dark theme (default):**
```yaml
theme:
  name: dark
```

**Use the light theme:**
```yaml
theme:
  name: light
```

**Customize specific colors on top of dark theme:**
```yaml
theme:
  name: dark              # Start with dark theme as base
  priority_high: "196"    # Override: use brighter red
  selection_bg: "33"      # Override: use different blue
  header_fg: "10"         # Override: use green headers
```

**Create a "custom" theme (still uses dark as base):**
```yaml
theme:
  name: mycustom          # Any name other than "dark"/"light" uses dark as base
  priority_high: "196"    # You must specify color overrides yourself
  selection_bg: "33"      # The name alone doesn't provide different colors
```

**Important:** The `name` field only chooses between two built-in palettes ("dark" or "light"). Any other name uses "dark" as the base. To actually customize colors, you must specify the individual color fields.

**Color values:**
- ANSI color codes: Numbers 0-255 (e.g., "9" for red, "12" for cyan)
- Standard colors: 0-15 work across all terminals
- Extended colors: 16-255 require 256-color terminal support
- See [ANSI color codes](https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit) for reference

## Custom Commands

Custom commands allow you to execute system commands with task data. Use the `{{.fieldname}}` template syntax to insert any task field into your command.

### Quick Example

```yaml
tui:
  custom_commands:
    O:  # Press 'O' (uppercase) to open URL
      name: "Open URL"
      command: "xdg-open {{.url}}"
      description: "Opens the task's URL in browser"
```

### Supported Fields

All task fields are available via `{{.fieldname}}`:
- **Standard fields**: `{{.id}}`, `{{.description}}`, `{{.project}}`, `{{.priority}}`, `{{.tags}}`, `{{.due}}`, `{{.uuid}}`, etc.
- **Custom UDAs**: `{{.url}}`, `{{.github}}`, `{{.contact}}`, or any field you've defined in Taskwarrior

### Platform-Specific Examples

**Linux:**
```yaml
custom_commands:
  O:
    name: "Open URL"
    command: "xdg-open {{.url}}"
```

**Termux (Android):**
```yaml
custom_commands:
  O:
    name: "Open URL"
    command: "termux-open-url {{.url}}"
```

**macOS:**
```yaml
custom_commands:
  O:
    name: "Open URL"
    command: "open {{.url}}"
```

**Windows:**
```yaml
custom_commands:
  O:
    name: "Open URL"
    command: "cmd /c start {{.url}}"
```

### Advanced Examples

**Multiple fields:**
```yaml
custom_commands:
  c:
    name: "Git Clone"
    command: "git clone {{.url}} ~/projects/{{.project}}"
    description: "Clone repository to project folder"
```

**Copy to clipboard:**
```yaml
custom_commands:
  "1":
    name: "Copy Description"
    command: "echo {{.description}} | xclip -selection clipboard"
    description: "Copy task description to clipboard"
```

Custom commands appear automatically in the help screen (`?`) under "Custom Commands" section.

**Note:** If a custom command uses a shortcut key that conflicts with a built-in internal shortcut (like `o` for opening annotation links, `s` for start/stop, etc.), wui will display a warning when exiting. The custom command will override the internal shortcut, but you'll be notified about this conflict. To silence these warnings, add `silence_shortcut_override_warnings: true` to your TUI config.

For complete documentation and more examples, see [`docs/custom-commands.md`](docs/custom-commands.md).

## CLI Flags

```
  --config string      Config file path (default: ~/.config/wui/config.yaml)
  --taskrc string      Taskrc file path (default: ~/.taskrc)
  --task-bin string    Task binary path (default: /usr/local/bin/task)
  --search string      Open in Search tab with the specified filter
  --log-level string   Log level: debug, info, warn, error (default: error)
  --log-format string  Log format: text, json (default: text)
```

### Logging Configuration

Logging level can be configured in three ways with the following priority:
1. **CLI flag** (highest): `--log-level debug`
2. **Environment variable**: `export WUI_LOG_LEVEL=info`
3. **Config file** (lowest): `log_level: debug` in `config.yaml`

Logs are written to `/tmp/wui.log` by default. You can change the log file location with the `WUI_LOG_FILE` environment variable.

Examples:
```bash
# Open with custom taskrc and debug logging
wui --taskrc ~/work/.taskrc --log-level debug

# Use environment variable for logging
export WUI_LOG_LEVEL=info
wui

# Open in Search tab with a pre-applied filter (useful in pipelines)
wui --search "project:Home +urgent"

# Search for tasks due today
wui --search "due:today"

# Search for completed tasks
wui --search "status:completed"
```

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

## Tab Sorting

Each tab can have custom sorting to display tasks in a specific order. Configure sorting using the `sort` and `reverse` fields in your tab configuration:

### Sort Methods

- **`alphabetic`** (aliases: `alpha`, `description`) - Sort tasks alphabetically by description (case-insensitive)
- **`due`** - Sort by due date (tasks without due date appear last)
- **`scheduled`** - Sort by scheduled date (tasks without scheduled date appear last)
- **`created`** (alias: `entry`) - Sort by creation date
- **`modified`** - Sort by modification date (tasks without modified date appear last)

### Reverse Order

Add `reverse: true` to invert the sort order (e.g., show newest/latest first instead of oldest/earliest).

### Examples

```yaml
tabs:
  # Sort by due date, earliest deadlines first
  - name: "Deadlines"
    filter: "status:pending +urgent"
    sort: "due"

  # Sort alphabetically
  - name: "All Tasks"
    filter: "status:pending"
    sort: "alphabetic"

  # Sort by creation date, newest first
  - name: "Recent"
    filter: "status:pending"
    sort: "created"
    reverse: true

  # Sort by last modified, most recent first
  - name: "Updated"
    filter: "status:pending"
    sort: "modified"
    reverse: true

  # No sorting (default behavior - uses Taskwarrior's order)
  - name: "Next"
    filter: "status:pending -WAITING"
```

**Note**: Completed tasks always appear after non-completed tasks, regardless of the sort method. Sorting is applied within each status group.

## Short View (Narrow Layout)

When the terminal width is below 80 columns, wui automatically switches to a compact **short view** where each task is displayed across multiple lines:

```
▶ 42  Fix login page crash
      DUE:  2026-02-20
      TAGS: +bug, +frontend
```

You can also force this layout at any terminal width with `force_small_screen: true` in your config.

### Configuring Short View Fields

Up to **3 fields** can be shown below the description. The default fields are **due date** and **tags**. Configure them with `narrow_view_fields` in your config:

```yaml
tui:
  narrow_view_fields:
    - name: due
      label: "DUE"
    - name: tags
      label: "TAGS"
    - name: project
      label: "PROJECT"
```

Any field supported by Taskwarrior works: `due`, `tags`, `project`, `priority`, `urgency`, `scheduled`, or any UDA.

To always use the short view:
```yaml
tui:
  force_small_screen: true
```

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
