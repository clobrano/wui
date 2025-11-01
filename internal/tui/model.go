package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
	"github.com/clobrano/wui/internal/tui/components"
)

// ViewMode represents the view layout mode
type ViewMode int

const (
	// ViewModeList shows only the task list
	ViewModeList ViewMode = iota
	// ViewModeListWithSidebar shows task list with detail sidebar
	ViewModeListWithSidebar
)

// String returns the string representation of ViewMode
func (v ViewMode) String() string {
	switch v {
	case ViewModeList:
		return "list"
	case ViewModeListWithSidebar:
		return "list_with_sidebar"
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
	tasks          []core.Task
	currentSection *core.Section

	// Grouping state (for Projects/Tags sections)
	groups        []core.TaskGroup // Current groups (when in group list view)
	selectedGroup *core.TaskGroup   // Selected group (when drilling into a group)
	inGroupView   bool              // true = showing group list, false = showing tasks

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

	// Confirm action tracking
	confirmAction string // "delete", "done", etc.
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

	// Create help component with keybindings from config
	var helpComponent components.Help
	if cfg.TUI != nil && cfg.TUI.Keybindings != nil {
		helpComponent = components.NewHelpWithKeybindings(80, 24, components.DefaultHelpStyles(), cfg.TUI.Keybindings)
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
		taskList:       taskList,
		sidebar:        components.NewSidebar(40, 24, styles.ToSidebarStyles()),       // Initial size, will be updated
		filter:         components.NewFilter(),
		modifyInput:    components.NewFilter(),
		annotateInput:  components.NewFilter(),
		newTaskInput:   components.NewFilter(),
		sections:       components.NewSectionsWithIndex(allSections, 80, styles.ToSectionsStyles(), initialSectionIndex), // Initial size, will be updated
		help:           helpComponent,         // Initial size, will be updated
		confirmAction:  "",
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

	return loadTasksCmd(m.service, filterToUse, isSearchTab)
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

		// Update sidebar with all tasks for dependency lookups
		m.sidebar.SetAllTasks(m.tasks)

		// If in Projects or Tags view and showing group list, compute groups
		if m.inGroupView {
			if m.sections.IsProjectsView() {
				m.groups = core.GroupByProject(m.tasks)
			} else if m.sections.IsTagsView() {
				m.groups = core.GroupByTag(m.tasks)
			}
			// Show group list in the task list component
			m.taskList.SetGroups(m.groups)
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

	case TaskModifiedMsg:
		if msg.Err != nil {
			m.errorMessage = "Task operation failed: " + msg.Err.Error()
			m.isLoading = false
			return m, nil
		}
		m.errorMessage = "" // Clear any previous error
		m.statusMessage = "Task updated successfully"
		m.isLoading = true
		// Refresh tasks
		isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"
		return m, loadTasksCmd(m.service, m.activeFilter, isSearchTab)

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

// handleNormalKeys handles keys in normal state
func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	keyPressed := msg.String()

	// Check configured keybindings
	if m.keyMatches(keyPressed, "quit") {
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

		// Otherwise toggle sidebar
		if m.viewMode == ViewModeList {
			m.viewMode = ViewModeListWithSidebar
		} else {
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

	// Update sections component width
	m.sections.SetSize(m.width)

	// Update help component size
	m.help.SetSize(m.width, m.height)

	// Calculate available height (subtract sections bar and footer with padding)
	// Footer has .Padding(1, 1) which adds 1 line top + 1 line bottom = 2 lines padding + 1 content = 3 total
	availableHeight := m.height - 4 // sections(1) + footer(3: 1 pad top + 1 content + 1 pad bottom)

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
	} else {
		// Full width task list
		m.taskList.SetSize(m.width, availableHeight)
	}

	// Update input component widths if in input mode
	if m.state == StateFilterInput || m.state == StateModifyInput ||
	   m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		// Calculate input width: available width - padding - prompt width - hint width - spacing
		availableWidth := m.width - 2 // Account for padding
		promptWidth := 12 // Approximate max prompt width ("New Task: " is longest)
		hintWidth := 35   // Approximate hint width "(Enter to apply, Esc to cancel)"
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
		m.updateComponentSizes()
		return m, nil

	case "enter":
		// Apply the filter
		filterText := m.filter.Value()
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

	default:
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
