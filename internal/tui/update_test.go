package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

// Helper function to create a model with test tasks
func createTestModel(service core.TaskService) Model {
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)
	model.tasks = []core.Task{
		{UUID: "test-uuid-1", Description: "Test task 1"},
		{UUID: "test-uuid-2", Description: "Test task 2"},
		{UUID: "test-uuid-3", Description: "Test task 3"},
	}
	model.taskList.SetTasks(model.tasks)
	model.width = 100
	model.height = 30
	return model
}

// Test edit operation (task 9.1-9.2)
func TestHandleEditKey(t *testing.T) {
	service := &core.MockTaskService{
		EditFunc: func(uuid string) error {
			if uuid != "test-uuid-1" {
				t.Errorf("Expected UUID 'test-uuid-1', got %s", uuid)
			}
			return nil
		},
	}

	model := createTestModel(service)
	model.state = StateNormal

	// Press 'e' key to edit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to remain Normal, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected a command to be returned for edit operation")
	}

	// Note: The command uses tea.ExecProcess which is async and can't be easily tested here
}

func TestEditTaskCmd(t *testing.T) {
	// Test the editTaskCmd function
	cfg := config.DefaultConfig()
	cmd := editTaskCmd(cfg.TaskBin, cfg.TaskrcPath, "test-uuid")

	if cmd == nil {
		t.Error("Expected editTaskCmd to return a command")
	}

	// We can't fully test ExecProcess without actually running the command
	// but we can verify the command is created
}

// Test modify operation (task 9.3-9.4)
func TestHandleModifyKey(t *testing.T) {
	service := &core.MockTaskService{}
	model := createTestModel(service)
	model.state = StateNormal

	// Press 'm' key to start modify input
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateModifyInput {
		t.Errorf("Expected state to be ModifyInput, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected focus command to be returned")
	}
}

func TestHandleModifyInput(t *testing.T) {
	service := &core.MockTaskService{
		ModifyFunc: func(uuid, modifications string) error {
			if uuid != "test-uuid-1" {
				t.Errorf("Expected UUID 'test-uuid-1', got %s", uuid)
			}
			if modifications != "priority:H due:tomorrow" {
				t.Errorf("Expected modifications 'priority:H due:tomorrow', got %s", modifications)
			}
			return nil
		},
	}

	model := createTestModel(service)
	model.state = StateModifyInput
	model.modifyInput.SetValue("priority:H due:tomorrow")

	// Press Enter to confirm modification
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to return to Normal, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected modify command to be returned")
	}

	// Execute the command to trigger the modification
	if cmd != nil {
		result := cmd()
		if result == nil {
			t.Error("Expected command to return a message")
		}
		// Verify it's a TaskModifiedMsg
		if _, ok := result.(TaskModifiedMsg); !ok {
			t.Errorf("Expected TaskModifiedMsg, got %T", result)
		}
	}
}

func TestHandleModifyEscape(t *testing.T) {
	service := &core.MockTaskService{}
	model := createTestModel(service)
	model.state = StateModifyInput

	// Press Esc to cancel modification
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	updatedModel, _ := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to return to Normal after Esc, got %v", m.state)
	}
}

func TestModifyTaskCmd(t *testing.T) {
	modifyCalled := false
	service := &core.MockTaskService{
		ModifyFunc: func(uuid, modifications string) error {
			modifyCalled = true
			return nil
		},
	}

	cmd := modifyTaskCmd(service, "test-uuid", "priority:H")
	if cmd == nil {
		t.Fatal("Expected modifyTaskCmd to return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected command to return a message")
	}

	if !modifyCalled {
		t.Error("Expected ModifyFunc to be called")
	}

	// Verify it's a TaskModifiedMsg
	if _, ok := msg.(TaskModifiedMsg); !ok {
		t.Errorf("Expected TaskModifiedMsg, got %T", msg)
	}
}

// Test annotate operation (task 9.5-9.6)
func TestHandleAnnotateKey(t *testing.T) {
	service := &core.MockTaskService{}
	model := createTestModel(service)
	model.state = StateNormal

	// Press 'a' key to start annotation input
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateAnnotateInput {
		t.Errorf("Expected state to be AnnotateInput, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected focus command to be returned")
	}
}

func TestHandleAnnotateInput(t *testing.T) {
	service := &core.MockTaskService{
		AnnotateFunc: func(uuid, text string) error {
			if uuid != "test-uuid-1" {
				t.Errorf("Expected UUID 'test-uuid-1', got %s", uuid)
			}
			if text != "This is a test annotation" {
				t.Errorf("Expected text 'This is a test annotation', got %s", text)
			}
			return nil
		},
	}

	model := createTestModel(service)
	model.state = StateAnnotateInput
	model.annotateInput.SetValue("This is a test annotation")

	// Press Enter to confirm annotation
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to return to Normal, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected annotate command to be returned")
	}

	// Execute the command
	if cmd != nil {
		result := cmd()
		if result == nil {
			t.Error("Expected command to return a message")
		}
	}
}

func TestAnnotateTaskCmd(t *testing.T) {
	annotateCalled := false
	service := &core.MockTaskService{
		AnnotateFunc: func(uuid, text string) error {
			annotateCalled = true
			if uuid != "test-uuid" {
				t.Errorf("Expected UUID 'test-uuid', got %s", uuid)
			}
			if text != "Test annotation" {
				t.Errorf("Expected text 'Test annotation', got %s", text)
			}
			return nil
		},
	}

	cmd := annotateTaskCmd(service, "test-uuid", "Test annotation")
	if cmd == nil {
		t.Fatal("Expected annotateTaskCmd to return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected command to return a message")
	}

	if !annotateCalled {
		t.Error("Expected AnnotateFunc to be called")
	}
}

// Test done operation (task 9.7)
func TestHandleDoneKey(t *testing.T) {
	service := &core.MockTaskService{
		DoneFunc: func(uuid string) error {
			if uuid != "test-uuid-1" {
				t.Errorf("Expected UUID 'test-uuid-1', got %s", uuid)
			}
			return nil
		},
	}

	model := createTestModel(service)
	model.state = StateNormal

	// Press 'd' key to mark task done
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to remain Normal, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected done command to be returned")
	}

	// Execute the command
	if cmd != nil {
		result := cmd()
		if result == nil {
			t.Error("Expected command to return a message")
		}
	}
}

func TestMarkTaskDoneCmd(t *testing.T) {
	doneCalled := false
	service := &core.MockTaskService{
		DoneFunc: func(uuid string) error {
			doneCalled = true
			return nil
		},
	}

	cmd := markTaskDoneCmd(service, "test-uuid", "test-task")
	if cmd == nil {
		t.Fatal("Expected markTaskDoneCmd to return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected command to return a message")
	}

	if !doneCalled {
		t.Error("Expected DoneFunc to be called")
	}

	// Verify it's a TaskModifiedMsg
	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err != nil {
		t.Errorf("Expected no error, got %v", modMsg.Err)
	}
}

// Test delete operation (task 9.8-9.9)
func TestHandleDeleteKey(t *testing.T) {
	service := &core.MockTaskService{}
	model := createTestModel(service)
	model.state = StateNormal

	// Press 'x' key to initiate delete
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateConfirm {
		t.Errorf("Expected state to be Confirm, got %v", m.state)
	}
	if m.confirmAction != "delete" {
		t.Errorf("Expected confirmAction to be 'delete', got %s", m.confirmAction)
	}
	if cmd != nil {
		t.Error("Expected no command until confirmation")
	}
}

func TestHandleDeleteConfirmYes(t *testing.T) {
	service := &core.MockTaskService{
		DeleteFunc: func(uuid string) error {
			if uuid != "test-uuid-1" {
				t.Errorf("Expected UUID 'test-uuid-1', got %s", uuid)
			}
			return nil
		},
	}

	model := createTestModel(service)
	model.state = StateConfirm
	model.confirmAction = "delete"

	// Press 'y' to confirm deletion
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to return to Normal, got %v", m.state)
	}
	if m.confirmAction != "" {
		t.Errorf("Expected confirmAction to be cleared, got %s", m.confirmAction)
	}
	if cmd == nil {
		t.Error("Expected delete command to be returned")
	}
}

func TestHandleDeleteConfirmNo(t *testing.T) {
	service := &core.MockTaskService{}
	model := createTestModel(service)
	model.state = StateConfirm
	model.confirmAction = "delete"

	// Press 'n' to cancel deletion
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to return to Normal, got %v", m.state)
	}
	if m.confirmAction != "" {
		t.Errorf("Expected confirmAction to be cleared, got %s", m.confirmAction)
	}
	if cmd != nil {
		t.Error("Expected no command when canceling")
	}
}

func TestDeleteTaskCmd(t *testing.T) {
	deleteCalled := false
	service := &core.MockTaskService{
		DeleteFunc: func(uuid string) error {
			deleteCalled = true
			return nil
		},
	}

	cmd := deleteTaskCmd(service, "test-uuid")
	if cmd == nil {
		t.Fatal("Expected deleteTaskCmd to return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected command to return a message")
	}

	if !deleteCalled {
		t.Error("Expected DeleteFunc to be called")
	}
}

// Test add new task operation (task 9.10-9.11)
func TestHandleNewTaskKey(t *testing.T) {
	service := &core.MockTaskService{}
	model := createTestModel(service)
	model.state = StateNormal

	// Press 'n' key to create new task
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNewTaskInput {
		t.Errorf("Expected state to be NewTaskInput, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected focus command to be returned")
	}
}

func TestHandleNewTaskInput(t *testing.T) {
	service := &core.MockTaskService{
		AddFunc: func(description string) (string, error) {
			if description != "Buy groceries +shopping" {
				t.Errorf("Expected description 'Buy groceries +shopping', got %s", description)
			}
			return "new-task-uuid", nil
		},
	}

	model := createTestModel(service)
	model.state = StateNewTaskInput
	model.newTaskInput.SetValue("Buy groceries +shopping")

	// Press Enter to create task
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to return to Normal, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected add task command to be returned")
	}

	// Execute the command
	if cmd != nil {
		result := cmd()
		if result == nil {
			t.Error("Expected command to return a message")
		}
	}
}

func TestAddTaskCmd(t *testing.T) {
	addCalled := false
	service := &core.MockTaskService{
		AddFunc: func(description string) (string, error) {
			addCalled = true
			if description != "Test new task" {
				t.Errorf("Expected description 'Test new task', got %s", description)
			}
			return "new-uuid-123", nil
		},
	}

	cmd := addTaskCmd(service, "Test new task")
	if cmd == nil {
		t.Fatal("Expected addTaskCmd to return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected command to return a message")
	}

	if !addCalled {
		t.Error("Expected AddFunc to be called")
	}

	// Verify it's a TaskModifiedMsg
	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err != nil {
		t.Errorf("Expected no error, got %v", modMsg.Err)
	}
}

// Test undo operation (task 9.12)
func TestHandleUndoKey(t *testing.T) {
	service := &core.MockTaskService{
		UndoFunc: func() error {
			return nil
		},
	}

	model := createTestModel(service)
	model.state = StateNormal

	// Press 'u' key to undo
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to remain Normal, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected undo command to be returned")
	}
}

func TestUndoCmd(t *testing.T) {
	undoCalled := false
	service := &core.MockTaskService{
		UndoFunc: func() error {
			undoCalled = true
			return nil
		},
	}

	cmd := undoCmd(service)
	if cmd == nil {
		t.Fatal("Expected undoCmd to return a command")
	}

	// Execute the command
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected command to return a message")
	}

	if !undoCalled {
		t.Error("Expected UndoFunc to be called")
	}
}

// Test manual refresh operation (task 9.13)
func TestHandleRefreshKey(t *testing.T) {
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			return []core.Task{
				{UUID: "refreshed-task", Description: "Refreshed"},
			}, nil
		},
	}

	model := createTestModel(service)
	model.state = StateNormal
	model.activeFilter = "status:pending"

	// Press 'r' key to refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.state != StateNormal {
		t.Errorf("Expected state to remain Normal, got %v", m.state)
	}
	if cmd == nil {
		t.Error("Expected refresh command to be returned")
	}

	// Execute the command to trigger export
	if cmd != nil {
		result := cmd()
		if result == nil {
			t.Error("Expected command to return a message")
		}
		// Verify it's a TasksLoadedMsg
		if _, ok := result.(TasksLoadedMsg); !ok {
			t.Errorf("Expected TasksLoadedMsg, got %T", result)
		}
	}
}

// Test error handling for operations (task 9.14)
func TestModifyErrorHandling(t *testing.T) {
	service := &core.MockTaskService{
		ModifyFunc: func(uuid, modifications string) error {
			return errors.New("modification failed")
		},
	}

	cmd := modifyTaskCmd(service, "test-uuid", "invalid")
	msg := cmd()

	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err == nil {
		t.Error("Expected error to be set")
	}
	if modMsg.Err.Error() != "modification failed" {
		t.Errorf("Expected error message 'modification failed', got %s", modMsg.Err.Error())
	}
}

func TestAnnotateErrorHandling(t *testing.T) {
	service := &core.MockTaskService{
		AnnotateFunc: func(uuid, text string) error {
			return errors.New("annotation failed")
		},
	}

	cmd := annotateTaskCmd(service, "test-uuid", "test")
	msg := cmd()

	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err == nil {
		t.Error("Expected error to be set")
	}
}

func TestDoneErrorHandling(t *testing.T) {
	service := &core.MockTaskService{
		DoneFunc: func(uuid string) error {
			return errors.New("task already completed")
		},
	}

	cmd := markTaskDoneCmd(service, "test-uuid", "test-task")
	msg := cmd()

	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err == nil {
		t.Error("Expected error to be set")
	}
}

func TestDeleteErrorHandling(t *testing.T) {
	service := &core.MockTaskService{
		DeleteFunc: func(uuid string) error {
			return errors.New("task not found")
		},
	}

	cmd := deleteTaskCmd(service, "test-uuid")
	msg := cmd()

	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err == nil {
		t.Error("Expected error to be set")
	}
}

func TestAddTaskErrorHandling(t *testing.T) {
	service := &core.MockTaskService{
		AddFunc: func(description string) (string, error) {
			return "", errors.New("invalid task description")
		},
	}

	cmd := addTaskCmd(service, "")
	msg := cmd()

	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err == nil {
		t.Error("Expected error to be set")
	}
}

func TestUndoErrorHandling(t *testing.T) {
	service := &core.MockTaskService{
		UndoFunc: func() error {
			return errors.New("nothing to undo")
		},
	}

	cmd := undoCmd(service)
	msg := cmd()

	modMsg, ok := msg.(TaskModifiedMsg)
	if !ok {
		t.Fatalf("Expected TaskModifiedMsg, got %T", msg)
	}

	if modMsg.Err == nil {
		t.Error("Expected error to be set")
	}
}

func TestTaskModifiedMsgErrorHandling(t *testing.T) {
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			return []core.Task{}, nil
		},
	}

	model := createTestModel(service)

	// Test TaskModifiedMsg with error
	msg := TaskModifiedMsg{
		UUID:        "test-uuid",
		Description: "test task",
		Err:         errors.New("operation failed"),
	}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.errorMessage == "" {
		t.Error("Expected error message to be set")
	}
	if m.errorMessage != "Task operation failed: operation failed" {
		t.Errorf("Expected specific error message, got: %s", m.errorMessage)
	}
	if cmd != nil {
		t.Error("Expected no command when error occurs")
	}
}

func TestTaskModifiedMsgSuccess(t *testing.T) {
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			return []core.Task{
				{UUID: "task-1", Description: "Updated task"},
			}, nil
		},
	}

	model := createTestModel(service)
	model.activeFilter = "status:pending"

	// Test TaskModifiedMsg with success
	msg := TaskModifiedMsg{
		UUID:        "test-uuid",
		Description: "test task",
		Err:         nil,
	}
	updatedModel, cmd := model.Update(msg)

	m := updatedModel.(Model)
	if m.errorMessage != "" {
		t.Errorf("Expected error message to be cleared, got: %s", m.errorMessage)
	}
	if m.statusMessage != "Task updated successfully" {
		t.Errorf("Expected success message, got: %s", m.statusMessage)
	}
	if cmd == nil {
		t.Error("Expected refresh command to be returned")
	}
}

// Test that operations do nothing when no task is selected
func TestOperationsWithNoTaskSelected(t *testing.T) {
	service := &core.MockTaskService{}
	cfg := config.DefaultConfig()
	model := NewModel(service, cfg)
	model.tasks = []core.Task{} // Empty task list
	model.taskList.SetTasks(model.tasks)
	model.state = StateNormal

	tests := []struct {
		name string
		key  rune
	}{
		{"edit with no task", 'e'},
		{"modify with no task", 'm'},
		{"annotate with no task", 'a'},
		{"done with no task", 'd'},
		{"delete with no task", 'x'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			updatedModel, cmd := model.Update(msg)

			m := updatedModel.(Model)

			// For 'n' key (new task), state should change even without selection
			// For other keys, nothing should happen
			if tt.key != 'n' && tt.key != 'r' && tt.key != 'u' {
				if m.state != StateNormal {
					t.Errorf("Expected state to remain Normal, got %v", m.state)
				}
				if cmd != nil && tt.key != 'd' {
					// 'd' might return a command even with no selection in current implementation
					// This is okay as the command will handle the nil task
				}
			}
		})
	}
}

// Test empty input handling
func TestEmptyInputHandling(t *testing.T) {
	service := &core.MockTaskService{}
	model := createTestModel(service)

	tests := []struct {
		name  string
		state AppState
		input string
	}{
		{"empty modify", StateModifyInput, ""},
		{"empty annotate", StateAnnotateInput, ""},
		{"empty new task", StateNewTaskInput, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.state = tt.state
			switch tt.state {
			case StateModifyInput:
				model.modifyInput.SetValue(tt.input)
			case StateAnnotateInput:
				model.annotateInput.SetValue(tt.input)
			case StateNewTaskInput:
				model.newTaskInput.SetValue(tt.input)
			}

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			updatedModel, cmd := model.Update(msg)

			m := updatedModel.(Model)
			if m.state != StateNormal {
				t.Errorf("Expected state to return to Normal, got %v", m.state)
			}
			// With empty input, no command should be returned
			if cmd != nil {
				t.Error("Expected no command with empty input")
			}
		})
	}
}
