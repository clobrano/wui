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
	Name string `yaml:"name"` // "dark", "light", or "custom"

	// Priority colors
	PriorityHigh   string `yaml:"priority_high"`
	PriorityMedium string `yaml:"priority_medium"`
	PriorityLow    string `yaml:"priority_low"`

	// Due date colors
	DueOverdue string `yaml:"due_overdue"`
	DueToday   string `yaml:"due_today"`
	DueSoon    string `yaml:"due_soon"`

	// Status colors
	StatusActive    string `yaml:"status_active"`
	StatusWaiting   string `yaml:"status_waiting"`
	StatusCompleted string `yaml:"status_completed"`

	// UI element colors
	HeaderFg       string `yaml:"header_fg"`
	FooterFg       string `yaml:"footer_fg"`
	SeparatorFg    string `yaml:"separator_fg"`
	SelectionBg    string `yaml:"selection_bg"`
	SelectionFg    string `yaml:"selection_fg"`
	SidebarBorder  string `yaml:"sidebar_border"`
	SidebarTitle   string `yaml:"sidebar_title"`
	LabelFg        string `yaml:"label_fg"`
	ValueFg        string `yaml:"value_fg"`
	DimFg          string `yaml:"dim_fg"`
	ErrorFg        string `yaml:"error_fg"`
	SuccessFg      string `yaml:"success_fg"`
	TagFg          string `yaml:"tag_fg"`

	// Section colors
	SectionActiveFg   string `yaml:"section_active_fg"`
	SectionActiveBg   string `yaml:"section_active_bg"`
	SectionInactiveFg string `yaml:"section_inactive_fg"`
}
