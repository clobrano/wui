package config

import (
	"os"
	"path/filepath"

	"k8s.io/utils/ptr"
)

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	taskrcPath := filepath.Join(homeDir, ".taskrc")

	return &Config{
		TaskBin:      "task", // Assumes task is in PATH
		TaskrcPath:   taskrcPath,
		LogLevel:     "error",
		TUI:          DefaultTUIConfig(),
		CalendarSync: DefaultCalendarSync(),
	}
}

// DefaultCalendarSync returns default Google Calendar sync configuration
func DefaultCalendarSync() *CalendarSync {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "wui")

	return &CalendarSync{
		Enabled:         false,
		CalendarName:    "",
		TaskFilter:      "status:pending or status:completed",
		CredentialsPath: filepath.Join(configDir, "credentials.json"),
		TokenPath:       filepath.Join(configDir, "token.json"),
		AutoSyncOnQuit:  false,
	}
}

// DefaultTUIConfig returns TUI configuration with defaults
func DefaultTUIConfig() *TUIConfig {
	return &TUIConfig{
		SidebarWidth:              33,           // Percentage of terminal width (33%)
		ScrollBuffer:              1,            // Number of tasks to keep visible above/below cursor
		InputMode:                 "floating",   // Default to floating window for input prompts
		ValidateTodosOnComplete:   ptr.To(true), // Prevent completing tasks with TODO: annotations
		ValidateBlockedOnComplete: ptr.To(true), // Prevent completing tasks blocked by other tasks
		Tabs:                      DefaultTabs(),
		Columns:                   DefaultColumns(),
		NarrowViewFields:          DefaultNarrowViewFields(),
		Keybindings:               DefaultKeybindings(),
		Theme:                     DefaultTheme(),
	}
}

// DefaultTabs returns the default tab list
//
// Note: The Search tab is always automatically prepended and should not be included here
//
// Special tab names:
//   - "Search" - Reserved, always auto-prepended as first tab (cannot be configured)
//   - "Projects" - Triggers grouped view by project (renaming breaks grouping behavior)
//   - "Tags" - Triggers grouped view by tag (renaming breaks grouping behavior)
func DefaultTabs() []Tab {
	return []Tab{
		{
			Name:   "Next",
			Filter: "( status:pending or status:active ) -WAITING",
			Sort:   "urgency",
		},
		{
			Name:   "Waiting",
			Filter: "status:waiting",
			Sort:   "urgency",
		},
		{
			Name:   "Projects",
			Filter: "status:pending or status:active",
			Sort:   "urgency",
		},
		{
			Name:   "Tags",
			Filter: "status:pending or status:active",
			Sort:   "urgency",
		},
		{
			Name:   "All",
			Filter: "status:pending or status:waiting or status:active",
			Sort:   "urgency",
		},
	}
}

// DefaultColumns returns the default column list
func DefaultColumns() Columns {
	return Columns{
		{Name: "id", Label: "ID"},
		{Name: "project", Label: "PROJECT"},
		{Name: "priority", Label: "P"},
		{Name: "due", Label: "DUE"},
		{Name: "dependency", Label: "D"},
		{Name: "description", Label: "DESCRIPTION"},
	}
}

// DefaultNarrowViewFields returns the default fields to display in narrow view
// These fields are shown below the description when terminal width < 80
// Returns empty by default - users must explicitly configure narrow view fields
func DefaultNarrowViewFields() Columns {
	return Columns{}
}

// DefaultKeybindings returns the default key mappings
func DefaultKeybindings() map[string]string {
	return map[string]string{
		// Navigation
		"quit":           "q",
		"help":           "?",
		"up":             "k",
		"down":           "j",
		"page_up":        "ctrl+u",
		"page_down":      "ctrl+d",
		"first":          "g",
		"last":           "G",
		"toggle_sidebar": "tab",

		// Sections
		"next_section": "L",
		"prev_section": "H",

		// Task operations
		"done":     "d",
		"delete":   "x",
		"edit":     "e",
		"modify":   "m",
		"annotate": "a",
		"new":      "n",
		"undo":     "u",
		"open_url": "o",

		// Filtering
		"filter":  "/",
		"refresh": "r",
	}
}

// GetInternalShortcuts returns a map of all internal shortcuts and their descriptions.
// This includes both configurable keybindings and hardcoded shortcuts.
// Used to detect when custom commands override internal functionality.
func GetInternalShortcuts(keybindings map[string]string) map[string]string {
	shortcuts := make(map[string]string)

	// Get configured keybindings (or defaults)
	getKey := func(action, defaultKey string) string {
		if keybindings != nil {
			if key, exists := keybindings[action]; exists {
				return key
			}
		}
		return defaultKey
	}

	// Configurable keybindings
	shortcuts[getKey("quit", "q")] = "quit"
	shortcuts[getKey("help", "?")] = "toggle help"
	shortcuts[getKey("up", "k")] = "move up"
	shortcuts[getKey("down", "j")] = "move down"
	shortcuts[getKey("page_up", "ctrl+u")] = "page up"
	shortcuts[getKey("page_down", "ctrl+d")] = "page down"
	shortcuts[getKey("first", "g")] = "jump to first"
	shortcuts[getKey("last", "G")] = "jump to last"
	shortcuts[getKey("toggle_sidebar", "tab")] = "toggle sidebar"
	shortcuts[getKey("next_section", "L")] = "next section"
	shortcuts[getKey("prev_section", "H")] = "previous section"
	shortcuts[getKey("done", "d")] = "mark done"
	shortcuts[getKey("delete", "x")] = "delete"
	shortcuts[getKey("edit", "e")] = "edit"
	shortcuts[getKey("modify", "m")] = "modify"
	shortcuts[getKey("annotate", "a")] = "annotate"
	shortcuts[getKey("new", "n")] = "new task"
	shortcuts[getKey("undo", "u")] = "undo"
	shortcuts[getKey("open_url", "o")] = "open URL/file from annotation"
	shortcuts[getKey("filter", "/")] = "filter"
	shortcuts[getKey("refresh", "r")] = "refresh"

	// Hardcoded shortcuts (not configurable)
	shortcuts["s"] = "start/stop task"
	shortcuts["M"] = "export markdown"

	return shortcuts
}

// DefaultTheme returns the default color theme (dark theme)
func DefaultTheme() *Theme {
	return &Theme{
		Name: "dark",

		// Priority colors
		PriorityHigh:   "9",  // Red
		PriorityMedium: "11", // Yellow
		PriorityLow:    "12", // Blue

		// Due date colors
		DueOverdue: "9",  // Red
		DueToday:   "11", // Yellow
		DueSoon:    "11", // Yellow

		// Status colors
		StatusActive:    "15", // White
		StatusWaiting:   "8",  // Dim gray
		StatusCompleted: "8",  // Dim gray

		// UI element colors
		HeaderFg:      "12",  // Bright cyan
		FooterFg:      "8",   // Dim gray
		SeparatorFg:   "8",   // Dim gray
		SelectionBg:   "250", // Light gray
		SelectionFg:   "16",  // Dark gray
		SidebarBorder: "8",   // Dim gray
		SidebarTitle:  "12",  // Bright cyan
		LabelFg:       "12",  // Bright cyan
		ValueFg:       "15",  // White
		DimFg:         "8",   // Dim gray
		ErrorFg:       "9",   // Red
		SuccessFg:     "10",  // Green
		TagFg:         "14",  // Cyan

		// Section colors
		SectionActiveFg:   "15",  // White
		SectionActiveBg:   "63",  // Purple/blue
		SectionInactiveFg: "246", // Dim
	}
}
