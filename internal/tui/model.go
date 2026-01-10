package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clobrano/wui/internal/calendar"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
	"github.com/clobrano/wui/internal/taskwarrior"
	"github.com/clobrano/wui/internal/tui/components"
)

// ViewMode represents the view layout mode
type ViewMode int

const (
	// ViewModeList shows only the task list
	ViewModeList ViewMode = iota
	// ViewModeListWithSidebar shows task list with detail sidebar
	ViewModeListWithSidebar
	// ViewModeSmall shows task list optimized for small screens (2 lines per task)
	ViewModeSmall
	// ViewModeSmallTaskDetail shows full-screen task details on small screens
	ViewModeSmallTaskDetail
)

// String returns the string representation of ViewMode
func (v ViewMode) String() string {
	switch v {
	case ViewModeList:
		return "list"
	case ViewModeListWithSidebar:
		return "list_with_sidebar"
	case ViewModeSmall:
		return "small"
	case ViewModeSmallTaskDetail:
		return "small_task_detail"
	default:
		return "unknown"
	}
}

// AppState represents the current application state
type AppState int

const (
	// StateNormal is the default state for navigation and task operations
	StateNormal AppState = iota
	// StateFilterInput is active when user is entering a filter
	StateFilterInput
	// StateHelp is active when help screen is shown
	StateHelp
	// StateConfirm is active when user needs to confirm an action
	StateConfirm
	// StateModifyInput is active when user is entering task modifications
	StateModifyInput
	// StateAnnotateInput is active when user is adding an annotation
	StateAnnotateInput
	// StateNewTaskInput is active when user is creating a new task
	StateNewTaskInput
)

// String returns the string representation of AppState
func (s AppState) String() string {
	switch s {
	case StateNormal:
		return "normal"
	case StateFilterInput:
		return "filter_input"
	case StateHelp:
		return "help"
	case StateConfirm:
		return "confirm"
	case StateModifyInput:
		return "modify_input"
	case StateAnnotateInput:
		return "annotate_input"
	case StateNewTaskInput:
		return "new_task_input"
	default:
		return "unknown"
	}
}

// Model represents the main TUI application model
type Model struct {
	// Core dependencies
	service core.TaskService
	config  *config.Config
	styles  *Styles // Centralized styling

	// Task data
	tasks            []core.Task
	currentSection   *core.Section
	projectSummaries []core.ProjectSummary // Project summaries for Projects tab

	// Grouping state (for Projects/Tags sections)
	groups        []core.TaskGroup // Current groups (when in group list view)
	selectedGroup *core.TaskGroup  // Selected group (when drilling into a group)
	inGroupView   bool             // true = showing group list, false = showing tasks

	// UI state
	viewMode ViewMode
	state    AppState

	// Current filter
	activeFilter string

	// Search tab filter (persists for the session)
	searchTabFilter string

	// Status and error messages
	statusMessage string
	errorMessage  string
	isLoading     bool // Indicates if an async operation is in progress

	// Terminal dimensions
	width  int
	height int

	// Components
	taskList      components.TaskList
	sidebar       components.Sidebar
	filter        components.Filter
	modifyInput   components.Filter // Reuse filter component for modify input
	annotateInput components.Filter // Reuse filter component for annotate input
	newTaskInput  components.Filter // Reuse filter component for new task input
	sections      components.Sections
	help          components.Help

	// Calendar autocompletion
	calendar           components.Calendar
	calendarActive     bool     // true when calendar picker is shown
	calendarFieldType  string   // "due" or "scheduled" - which field is being completed
	calendarInsertPos  int      // position in input where date should be inserted
	calendarInputState AppState // which input state triggered the calendar

	// Time picker autocompletion
	timePicker           components.TimePicker
	timePickerActive     bool     // true when time picker is shown
	timePickerInsertPos  int      // position in input where time should be inserted
	timePickerInputState AppState // which input state triggered the time picker

	// List picker autocompletion (for projects and tags)
	listPicker             components.ListPicker
	listPickerActive       bool     // true when list picker is shown
	listPickerType         string   // "project" or "tag" - which type is being completed
	listPickerInsertPos    int      // position in input where selection should be inserted
	listPickerFilterPrefix string   // the partial text that was typed before TAB (needs to be removed)
	listPickerInputState   AppState // which input state triggered the list picker
	availableProjects      []string // all unique projects from loaded tasks
	availableTags          []string // all unique tags from loaded tasks

	// Confirm action tracking
	confirmAction string // "delete", "done", etc.

	// Calendar sync state
	syncingBeforeQuit bool                  // true when syncing before quit
	syncWarnings      *calendar.SyncResult // warnings to print after quit
}

// NewModel creates a new TUI model
func NewModel(service core.TaskService, cfg *config.Config) Model {
	// Create styles from config theme
	var theme Theme
	if cfg.TUI != nil && cfg.TUI.Theme != nil {
		theme = ThemeFromConfig(cfg.TUI.Theme)
	} else {
		theme = DefaultDarkTheme()
	}
	styles := NewStyles(theme)

	// Get sections from config tabs
	var allSections []core.Section

	// Always prepend the special Search section (non-configurable)
	searchSection := core.Section{
		Name:        "Search",
		Filter:      "",
		Description: "Search across all tasks",
	}
	allSections = append(allSections, searchSection)

	// Add user-configured or default sections
	if cfg.TUI != nil && len(cfg.TUI.Tabs) > 0 {
		// Convert config.Tab to core.Tab
		var coreTabs []core.Tab
		for _, t := range cfg.TUI.Tabs {
			// Skip if user tries to add a Search tab (it's always first)
			if t.Name == "Search" {
				continue
			}
			coreTabs = append(coreTabs, core.Tab{
				Name:   t.Name,
				Filter: t.Filter,
			})
		}
		allSections = append(allSections, core.TabsToSections(coreTabs)...)
	} else {
		allSections = append(allSections, core.DefaultSections()...)
	}

	taskList := components.NewTaskList(80, 24, cfg.TUI.Columns, styles.ToTaskListStyles())
	taskList.SetScrollBuffer(cfg.TUI.ScrollBuffer)

	// Determine initial section: Search tab if --search flag provided, otherwise "Next" tab
	initialSectionIndex := 1 // Default to "Next" tab (index 1)
	initialSearchFilter := ""

	if cfg.InitialSearchFilter != "" {
		// Open in Search tab with the provided filter
		initialSectionIndex = 0
		initialSearchFilter = cfg.InitialSearchFilter
	} else if len(allSections) <= 1 {
		initialSectionIndex = 0 // Fallback to first section if only Search exists
	}

	// Create help component with keybindings and custom commands from config
	var helpComponent components.Help
	if cfg.TUI != nil {
		// Convert config.CustomCommand to components.CustomCommand
		var customCmds map[string]components.CustomCommand
		if len(cfg.TUI.CustomCommands) > 0 {
			customCmds = make(map[string]components.CustomCommand)
			for key, cmd := range cfg.TUI.CustomCommands {
				customCmds[key] = components.CustomCommand{
					Name:        cmd.Name,
					Command:     cmd.Command,
					Description: cmd.Description,
				}
			}
		}

		if cfg.TUI.Keybindings != nil || len(customCmds) > 0 {
			helpComponent = components.NewHelpWithCustomCommands(80, 24, components.DefaultHelpStyles(), cfg.TUI.Keybindings, customCmds)
		} else {
			helpComponent = components.NewHelp(80, 24, components.DefaultHelpStyles())
		}
	} else {
		helpComponent = components.NewHelp(80, 24, components.DefaultHelpStyles())
	}

	m := Model{
		service:         service,
		config:          cfg,
		styles:          styles,
		tasks:           []core.Task{},
		viewMode:        ViewModeList,
		state:           StateNormal,
		currentSection:  &allSections[initialSectionIndex],
		activeFilter:    allSections[initialSectionIndex].Filter,
		searchTabFilter: initialSearchFilter, // Set from --search flag if provided
		groups:          []core.TaskGroup{},
		selectedGroup:   nil,
		inGroupView:     false,
		statusMessage:   "",
		errorMessage:    "",
		taskList:        taskList,
		sidebar:         components.NewSidebar(40, 24, styles.ToSidebarStyles()), // Initial size, will be updated
		filter:          components.NewFilter(),
		modifyInput:     components.NewFilter(),
		annotateInput:   components.NewFilter(),
		newTaskInput:    components.NewFilter(),
		sections:        components.NewSectionsWithIndex(allSections, 80, styles.ToSectionsStyles(), initialSectionIndex), // Initial size, will be updated
		help:            helpComponent,                                                                                    // Initial size, will be updated
		confirmAction:   "",
	}

	// Set custom empty message for Search tab if starting there
	if initialSectionIndex == 0 {
		m.taskList.SetEmptyMessage("Search across all tasks\n\nPress / to enter a search filter\n\nExamples:\n  • bug                    - search for 'bug' in all tasks\n  • project:home           - tasks in 'home' project\n  • status:completed       - completed tasks only\n  • +urgent due.before:eom - urgent tasks due before end of month")
	}

	return m
}

// Init initializes the model and returns the initial command
func (m Model) Init() tea.Cmd {
	// Set loading state and load tasks with the current section's filter
	m.isLoading = true
	isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"

	// If starting in Search tab with a filter, use the search filter
	filterToUse := m.activeFilter
	if isSearchTab && m.searchTabFilter != "" {
		filterToUse = m.searchTabFilter
	}

	// Load both tasks and autocomplete data in parallel
	return tea.Batch(
		loadTasksCmd(m.service, filterToUse, isSearchTab),
		loadAllProjectsAndTagsCmd(m.service),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update component sizes
		m.updateComponentSizes()
		return m, nil

	case components.SectionChangedMsg:
		// Section changed - load tasks with new filter
		m.currentSection = &msg.Section
		m.errorMessage = ""
		m.statusMessage = ""
		m.isLoading = true

		// Reset grouping state when switching sections
		m.selectedGroup = nil
		m.groups = []core.TaskGroup{}

		// Determine if we should show groups
		if m.sections.IsProjectsView() || m.sections.IsTagsView() {
			m.inGroupView = true
		} else {
			m.inGroupView = false
		}

		// Set custom empty message for Search tab
		isSearchTab := msg.Section.Name == "Search"
		if isSearchTab {
			// Restore the saved Search tab filter (if any)
			m.activeFilter = m.searchTabFilter
			m.taskList.SetEmptyMessage("Search across all tasks\n\nPress / to enter a search filter\n\nExamples:\n  • bug                    - search for 'bug' in all tasks\n  • project:home           - tasks in 'home' project\n  • status:completed       - completed tasks only\n  • +urgent due.before:eom - urgent tasks due before end of month")
		} else {
			// Use the section's default filter
			m.activeFilter = msg.Section.Filter
			m.taskList.SetEmptyMessage("") // Reset to default message
		}

		return m, loadTasksCmd(m.service, m.activeFilter, isSearchTab)

	case TasksLoadedMsg:
		m.isLoading = false
		if msg.Err != nil {
			m.errorMessage = "Failed to load tasks: " + msg.Err.Error()
			// If error was from filter, reopen filter input
			if m.state == StateNormal && m.activeFilter != "" {
				m.state = StateFilterInput
				return m, m.filter.Focus()
			}
			return m, nil
		}
		m.tasks = msg.Tasks
		m.errorMessage = ""

		// Note: Projects/tags for autocompletion are loaded separately via AutocompleteDataLoadedMsg
		// to ensure we have ALL projects/tags, not just those in the current filtered view

		// Update sidebar with all tasks for dependency lookups
		m.sidebar.SetAllTasks(m.tasks)

		// If in Projects or Tags view and showing group list, compute groups
		if m.inGroupView {
			if m.sections.IsProjectsView() {
				// For Projects view, load summaries to get completion percentages
				// We'll build groups when summaries arrive
				m.isLoading = true
				return m, loadProjectSummaryCmd(m.service)
			} else if m.sections.IsTagsView() {
				m.groups = core.GroupByTag(m.tasks)
				// Show group list in the task list component
				m.taskList.SetGroups(m.groups)
			}
		} else {
			// Normal view or drilling into a group
			// Update task list component with actual tasks
			m.taskList.SetTasks(m.tasks)
		}

		// Update task count in sections component
		m.sections.SetTaskCount(len(m.tasks))

		// Update sidebar with selected task (only if not in group view)
		if !m.inGroupView {
			m.updateSidebar()
		}

		return m, nil

	case ProjectSummaryLoadedMsg:
		m.isLoading = false
		if msg.Err != nil {
			m.errorMessage = "Failed to load project summary: " + msg.Err.Error()
			return m, nil
		}

		m.projectSummaries = msg.Summaries

		// Build project groups using hierarchy with percentages
		m.groups = core.GroupProjectsByHierarchy(m.projectSummaries, m.tasks)

		// Show group list in the task list component
		m.taskList.SetGroups(m.groups)

		return m, nil

	case TaskModifiedMsg:
		if msg.Err != nil {
			m.errorMessage = "Task operation failed: " + msg.Err.Error()
			m.isLoading = false
			return m, nil
		}
		m.errorMessage = "" // Clear any previous error
		m.statusMessage = "Task updated successfully"
		m.isLoading = true
		// Refresh tasks and autocomplete data
		isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"
		return m, tea.Batch(
			loadTasksCmd(m.service, m.activeFilter, isSearchTab),
			loadAllProjectsAndTagsCmd(m.service),
		)

	case ErrorMsg:
		m.errorMessage = msg.Err.Error()
		return m, nil

	case RefreshMsg:
		isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"
		return m, loadTasksCmd(m.service, m.activeFilter, isSearchTab)

	case StatusMsg:
		if msg.IsError {
			m.errorMessage = msg.Message
		} else {
			m.statusMessage = msg.Message
		}
		return m, nil

	case CalendarSyncCompletedMsg:
		m.isLoading = false
		m.syncingBeforeQuit = false
		if msg.Err != nil {
			m.errorMessage = "Calendar sync failed: " + msg.Err.Error()
			// Don't quit on sync error, let user see the error
			return m, nil
		}

		// Build status message with result details
		if msg.Result != nil {
			m.statusMessage = fmt.Sprintf("Calendar synced: %d created, %d updated", msg.Result.Created, msg.Result.Updated)
			if len(msg.Result.Warnings) > 0 {
				m.statusMessage += fmt.Sprintf(", %d warnings - see output after quit", len(msg.Result.Warnings))
			}
		} else {
			m.statusMessage = "Calendar synced successfully"
		}

		// Store warnings to print after quit
		m.syncWarnings = msg.Result

		// Quit after successful sync
		return m, tea.Quit

	case AutocompleteDataLoadedMsg:
		if msg.Err != nil {
			// Don't show error to user, just use empty lists
			m.availableProjects = []string{}
			m.availableTags = []string{}
		} else {
			m.availableProjects = msg.Projects
			m.availableTags = msg.Tags
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input based on current state
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys (work in any state)
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	// If calendar is active, handle calendar input
	if m.calendarActive {
		var cmd tea.Cmd

		switch msg.String() {
		case "enter":
			// Select date and insert into input
			selectedDate := m.calendar.GetSelectedDate()
			m.insertDateFromCalendar(selectedDate)
			m.deactivateCalendar()
			return m, nil

		case "esc":
			// Cancel calendar
			m.deactivateCalendar()
			return m, nil

		default:
			// Delegate to calendar component
			m.calendar, cmd = m.calendar.Update(msg)
			return m, cmd
		}
	}

	// If time picker is active, handle time picker input
	if m.timePickerActive {
		var cmd tea.Cmd

		switch msg.String() {
		case "enter":
			// Select time and insert into input
			m.insertTimeFromPicker()
			m.deactivateTimePicker()
			return m, nil

		case "esc":
			// Cancel time picker
			m.deactivateTimePicker()
			return m, nil

		default:
			// Delegate to time picker component
			m.timePicker, cmd = m.timePicker.Update(msg)
			return m, cmd
		}
	}

	// If list picker is active, handle list picker input
	if m.listPickerActive {
		var cmd tea.Cmd

		switch msg.String() {
		case "enter":
			// Select item and insert into input
			if m.listPicker.HasItems() {
				m.insertSelectionFromListPicker()
			}
			m.deactivateListPicker()
			return m, nil

		case "esc":
			// Cancel list picker
			m.deactivateListPicker()
			return m, nil

		default:
			// Delegate to list picker component
			m.listPicker, cmd = m.listPicker.Update(msg)
			return m, cmd
		}
	}

	// State-specific handling
	switch m.state {
	case StateNormal:
		return m.handleNormalKeys(msg)
	case StateFilterInput:
		return m.handleFilterKeys(msg)
	case StateHelp:
		return m.handleHelpKeys(msg)
	case StateConfirm:
		return m.handleConfirmKeys(msg)
	case StateModifyInput:
		return m.handleModifyKeys(msg)
	case StateAnnotateInput:
		return m.handleAnnotateKeys(msg)
	case StateNewTaskInput:
		return m.handleNewTaskKeys(msg)
	}

	return m, nil
}

// keyMatches checks if the pressed key matches the configured keybinding for the given action
func (m Model) keyMatches(keyPressed string, action string) bool {
	if m.config == nil || m.config.TUI == nil || m.config.TUI.Keybindings == nil {
		return false
	}
	configuredKey, exists := m.config.TUI.Keybindings[action]
	return exists && configuredKey == keyPressed
}

// detectDateFieldContext checks if cursor is positioned after a date field keyword
// Returns (fieldType, insertPosition, found) where fieldType is "due" or "scheduled"
func detectDateFieldContext(input string, cursorPos int) (string, int, bool) {
	// Get text before cursor
	textBefore := input[:cursorPos]

	// Look for date field keywords at the end
	// Support: "due:", "scheduled:", "sched:", "sch:"
	dateFields := []struct {
		keyword   string
		fieldType string
	}{
		{"due:", "due"},
		{"scheduled:", "scheduled"},
		{"sched:", "scheduled"},
		{"sch:", "scheduled"},
	}

	for _, field := range dateFields {
		if strings.HasSuffix(textBefore, field.keyword) {
			return field.fieldType, cursorPos, true
		}

		// Also check if there's a space after the keyword and cursor is in that space
		// e.g., "due: " with cursor after the space
		if len(textBefore) > len(field.keyword) {
			lastWord := ""
			parts := strings.Fields(textBefore)
			if len(parts) > 0 {
				lastWord = parts[len(parts)-1]
				if lastWord == field.keyword[:len(field.keyword)-1] {
					// Found keyword without colon in last word, check if there's a colon after
					afterLastWord := textBefore[strings.LastIndex(textBefore, lastWord):]
					if strings.HasPrefix(afterLastWord, field.keyword) {
						return field.fieldType, cursorPos, true
					}
				}
			}
		}
	}

	return "", 0, false
}

// detectProjectFieldContext checks if cursor is positioned after a project field keyword
// Returns (filterPrefix, insertPosition, found) where filterPrefix is what user already typed
func detectProjectFieldContext(input string, cursorPos int) (string, int, bool) {
	// Get text before cursor
	textBefore := input[:cursorPos]

	// Look for project field keywords
	// Support: "project:", "proj:", "pro:"
	projectKeywords := []string{"project:", "proj:", "pro:"}

	for _, keyword := range projectKeywords {
		// Find the keyword in the text
		idx := strings.LastIndex(textBefore, keyword)
		if idx == -1 {
			continue
		}

		// Check if keyword is at a word boundary (start or after space)
		if idx > 0 {
			charBefore := textBefore[idx-1]
			if charBefore != ' ' && charBefore != '\t' {
				continue
			}
		}

		// Extract text after the keyword up to cursor
		afterKeyword := textBefore[idx+len(keyword):]

		// Check if there's only valid project characters (no spaces or other field separators)
		if strings.ContainsAny(afterKeyword, " \t") {
			continue
		}

		// Return the filter prefix (what user already typed) and insert position
		return afterKeyword, idx + len(keyword), true
	}

	return "", 0, false
}

// detectTagFieldContext checks if cursor is positioned after a tag prefix "+"
// Returns (filterPrefix, insertPosition, found) where filterPrefix is what user already typed
func detectTagFieldContext(input string, cursorPos int) (string, int, bool) {
	// Get text before cursor
	textBefore := input[:cursorPos]

	// Find the last "+" in the text
	idx := strings.LastIndex(textBefore, "+")
	if idx == -1 {
		return "", 0, false
	}

	// Check if "+" is at a word boundary (start or after space)
	if idx > 0 {
		charBefore := textBefore[idx-1]
		if charBefore != ' ' && charBefore != '\t' {
			return "", 0, false
		}
	}

	// Extract text after the "+" up to cursor
	afterPlus := textBefore[idx+1:]

	// Check if there's only valid tag characters (no spaces or other separators)
	if strings.ContainsAny(afterPlus, " \t") {
		return "", 0, false
	}

	// Return the filter prefix (what user already typed) and insert position
	return afterPlus, idx + 1, true
}

// activateCalendar activates the calendar picker for date selection
func (m *Model) activateCalendar(fieldType string, insertPos int, inputState AppState) {
	m.calendar = components.NewCalendar(time.Now())
	m.calendarActive = true
	m.calendarFieldType = fieldType
	m.calendarInsertPos = insertPos
	m.calendarInputState = inputState
}

// deactivateCalendar closes the calendar picker
func (m *Model) deactivateCalendar() {
	m.calendarActive = false
	m.calendarFieldType = ""
	m.calendarInsertPos = 0
}

// insertDateFromCalendar inserts the selected date into the appropriate input field
func (m *Model) insertDateFromCalendar(selectedDate time.Time) {
	dateStr := selectedDate.Format("2006-01-02")

	var currentValue string
	var inputComponent *components.Filter

	switch m.calendarInputState {
	case StateModifyInput:
		inputComponent = &m.modifyInput
		currentValue = m.modifyInput.Value()
	case StateNewTaskInput:
		inputComponent = &m.newTaskInput
		currentValue = m.newTaskInput.Value()
	case StateFilterInput:
		inputComponent = &m.filter
		currentValue = m.filter.Value()
	default:
		return
	}

	// Insert date at the saved position
	newValue := currentValue[:m.calendarInsertPos] + dateStr + currentValue[m.calendarInsertPos:]
	inputComponent.SetValue(newValue)

	// Move cursor to after the inserted date
	inputComponent.SetCursor(m.calendarInsertPos + len(dateStr))
}

// detectCompleteDateContext checks if cursor is positioned after a complete date (YYYY-MM-DD)
// Returns (insertPosition, found) where insertPosition is where to insert the time
func detectCompleteDateContext(input string, cursorPos int) (int, bool) {
	// Get text before cursor
	textBefore := input[:cursorPos]

	// Look for date pattern at the end: YYYY-MM-DD
	// The pattern should be preceded by due:/scheduled:/sched:/sch:
	if len(textBefore) < 10 {
		return 0, false
	}

	// Check if the last 10 characters match YYYY-MM-DD pattern
	last10 := textBefore[len(textBefore)-10:]

	// Simple validation: check format DDDD-DD-DD where D is digit
	if len(last10) == 10 &&
		last10[4] == '-' && last10[7] == '-' &&
		isDigit(last10[0]) && isDigit(last10[1]) && isDigit(last10[2]) && isDigit(last10[3]) &&
		isDigit(last10[5]) && isDigit(last10[6]) &&
		isDigit(last10[8]) && isDigit(last10[9]) {

		// Make sure it's preceded by a date field keyword
		if len(textBefore) > 10 {
			before := textBefore[:len(textBefore)-10]
			if strings.HasSuffix(before, "due:") ||
				strings.HasSuffix(before, "scheduled:") ||
				strings.HasSuffix(before, "sched:") ||
				strings.HasSuffix(before, "sch:") {
				return cursorPos, true
			}
		}
	}

	return 0, false
}

// isDigit checks if a byte is an ASCII digit
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// activateTimePicker activates the time picker for time selection
func (m *Model) activateTimePicker(insertPos int, inputState AppState) {
	m.timePicker = components.NewTimePicker()
	m.timePickerActive = true
	m.timePickerInsertPos = insertPos
	m.timePickerInputState = inputState
}

// deactivateTimePicker closes the time picker
func (m *Model) deactivateTimePicker() {
	m.timePickerActive = false
	m.timePickerInsertPos = 0
}

// insertTimeFromPicker inserts the selected time into the appropriate input field
func (m *Model) insertTimeFromPicker() {
	timeStr := "T" + m.timePicker.GetFormattedTime()

	var currentValue string
	var inputComponent *components.Filter

	switch m.timePickerInputState {
	case StateModifyInput:
		inputComponent = &m.modifyInput
		currentValue = m.modifyInput.Value()
	case StateNewTaskInput:
		inputComponent = &m.newTaskInput
		currentValue = m.newTaskInput.Value()
	case StateFilterInput:
		inputComponent = &m.filter
		currentValue = m.filter.Value()
	default:
		return
	}

	// Insert time at the saved position (after the date)
	newValue := currentValue[:m.timePickerInsertPos] + timeStr + currentValue[m.timePickerInsertPos:]
	inputComponent.SetValue(newValue)

	// Move cursor to after the inserted time
	inputComponent.SetCursor(m.timePickerInsertPos + len(timeStr))
}

// activateListPicker activates the list picker for project or tag selection
func (m *Model) activateListPicker(pickerType string, filter string, insertPos int, inputState AppState) {
	var items []string
	var title string

	if pickerType == "project" {
		items = m.availableProjects
		title = "Projects"
	} else if pickerType == "tag" {
		items = m.availableTags
		title = "Tags"
	}

	m.listPicker = components.NewListPicker(title, items, filter)
	m.listPickerActive = true
	m.listPickerType = pickerType
	m.listPickerInsertPos = insertPos
	m.listPickerFilterPrefix = filter
	m.listPickerInputState = inputState
}

// deactivateListPicker closes the list picker
func (m *Model) deactivateListPicker() {
	m.listPickerActive = false
	m.listPickerType = ""
	m.listPickerInsertPos = 0
	m.listPickerFilterPrefix = ""
}

// insertSelectionFromListPicker inserts the selected item into the appropriate input field
func (m *Model) insertSelectionFromListPicker() {
	selectedItem := m.listPicker.SelectedItem()
	if selectedItem == "" {
		return
	}

	var currentValue string
	var inputComponent *components.Filter

	switch m.listPickerInputState {
	case StateModifyInput:
		inputComponent = &m.modifyInput
		currentValue = m.modifyInput.Value()
	case StateNewTaskInput:
		inputComponent = &m.newTaskInput
		currentValue = m.newTaskInput.Value()
	case StateFilterInput:
		inputComponent = &m.filter
		currentValue = m.filter.Value()
	default:
		return
	}

	// Find the insertion position - we need to remove any filter text that was already typed
	// and replace it with the selected item

	// Get the text before and after the insertion point
	beforeInsert := currentValue[:m.listPickerInsertPos]
	afterInsert := currentValue[m.listPickerInsertPos:]

	// Remove the filter prefix from afterInsert if it starts with it
	// This prevents duplication like "+G" + "GA" + "G" = "+GAG"
	if m.listPickerFilterPrefix != "" && strings.HasPrefix(afterInsert, m.listPickerFilterPrefix) {
		afterInsert = afterInsert[len(m.listPickerFilterPrefix):]
	}

	// Build the new value with the selected item
	newValue := beforeInsert + selectedItem + afterInsert
	inputComponent.SetValue(newValue)

	// Move cursor to after the inserted selection
	inputComponent.SetCursor(m.listPickerInsertPos + len(selectedItem))
}

// handleNormalKeys handles keys in normal state
func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	keyPressed := msg.String()

	// Check configured keybindings
	if m.keyMatches(keyPressed, "quit") {
		// Check if auto-sync before quit is enabled
		if m.config.CalendarSync != nil &&
			m.config.CalendarSync.Enabled &&
			m.config.CalendarSync.AutoSyncOnQuit &&
			!m.syncingBeforeQuit {
			// Trigger calendar sync before quitting
			m.syncingBeforeQuit = true
			m.isLoading = true
			m.statusMessage = "Syncing calendar before quit..."
			return m, calendarSyncCmd(m.config, m.service)
		}
		return m, tea.Quit
	}

	if m.keyMatches(keyPressed, "help") {
		m.state = StateHelp
		return m, nil
	}

	// Space key for multi-select (not configurable)
	if keyPressed == " " {
		// Toggle selection on current task
		if !m.inGroupView {
			m.taskList.ToggleSelection()
		}
		return m, nil
	}

	// Escape key for clearing selections/going back (not configurable)
	if keyPressed == "esc" {
		// If in small screen task detail view, go back to task list
		if m.viewMode == ViewModeSmallTaskDetail {
			m.viewMode = ViewModeSmall
			m.updateComponentSizes()
			return m, nil
		}
		// Clear selections if any exist
		if m.taskList.HasSelections() {
			m.taskList.ClearSelection()
			return m, nil
		}
		// Go back to group list if we drilled into a group
		if !m.inGroupView && m.selectedGroup != nil && (m.sections.IsProjectsView() || m.sections.IsTagsView()) {
			m.inGroupView = true
			m.selectedGroup = nil
			// Recompute groups from all tasks and display them
			if m.sections.IsProjectsView() {
				m.groups = core.GroupByProject(m.tasks)
			} else if m.sections.IsTagsView() {
				m.groups = core.GroupByTag(m.tasks)
			}
			m.taskList.SetGroups(m.groups)
			return m, nil
		}
		return m, nil
	}

	if m.keyMatches(keyPressed, "filter") {
		// Activate filter input
		m.state = StateFilterInput
		// Add trailing space to make it easier to extend the filter
		filterValue := m.activeFilter
		if filterValue != "" {
			filterValue += " "
		}
		m.filter.SetValue(filterValue)
		m.updateComponentSizes()
		return m, m.filter.Focus()
	}

	if m.keyMatches(keyPressed, "refresh") {
		m.isLoading = true
		isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"
		return m, loadTasksCmd(m.service, m.activeFilter, isSearchTab)
	}

	// Enter key for sidebar toggle/group drill-down (not configurable)
	if keyPressed == "enter" {
		// If in group view, drill into selected group
		if m.inGroupView && len(m.groups) > 0 {
			// Get the selected group index from task list cursor
			selectedIndex := m.taskList.Cursor()
			if selectedIndex >= 0 && selectedIndex < len(m.groups) {
				m.selectedGroup = &m.groups[selectedIndex]
				m.inGroupView = false
				// Set tasks to the tasks in this group
				m.taskList.SetTasks(m.selectedGroup.Tasks)
				m.updateSidebar()
			}
			return m, nil
		}

		// In small screen mode, Enter shows full-screen task details
		if m.viewMode == ViewModeSmall {
			m.viewMode = ViewModeSmallTaskDetail
			m.updateComponentSizes()
			m.updateSidebar()
			return m, nil
		}

		// Otherwise toggle sidebar (for normal screens)
		if m.viewMode == ViewModeList {
			m.viewMode = ViewModeListWithSidebar
		} else if m.viewMode == ViewModeListWithSidebar {
			m.viewMode = ViewModeList
		}
		m.updateComponentSizes()
		m.updateSidebar()
		return m, nil
	}

	if m.keyMatches(keyPressed, "done") {
		// Mark task(s) done
		selectedTasks := m.taskList.GetSelectedTasks()
		if len(selectedTasks) > 0 {
			m.taskList.ClearSelection()
			return m, markTasksDoneCmd(m.service, selectedTasks)
		}
		return m, nil
	}

	// Start/stop task (not in default config, but 's' is commonly used)
	if keyPressed == "s" {
		// Toggle start/stop on task(s)
		selectedTasks := m.taskList.GetSelectedTasks()
		if len(selectedTasks) > 0 {
			m.taskList.ClearSelection()
			return m, toggleStartStopCmd(m.service, selectedTasks)
		}
		return m, nil
	}

	if m.keyMatches(keyPressed, "delete") {
		// Delete task(s) (with confirmation)
		selectedTasks := m.taskList.GetSelectedTasks()
		if len(selectedTasks) > 0 {
			m.state = StateConfirm
			m.confirmAction = "delete"
			return m, nil
		}
		return m, nil
	}

	if m.keyMatches(keyPressed, "undo") {
		// Undo last operation
		return m, undoCmd(m.service)
	}

	if m.keyMatches(keyPressed, "new") {
		// New task
		m.state = StateNewTaskInput
		m.newTaskInput.SetValue("")
		m.updateComponentSizes()
		return m, m.newTaskInput.Focus()
	}

	if m.keyMatches(keyPressed, "modify") {
		// Modify task(s)
		selectedTasks := m.taskList.GetSelectedTasks()
		if len(selectedTasks) > 0 {
			m.state = StateModifyInput
			m.modifyInput.SetValue("")
			m.updateComponentSizes()
			return m, m.modifyInput.Focus()
		}
		return m, nil
	}

	// Export to markdown (not in default config, but 'M' is commonly used)
	if keyPressed == "M" {
		// Export task(s) to markdown
		selectedTasks := m.taskList.GetSelectedTasks()
		if len(selectedTasks) > 0 {
			m.taskList.ClearSelection()
			return m, exportMarkdownCmd(selectedTasks)
		}
		return m, nil
	}

	if m.keyMatches(keyPressed, "annotate") {
		// Add annotation to task(s)
		selectedTasks := m.taskList.GetSelectedTasks()
		if len(selectedTasks) > 0 {
			m.state = StateAnnotateInput
			m.annotateInput.SetValue("")
			m.updateComponentSizes()
			return m, m.annotateInput.Focus()
		}
		return m, nil
	}

	if m.keyMatches(keyPressed, "edit") {
		// Edit task (suspend TUI)
		selectedTask := m.taskList.SelectedTask()
		if selectedTask != nil {
			return m, editTaskCmd(m.config.TaskBin, m.config.TaskrcPath, selectedTask.UUID)
		}
		return m, nil
	}

	// Section navigation (not configurable - uses tab, h/l, arrows, and number keys)
	if keyPressed == "tab" || keyPressed == "shift+tab" || keyPressed == "h" || keyPressed == "l" ||
		keyPressed == "left" || keyPressed == "right" {
		// Delegate section navigation to sections component
		m.sections, cmd = m.sections.Update(msg)
		return m, cmd
	}

	// Number keys for quick navigation (1-5 for sections, 1-9 for tasks)
	if keyPressed == "1" || keyPressed == "2" || keyPressed == "3" || keyPressed == "4" || keyPressed == "5" {
		// Check if it's a section navigation (1-5 for sections)
		// or task navigation (1-9 for tasks)
		// Section navigation takes priority if there are sections
		sectionCount := len(m.sections.Items)
		if sectionCount > 0 {
			key := keyPressed[0] - '0'
			if int(key) <= sectionCount {
				// It's a section navigation
				m.sections, cmd = m.sections.Update(msg)
				return m, cmd
			}
		}
		// Otherwise, fall through to task list navigation
		m.taskList, cmd = m.taskList.Update(msg)
		m.updateSidebar()
		return m, cmd
	}

	// Number keys 6-9 are only for task navigation
	if keyPressed == "6" || keyPressed == "7" || keyPressed == "8" || keyPressed == "9" {
		m.taskList, cmd = m.taskList.Update(msg)
		m.updateSidebar()
		return m, cmd
	}

	// Check for custom commands (user-configured in config)
	if m.config.TUI != nil && m.config.TUI.CustomCommands != nil {
		if customCmd, exists := m.config.TUI.CustomCommands[keyPressed]; exists {
			// Skip if in group view
			if !m.inGroupView {
				selectedTask := m.taskList.SelectedTask()
				if selectedTask != nil {
					return m, executeCustomCommand(customCmd, selectedTask)
				} else {
					m.statusMessage = "No task selected"
				}
			}
			return m, nil
		}
	}

	// Navigation keys - check both configured keys and arrow keys
	if m.keyMatches(keyPressed, "up") || m.keyMatches(keyPressed, "down") ||
		m.keyMatches(keyPressed, "first") || m.keyMatches(keyPressed, "last") ||
		m.keyMatches(keyPressed, "page_up") || m.keyMatches(keyPressed, "page_down") ||
		keyPressed == "up" || keyPressed == "down" {
		// Delegate navigation to task list component
		m.taskList, cmd = m.taskList.Update(msg)
		m.updateSidebar()
		return m, cmd
	}

	// If sidebar is visible, check for sidebar scrolling keys (not configurable)
	if m.viewMode == ViewModeListWithSidebar {
		if keyPressed == "ctrl+d" || keyPressed == "ctrl+u" || keyPressed == "ctrl+f" ||
			keyPressed == "ctrl+b" || keyPressed == "J" || keyPressed == "K" ||
			keyPressed == "pgdown" || keyPressed == "pgup" {
			m.sidebar, cmd = m.sidebar.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// updateComponentSizes updates the sizes of all components based on terminal dimensions
func (m *Model) updateComponentSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	// Detect small screen and auto-switch layout mode
	// Don't auto-switch if we're in task detail view
	if m.width < 80 {
		if m.viewMode != ViewModeSmallTaskDetail {
			m.viewMode = ViewModeSmall
		}
	} else {
		// Switch back to normal mode if we were in small screen mode
		if m.viewMode == ViewModeSmall {
			m.viewMode = ViewModeList
		} else if m.viewMode == ViewModeSmallTaskDetail {
			// If we're in task detail on small screen but screen got bigger, go back to list
			m.viewMode = ViewModeList
		}
	}

	// Update sections component width
	m.sections.SetSize(m.width)

	// Update help component size
	m.help.SetSize(m.width, m.height)

	// Calculate available height (subtract sections bar, footer, bottom border, and spacing)
	// Note: The actual space calculation must match view.go's trimming logic
	// - Sections bar: 1 line
	// - Footer: 3 lines (1 pad top + 1 content + 1 pad bottom)
	// - Bottom border: 1 line (added in renderTaskListWithComponents)
	// - Additional spacing/margins: 2 lines (empirically determined to prevent trimming)
	availableHeight := m.height - 7 // sections(1) + footer(3) + bottom border(1) + spacing(2)

	// If in input mode, subtract input prompt area (2 lines: separator + input)
	if m.state == StateFilterInput || m.state == StateModifyInput ||
		m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		availableHeight -= 2
	}

	if m.viewMode == ViewModeListWithSidebar {
		// Split view: task list and sidebar
		// Calculate sidebar width from config percentage (default 33%)
		sidebarWidthPercent := m.config.TUI.SidebarWidth
		if sidebarWidthPercent <= 0 || sidebarWidthPercent > 100 {
			sidebarWidthPercent = 33 // Default to 33% if invalid
		}
		sidebarWidth := (m.width * sidebarWidthPercent) / 100
		if sidebarWidth < 30 {
			sidebarWidth = 30
		}
		taskListWidth := m.width - sidebarWidth

		m.taskList.SetSize(taskListWidth, availableHeight)
		m.sidebar.SetSize(sidebarWidth, availableHeight)
	} else if m.viewMode == ViewModeSmallTaskDetail {
		// Full screen task detail view (using sidebar component)
		m.sidebar.SetSize(m.width, availableHeight)
	} else {
		// Full width task list (ViewModeList or ViewModeSmall)
		m.taskList.SetSize(m.width, availableHeight)
	}

	// Update input component widths if in input mode
	if m.state == StateFilterInput || m.state == StateModifyInput ||
		m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		// Calculate input width: available width - padding - prompt width - hint width - spacing
		availableWidth := m.width - 2 // Account for padding
		promptWidth := 12             // Approximate max prompt width ("New Task: " is longest)
		hintWidth := 35               // Approximate hint width "(Enter to apply, Esc to cancel)"
		spacing := 2

		inputWidth := availableWidth - promptWidth - hintWidth - spacing
		if inputWidth < 20 {
			inputWidth = 20 // Minimum width
		}

		// Set width for all input components
		m.filter.SetWidth(inputWidth)
		m.modifyInput.SetWidth(inputWidth)
		m.annotateInput.SetWidth(inputWidth)
		m.newTaskInput.SetWidth(inputWidth)
	}
}

// updateSidebar updates the sidebar with the currently selected task
func (m *Model) updateSidebar() {
	selectedTask := m.taskList.SelectedTask()
	m.sidebar.SetTask(selectedTask)
}

// handleFilterKeys handles keys in filter input state
func (m Model) handleFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.state = StateNormal
		m.filter.Blur()
		m.filter.ResetHistoryNavigation()
		m.updateComponentSizes()
		return m, nil

	case "enter":
		// Apply the filter
		filterText := m.filter.Value()

		// Add to history before clearing
		m.filter.AddToHistory(filterText)
		m.filter.ResetHistoryNavigation()

		m.state = StateNormal
		m.filter.Blur()
		m.activeFilter = filterText
		m.isLoading = true
		m.updateComponentSizes()

		// Check if we're in the Search tab
		isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"

		// Save the filter if we're in the Search tab (for session persistence)
		if isSearchTab {
			m.searchTabFilter = filterText
		}

		// Load tasks with new filter
		return m, loadTasksCmd(m.service, filterText, isSearchTab)

	case "up":
		// Navigate to previous command in history
		m.filter.NavigateHistoryUp()
		return m, nil

	case "down":
		// Navigate to next command in history
		m.filter.NavigateHistoryDown()
		return m, nil

	case "tab":
		currentValue := m.filter.Value()
		cursorPos := m.filter.CursorPosition()

		// Check for project field context
		if filter, insertPos, found := detectProjectFieldContext(currentValue, cursorPos); found {
			// Activate list picker for projects
			m.activateListPicker("project", filter, insertPos, StateFilterInput)
			return m, nil
		}

		// Check for tag field context
		if filter, insertPos, found := detectTagFieldContext(currentValue, cursorPos); found {
			// Activate list picker for tags
			m.activateListPicker("tag", filter, insertPos, StateFilterInput)
			return m, nil
		}

		// Check if cursor is after a complete date (for time picker)
		if insertPos, found := detectCompleteDateContext(currentValue, cursorPos); found {
			// Activate time picker
			m.activateTimePicker(insertPos, StateFilterInput)
			return m, nil
		}

		// Check if cursor is after a date field keyword (for calendar)
		if fieldType, insertPos, found := detectDateFieldContext(currentValue, cursorPos); found {
			// Activate calendar picker
			m.activateCalendar(fieldType, insertPos, StateFilterInput)
			return m, nil
		}

		// If not in any field context, let the input handle it normally
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd

	default:
		// When user types, reset history navigation
		m.filter.ResetHistoryNavigation()
		// Delegate to filter component for text input
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}
}

// handleHelpKeys handles keys in help state
func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc", "q", "?":
		m.state = StateNormal
		return m, nil
	default:
		// Delegate to help component for scrolling
		m.help, cmd = m.help.Update(msg)
		return m, cmd
	}
}

// handleConfirmKeys handles keys in confirm state
func (m Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		m.state = StateNormal
		m.confirmAction = ""
		return m, nil
	case "y", "Y":
		// Execute the confirmed action
		m.state = StateNormal
		selectedTasks := m.taskList.GetSelectedTasks()

		if m.confirmAction == "delete" && len(selectedTasks) > 0 {
			m.confirmAction = ""
			m.taskList.ClearSelection()
			return m, deleteTasksCmd(m.service, selectedTasks)
		}

		m.confirmAction = ""
		return m, nil
	}
	return m, nil
}

// handleModifyKeys handles keys in modify input state
func (m Model) handleModifyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.state = StateNormal
		m.modifyInput.Blur()
		m.updateComponentSizes()
		return m, nil

	case "enter":
		// Apply modifications
		modifications := m.modifyInput.Value()
		selectedTasks := m.taskList.GetSelectedTasks()
		m.state = StateNormal
		m.modifyInput.Blur()
		m.updateComponentSizes()

		if len(selectedTasks) > 0 && modifications != "" {
			m.taskList.ClearSelection()
			return m, modifyTasksCmd(m.service, selectedTasks, modifications)
		}
		return m, nil

	case "tab":
		currentValue := m.modifyInput.Value()
		cursorPos := m.modifyInput.CursorPosition()

		// Check for project field context
		if filter, insertPos, found := detectProjectFieldContext(currentValue, cursorPos); found {
			// Activate list picker for projects
			m.activateListPicker("project", filter, insertPos, StateModifyInput)
			return m, nil
		}

		// Check for tag field context
		if filter, insertPos, found := detectTagFieldContext(currentValue, cursorPos); found {
			// Activate list picker for tags
			m.activateListPicker("tag", filter, insertPos, StateModifyInput)
			return m, nil
		}

		// Check if cursor is after a complete date (for time picker)
		if insertPos, found := detectCompleteDateContext(currentValue, cursorPos); found {
			// Activate time picker
			m.activateTimePicker(insertPos, StateModifyInput)
			return m, nil
		}

		// Check if cursor is after a date field keyword (for calendar)
		if fieldType, insertPos, found := detectDateFieldContext(currentValue, cursorPos); found {
			// Activate calendar picker
			m.activateCalendar(fieldType, insertPos, StateModifyInput)
			return m, nil
		}

		// If not in any field context, let the input handle it normally
		m.modifyInput, cmd = m.modifyInput.Update(msg)
		return m, cmd

	default:
		// Delegate to input component for text input
		m.modifyInput, cmd = m.modifyInput.Update(msg)
		return m, cmd
	}
}

// handleAnnotateKeys handles keys in annotate input state
func (m Model) handleAnnotateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.state = StateNormal
		m.annotateInput.Blur()
		m.updateComponentSizes()
		return m, nil

	case "enter":
		// Add annotation
		text := m.annotateInput.Value()
		selectedTasks := m.taskList.GetSelectedTasks()
		m.state = StateNormal
		m.annotateInput.Blur()
		m.updateComponentSizes()

		if len(selectedTasks) > 0 && text != "" {
			m.taskList.ClearSelection()
			return m, annotateTasksCmd(m.service, selectedTasks, text)
		}
		return m, nil

	default:
		// Delegate to input component for text input
		m.annotateInput, cmd = m.annotateInput.Update(msg)
		return m, cmd
	}
}

// handleNewTaskKeys handles keys in new task input state
func (m Model) handleNewTaskKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.state = StateNormal
		m.newTaskInput.Blur()
		m.updateComponentSizes()
		return m, nil

	case "enter":
		// Add new task
		description := m.newTaskInput.Value()
		m.state = StateNormal
		m.newTaskInput.Blur()
		m.updateComponentSizes()

		if description != "" {
			return m, addTaskCmd(m.service, description)
		}
		return m, nil

	case "tab":
		currentValue := m.newTaskInput.Value()
		cursorPos := m.newTaskInput.CursorPosition()

		// Check for project field context
		if filter, insertPos, found := detectProjectFieldContext(currentValue, cursorPos); found {
			// Activate list picker for projects
			m.activateListPicker("project", filter, insertPos, StateNewTaskInput)
			return m, nil
		}

		// Check for tag field context
		if filter, insertPos, found := detectTagFieldContext(currentValue, cursorPos); found {
			// Activate list picker for tags
			m.activateListPicker("tag", filter, insertPos, StateNewTaskInput)
			return m, nil
		}

		// Check if cursor is after a complete date (for time picker)
		if insertPos, found := detectCompleteDateContext(currentValue, cursorPos); found {
			// Activate time picker
			m.activateTimePicker(insertPos, StateNewTaskInput)
			return m, nil
		}

		// Check if cursor is after a date field keyword (for calendar)
		if fieldType, insertPos, found := detectDateFieldContext(currentValue, cursorPos); found {
			// Activate calendar picker
			m.activateCalendar(fieldType, insertPos, StateNewTaskInput)
			return m, nil
		}

		// If not in any field context, let the input handle it normally
		m.newTaskInput, cmd = m.newTaskInput.Update(msg)
		return m, cmd

	default:
		// Delegate to input component for text input
		m.newTaskInput, cmd = m.newTaskInput.Update(msg)
		return m, cmd
	}
}

// loadTasksCmd creates a command to load tasks asynchronously
func loadTasksCmd(service core.TaskService, filter string, isSearchTab bool) tea.Cmd {
	return func() tea.Msg {
		// If filter is empty, return empty task list
		// This shows nothing until user enters a search query
		if filter == "" {
			return TasksLoadedMsg{
				Tasks: []core.Task{},
				Err:   nil,
			}
		}

		// In Search tab, we want to search ALL tasks in the database by default
		// unless the user explicitly filters by status
		actualFilter := filter
		if isSearchTab {
			// Check if user already specified a status filter
			// Common patterns: "status:", "status.not:", "status.is:"
			hasStatusFilter := strings.Contains(filter, "status:")

			// If no status filter specified, search across ALL tasks
			// By using "status.any:" we tell taskwarrior to search all statuses
			if !hasStatusFilter {
				// Prepend status.any: to search all tasks regardless of status
				// This searches pending, completed, deleted, waiting, and recurring tasks
				actualFilter = "status.any: " + filter
			}
		}

		tasks, err := service.Export(actualFilter)
		return TasksLoadedMsg{
			Tasks: tasks,
			Err:   err,
		}
	}
}

// loadProjectSummaryCmd creates a command to load project summaries asynchronously
func loadProjectSummaryCmd(service core.TaskService) tea.Cmd {
	return func() tea.Msg {
		summaries, err := service.GetProjectSummary()
		return ProjectSummaryLoadedMsg{
			Summaries: summaries,
			Err:       err,
		}
	}
}

// markTaskDoneCmd creates a command to mark a task as done
func markTaskDoneCmd(service core.TaskService, uuid string) tea.Cmd {
	return func() tea.Msg {
		err := service.Done(uuid)
		return TaskModifiedMsg{
			Err: err,
		}
	}
}

// deleteTaskCmd creates a command to delete a task
func deleteTaskCmd(service core.TaskService, uuid string) tea.Cmd {
	return func() tea.Msg {
		err := service.Delete(uuid)
		return TaskModifiedMsg{
			Err: err,
		}
	}
}

// undoCmd creates a command to undo the last operation
func undoCmd(service core.TaskService) tea.Cmd {
	return func() tea.Msg {
		err := service.Undo()
		return TaskModifiedMsg{
			Err: err,
		}
	}
}

// modifyTaskCmd creates a command to modify a task
func modifyTaskCmd(service core.TaskService, uuid, modifications string) tea.Cmd {
	return func() tea.Msg {
		err := service.Modify(uuid, modifications)
		return TaskModifiedMsg{
			Err: err,
		}
	}
}

// annotateTaskCmd creates a command to add an annotation to a task
func annotateTaskCmd(service core.TaskService, uuid, text string) tea.Cmd {
	return func() tea.Msg {
		err := service.Annotate(uuid, text)
		return TaskModifiedMsg{
			Err: err,
		}
	}
}

// addTaskCmd creates a command to add a new task
func addTaskCmd(service core.TaskService, description string) tea.Cmd {
	return func() tea.Msg {
		_, err := service.Add(description)
		if err != nil {
			return TaskModifiedMsg{Err: err}
		}
		// Return success
		return TaskModifiedMsg{Err: nil}
	}
}

// editTaskCmd creates a command to edit a task (suspends TUI)
func editTaskCmd(taskBin, taskrcPath, uuid string) tea.Cmd {
	c := exec.Command(taskBin, uuid, "edit")
	if taskrcPath != "" {
		c.Env = append(os.Environ(), fmt.Sprintf("TASKRC=%s", taskrcPath))
	}

	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return TaskModifiedMsg{Err: err}
		}
		// Return success - will trigger refresh
		return TaskModifiedMsg{Err: nil}
	})
}

// startTaskCmd creates a command to start a task
func startTaskCmd(service core.TaskService, uuid string) tea.Cmd {
	return func() tea.Msg {
		err := service.Start(uuid)
		return TaskModifiedMsg{
			Err: err,
		}
	}
}

// stopTaskCmd creates a command to stop a task
func stopTaskCmd(service core.TaskService, uuid string) tea.Cmd {
	return func() tea.Msg {
		err := service.Stop(uuid)
		return TaskModifiedMsg{
			Err: err,
		}
	}
}

// markTasksDoneCmd creates a command to mark multiple tasks as done
func markTasksDoneCmd(service core.TaskService, tasks []core.Task) tea.Cmd {
	return func() tea.Msg {
		var firstErr error
		for _, task := range tasks {
			err := service.Done(task.UUID)
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return TaskModifiedMsg{
			Err: firstErr,
		}
	}
}

// deleteTasksCmd creates a command to delete multiple tasks
func deleteTasksCmd(service core.TaskService, tasks []core.Task) tea.Cmd {
	return func() tea.Msg {
		var firstErr error
		for _, task := range tasks {
			err := service.Delete(task.UUID)
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return TaskModifiedMsg{
			Err: firstErr,
		}
	}
}

// modifyTasksCmd creates a command to modify multiple tasks
func modifyTasksCmd(service core.TaskService, tasks []core.Task, modifications string) tea.Cmd {
	return func() tea.Msg {
		var firstErr error
		for _, task := range tasks {
			err := service.Modify(task.UUID, modifications)
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return TaskModifiedMsg{
			Err: firstErr,
		}
	}
}

// annotateTasksCmd creates a command to add an annotation to multiple tasks
func annotateTasksCmd(service core.TaskService, tasks []core.Task, text string) tea.Cmd {
	return func() tea.Msg {
		var firstErr error
		for _, task := range tasks {
			err := service.Annotate(task.UUID, text)
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return TaskModifiedMsg{
			Err: firstErr,
		}
	}
}

// toggleStartStopCmd creates a command to toggle start/stop on multiple tasks
func toggleStartStopCmd(service core.TaskService, tasks []core.Task) tea.Cmd {
	return func() tea.Msg {
		var firstErr error
		for _, task := range tasks {
			var err error
			// If task is started (has Start field), stop it; otherwise start it
			if task.Start != nil {
				err = service.Stop(task.UUID)
			} else {
				err = service.Start(task.UUID)
			}
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return TaskModifiedMsg{
			Err: firstErr,
		}
	}
}

// exportMarkdownCmd exports task(s) to markdown format and copies to clipboard
func exportMarkdownCmd(tasks []core.Task) tea.Cmd {
	return func() tea.Msg {
		var markdowns []string
		for _, task := range tasks {
			markdowns = append(markdowns, task.ToMarkdown())
		}
		markdown := strings.Join(markdowns, "\n")

		// Try to copy to clipboard
		err := clipboard.WriteAll(markdown)

		if err != nil {
			return StatusMsg{
				Message: "Failed to copy to clipboard: " + markdown,
				IsError: true,
			}
		}

		return StatusMsg{
			Message: "Task exported to clipboard as markdown ✓",
			IsError: false,
		}
	}
}

// calendarSyncCmd creates a command to sync tasks to Google Calendar
func calendarSyncCmd(cfg *config.Config, service core.TaskService) tea.Cmd {
	return func() tea.Msg {
		// Validate calendar configuration
		if cfg.CalendarSync == nil {
			return CalendarSyncCompletedMsg{
				Err: fmt.Errorf("calendar sync not configured"),
			}
		}

		calendarName := cfg.CalendarSync.CalendarName
		taskFilter := cfg.CalendarSync.TaskFilter
		credentialsPath := cfg.CalendarSync.CredentialsPath
		tokenPath := cfg.CalendarSync.TokenPath

		// Validate required fields
		if calendarName == "" {
			return CalendarSyncCompletedMsg{
				Err: fmt.Errorf("calendar name is not configured"),
			}
		}
		if taskFilter == "" {
			return CalendarSyncCompletedMsg{
				Err: fmt.Errorf("task filter is not configured"),
			}
		}

		// We need access to the taskwarrior client to create the calendar sync client
		// Since the service interface doesn't expose the underlying client,
		// we need to recreate it from the config
		// This is necessary because the calendar sync requires the taskwarrior.Client type
		taskClient, err := createTaskClient(cfg)
		if err != nil {
			return CalendarSyncCompletedMsg{
				Err: fmt.Errorf("failed to create task client: %w", err),
			}
		}

		// Perform the calendar sync
		result, err := performCalendarSync(taskClient, credentialsPath, tokenPath, calendarName, taskFilter)
		return CalendarSyncCompletedMsg{
			Result: result,
			Err:    err,
		}
	}
}

// Helper function to create taskwarrior client from config
func createTaskClient(cfg *config.Config) (*taskwarrior.Client, error) {
	return taskwarrior.NewClient(cfg.TaskBin, cfg.TaskrcPath)
}

// Helper function to perform calendar sync
func performCalendarSync(taskClient *taskwarrior.Client, credentialsPath, tokenPath, calendarName, taskFilter string) (*calendar.SyncResult, error) {
	ctx := context.Background()

	// Create sync client
	syncClient, err := calendar.NewSyncClient(ctx, taskClient, credentialsPath, tokenPath, calendarName, taskFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync client: %w", err)
	}

	// Perform sync
	result, err := syncClient.Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	return result, nil
}

// extractUniqueProjects extracts all unique projects from tasks
func extractUniqueProjects(tasks []core.Task) []string {
	projectMap := make(map[string]bool)
	for _, task := range tasks {
		if task.Project != "" {
			projectMap[task.Project] = true
		}
	}

	projects := make([]string, 0, len(projectMap))
	for project := range projectMap {
		projects = append(projects, project)
	}

	// Sort alphabetically for consistent ordering
	// Using a simple bubble sort to avoid importing sort package
	for i := 0; i < len(projects); i++ {
		for j := i + 1; j < len(projects); j++ {
			if projects[i] > projects[j] {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
	}

	return projects
}

// extractUniqueTags extracts all unique tags from tasks
func extractUniqueTags(tasks []core.Task) []string {
	tagMap := make(map[string]bool)
	for _, task := range tasks {
		for _, tag := range task.Tags {
			tagMap[tag] = true
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	// Sort alphabetically for consistent ordering
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i] > tags[j] {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	return tags
}

// updateAvailableProjectsAndTags updates the cached lists of available projects and tags
// This should be called with ALL tasks, not just filtered ones
func (m *Model) updateAvailableProjectsAndTags() {
	m.availableProjects = extractUniqueProjects(m.tasks)
	m.availableTags = extractUniqueTags(m.tasks)
}

// loadAllProjectsAndTagsCmd creates a command to load all projects and tags for autocompletion
// This loads from ALL tasks (status:pending) to ensure complete autocomplete lists
func loadAllProjectsAndTagsCmd(service core.TaskService) tea.Cmd {
	return func() tea.Msg {
		// Get all pending tasks to extract projects/tags
		tasks, err := service.Export("status:pending")
		if err != nil {
			return AutocompleteDataLoadedMsg{Err: err}
		}

		projects := extractUniqueProjects(tasks)
		tags := extractUniqueTags(tasks)

		return AutocompleteDataLoadedMsg{
			Projects: projects,
			Tags:     tags,
			Err:      nil,
		}
	}
}

// expandCommandTemplate expands a command template with task data
// Template format: "xdg-open {{.url}}" or "echo {{.description}} {{.project}}"
// Supports any task field accessible via GetProperty()
func expandCommandTemplate(template string, task *core.Task) (string, error) {
	if task == nil {
		return "", fmt.Errorf("no task selected")
	}

	result := template

	// Find all {{.field}} patterns
	start := 0
	for {
		// Find opening {{
		openIdx := strings.Index(result[start:], "{{.")
		if openIdx == -1 {
			break
		}
		openIdx += start

		// Find closing }}
		closeIdx := strings.Index(result[openIdx:], "}}")
		if closeIdx == -1 {
			return "", fmt.Errorf("unclosed template placeholder in command")
		}
		closeIdx += openIdx

		// Extract field name (without {{. and }})
		fieldName := result[openIdx+3 : closeIdx]

		// Get field value from task
		value, exists := task.GetProperty(fieldName)
		if !exists {
			return "", fmt.Errorf("field '%s' not found in task", fieldName)
		}

		// Replace {{.field}} with value
		placeholder := result[openIdx : closeIdx+2]
		result = strings.Replace(result, placeholder, value, 1)

		// Continue searching after the replacement
		start = openIdx + len(value)
	}

	return result, nil
}

// executeCustomCommand executes a custom command with template expansion
func executeCustomCommand(cmd config.CustomCommand, task *core.Task) tea.Cmd {
	return func() tea.Msg {
		// Expand template
		expandedCmd, err := expandCommandTemplate(cmd.Command, task)
		if err != nil {
			return StatusMsg{
				Message: fmt.Sprintf("Command expansion failed: %s", err.Error()),
				IsError: true,
			}
		}

		// Parse command into parts (handle quoted arguments properly)
		parts, err := parseCommandLine(expandedCmd)
		if err != nil {
			return StatusMsg{
				Message: fmt.Sprintf("Command parsing failed: %s", err.Error()),
				IsError: true,
			}
		}

		if len(parts) == 0 {
			return StatusMsg{
				Message: "Empty command after expansion",
				IsError: true,
			}
		}

		// Execute command
		execCmd := exec.Command(parts[0], parts[1:]...)
		err = execCmd.Start()
		if err != nil {
			return StatusMsg{
				Message: fmt.Sprintf("Command execution failed: %s", err.Error()),
				IsError: true,
			}
		}

		return StatusMsg{
			Message: fmt.Sprintf("Executed: %s", cmd.Name),
			IsError: false,
		}
	}
}

// parseCommandLine parses a command line string into parts, respecting quotes
// Simple parser that handles: cmd arg1 "arg with spaces" arg3
func parseCommandLine(cmdLine string) ([]string, error) {
	var parts []string
	var current strings.Builder
	inQuote := false
	escaped := false

	for i, r := range cmdLine {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '"':
			inQuote = !inQuote
		case r == ' ' && !inQuote:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}

		// Check for unterminated quote at end
		if i == len(cmdLine)-1 && inQuote {
			return nil, fmt.Errorf("unterminated quote in command")
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts, nil
}
