package tui

import (
	"fmt"
	"os"
	"os/exec"

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
	taskList    components.TaskList
	sidebar     components.Sidebar
	filter      components.Filter
	modifyInput components.Filter // Reuse filter component for modify input
	annotateInput components.Filter // Reuse filter component for annotate input
	newTaskInput components.Filter // Reuse filter component for new task input
	// sections  components.Sections
	// help      components.Help

	// Confirm action tracking
	confirmAction string // "delete", "done", etc.
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
		modifyInput:    components.NewFilter(),
		annotateInput:  components.NewFilter(),
		newTaskInput:   components.NewFilter(),
		confirmAction:  "",
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
		m.errorMessage = "" // Clear any previous error
		m.statusMessage = "Task updated successfully"
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
	case StateModifyInput:
		return m.handleModifyKeys(msg)
	case StateAnnotateInput:
		return m.handleAnnotateKeys(msg)
	case StateNewTaskInput:
		return m.handleNewTaskKeys(msg)
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
		m.updateComponentSizes()
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

	case "d":
		// Mark task done
		selectedTask := m.taskList.SelectedTask()
		if selectedTask != nil {
			return m, markTaskDoneCmd(m.service, selectedTask.UUID)
		}
		return m, nil

	case "x":
		// Delete task (with confirmation)
		selectedTask := m.taskList.SelectedTask()
		if selectedTask != nil {
			m.state = StateConfirm
			m.confirmAction = "delete"
			return m, nil
		}
		return m, nil

	case "u":
		// Undo last operation
		return m, undoCmd(m.service)

	case "n":
		// New task
		m.state = StateNewTaskInput
		m.newTaskInput.SetValue("")
		m.updateComponentSizes()
		return m, m.newTaskInput.Focus()

	case "m":
		// Modify task
		selectedTask := m.taskList.SelectedTask()
		if selectedTask != nil {
			m.state = StateModifyInput
			m.modifyInput.SetValue("")
			m.updateComponentSizes()
			return m, m.modifyInput.Focus()
		}
		return m, nil

	case "a":
		// Add annotation
		selectedTask := m.taskList.SelectedTask()
		if selectedTask != nil {
			m.state = StateAnnotateInput
			m.annotateInput.SetValue("")
			m.updateComponentSizes()
			return m, m.annotateInput.Focus()
		}
		return m, nil

	case "e":
		// Edit task (suspend TUI)
		selectedTask := m.taskList.SelectedTask()
		if selectedTask != nil {
			return m, editTaskCmd(m.config.TaskBin, m.config.TaskrcPath, selectedTask.UUID)
		}
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

	// If in input mode, subtract input prompt area (2 lines: separator + input)
	if m.state == StateFilterInput || m.state == StateModifyInput ||
	   m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		availableHeight -= 2
	}

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
		m.updateComponentSizes()
		return m, nil

	case "enter":
		// Apply the filter
		filterText := m.filter.Value()
		m.state = StateNormal
		m.filter.Blur()
		m.activeFilter = filterText
		m.updateComponentSizes()

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
		m.confirmAction = ""
		return m, nil
	case "y", "Y":
		// Execute the confirmed action
		m.state = StateNormal
		selectedTask := m.taskList.SelectedTask()

		if m.confirmAction == "delete" && selectedTask != nil {
			m.confirmAction = ""
			return m, deleteTaskCmd(m.service, selectedTask.UUID)
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
		selectedTask := m.taskList.SelectedTask()
		m.state = StateNormal
		m.modifyInput.Blur()
		m.updateComponentSizes()

		if selectedTask != nil && modifications != "" {
			return m, modifyTaskCmd(m.service, selectedTask.UUID, modifications)
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
		selectedTask := m.taskList.SelectedTask()
		m.state = StateNormal
		m.annotateInput.Blur()
		m.updateComponentSizes()

		if selectedTask != nil && text != "" {
			return m, annotateTaskCmd(m.service, selectedTask.UUID, text)
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
func loadTasksCmd(service core.TaskService, filter string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := service.Export(filter)
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
