package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

func TestNewModel(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()

	model := NewModel(service, cfg)

	if model.service == nil {
		t.Error("Expected service to be set")
	}
	if model.config == nil {
		t.Error("Expected config to be set")
	}
	if model.tasks == nil {
		t.Error("Expected tasks slice to be initialized")
	}
	if model.state != StateNormal {
		t.Errorf("Expected initial state to be Normal, got %v", model.state)
	}
	if model.viewMode != ViewModeList {
		t.Errorf("Expected initial view mode to be List, got %v", model.viewMode)
	}
}

func TestModelInit(t *testing.T) {
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			return []core.Task{
				{UUID: "task-1", Description: "Test task"},
			}, nil
		},
	}
	cfg := config.DefaultConfig()

	model := NewModel(service, cfg)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Expected Init to return a command")
	}

	// Execute the command to get the message
	msg := cmd()
	if msg == nil {
		t.Error("Expected command to return a message")
	}

	// Check it's a TasksLoadedMsg
	if _, ok := msg.(TasksLoadedMsg); !ok {
		t.Errorf("Expected TasksLoadedMsg, got %T", msg)
	}
}

func TestViewMode(t *testing.T) {
	tests := []struct {
		name     string
		viewMode ViewMode
		expected string
	}{
		{"List mode", ViewModeList, "list"},
		{"List with sidebar mode", ViewModeListWithSidebar, "list_with_sidebar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.viewMode.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.viewMode.String())
			}
		})
	}
}

func TestAppState(t *testing.T) {
	tests := []struct {
		name     string
		state    AppState
		expected string
	}{
		{"Normal state", StateNormal, "normal"},
		{"Filter input state", StateFilterInput, "filter_input"},
		{"Help state", StateHelp, "help"},
		{"Confirm state", StateConfirm, "confirm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.state.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.state.String())
			}
		})
	}
}

func TestModelStruct(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	sections := core.DefaultSections()

	model := Model{
		service:        service,
		config:         cfg,
		tasks:          []core.Task{},
		viewMode:       ViewModeList,
		state:          StateNormal,
		currentSection: &sections[0],
		statusMessage:  "",
		errorMessage:   "",
	}

	if model.service != service {
		t.Error("Expected service to match")
	}
	if model.config != cfg {
		t.Error("Expected config to match")
	}
	if len(model.tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(model.tasks))
	}
	if model.viewMode != ViewModeList {
		t.Errorf("Expected ViewModeList, got %v", model.viewMode)
	}
	if model.state != StateNormal {
		t.Errorf("Expected StateNormal, got %v", model.state)
	}
}

func TestModelUpdateWithTasksLoaded(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	tasks := []core.Task{
		{UUID: "task-1", Description: "Task 1"},
		{UUID: "task-2", Description: "Task 2"},
	}

	msg := TasksLoadedMsg{Tasks: tasks, Err: nil}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(Model)
	if len(m.tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(m.tasks))
	}
	if m.tasks[0].UUID != "task-1" {
		t.Errorf("Expected first task UUID 'task-1', got %s", m.tasks[0].UUID)
	}
}

func TestModelUpdateWithError(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	msg := ErrorMsg{Err: tea.ErrProgramKilled}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(Model)
	if m.errorMessage == "" {
		t.Error("Expected error message to be set")
	}
}

func TestLoadTasksCmd(t *testing.T) {
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			return []core.Task{
				{UUID: "test-task", Description: "Test"},
			}, nil
		},
	}

	cmd := loadTasksCmd(service, "status:pending", false)
	if cmd == nil {
		t.Fatal("Expected loadTasksCmd to return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected command to return a message")
	}

	// Check it's a TasksLoadedMsg
	loadedMsg, ok := msg.(TasksLoadedMsg)
	if !ok {
		t.Fatalf("Expected TasksLoadedMsg, got %T", msg)
	}

	if len(loadedMsg.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loadedMsg.Tasks))
	}
	if loadedMsg.Err != nil {
		t.Errorf("Expected no error, got %v", loadedMsg.Err)
	}
}

func TestLoadTasksCmdWithError(t *testing.T) {
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			return nil, tea.ErrProgramKilled
		},
	}

	cmd := loadTasksCmd(service, "status:pending", false)
	msg := cmd()

	loadedMsg, ok := msg.(TasksLoadedMsg)
	if !ok {
		t.Fatalf("Expected TasksLoadedMsg, got %T", msg)
	}

	if loadedMsg.Err == nil {
		t.Error("Expected error to be set")
	}
	if len(loadedMsg.Tasks) != 0 {
		t.Errorf("Expected 0 tasks on error, got %d", len(loadedMsg.Tasks))
	}
}

func TestNewModelWithSearchFilter(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	cfg.InitialSearchFilter = "project:Home +urgent"

	model := NewModel(service, cfg)

	// Should start in Search tab (index 0)
	if model.currentSection == nil || model.currentSection.Name != "Search" {
		t.Errorf("Expected to start in Search tab, got %v", model.currentSection)
	}

	// Should have the filter saved in searchTabFilter
	if model.searchTabFilter != "project:Home +urgent" {
		t.Errorf("Expected searchTabFilter to be 'project:Home +urgent', got %s", model.searchTabFilter)
	}
}

func TestNewModelWithoutSearchFilter(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	// No InitialSearchFilter set

	model := NewModel(service, cfg)

	// Should NOT start in Search tab (should start in "Next" tab, index 1)
	if model.currentSection == nil || model.currentSection.Name == "Search" {
		t.Errorf("Expected NOT to start in Search tab, got %v", model.currentSection)
	}

	// searchTabFilter should be empty
	if model.searchTabFilter != "" {
		t.Errorf("Expected searchTabFilter to be empty, got %s", model.searchTabFilter)
	}
}

func TestInitWithSearchFilter(t *testing.T) {
	filterCalled := ""
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			filterCalled = filter
			return []core.Task{
				{UUID: "task-1", Description: "Test task"},
			}, nil
		},
	}
	cfg := config.DefaultConfig()
	cfg.InitialSearchFilter = "due:today"

	model := NewModel(service, cfg)
	cmd := model.Init()

	if cmd == nil {
		t.Fatal("Expected Init to return a command")
	}

	// Execute the command
	msg := cmd()

	// Check it's a TasksLoadedMsg
	if _, ok := msg.(TasksLoadedMsg); !ok {
		t.Errorf("Expected TasksLoadedMsg, got %T", msg)
	}

	// Verify the filter was used (note: Search tab prepends "status.any:" automatically)
	if filterCalled != "status.any: due:today" {
		t.Errorf("Expected filter 'status.any: due:today' to be used, got '%s'", filterCalled)
	}
}

func TestInitWithSearchFilterWithStatus(t *testing.T) {
	filterCalled := ""
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			filterCalled = filter
			return []core.Task{
				{UUID: "task-1", Description: "Test task"},
			}, nil
		},
	}
	cfg := config.DefaultConfig()
	cfg.InitialSearchFilter = "status:completed due:today"

	model := NewModel(service, cfg)
	cmd := model.Init()

	if cmd == nil {
		t.Fatal("Expected Init to return a command")
	}

	// Execute the command
	msg := cmd()

	// Check it's a TasksLoadedMsg
	if _, ok := msg.(TasksLoadedMsg); !ok {
		t.Errorf("Expected TasksLoadedMsg, got %T", msg)
	}

	// Verify the filter was used (note: status.any: should NOT be prepended when status: is already present)
	if filterCalled != "status:completed due:today" {
		t.Errorf("Expected filter 'status:completed due:today' to be used, got '%s'", filterCalled)
	}
}
