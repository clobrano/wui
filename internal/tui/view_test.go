package tui

import (
	"strings"
	"testing"

	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

func TestView(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	// Set dimensions
	model.width = 80
	model.height = 24

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain section tabs
	if !strings.Contains(view, "Next") {
		t.Error("Expected view to contain section tabs (Next)")
	}
}

func TestViewWithTasks(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	model.width = 80
	model.height = 24
	model.tasks = []core.Task{
		{ID: 1, UUID: "task-1", Description: "Test task 1", Project: "TestProject", Status: "pending"},
		{ID: 2, UUID: "task-2", Description: "Test task 2", Project: "TestProject", Status: "pending"},
	}
	// Update component sizes first
	model.updateComponentSizes()
	// Set tasks in the task list component
	model.taskList.SetTasks(model.tasks)

	view := model.View()
	if !strings.Contains(view, "Test task 1") {
		t.Error("Expected view to contain task description")
	}
}

func TestViewHelp(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	model.width = 80
	model.height = 24
	model.state = StateHelp

	view := model.View()
	if !strings.Contains(view, "Help") {
		t.Error("Expected view to contain help text")
	}
	if !strings.Contains(view, "Navigation") {
		t.Error("Expected view to contain navigation help")
	}
}

func TestViewLoading(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	// No dimensions set
	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Expected view to show loading state")
	}
}

func TestViewWithSidebar(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	model.width = 120
	model.height = 30
	model.viewMode = ViewModeListWithSidebar
	model.tasks = []core.Task{
		{ID: 1, UUID: "task-1", Description: "Test task", Project: "Project1", Status: "pending"},
	}
	model.taskList.SetTasks(model.tasks)
	model.updateSidebar()

	view := model.View()
	// The sidebar should show task details
	if !strings.Contains(view, "Task #") {
		t.Error("Expected view to contain sidebar task title")
	}
}

func TestRenderFooterWithError(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	model.errorMessage = "Test error"
	model.state = StateNormal

	footer := model.renderFooter()
	if !strings.Contains(footer, "Test error") {
		t.Error("Expected footer to contain error message")
	}
}

func TestRenderFooterWithStatus(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	model.statusMessage = "Task completed"
	model.state = StateNormal

	footer := model.renderFooter()
	if !strings.Contains(footer, "Task completed") {
		t.Error("Expected footer to contain status message")
	}
}
