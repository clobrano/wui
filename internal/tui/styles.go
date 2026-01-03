package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/tui/components"
	"github.com/muesli/termenv"
)

// ColorScheme defines the color palette for the TUI
type ColorScheme struct {
	// Priority colors
	PriorityHigh   lipgloss.Color
	PriorityMedium lipgloss.Color
	PriorityLow    lipgloss.Color

	// Due date colors
	DueOverdue lipgloss.Color
	DueToday   lipgloss.Color
	DueSoon    lipgloss.Color

	// Status colors
	StatusActive    lipgloss.Color
	StatusWaiting   lipgloss.Color
	StatusCompleted lipgloss.Color

	// UI element colors
	HeaderFg      lipgloss.Color
	FooterFg      lipgloss.Color
	SeparatorFg   lipgloss.Color
	SelectionBg   lipgloss.Color
	SelectionFg   lipgloss.Color
	SidebarBorder lipgloss.Color
	SidebarTitle  lipgloss.Color
	LabelFg       lipgloss.Color
	ValueFg       lipgloss.Color
	DimFg         lipgloss.Color
	ErrorFg       lipgloss.Color
	SuccessFg     lipgloss.Color
	TagFg         lipgloss.Color

	// Section colors
	SectionActiveFg   lipgloss.Color
	SectionActiveBg   lipgloss.Color
	SectionInactiveFg lipgloss.Color
}

// Theme represents a complete visual theme
type Theme struct {
	Name   string
	Colors ColorScheme
}

// DefaultDarkTheme returns the default dark color scheme
func DefaultDarkTheme() Theme {
	return Theme{
		Name: "dark",
		Colors: ColorScheme{
			// Priority colors (standard terminal colors)
			PriorityHigh:   lipgloss.Color("9"),  // Red
			PriorityMedium: lipgloss.Color("11"), // Yellow
			PriorityLow:    lipgloss.Color("12"), // Blue

			// Due date colors
			DueOverdue: lipgloss.Color("9"),  // Red
			DueToday:   lipgloss.Color("11"), // Yellow/Orange
			DueSoon:    lipgloss.Color("11"), // Yellow

			// Status colors
			StatusActive:    lipgloss.Color("15"), // White/default
			StatusWaiting:   lipgloss.Color("8"),  // Dim gray
			StatusCompleted: lipgloss.Color("8"),  // Dim gray

			// UI element colors
			HeaderFg:      lipgloss.Color("12"),  // Bright cyan
			FooterFg:      lipgloss.Color("246"), // Light gray (more visible)
			SeparatorFg:   lipgloss.Color("8"),   // Dim gray
			SelectionBg:   lipgloss.Color("12"),  // Cyan background
			SelectionFg:   lipgloss.Color("0"),   // Black foreground
			SidebarBorder: lipgloss.Color("8"),   // Dim gray
			SidebarTitle:  lipgloss.Color("12"),  // Bright cyan
			LabelFg:       lipgloss.Color("12"),  // Bright cyan
			ValueFg:       lipgloss.Color("15"),  // White
			DimFg:         lipgloss.Color("8"),   // Dim gray
			ErrorFg:       lipgloss.Color("9"),   // Red
			SuccessFg:     lipgloss.Color("10"),  // Green
			TagFg:         lipgloss.Color("14"),  // Cyan

			// Section colors
			SectionActiveFg:   lipgloss.Color("15"),  // White
			SectionActiveBg:   lipgloss.Color("63"),  // Purple/blue
			SectionInactiveFg: lipgloss.Color("246"), // Dim
		},
	}
}

// DefaultLightTheme returns a light color scheme
func DefaultLightTheme() Theme {
	return Theme{
		Name: "light",
		Colors: ColorScheme{
			// Priority colors
			PriorityHigh:   lipgloss.Color("1"), // Dark red
			PriorityMedium: lipgloss.Color("3"), // Dark yellow
			PriorityLow:    lipgloss.Color("4"), // Dark blue

			// Due date colors
			DueOverdue: lipgloss.Color("1"), // Dark red
			DueToday:   lipgloss.Color("3"), // Dark yellow
			DueSoon:    lipgloss.Color("3"), // Dark yellow

			// Status colors
			StatusActive:    lipgloss.Color("0"), // Black/default
			StatusWaiting:   lipgloss.Color("8"), // Gray
			StatusCompleted: lipgloss.Color("8"), // Gray

			// UI element colors
			HeaderFg:      lipgloss.Color("4"),   // Dark blue
			FooterFg:      lipgloss.Color("240"), // Darker gray (more visible on light background)
			SeparatorFg:   lipgloss.Color("8"),   // Gray
			SelectionBg:   lipgloss.Color("12"),  // Light cyan
			SelectionFg:   lipgloss.Color("0"),   // Black
			SidebarBorder: lipgloss.Color("8"),   // Gray
			SidebarTitle:  lipgloss.Color("4"),   // Dark blue
			LabelFg:       lipgloss.Color("4"),   // Dark blue
			ValueFg:       lipgloss.Color("0"),   // Black
			DimFg:         lipgloss.Color("8"),   // Gray
			ErrorFg:       lipgloss.Color("1"),   // Dark red
			SuccessFg:     lipgloss.Color("2"),   // Dark green
			TagFg:         lipgloss.Color("6"),   // Dark cyan

			// Section colors
			SectionActiveFg:   lipgloss.Color("15"),  // White
			SectionActiveBg:   lipgloss.Color("4"),   // Dark blue
			SectionInactiveFg: lipgloss.Color("240"), // Dim
		},
	}
}

// Styles holds all the lipgloss styles for the TUI
type Styles struct {
	theme Theme

	// Pre-computed styles
	Header           lipgloss.Style
	Footer           lipgloss.Style
	Separator        lipgloss.Style
	Selection        lipgloss.Style
	Normal           lipgloss.Style
	SidebarBorder    lipgloss.Style
	SidebarTitle     lipgloss.Style
	Label            lipgloss.Style
	Value            lipgloss.Style
	Dim              lipgloss.Style
	Error            lipgloss.Style
	Success          lipgloss.Style
	Tag              lipgloss.Style
	LoadingIndicator lipgloss.Style
	SectionActive    lipgloss.Style
	SectionInactive  lipgloss.Style
	SectionCount     lipgloss.Style
	TasklistHeader   lipgloss.Style
	GroupHeader      lipgloss.Style
	InputPrompt      lipgloss.Style
	InputHint        lipgloss.Style
}

// NewStyles creates a new Styles instance with the given theme
func NewStyles(theme Theme) *Styles {
	s := &Styles{
		theme: theme,
	}

	// Build all styles from theme
	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Colors.HeaderFg).
		Padding(0, 1)

	s.Footer = lipgloss.NewStyle().
		Foreground(theme.Colors.FooterFg).
		Padding(1, 1)

	s.Separator = lipgloss.NewStyle().
		Foreground(theme.Colors.SeparatorFg)

	s.Selection = lipgloss.NewStyle().
		Background(theme.Colors.SelectionBg).
		Foreground(theme.Colors.SelectionFg)

	s.Normal = lipgloss.NewStyle()

	s.SidebarBorder = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Colors.SidebarBorder).
		Padding(1, 2)

	s.SidebarTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Colors.SidebarTitle)

	s.Label = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Colors.LabelFg)

	s.Value = lipgloss.NewStyle().
		Foreground(theme.Colors.ValueFg)

	s.Dim = lipgloss.NewStyle().
		Foreground(theme.Colors.DimFg)

	s.Error = lipgloss.NewStyle().
		Foreground(theme.Colors.ErrorFg).
		Bold(true)

	s.Success = lipgloss.NewStyle().
		Foreground(theme.Colors.SuccessFg)

	s.LoadingIndicator = lipgloss.NewStyle().
		Foreground(theme.Colors.HeaderFg).
		Bold(true)

	s.Tag = lipgloss.NewStyle().
		Foreground(theme.Colors.TagFg)

	s.SectionActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Colors.SectionActiveFg).
		Background(theme.Colors.SectionActiveBg).
		Padding(0, 1)

	s.SectionInactive = lipgloss.NewStyle().
		Foreground(theme.Colors.SectionInactiveFg).
		Padding(0, 1)

	s.SectionCount = lipgloss.NewStyle().
		Foreground(theme.Colors.DimFg).
		Padding(0, 1)

	s.TasklistHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Colors.HeaderFg)

	s.GroupHeader = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Colors.HeaderFg)

	s.InputPrompt = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Colors.HeaderFg)

	s.InputHint = lipgloss.NewStyle().
		Foreground(theme.Colors.DimFg)

	return s
}

// GetTheme returns the current theme
func (s *Styles) GetTheme() Theme {
	return s.theme
}

// Priority returns a style for the given priority level
func (s *Styles) Priority(priority string) lipgloss.Style {
	switch priority {
	case "H":
		return lipgloss.NewStyle().Foreground(s.theme.Colors.PriorityHigh)
	case "M":
		return lipgloss.NewStyle().Foreground(s.theme.Colors.PriorityMedium)
	case "L":
		return lipgloss.NewStyle().Foreground(s.theme.Colors.PriorityLow)
	default:
		return lipgloss.NewStyle()
	}
}

// DueDate returns a style for a due date based on its urgency
func (s *Styles) DueDate(task interface {
	IsOverdue() bool
	IsDueToday() bool
	IsDueSoon() bool
}) lipgloss.Style {
	if task.IsOverdue() {
		return lipgloss.NewStyle().Foreground(s.theme.Colors.DueOverdue)
	}
	if task.IsDueToday() {
		return lipgloss.NewStyle().Foreground(s.theme.Colors.DueToday)
	}
	if task.IsDueSoon() {
		return lipgloss.NewStyle().Foreground(s.theme.Colors.DueSoon)
	}
	return lipgloss.NewStyle()
}

// TaskStatus returns a style for a task based on its status
func (s *Styles) TaskStatus(status string) lipgloss.Style {
	switch status {
	case "waiting":
		return lipgloss.NewStyle().Foreground(s.theme.Colors.StatusWaiting)
	case "completed":
		return lipgloss.NewStyle().
			Foreground(s.theme.Colors.StatusCompleted).
			Strikethrough(true)
	case "active", "pending":
		return lipgloss.NewStyle().Foreground(s.theme.Colors.StatusActive)
	default:
		return lipgloss.NewStyle()
	}
}

// DetectColorProfile returns the terminal's color capability
func DetectColorProfile() termenv.Profile {
	return lipgloss.ColorProfile()
}

// ColorProfileName returns a human-readable name for the color profile
func ColorProfileName(profile termenv.Profile) string {
	switch profile {
	case termenv.TrueColor:
		return "TrueColor (24-bit)"
	case termenv.ANSI256:
		return "256 colors"
	case termenv.ANSI:
		return "16 colors (ANSI)"
	case termenv.Ascii:
		return "No color (ASCII)"
	default:
		return "Unknown"
	}
}

// ThemeFromConfig converts a config.Theme to a tui.Theme
func ThemeFromConfig(cfgTheme *config.Theme) Theme {
	if cfgTheme == nil {
		return DefaultDarkTheme()
	}

	// Start with a base theme based on the name
	var base Theme
	switch cfgTheme.Name {
	case "light":
		base = DefaultLightTheme()
	case "dark", "":
		base = DefaultDarkTheme()
	default:
		// For custom theme names, start with dark as base
		base = DefaultDarkTheme()
		base.Name = cfgTheme.Name
	}

	// Apply color overrides from config
	// Only override if the config value is non-empty
	if cfgTheme.PriorityHigh != "" {
		base.Colors.PriorityHigh = lipgloss.Color(cfgTheme.PriorityHigh)
	}
	if cfgTheme.PriorityMedium != "" {
		base.Colors.PriorityMedium = lipgloss.Color(cfgTheme.PriorityMedium)
	}
	if cfgTheme.PriorityLow != "" {
		base.Colors.PriorityLow = lipgloss.Color(cfgTheme.PriorityLow)
	}
	if cfgTheme.DueOverdue != "" {
		base.Colors.DueOverdue = lipgloss.Color(cfgTheme.DueOverdue)
	}
	if cfgTheme.DueToday != "" {
		base.Colors.DueToday = lipgloss.Color(cfgTheme.DueToday)
	}
	if cfgTheme.DueSoon != "" {
		base.Colors.DueSoon = lipgloss.Color(cfgTheme.DueSoon)
	}
	if cfgTheme.StatusActive != "" {
		base.Colors.StatusActive = lipgloss.Color(cfgTheme.StatusActive)
	}
	if cfgTheme.StatusWaiting != "" {
		base.Colors.StatusWaiting = lipgloss.Color(cfgTheme.StatusWaiting)
	}
	if cfgTheme.StatusCompleted != "" {
		base.Colors.StatusCompleted = lipgloss.Color(cfgTheme.StatusCompleted)
	}
	if cfgTheme.HeaderFg != "" {
		base.Colors.HeaderFg = lipgloss.Color(cfgTheme.HeaderFg)
	}
	if cfgTheme.FooterFg != "" {
		base.Colors.FooterFg = lipgloss.Color(cfgTheme.FooterFg)
	}
	if cfgTheme.SeparatorFg != "" {
		base.Colors.SeparatorFg = lipgloss.Color(cfgTheme.SeparatorFg)
	}
	if cfgTheme.SelectionBg != "" {
		base.Colors.SelectionBg = lipgloss.Color(cfgTheme.SelectionBg)
	}
	if cfgTheme.SelectionFg != "" {
		base.Colors.SelectionFg = lipgloss.Color(cfgTheme.SelectionFg)
	}
	if cfgTheme.SidebarBorder != "" {
		base.Colors.SidebarBorder = lipgloss.Color(cfgTheme.SidebarBorder)
	}
	if cfgTheme.SidebarTitle != "" {
		base.Colors.SidebarTitle = lipgloss.Color(cfgTheme.SidebarTitle)
	}
	if cfgTheme.LabelFg != "" {
		base.Colors.LabelFg = lipgloss.Color(cfgTheme.LabelFg)
	}
	if cfgTheme.ValueFg != "" {
		base.Colors.ValueFg = lipgloss.Color(cfgTheme.ValueFg)
	}
	if cfgTheme.DimFg != "" {
		base.Colors.DimFg = lipgloss.Color(cfgTheme.DimFg)
	}
	if cfgTheme.ErrorFg != "" {
		base.Colors.ErrorFg = lipgloss.Color(cfgTheme.ErrorFg)
	}
	if cfgTheme.SuccessFg != "" {
		base.Colors.SuccessFg = lipgloss.Color(cfgTheme.SuccessFg)
	}
	if cfgTheme.TagFg != "" {
		base.Colors.TagFg = lipgloss.Color(cfgTheme.TagFg)
	}
	if cfgTheme.SectionActiveFg != "" {
		base.Colors.SectionActiveFg = lipgloss.Color(cfgTheme.SectionActiveFg)
	}
	if cfgTheme.SectionActiveBg != "" {
		base.Colors.SectionActiveBg = lipgloss.Color(cfgTheme.SectionActiveBg)
	}
	if cfgTheme.SectionInactiveFg != "" {
		base.Colors.SectionInactiveFg = lipgloss.Color(cfgTheme.SectionInactiveFg)
	}

	return base
}

// ToTaskListStyles converts Styles to component-specific TaskListStyles
func (s *Styles) ToTaskListStyles() components.TaskListStyles {
	return components.TaskListStyles{
		Header:          s.TasklistHeader,
		Separator:       s.Separator,
		Selection:       s.Selection,
		PriorityHigh:    s.theme.Colors.PriorityHigh,
		PriorityMedium:  s.theme.Colors.PriorityMedium,
		PriorityLow:     s.theme.Colors.PriorityLow,
		DueOverdue:      s.theme.Colors.DueOverdue,
		TagColor:        s.theme.Colors.TagFg,
		StatusCompleted: s.theme.Colors.StatusCompleted,
		StatusWaiting:   s.theme.Colors.StatusWaiting,
		StatusActive:    s.theme.Colors.StatusActive,
	}
}

// ToSidebarStyles converts Styles to component-specific SidebarStyles
func (s *Styles) ToSidebarStyles() components.SidebarStyles {
	return components.SidebarStyles{
		Border:         s.SidebarBorder,
		Title:          s.SidebarTitle,
		Label:          s.Label,
		Value:          s.Value,
		Dim:            s.Dim,
		PriorityHigh:   s.theme.Colors.PriorityHigh,
		PriorityMedium: s.theme.Colors.PriorityMedium,
		PriorityLow:    s.theme.Colors.PriorityLow,
		DueOverdue:     s.theme.Colors.DueOverdue,
		StatusPending:  s.theme.Colors.StatusActive,
		StatusActive:   s.theme.Colors.StatusActive,
		StatusDone:     s.theme.Colors.SuccessFg,
		StatusWaiting:  s.theme.Colors.StatusWaiting,
		Tag:            s.theme.Colors.TagFg,
	}
}

// ToSectionsStyles converts Styles to component-specific SectionsStyles
func (s *Styles) ToSectionsStyles() components.SectionsStyles {
	return components.SectionsStyles{
		Active:   s.SectionActive,
		Inactive: s.SectionInactive,
		Count:    s.SectionCount,
	}
}
