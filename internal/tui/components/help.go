package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// Keybinding represents a keyboard shortcut and its description
type Keybinding struct {
	Keys        []string // Multiple keys that perform the same action (e.g., ["j", "↓"])
	Description string
}

// KeybindingGroup represents a group of related keybindings
type KeybindingGroup struct {
	Title    string
	Bindings []Keybinding
}

// Help component displays the help screen with keybinding reference
type Help struct {
	viewport viewport.Model
	groups   []KeybindingGroup
	width    int
	height   int
	styles   HelpStyles
}

// HelpStyles contains styling for the help component
type HelpStyles struct {
	Title       lipgloss.Style
	GroupTitle  lipgloss.Style
	Key         lipgloss.Style
	Description lipgloss.Style
	Border      lipgloss.Style
}

// DefaultHelpStyles returns the default help styles
func DefaultHelpStyles() HelpStyles {
	return HelpStyles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Padding(1, 0),
		GroupTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14")).
			Padding(1, 0, 0, 0),
		Key: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10")).
			Width(12),
		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 2),
	}
}

// NewHelp creates a new help component with default keybindings
func NewHelp(width, height int, styles HelpStyles) Help {
	groups := defaultKeybindingGroups()

	vp := viewport.New(width-4, height-4) // Account for border padding
	vp.SetContent(renderHelpContent(groups, styles))

	return Help{
		viewport: vp,
		groups:   groups,
		width:    width,
		height:   height,
		styles:   styles,
	}
}

// defaultKeybindingGroups returns the default keybinding groups
func defaultKeybindingGroups() []KeybindingGroup {
	return []KeybindingGroup{
		{
			Title: "Task Navigation",
			Bindings: []Keybinding{
				{Keys: []string{"j", "↓"}, Description: "Move down"},
				{Keys: []string{"k", "↑"}, Description: "Move up"},
				{Keys: []string{"g"}, Description: "Jump to first task"},
				{Keys: []string{"G"}, Description: "Jump to last task"},
				{Keys: []string{"1-9"}, Description: "Quick jump to task"},
			},
		},
		{
			Title: "Section Navigation",
			Bindings: []Keybinding{
				{Keys: []string{"Tab", "l", "→"}, Description: "Next section"},
				{Keys: []string{"Shift+Tab", "h", "←"}, Description: "Previous section"},
				{Keys: []string{"1-5"}, Description: "Jump to section"},
			},
		},
		{
			Title: "Multi-Select",
			Bindings: []Keybinding{
				{Keys: []string{"Space"}, Description: "Toggle task selection"},
				{Keys: []string{"Esc"}, Description: "Clear all selections"},
			},
		},
		{
			Title: "Task Actions",
			Bindings: []Keybinding{
				{Keys: []string{"d"}, Description: "Mark task(s) done"},
				{Keys: []string{"s"}, Description: "Start/Stop task(s)"},
				{Keys: []string{"x"}, Description: "Delete task(s)"},
				{Keys: []string{"e"}, Description: "Edit task in $EDITOR"},
				{Keys: []string{"n"}, Description: "Create new task"},
				{Keys: []string{"m"}, Description: "Modify task(s) (quick edit)"},
				{Keys: []string{"M"}, Description: "Export task(s) as markdown (to clipboard)"},
				{Keys: []string{"a"}, Description: "Add annotation to task(s)"},
				{Keys: []string{"u"}, Description: "Undo last operation"},
			},
		},
		{
			Title: "View Controls",
			Bindings: []Keybinding{
				{Keys: []string{"Enter"}, Description: "Toggle sidebar / Drill into group"},
				{Keys: []string{"Esc"}, Description: "Close sidebar / Back to group list"},
				{Keys: []string{"/"}, Description: "Filter tasks"},
				{Keys: []string{"r"}, Description: "Refresh task list"},
			},
		},
		{
			Title: "Sidebar Scrolling (when sidebar is open)",
			Bindings: []Keybinding{
				{Keys: []string{"J"}, Description: "Scroll down one line"},
				{Keys: []string{"K"}, Description: "Scroll up one line"},
				{Keys: []string{"Ctrl+d"}, Description: "Jump to bottom"},
				{Keys: []string{"Ctrl+u"}, Description: "Jump to top"},
				{Keys: []string{"Ctrl+f", "PgDn"}, Description: "Scroll down full page"},
				{Keys: []string{"Ctrl+b", "PgUp"}, Description: "Scroll up full page"},
			},
		},
		{
			Title: "Other",
			Bindings: []Keybinding{
				{Keys: []string{"?"}, Description: "Toggle this help screen"},
				{Keys: []string{"q"}, Description: "Quit wui"},
				{Keys: []string{"Ctrl+c"}, Description: "Force quit"},
			},
		},
	}
}

// SetKeybindings sets custom keybinding groups
func (h *Help) SetKeybindings(groups []KeybindingGroup) {
	h.groups = groups
	h.viewport.SetContent(renderHelpContent(groups, h.styles))
}

// SetSize updates the size of the help component
func (h *Help) SetSize(width, height int) {
	h.width = width
	h.height = height
	h.viewport.Width = width - 4
	h.viewport.Height = height - 4
}

// Update handles messages for the help component
func (h Help) Update(msg tea.Msg) (Help, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			h.viewport.LineDown(1)
		case "k", "up":
			h.viewport.LineUp(1)
		case "ctrl+d":
			h.viewport.HalfViewDown()
		case "ctrl+u":
			h.viewport.HalfViewUp()
		case "ctrl+f", "pgdown":
			h.viewport.ViewDown()
		case "ctrl+b", "pgup":
			h.viewport.ViewUp()
		case "g", "home":
			h.viewport.GotoTop()
		case "G", "end":
			h.viewport.GotoBottom()
		}
	case tea.WindowSizeMsg:
		h.SetSize(msg.Width, msg.Height)
	}

	h.viewport, cmd = h.viewport.Update(msg)
	return h, cmd
}

// View renders the help screen
func (h Help) View() string {
	if h.width == 0 || h.height == 0 {
		return ""
	}

	// Render viewport content in a bordered box
	content := h.viewport.View()

	return h.styles.Border.
		Width(h.width - 4).
		Height(h.height - 4).
		Render(content)
}

// renderHelpContent renders the help content as a string
func renderHelpContent(groups []KeybindingGroup, styles HelpStyles) string {
	var sections []string

	// Title
	title := styles.Title.Render("Keyboard Shortcuts")
	sections = append(sections, title)

	// Render each group
	for _, group := range groups {
		groupTitle := styles.GroupTitle.Render(group.Title + ":")
		sections = append(sections, groupTitle)

		for _, binding := range group.Bindings {
			// Join multiple keys with "/"
			keyStr := strings.Join(binding.Keys, "/")
			key := styles.Key.Render("  " + keyStr)
			desc := styles.Description.Render(binding.Description)
			line := lipgloss.JoinHorizontal(lipgloss.Left, key, desc)
			sections = append(sections, line)
		}

		// Add spacing between groups
		sections = append(sections, "")
	}

	// Add footer
	footer := styles.Description.Render("Press ? or Esc to close • Use j/k or ↑/↓ to scroll")
	sections = append(sections, "", footer)

	return strings.Join(sections, "\n")
}
