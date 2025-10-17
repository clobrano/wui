package tui

import (
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

	// Components
	taskList components.TaskList
	sidebar  components.Sidebar
	filter   components.Filter
	// sections  components.Sections
	// help      components.Help
}

// NewModel creates a new TUI model
func NewModel(service core.TaskService, cfg *config.Config) Model {
	sections := core.DefaultSections()

	return Model{
		service:        service,
		config:         cfg,
		tasks:          []core.Task{},
		viewMode:       ViewModeList,
		state:          StateNormal,
		currentSection: &sections[0], // Start with first section (Next)
		activeFilter:   sections[0].Filter,
		statusMessage:  "",
		errorMessage:   "",
		taskList:       components.NewTaskList(80, 24), // Initial size, will be updated
		sidebar:        components.NewSidebar(40, 24),  // Initial size, will be updated
		filter:         components.NewFilter(),
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

		// Update component sizes
		m.updateComponentSizes()
		return m, nil

	case TasksLoadedMsg:
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

		// Update task list component
		m.taskList.SetTasks(m.tasks)

		// Update sidebar with selected task
		m.updateSidebar()

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
	var cmd tea.Cmd

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "?":
		m.state = StateHelp
		return m, nil

	case "/":
		// Activate filter input
		m.state = StateFilterInput
		m.filter.SetValue(m.activeFilter)
		return m, m.filter.Focus()

	case "r":
		return m, loadTasksCmd(m.service, m.activeFilter)

	case "tab":
		// Toggle sidebar
		if m.viewMode == ViewModeList {
			m.viewMode = ViewModeListWithSidebar
		} else {
			m.viewMode = ViewModeList
		}
		m.updateComponentSizes()
		return m, nil

	case "j", "down", "k", "up", "g", "G", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Delegate navigation to task list component
		m.taskList, cmd = m.taskList.Update(msg)
		m.updateSidebar()
		return m, cmd
	}

	// If sidebar is visible, check for sidebar scrolling keys
	if m.viewMode == ViewModeListWithSidebar {
		switch msg.String() {
		case "ctrl+d", "ctrl+u", "ctrl+f", "ctrl+b":
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

	// Calculate available height (subtract header and footer)
	availableHeight := m.height - 3 // header(1) + footer(2)

	if m.viewMode == ViewModeListWithSidebar {
		// Split view: task list and sidebar
		sidebarWidth := m.width / 3
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
		return m, nil

	case "enter":
		// Apply the filter
		filterText := m.filter.Value()
		m.state = StateNormal
		m.filter.Blur()
		m.activeFilter = filterText

		// Load tasks with new filter
		return m, loadTasksCmd(m.service, filterText)

	default:
		// Delegate to filter component for text input
		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}
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
