package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/core"
)

// SidebarStyles holds the styles needed for rendering the sidebar
type SidebarStyles struct {
	Border        lipgloss.Style
	Title         lipgloss.Style
	Label         lipgloss.Style
	Value         lipgloss.Style
	Dim           lipgloss.Style
	PriorityHigh  lipgloss.Color
	PriorityMedium lipgloss.Color
	PriorityLow   lipgloss.Color
	DueOverdue    lipgloss.Color
	StatusPending lipgloss.Color
	StatusActive  lipgloss.Color
	StatusDone    lipgloss.Color
	StatusWaiting lipgloss.Color
	Tag           lipgloss.Color
}

// Sidebar displays detailed information about a task
type Sidebar struct {
	task     *core.Task
	allTasks []core.Task // All tasks for dependency lookups
	width    int
	height   int
	offset   int // Scroll offset for content
	styles   SidebarStyles
}

// NewSidebar creates a new sidebar component
func NewSidebar(width, height int, styles SidebarStyles) Sidebar {
	return Sidebar{
		task:   nil,
		width:  width,
		height: height,
		offset: 0,
		styles: styles,
	}
}

// SetTask updates the task being displayed
func (s *Sidebar) SetTask(task *core.Task) {
	s.task = task
	s.offset = 0 // Reset scroll when task changes
}

// SetAllTasks updates the list of all tasks for dependency lookups
func (s *Sidebar) SetAllTasks(tasks []core.Task) {
	s.allTasks = tasks
}

// SetSize updates the sidebar dimensions
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// Update handles messages for the sidebar
func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	if s.task == nil {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Use pointer to modify the sidebar in place
		s.handleKey(msg)
		return s, nil
	}
	return s, nil
}

// handleKey processes keyboard input for scrolling
func (s *Sidebar) handleKey(msg tea.KeyMsg) {
	// Only handle scrolling keys when sidebar is active
	switch msg.String() {
	case "J": // J (shift+j) for line down
		s.scrollDown(1)
	case "K": // K (shift+k) for line up
		s.scrollUp(1)
	case "ctrl+d": // Jump to bottom
		s.scrollToBottom()
	case "ctrl+u": // Jump to top
		s.scrollToTop()
	case "ctrl+f", "pgdown": // Full page down
		s.scrollDown(s.height)
	case "ctrl+b", "pgup": // Full page up
		s.scrollUp(s.height)
	}
}

// scrollDown scrolls the sidebar content down
func (s *Sidebar) scrollDown(amount int) {
	// Calculate actual content lines by rendering
	sections := s.renderContent()
	lines := strings.Split(sections, "\n")
	totalLines := len(lines)

	contentHeight := s.height - 10
	if contentHeight < 1 {
		contentHeight = 1
	}

	maxOffset := totalLines - contentHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	s.offset += amount
	if s.offset > maxOffset {
		s.offset = maxOffset
	}
}

// scrollUp scrolls the sidebar content up
func (s *Sidebar) scrollUp(amount int) {
	s.offset -= amount
	if s.offset < 0 {
		s.offset = 0
	}
}

// scrollToBottom scrolls to the bottom of the sidebar content
func (s *Sidebar) scrollToBottom() {
	// Set to a very large number - View() will clamp it
	s.offset = 9999
}

// scrollToTop scrolls to the top of the sidebar content
func (s *Sidebar) scrollToTop() {
	s.offset = 0
}

// contentLineCount estimates the number of content lines
func (s *Sidebar) contentLineCount() int {
	if s.task == nil {
		return 0
	}

	// Rough estimate based on content sections
	lines := 10 // Base fields
	lines += len(s.task.Tags)
	lines += len(s.task.Annotations) * 3
	lines += len(s.task.Depends) * 2
	lines += len(s.task.UDAs)

	// Description wrapping
	if len(s.task.Description) > s.width-6 {
		lines += len(s.task.Description) / (s.width - 6)
	}

	return lines
}

// View renders the sidebar
func (s Sidebar) View() string {
	if s.task == nil {
		return s.renderEmpty()
	}

	// Render all content sections
	sections := s.renderContent()

	// Apply scrolling offset
	lines := strings.Split(sections, "\n")

	// Account for border (2 lines) + padding (2 lines from Padding(1, 2)) = 4 lines total
	// Add extra safety margin to account for lipgloss wrapping long lines
	contentHeight := s.height - 10
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Clamp offset to ensure we can fill the screen
	// If offset is too large, show the last contentHeight lines
	maxOffset := len(lines) - contentHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := s.offset
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	visibleLines := lines[offset:]
	if len(visibleLines) > contentHeight {
		visibleLines = visibleLines[:contentHeight]
	}

	// Fill remaining space to match content height
	for len(visibleLines) < contentHeight {
		visibleLines = append(visibleLines, "")
	}

	content := strings.Join(visibleLines, "\n")

	// Apply sidebar styling with explicit height to ensure border closes at bottom
	return s.styles.Border.
		Width(s.width - 4).
		MaxHeight(s.height).
		Render(content)
}

// renderEmpty renders the empty state
func (s Sidebar) renderEmpty() string {
	// Fill empty content to match expected height
	contentHeight := s.height - 6 // border(2) + padding(2) + safety margin(2)
	if contentHeight < 1 {
		contentHeight = 1
	}

	var lines []string
	lines = append(lines, s.styles.Dim.Render("No task selected"))
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}

	return s.styles.Border.
		Width(s.width - 4).
		MaxHeight(s.height). // Use MaxHeight to prevent overflow
		Render(strings.Join(lines, "\n"))
}

// renderContent renders all task details
func (s Sidebar) renderContent() string {
	var sections []string

	// Title
	title := fmt.Sprintf("Task #%d", s.task.ID)
	if s.task.ID == 0 {
		title = "Task Details"
	}
	sections = append(sections, s.styles.Title.Render(title))
	sections = append(sections, "")

	// UUID
	sections = append(sections, s.renderField("UUID", s.task.UUID))

	// Description (wrapped)
	sections = append(sections, s.renderWrappedField("Description", s.task.Description))

	// Project
	project := s.task.Project
	if project == "" {
		project = "-"
	}
	sections = append(sections, s.renderField("Project", project))

	// Status
	sections = append(sections, s.renderStatusField())

	// Priority
	priority := s.task.Priority
	if priority == "" {
		priority = "-"
	}
	sections = append(sections, s.renderPriorityField(priority))

	// Tags
	if len(s.task.Tags) > 0 {
		sections = append(sections, "")
		sections = append(sections, s.renderTags())
	}

	// Virtual Tags (based on status)
	virtualTags := s.getVirtualTags()
	if len(virtualTags) > 0 {
		// Always add spacing before virtual tags
		sections = append(sections, "")
		sections = append(sections, s.renderVirtualTags(virtualTags))
	}

	// Dates section
	sections = append(sections, "")
	sections = append(sections, s.renderDates())

	// Dependencies
	if len(s.task.Depends) > 0 {
		sections = append(sections, "")
		sections = append(sections, s.renderDependencies())
	}

	// Annotations
	if len(s.task.Annotations) > 0 {
		sections = append(sections, "")
		sections = append(sections, s.renderAnnotations())
	}

	// UDAs (User Defined Attributes)
	if len(s.task.UDAs) > 0 {
		sections = append(sections, "")
		sections = append(sections, s.renderUDAs())
	}

	// Urgency
	sections = append(sections, "")
	sections = append(sections, s.renderField("Urgency", fmt.Sprintf("%.2f", s.task.Urgency)))

	return strings.Join(sections, "\n")
}

// renderField renders a simple label: value field
func (s Sidebar) renderField(label, value string) string {
	return fmt.Sprintf("%s: %s", s.styles.Label.Render(label), value)
}

// renderWrappedField renders a field with wrapped text
func (s Sidebar) renderWrappedField(label, value string) string {
	// Wrap text to fit width
	// Account for border width (-4) and padding (-4) = -8 total
	maxWidth := s.width - 8
	wrapped := wrapText(value, maxWidth)

	return fmt.Sprintf("%s:\n%s", s.styles.Label.Render(label), wrapped)
}

// renderStatusField renders the status with color coding
func (s Sidebar) renderStatusField() string {
	statusStyle := lipgloss.NewStyle()
	statusText := s.task.Status

	// Check if task is started (has Start field set)
	if s.task.Start != nil {
		statusStyle = statusStyle.Foreground(s.styles.StatusActive).Bold(true)
		statusText = s.task.Status + " (started)"
	} else {
		switch s.task.Status {
		case "pending":
			statusStyle = statusStyle.Foreground(s.styles.StatusPending)
		case "completed":
			statusStyle = statusStyle.Foreground(s.styles.StatusDone)
		case "deleted":
			statusStyle = statusStyle.Foreground(s.styles.Dim.GetForeground())
		case "waiting":
			statusStyle = statusStyle.Foreground(s.styles.StatusWaiting)
		}
	}

	return fmt.Sprintf("%s: %s", s.styles.Label.Render("Status"), statusStyle.Render(statusText))
}

// renderPriorityField renders priority with color coding
func (s Sidebar) renderPriorityField(priority string) string {
	priorityStyle := lipgloss.NewStyle()
	switch priority {
	case "H":
		priorityStyle = priorityStyle.Foreground(s.styles.PriorityHigh)
	case "M":
		priorityStyle = priorityStyle.Foreground(s.styles.PriorityMedium)
	case "L":
		priorityStyle = priorityStyle.Foreground(s.styles.PriorityLow)
	}

	return fmt.Sprintf("%s: %s", s.styles.Label.Render("Priority"), priorityStyle.Render(priority))
}

// renderTags renders the tags section
func (s Sidebar) renderTags() string {
	tagStyle := lipgloss.NewStyle().
		Foreground(s.styles.Tag)

	tags := make([]string, len(s.task.Tags))
	for i, tag := range s.task.Tags {
		tags[i] = tagStyle.Render("+" + tag)
	}

	return fmt.Sprintf("%s: %s", s.styles.Label.Render("Tags"), strings.Join(tags, " "))
}

// getVirtualTags returns virtual tags based on task state
// Only shows the most relevant virtual tags (ACTIVE, WAITING)
func (s Sidebar) getVirtualTags() []string {
	var virtualTags []string

	// A task is ACTIVE if it has a start time set
	if s.task.Start != nil {
		virtualTags = append(virtualTags, "ACTIVE")
	}

	// A task is WAITING if status is waiting
	if s.task.Status == "waiting" {
		virtualTags = append(virtualTags, "WAITING")
	}

	return virtualTags
}

// renderVirtualTags renders virtual tags
func (s Sidebar) renderVirtualTags(virtualTags []string) string {
	// Use tag color for virtual tags to make them visible
	tagStyle := lipgloss.NewStyle().
		Foreground(s.styles.Tag)

	styledTags := make([]string, len(virtualTags))
	for i, tag := range virtualTags {
		styledTags[i] = tagStyle.Render("+" + tag)
	}

	return fmt.Sprintf("%s: %s", s.styles.Label.Render("Virtual"), strings.Join(styledTags, " "))
}

// renderDates renders all date fields
func (s Sidebar) renderDates() string {
	var lines []string

	lines = append(lines, s.styles.Label.Underline(true).Render("Dates"))

	// Due date
	if s.task.Due != nil {
		dueStr := formatDateWithRelative(*s.task.Due)
		style := lipgloss.NewStyle()
		if s.task.IsOverdue() {
			style = style.Foreground(s.styles.DueOverdue)
		}
		lines = append(lines, fmt.Sprintf("  Due: %s", style.Render(dueStr)))
	}

	// Scheduled date
	if s.task.Scheduled != nil {
		lines = append(lines, fmt.Sprintf("  Scheduled: %s", formatDateWithRelative(*s.task.Scheduled)))
	}

	// Wait date
	if s.task.Wait != nil {
		lines = append(lines, fmt.Sprintf("  Wait: %s", formatDateWithRelative(*s.task.Wait)))
	}

	// Start date (when task was started)
	if s.task.Start != nil {
		lines = append(lines, fmt.Sprintf("  Started: %s", formatDateWithRelative(*s.task.Start)))
	}

	// Entry date
	lines = append(lines, fmt.Sprintf("  Created: %s", formatDateWithRelative(s.task.Entry)))

	// Modified date
	if s.task.Modified != nil {
		lines = append(lines, fmt.Sprintf("  Modified: %s", formatDateWithRelative(*s.task.Modified)))
	}

	// End date (for completed tasks)
	if s.task.End != nil {
		lines = append(lines, fmt.Sprintf("  Completed: %s", formatDateWithRelative(*s.task.End)))
	}

	return strings.Join(lines, "\n")
}

// renderDependencies renders task dependencies
func (s Sidebar) renderDependencies() string {
	var lines []string
	lines = append(lines, s.styles.Label.Underline(true).Render("Dependencies"))

	if len(s.task.Depends) > 0 {
		lines = append(lines, "  Blocked by:")
		for _, uuid := range s.task.Depends {
			// Look up the dependent task
			depTask := s.findTaskByUUID(uuid)

			if depTask != nil {
				// Check if task is completed
				if depTask.Status == "completed" {
					// Show with "x" prefix for completed tasks
					lines = append(lines, fmt.Sprintf("    - x %s", depTask.Description))
				} else {
					// Show ID and description for non-completed tasks
					lines = append(lines, fmt.Sprintf("    - #%d: %s", depTask.ID, depTask.Description))
				}
			} else {
				// Show shortened UUID if task not found
				shortUUID := uuid
				if len(shortUUID) > 8 {
					shortUUID = shortUUID[:8]
				}
				lines = append(lines, fmt.Sprintf("    - %s", shortUUID))
			}
		}
	} else {
		lines = append(lines, "  None")
	}

	return strings.Join(lines, "\n")
}

// findTaskByUUID finds a task in allTasks by its UUID
func (s Sidebar) findTaskByUUID(uuid string) *core.Task {
	for i := range s.allTasks {
		if s.allTasks[i].UUID == uuid {
			return &s.allTasks[i]
		}
	}
	return nil
}

// renderAnnotations renders task annotations
func (s Sidebar) renderAnnotations() string {
	var lines []string
	lines = append(lines, s.styles.Label.Underline(true).Render("Annotations"))

	for _, ann := range s.task.Annotations {
		dateStr := formatDateWithRelative(ann.Entry)
		lines = append(lines, fmt.Sprintf("  [%s]", dateStr))

		// Wrap annotation text
		// Account for border width (-4) and padding (-4) = -8 total
		wrapped := wrapText(ann.Description, s.width-8)
		for _, line := range strings.Split(wrapped, "\n") {
			lines = append(lines, fmt.Sprintf("  %s", line))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderUDAs renders user-defined attributes
func (s Sidebar) renderUDAs() string {
	var lines []string
	lines = append(lines, s.styles.Label.Underline(true).Render("Custom Fields"))

	// Sort keys for consistent display
	for key, value := range s.task.UDAs {
		lines = append(lines, fmt.Sprintf("  %s: %s", key, value))
	}

	return strings.Join(lines, "\n")
}

// formatDateWithRelative formats a date with relative time
func formatDateWithRelative(t time.Time) string {
	// Convert to local timezone for display
	localTime := t.Local()

	// Format: "2006-01-02 15:04 (2 days ago)"
	dateStr := localTime.Format("2006-01-02 15:04")
	relativeStr := formatRelativeTime(t)

	if relativeStr != "" {
		return fmt.Sprintf("%s (%s)", dateStr, relativeStr)
	}
	return dateStr
}

// formatRelativeTime returns a human-readable relative time string
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		diff = -diff
		// Future dates
		if diff < time.Minute {
			return "in moments"
		} else if diff < time.Hour {
			mins := int(diff.Minutes())
			if mins == 1 {
				return "in 1 minute"
			}
			return fmt.Sprintf("in %d minutes", mins)
		} else if diff < 24*time.Hour {
			hours := int(diff.Hours())
			if hours == 1 {
				return "in 1 hour"
			}
			return fmt.Sprintf("in %d hours", hours)
		} else if diff < 36*time.Hour {
			// 24-36 hours = tomorrow
			return "tomorrow"
		} else if diff < 7*24*time.Hour {
			days := int(diff.Hours() / 24)
			return fmt.Sprintf("in %d days", days)
		}
		return ""
	}

	// Past dates
	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	return ""
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}
