package config

// TUIConfig contains TUI-specific configuration
type TUIConfig struct {
	SidebarWidth int                `yaml:"sidebar_width"`
	Bookmarks    []Bookmark         `yaml:"bookmarks"`
	Columns      []string           `yaml:"columns"`
	Keybindings  map[string]string  `yaml:"keybindings"`
	Theme        *Theme             `yaml:"theme"`
}

// Bookmark represents a saved filter
type Bookmark struct {
	Name   string `yaml:"name"`
	Filter string `yaml:"filter"`
}

// Theme contains color and style configuration
type Theme struct {
	// Priority colors
	PriorityHigh   string `yaml:"priority_high"`
	PriorityMedium string `yaml:"priority_medium"`
	PriorityLow    string `yaml:"priority_low"`

	// Due date colors
	Overdue  string `yaml:"overdue"`
	DueToday string `yaml:"due_today"`
	DueSoon  string `yaml:"due_soon"`

	// UI elements
	Selected string `yaml:"selected"`
	Border   string `yaml:"border"`
}
