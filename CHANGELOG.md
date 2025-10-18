# Changelog

All notable changes to wui will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of wui (Warrior UI)
- Modern TUI built with bubbletea framework
- Multiple view sections (Next, Waiting, Projects, Tags, All)
- Detailed task sidebar with full metadata display
- Task operations: done, start/stop, delete, edit, modify, annotate, undo
- Quick task navigation with j/k and number keys (1-9)
- Section navigation with Tab and h/l keys
- Filter support with Taskwarrior syntax
- Grouped views by project and tag with drill-down
- Configuration file support (~/.config/wui/config.yaml)
- Filter bookmarks for custom sections
- Configurable keybindings and themes (dark/light)
- UDA (User Defined Attributes) support from .taskrc
- Help screen (press ?) with comprehensive keybinding reference
- Status bar with error handling and loading indicators
- CLI flags for custom config, taskrc, and task binary paths
- Logging support (debug, info, warn, error levels)

### Features

#### Task Management
- Mark tasks done (`d`)
- Start/stop tasks (`s`)
- Delete tasks with confirmation (`x`)
- Edit tasks in $EDITOR (`e`)
- Quick modify with Taskwarrior syntax (`m`)
- Add annotations (`a`)
- Create new tasks (`n`)
- Undo operations (`u`)

#### Navigation
- Vim-style navigation (j/k, g/G, h/l)
- Arrow key support
- Quick jump to visible tasks (1-9)
- Quick jump to sections (1-5)
- Sidebar scrolling (Ctrl+d/u/f/b)

#### Views
- Next: Ready-to-work tasks
- Waiting: Blocked or scheduled tasks
- Projects: Grouped by project with task counts
- Tags: Grouped by tag with task counts
- Custom: Via filter bookmarks

#### Display
- Configurable columns (ID, PROJECT, P, DUE, TAGS, DESCRIPTION)
- Color-coded priorities (High/Medium/Low)
- Color-coded due dates (overdue/today/soon)
- Task metadata: UUID, description, project, status, priority, dates, dependencies, annotations, UDAs
- Responsive layout adapting to terminal size

#### Configuration
- YAML configuration file
- Respects .taskrc settings and UDAs
- Customizable sidebar width
- Filter bookmarks
- Theme customization

### Technical
- Built with Go 1.21+
- Uses bubbletea for TUI framework
- Uses lipgloss for styling
- Uses bubbles for UI components
- Clean layered architecture (Core/Adapter/Presentation)
- Comprehensive test coverage
- Error handling with user-friendly messages
- Loading indicators for async operations

## Version History

### [0.1.0] - TBD
- Initial MVP release
- All features listed above

---

## Release Notes Template

### [X.Y.Z] - YYYY-MM-DD

#### Added
- New features

#### Changed
- Changes in existing functionality

#### Deprecated
- Soon-to-be removed features

#### Removed
- Now removed features

#### Fixed
- Bug fixes

#### Security
- Vulnerability fixes
