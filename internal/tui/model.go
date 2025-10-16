package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
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
	default:
		return "unknown"
	}
}

// Model represents the main TUI application model
type Model struct {
	// Core dependencies
	service core.TaskService
	config  *config.Config

	// Task data
	tasks          []core.Task
	selectedIndex  int
	currentSection *core.Section

	// UI state
	viewMode ViewMode
	state    AppState

	// Current filter
	activeFilter string

	// Status and error messages
	statusMessage string
	errorMessage  string

	// Terminal dimensions
	width  int
	height int

	// TODO: Components will be added later
	// taskList  components.TaskList
	// sidebar   components.Sidebar
	// sections  components.Sections
	// filter    components.Filter
	// help      components.Help
}

// NewModel creates a new TUI model
func NewModel(service core.TaskService, cfg *config.Config) Model {
	sections := core.DefaultSections()

	return Model{
		service:        service,
		config:         cfg,
		tasks:          []core.Task{},
		selectedIndex:  0,
		viewMode:       ViewModeList,
		state:          StateNormal,
		currentSection: &sections[0], // Start with first section (Next)
		activeFilter:   sections[0].Filter,
		statusMessage:  "",
		errorMessage:   "",
	}
}

// Init initializes the model and returns the initial command
func (m Model) Init() tea.Cmd {
	// Load tasks with the current section's filter
	return loadTasksCmd(m.service, m.activeFilter)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TasksLoadedMsg:
		if msg.Err != nil {
			m.errorMessage = "Failed to load tasks: " + msg.Err.Error()
			return m, nil
		}
		m.tasks = msg.Tasks
		m.errorMessage = ""
		// Reset selection if out of bounds
		if m.selectedIndex >= len(m.tasks) {
			m.selectedIndex = 0
		}
		return m, nil

	case TaskModifiedMsg:
		if msg.Err != nil {
			m.errorMessage = "Task operation failed: " + msg.Err.Error()
			return m, nil
		}
		m.statusMessage = "Task updated"
		// Refresh tasks
		return m, loadTasksCmd(m.service, m.activeFilter)

	case ErrorMsg:
		m.errorMessage = msg.Err.Error()
		return m, nil

	case RefreshMsg:
		return m, loadTasksCmd(m.service, m.activeFilter)

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
	}

	return m, nil
}

// handleNormalKeys handles keys in normal state
func (m Model) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "?":
		m.state = StateHelp
		return m, nil

	case "r":
		return m, loadTasksCmd(m.service, m.activeFilter)

	case "j", "down":
		if m.selectedIndex < len(m.tasks)-1 {
			m.selectedIndex++
		}
		return m, nil

	case "k", "up":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
		return m, nil

	case "g":
		m.selectedIndex = 0
		return m, nil

	case "G":
		if len(m.tasks) > 0 {
			m.selectedIndex = len(m.tasks) - 1
		}
		return m, nil

	case "tab":
		// Toggle sidebar
		if m.viewMode == ViewModeList {
			m.viewMode = ViewModeListWithSidebar
		} else {
			m.viewMode = ViewModeList
		}
		return m, nil
	}

	return m, nil
}

// handleFilterKeys handles keys in filter input state
func (m Model) handleFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateNormal
		return m, nil
	case "enter":
		// Apply filter (will be implemented with filter component)
		m.state = StateNormal
		return m, nil
	}
	return m, nil
}

// handleHelpKeys handles keys in help state
func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		m.state = StateNormal
		return m, nil
	}
	return m, nil
}

// handleConfirmKeys handles keys in confirm state
func (m Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		m.state = StateNormal
		return m, nil
	case "y", "Y":
		// Confirm action (specific action will be tracked separately)
		m.state = StateNormal
		return m, nil
	}
	return m, nil
}

// loadTasksCmd creates a command to load tasks asynchronously
func loadTasksCmd(service core.TaskService, filter string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := service.Export(filter)
		return TasksLoadedMsg{
			Tasks: tasks,
			Err:   err,
		}
	}
}
