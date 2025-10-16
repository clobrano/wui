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

	// Should contain title
	if !strings.Contains(view, "wui") {
		t.Error("Expected view to contain 'wui'")
	}
}

func TestViewWithTasks(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	model.width = 80
	model.height = 24
	model.tasks = []core.Task{
		{UUID: "task-1", Description: "Test task 1", Project: "TestProject"},
		{UUID: "task-2", Description: "Test task 2", Project: "TestProject"},
	}

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
		{UUID: "task-1", Description: "Test task", Project: "Project1", Status: "pending"},
	}
	model.selectedIndex = 0

	view := model.View()
	if !strings.Contains(view, "Task Details") {
		t.Error("Expected view to contain sidebar")
	}
}

func TestRenderTaskLine(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	task := core.Task{
		UUID:        "abc-123-def",
		Description: "Test task",
		Project:     "TestProject",
	}

	line := model.renderTaskLine(task, false)
	if !strings.Contains(line, "Test task") {
		t.Error("Expected line to contain task description")
	}
	if !strings.Contains(line, "abc-123") {
		t.Error("Expected line to contain shortened UUID")
	}
}

func TestRenderTaskLineSelected(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)

	task := core.Task{
		UUID:        "abc-123",
		Description: "Selected task",
	}

	line := model.renderTaskLine(task, true)
	if !strings.Contains(line, "■") {
		t.Error("Expected selected line to contain cursor")
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
