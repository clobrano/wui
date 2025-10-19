package config

import (
	"os"
	"path/filepath"
)

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	taskrcPath := filepath.Join(homeDir, ".taskrc")

	return &Config{
		TaskBin:    "task", // Assumes task is in PATH
		TaskrcPath: taskrcPath,
		TUI:        DefaultTUIConfig(),
	}
}

// DefaultTUIConfig returns TUI configuration with defaults
func DefaultTUIConfig() *TUIConfig {
	return &TUIConfig{
		SidebarWidth: 33, // Percentage of terminal width (33%)
		Tabs:         DefaultTabs(),
		Columns:      DefaultColumns(),
		Keybindings:  DefaultKeybindings(),
		Theme:        DefaultTheme(),
	}
}

// DefaultTabs returns the default tab list
func DefaultTabs() []Tab {
	return []Tab{
		{
			Name:   "Next",
			Filter: "( status:pending or status:active ) -WAITING",
		},
		{
			Name:   "Waiting",
			Filter: "status:waiting",
		},
		{
			Name:   "Projects",
			Filter: "status:pending or status:active",
		},
		{
			Name:   "Tags",
			Filter: "status:pending or status:active",
		},
		{
			Name:   "All",
			Filter: "status:pending or status:waiting or status:active",
		},
	}
}

// DefaultColumns returns the default column list
func DefaultColumns() []string {
	return []string{
		"id",
		"project",
		"priority",
		"due",
		"description",
	}
}

// DefaultKeybindings returns the default key mappings
func DefaultKeybindings() map[string]string {
	return map[string]string{
		// Navigation
		"quit":          "q",
		"help":          "?",
		"up":            "k",
		"down":          "j",
		"page_up":       "ctrl+u",
		"page_down":     "ctrl+d",
		"first":         "g",
		"last":          "G",
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

		// Filtering
		"filter":  "/",
		"refresh": "r",
	}
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
		HeaderFg:       "12", // Bright cyan
		FooterFg:       "8",  // Dim gray
		SeparatorFg:    "8",  // Dim gray
		SelectionBg:    "12", // Cyan
		SelectionFg:    "0",  // Black
		SidebarBorder:  "8",  // Dim gray
		SidebarTitle:   "12", // Bright cyan
		LabelFg:        "12", // Bright cyan
		ValueFg:        "15", // White
		DimFg:          "8",  // Dim gray
		ErrorFg:        "9",  // Red
		SuccessFg:      "10", // Green
		TagFg:          "14", // Cyan

		// Section colors
		SectionActiveFg:   "15",  // White
		SectionActiveBg:   "63",  // Purple/blue
		SectionInactiveFg: "246", // Dim
	}
}
