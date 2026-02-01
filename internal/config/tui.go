package config

import "gopkg.in/yaml.v3"

// TUIConfig contains TUI-specific configuration
type TUIConfig struct {
	SidebarWidth                    int                      `yaml:"sidebar_width"`
	ScrollBuffer                    int                      `yaml:"scroll_buffer"`
	InputMode                       string                   `yaml:"input_mode"` // "floating" or "bottom" - controls how input prompts are displayed
	SilenceShortcutOverrideWarnings bool                     `yaml:"silence_shortcut_override_warnings,omitempty"`
	ValidateTodosOnComplete         *bool                    `yaml:"validate_todos_on_complete,omitempty"` // Prevent completing tasks with TODO: annotations (default: true)
	Tabs                            []Tab                    `yaml:"tabs"`
	Columns                         Columns                  `yaml:"columns"`
	NarrowViewFields                Columns                  `yaml:"narrow_view_fields"` // Fields to display below description in narrow view (terminal width < 80)
	Keybindings                     map[string]string        `yaml:"keybindings"`
	Theme                           *Theme                   `yaml:"theme"`
	CustomCommands                  map[string]CustomCommand `yaml:"custom_commands,omitempty"`
}

// CustomCommand represents a user-defined command that can be executed with task data
type CustomCommand struct {
	Name        string `yaml:"name"`        // Display name for the command
	Command     string `yaml:"command"`     // Command template with {{.field}} placeholders
	Description string `yaml:"description"` // Optional description for help text
}

// Tab represents a section/tab in the UI
type Tab struct {
	Name    string `yaml:"name"`
	Filter  string `yaml:"filter"`
	Sort    string `yaml:"sort,omitempty"`    // Sorting method: "alphabetic", "due", "scheduled", "created", "modified" (default: none)
	Reverse bool   `yaml:"reverse,omitempty"` // Reverse sort order
}

// Column represents a table column configuration
type Column struct {
	Name   string `yaml:"name"`             // Taskwarrior property name (e.g., "id", "project", "priority")
	Label  string `yaml:"label"`            // Display label for the column header
	Length int    `yaml:"length,omitempty"` // Maximum width in characters (0 = use default/dynamic)
}

// Columns is a wrapper type that supports both old ([]string) and new ([]Column) formats
type Columns []Column

// GetDefaultLabels returns default labels for all known taskwarrior properties
func GetDefaultLabels() map[string]string {
	return map[string]string{
		// Core fields
		"id":          "ID",
		"uuid":        "UUID",
		"description": "DESCRIPTION",
		"project":     "PROJECT",
		"priority":    "P",
		"status":      "STATUS",
		"tags":        "TAGS",
		// Date fields
		"due":       "DUE",
		"scheduled": "SCHEDULED",
		"wait":      "WAIT",
		"start":     "START",
		"entry":     "ENTRY",
		"modified":  "MODIFIED",
		"end":       "END",
		// Other fields
		"urgency":    "URG",
		"annotation": "A",
		"dependency": "D",
	}
}

// UnmarshalYAML implements custom unmarshaling to support backward compatibility
// Accepts both:
//   - Old format: ["id", "project", "priority"]
//   - New format: [{name: "id", label: "#"}, {name: "project", label: "Project"}]
func (c *Columns) UnmarshalYAML(node *yaml.Node) error {
	// Try to unmarshal as []Column (new format)
	var columns []Column
	if err := node.Decode(&columns); err == nil {
		*c = columns
		return nil
	}

	// Fall back to old format: []string
	var oldFormat []string
	if err := node.Decode(&oldFormat); err != nil {
		return err
	}

	// Convert old format to new format with default labels
	defaultLabels := GetDefaultLabels()

	columns = make([]Column, len(oldFormat))
	for i, name := range oldFormat {
		label := defaultLabels[name]
		if label == "" {
			// For unknown columns, use uppercase of the name
			label = name
		}
		columns[i] = Column{Name: name, Label: label}
	}

	*c = columns
	return nil
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
	HeaderFg      string `yaml:"header_fg"`
	FooterFg      string `yaml:"footer_fg"`
	SeparatorFg   string `yaml:"separator_fg"`
	SelectionBg   string `yaml:"selection_bg"`
	SelectionFg   string `yaml:"selection_fg"`
	SidebarBorder string `yaml:"sidebar_border"`
	SidebarTitle  string `yaml:"sidebar_title"`
	LabelFg       string `yaml:"label_fg"`
	ValueFg       string `yaml:"value_fg"`
	DimFg         string `yaml:"dim_fg"`
	ErrorFg       string `yaml:"error_fg"`
	SuccessFg     string `yaml:"success_fg"`
	TagFg         string `yaml:"tag_fg"`

	// Section colors
	SectionActiveFg   string `yaml:"section_active_fg"`
	SectionActiveBg   string `yaml:"section_active_bg"`
	SectionInactiveFg string `yaml:"section_inactive_fg"`
}
