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
	Border              lipgloss.Style
	Title               lipgloss.Style
	Label               lipgloss.Style
	Value               lipgloss.Style
	Dim                 lipgloss.Style
	AnnotationTimestamp lipgloss.Style
	PriorityHigh        lipgloss.Color
	PriorityMedium lipgloss.Color
	PriorityLow    lipgloss.Color
	DueOverdue     lipgloss.Color
	StatusPending  lipgloss.Color
	StatusActive   lipgloss.Color
	StatusDone     lipgloss.Color
	StatusWaiting  lipgloss.Color
	Tag            lipgloss.Color
}

// Sidebar displays detailed information about a task
type Sidebar struct {
	task     *core.Task
	allTasks []core.Task // All tasks for dependency lookups
	width    int
	height   int
	offset   int // Scroll offset for main content (left panel)
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
		s.handleKey(msg)
		return s, nil
	}
	return s, nil
}

// handleKey processes keyboard input for scrolling
func (s *Sidebar) handleKey(msg tea.KeyMsg) {
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

// scrollDown scrolls the main content down
func (s *Sidebar) scrollDown(amount int) {
	// Use approximate left panel width for content measurement
	leftWidth := s.width - rightPanelWidth(s.width)
	contentWidth := leftWidth - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	lines := strings.Split(s.renderMainContent(contentWidth), "\n")
	totalLines := len(lines)

	contentHeight := s.height - titleHeight()
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

// scrollUp scrolls the main content up
func (s *Sidebar) scrollUp(amount int) {
	s.offset -= amount
	if s.offset < 0 {
		s.offset = 0
	}
}

// scrollToBottom scrolls to the bottom of the main content
func (s *Sidebar) scrollToBottom() {
	s.offset = 9999
}

// scrollToTop scrolls to the top of the main content
func (s *Sidebar) scrollToTop() {
	s.offset = 0
}

// titleHeight returns the number of lines the title section occupies
func titleHeight() int {
	return 2 // title line + separator line
}

// rightPanelWidth returns the width of the right metadata panel
func rightPanelWidth(totalWidth int) int {
	w := totalWidth * 30 / 100
	if w < 26 {
		w = 26
	}
	if w > 46 {
		w = 46
	}
	return w
}

// View renders the task detail page
func (s Sidebar) View() string {
	if s.task == nil {
		return s.renderEmpty()
	}

	rightWidth := rightPanelWidth(s.width)
	leftWidth := s.width - rightWidth

	// Title occupies 2 lines (content + separator)
	th := titleHeight()
	contentHeight := s.height - th
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Left panel inner width (Padding(0,2) adds 4 to outer width)
	leftContentWidth := leftWidth - 4
	if leftContentWidth < 10 {
		leftContentWidth = 10
	}

	// Render title (full width)
	title := s.renderTitle()

	// Render and scroll main content (dependencies + annotations)
	mainContent := s.renderMainContent(leftContentWidth)
	mainLines := strings.Split(mainContent, "\n")

	maxOffset := len(mainLines) - contentHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := s.offset
	if offset > maxOffset {
		offset = maxOffset
	}

	visibleLines := mainLines[offset:]
	if len(visibleLines) > contentHeight {
		visibleLines = visibleLines[:contentHeight]
	}
	for len(visibleLines) < contentHeight {
		visibleLines = append(visibleLines, "")
	}

	leftPanel := lipgloss.NewStyle().
		Width(leftContentWidth).
		Height(contentHeight).
		Padding(0, 2).
		Render(strings.Join(visibleLines, "\n"))

	// Right panel: proper sidebar with a full-height │ separator on the left.
	// Overhead: 1 (│) + 1 (left pad) + 1 (right pad) = 3 chars.
	rightInnerWidth := rightWidth - 3
	if rightInnerWidth < 8 {
		rightInnerWidth = 8
	}

	// Pad metadata content to fill the full height so the separator extends top-to-bottom.
	metaLines := strings.Split(s.renderMetadata(), "\n")
	for len(metaLines) < contentHeight {
		metaLines = append(metaLines, "")
	}
	if len(metaLines) > contentHeight {
		metaLines = metaLines[:contentHeight]
	}

	sidebarStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.Border{Left: "│"}).
		BorderLeft(true).
		BorderForeground(s.styles.Dim.GetForeground()).
		PaddingLeft(1).
		PaddingRight(1)

	rightPanel := sidebarStyle.
		Width(rightInnerWidth).
		Render(strings.Join(metaLines, "\n"))

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	return lipgloss.JoinVertical(lipgloss.Left, title, body)
}

// renderEmpty renders the empty state (no task selected)
func (s Sidebar) renderEmpty() string {
	contentHeight := s.height - 6
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
		MaxHeight(s.height).
		Render(strings.Join(lines, "\n"))
}

// renderTitle renders the task title bar: "#ID  Description" + separator
func (s Sidebar) renderTitle() string {
	idStr := s.styles.Label.Render(fmt.Sprintf("#%d", s.task.ID))
	if s.task.ID == 0 {
		idStr = s.styles.Label.Render("Task")
	}
	titleLine := idStr + "  " + s.task.Description
	separator := s.styles.Dim.Render(strings.Repeat("─", s.width))
	return titleLine + "\n" + separator
}

// renderMainContent renders the scrollable left panel (dependencies + annotations)
func (s Sidebar) renderMainContent(contentWidth int) string {
	var sections []string

	if len(s.task.Depends) > 0 {
		sections = append(sections, s.renderDependencies())
		sections = append(sections, "")
	}

	if len(s.task.Annotations) > 0 {
		sections = append(sections, s.renderAnnotations(contentWidth))
	}

	if len(sections) == 0 {
		sections = append(sections, s.styles.Dim.Render("No dependencies or annotations"))
	}

	return strings.Join(sections, "\n")
}

// renderMetadata renders the right sidebar metadata panel (fixed, not scrollable)
func (s Sidebar) renderMetadata() string {
	var lines []string

	// UUID (show short form when task has no numeric ID, i.e. completed/deleted)
	if s.task.ID == 0 {
		shortUUID, _ := s.task.GetProperty("short_uuid")
		lines = append(lines, s.renderField("UUID", shortUUID))
	}

	// Status
	lines = append(lines, s.renderStatusField())

	// Project
	project := s.task.Project
	if project == "" {
		project = "-"
	}
	lines = append(lines, s.renderField("Project", project))

	// Priority
	priority := s.task.Priority
	if priority == "" {
		priority = "-"
	}
	lines = append(lines, s.renderPriorityField(priority))

	// Tags
	if len(s.task.Tags) > 0 {
		lines = append(lines, s.renderTags())
	}

	// Virtual tags
	virtualTags := s.getVirtualTags()
	if len(virtualTags) > 0 {
		lines = append(lines, s.renderVirtualTags(virtualTags))
	}

	// Dates
	lines = append(lines, "")
	lines = append(lines, s.renderDatesCompact())

	// Urgency
	lines = append(lines, "")
	lines = append(lines, s.renderField("Urgency", fmt.Sprintf("%.2f", s.task.Urgency)))

	// UDAs
	if len(s.task.UDAs) > 0 {
		lines = append(lines, "")
		lines = append(lines, s.renderUDAs())
	}

	return strings.Join(lines, "\n")
}

// renderField renders a simple label: value field
func (s Sidebar) renderField(label, value string) string {
	return fmt.Sprintf("%s: %s", s.styles.Label.Render(label), value)
}

// renderStatusField renders the status with color coding
func (s Sidebar) renderStatusField() string {
	statusStyle := lipgloss.NewStyle()
	statusText := s.task.Status

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
	tagStyle := lipgloss.NewStyle().Foreground(s.styles.Tag)

	tags := make([]string, len(s.task.Tags))
	for i, tag := range s.task.Tags {
		tags[i] = tagStyle.Render("+" + tag)
	}

	return fmt.Sprintf("%s: %s", s.styles.Label.Render("Tags"), strings.Join(tags, " "))
}

// getVirtualTags returns virtual tags based on task state
func (s Sidebar) getVirtualTags() []string {
	var virtualTags []string
	if s.task.Start != nil {
		virtualTags = append(virtualTags, "ACTIVE")
	}
	if s.task.Status == "waiting" {
		virtualTags = append(virtualTags, "WAITING")
	}
	return virtualTags
}

// renderVirtualTags renders virtual tags
func (s Sidebar) renderVirtualTags(virtualTags []string) string {
	tagStyle := lipgloss.NewStyle().Foreground(s.styles.Tag)

	styledTags := make([]string, len(virtualTags))
	for i, tag := range virtualTags {
		styledTags[i] = tagStyle.Render("+" + tag)
	}

	return fmt.Sprintf("%s: %s", s.styles.Label.Render("Virtual"), strings.Join(styledTags, " "))
}

// renderDatesCompact renders date fields in a compact format for the right panel
func (s Sidebar) renderDatesCompact() string {
	var lines []string

	lines = append(lines, s.styles.Label.Underline(true).Render("Dates"))

	if s.task.Due != nil {
		style := lipgloss.NewStyle()
		if s.task.IsOverdue() {
			style = style.Foreground(s.styles.DueOverdue)
		}
		lines = append(lines, fmt.Sprintf("  Due: %s", style.Render(formatCompactDate(*s.task.Due))))
	}
	if s.task.Scheduled != nil {
		lines = append(lines, fmt.Sprintf("  Sched: %s", formatCompactDate(*s.task.Scheduled)))
	}
	if s.task.Wait != nil {
		lines = append(lines, fmt.Sprintf("  Wait: %s", formatCompactDate(*s.task.Wait)))
	}
	if s.task.Start != nil {
		lines = append(lines, fmt.Sprintf("  Started: %s", formatCompactDate(*s.task.Start)))
	}
	lines = append(lines, fmt.Sprintf("  Created: %s", formatCompactDate(s.task.Entry)))
	if s.task.Modified != nil {
		lines = append(lines, fmt.Sprintf("  Modified: %s", formatCompactDate(*s.task.Modified)))
	}
	if s.task.End != nil {
		lines = append(lines, fmt.Sprintf("  Done: %s", formatCompactDate(*s.task.End)))
	}

	return strings.Join(lines, "\n")
}

// renderDependencies renders task dependencies
func (s Sidebar) renderDependencies() string {
	var lines []string
	lines = append(lines, s.styles.Label.Underline(true).Render("Dependencies"))

	for _, uuid := range s.task.Depends {
		depTask := s.findTaskByUUID(uuid)
		if depTask != nil {
			if depTask.Status == "completed" {
				lines = append(lines, fmt.Sprintf("  ✓ %s", depTask.Description))
			} else {
				lines = append(lines, fmt.Sprintf("  ○ #%d: %s", depTask.ID, depTask.Description))
			}
		} else {
			shortUUID := uuid
			if len(shortUUID) > 8 {
				shortUUID = shortUUID[:8]
			}
			lines = append(lines, fmt.Sprintf("  ○ %s", shortUUID))
		}
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
func (s Sidebar) renderAnnotations(contentWidth int) string {
	var lines []string
	lines = append(lines, s.styles.Label.Underline(true).Render("Annotations"))

	for _, ann := range s.task.Annotations {
		dateStr := formatDateWithRelative(ann.Entry)
		lines = append(lines, "  "+s.styles.AnnotationTimestamp.Render("["+dateStr+"]"))

		wrapped := wrapText(ann.Description, contentWidth-4)
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

	for key, value := range s.task.UDAs {
		lines = append(lines, fmt.Sprintf("  %s: %s", key, value))
	}

	return strings.Join(lines, "\n")
}

// formatCompactDate formats a date compactly: relative time or date only
func formatCompactDate(t time.Time) string {
	relative := formatRelativeTime(t)
	if relative != "" {
		return relative
	}
	return t.Local().Format("2006-01-02")
}

// formatDateWithRelative formats a date with relative time
func formatDateWithRelative(t time.Time) string {
	localTime := t.Local()
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
			return "tomorrow"
		} else if diff < 7*24*time.Hour {
			days := int(diff.Hours() / 24)
			return fmt.Sprintf("in %d days", days)
		}
		return ""
	}

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
