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
		SidebarWidth: 40,
		Bookmarks:    []Bookmark{},
		Columns:      DefaultColumns(),
		Keybindings:  DefaultKeybindings(),
		Theme:        DefaultTheme(),
	}
}

// DefaultColumns returns the default column list
func DefaultColumns() []string {
	return []string{
		"id",
		"project",
		"description",
		"due",
		"priority",
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

// DefaultTheme returns the default color theme
func DefaultTheme() *Theme {
	return &Theme{
		// Priority colors
		PriorityHigh:   "red",
		PriorityMedium: "yellow",
		PriorityLow:    "blue",

		// Due date colors
		Overdue:  "red",
		DueToday: "orange",
		DueSoon:  "yellow",

		// UI elements
		Selected: "reverse",
		Border:   "gray",
	}
}
