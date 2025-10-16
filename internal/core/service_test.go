package core

import (
	"errors"
	"testing"
)

// MockTaskService is a mock implementation of TaskService for testing
type MockTaskService struct {
	ExportFunc   func(filter string) ([]Task, error)
	ModifyFunc   func(uuid, modifications string) error
	AnnotateFunc func(uuid, text string) error
	DoneFunc     func(uuid string) error
	DeleteFunc   func(uuid string) error
	AddFunc      func(description string) (string, error)
	UndoFunc     func() error
	EditFunc     func(uuid string) error
}

func (m *MockTaskService) Export(filter string) ([]Task, error) {
	if m.ExportFunc != nil {
		return m.ExportFunc(filter)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTaskService) Modify(uuid, modifications string) error {
	if m.ModifyFunc != nil {
		return m.ModifyFunc(uuid, modifications)
	}
	return errors.New("not implemented")
}

func (m *MockTaskService) Annotate(uuid, text string) error {
	if m.AnnotateFunc != nil {
		return m.AnnotateFunc(uuid, text)
	}
	return errors.New("not implemented")
}

func (m *MockTaskService) Done(uuid string) error {
	if m.DoneFunc != nil {
		return m.DoneFunc(uuid)
	}
	return errors.New("not implemented")
}

func (m *MockTaskService) Delete(uuid string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(uuid)
	}
	return errors.New("not implemented")
}

func (m *MockTaskService) Add(description string) (string, error) {
	if m.AddFunc != nil {
		return m.AddFunc(description)
	}
	return "", errors.New("not implemented")
}

func (m *MockTaskService) Undo() error {
	if m.UndoFunc != nil {
		return m.UndoFunc()
	}
	return errors.New("not implemented")
}

func (m *MockTaskService) Edit(uuid string) error {
	if m.EditFunc != nil {
		return m.EditFunc(uuid)
	}
	return errors.New("not implemented")
}

// TestTaskServiceInterface verifies that MockTaskService implements TaskService
func TestTaskServiceInterface(t *testing.T) {
	var _ TaskService = (*MockTaskService)(nil)
}

func TestMockTaskServiceExport(t *testing.T) {
	expectedTasks := []Task{
		{UUID: "abc-123", Description: "Test task"},
	}

	mock := &MockTaskService{
		ExportFunc: func(filter string) ([]Task, error) {
			if filter == "status:pending" {
				return expectedTasks, nil
			}
			return nil, errors.New("invalid filter")
		},
	}

	// Test successful export
	tasks, err := mock.Export("status:pending")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].UUID != "abc-123" {
		t.Errorf("Expected UUID 'abc-123', got %s", tasks[0].UUID)
	}

	// Test error case
	_, err = mock.Export("invalid")
	if err == nil {
		t.Error("Expected error for invalid filter")
	}
}

func TestMockTaskServiceModify(t *testing.T) {
	modified := false
	mock := &MockTaskService{
		ModifyFunc: func(uuid, modifications string) error {
			if uuid == "abc-123" && modifications == "priority:H" {
				modified = true
				return nil
			}
			return errors.New("invalid modification")
		},
	}

	err := mock.Modify("abc-123", "priority:H")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !modified {
		t.Error("Expected modify to be called")
	}
}

func TestMockTaskServiceAnnotate(t *testing.T) {
	annotated := false
	mock := &MockTaskService{
		AnnotateFunc: func(uuid, text string) error {
			if uuid == "abc-123" && text == "Test note" {
				annotated = true
				return nil
			}
			return errors.New("invalid annotation")
		},
	}

	err := mock.Annotate("abc-123", "Test note")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !annotated {
		t.Error("Expected annotate to be called")
	}
}

func TestMockTaskServiceDone(t *testing.T) {
	done := false
	mock := &MockTaskService{
		DoneFunc: func(uuid string) error {
			if uuid == "abc-123" {
				done = true
				return nil
			}
			return errors.New("task not found")
		},
	}

	err := mock.Done("abc-123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !done {
		t.Error("Expected done to be called")
	}
}

func TestMockTaskServiceDelete(t *testing.T) {
	deleted := false
	mock := &MockTaskService{
		DeleteFunc: func(uuid string) error {
			if uuid == "abc-123" {
				deleted = true
				return nil
			}
			return errors.New("task not found")
		},
	}

	err := mock.Delete("abc-123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !deleted {
		t.Error("Expected delete to be called")
	}
}

func TestMockTaskServiceAdd(t *testing.T) {
	mock := &MockTaskService{
		AddFunc: func(description string) (string, error) {
			if description == "New task" {
				return "new-uuid-123", nil
			}
			return "", errors.New("invalid description")
		},
	}

	uuid, err := mock.Add("New task")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if uuid != "new-uuid-123" {
		t.Errorf("Expected UUID 'new-uuid-123', got %s", uuid)
	}

	// Test error case
	_, err = mock.Add("")
	if err == nil {
		t.Error("Expected error for empty description")
	}
}

func TestMockTaskServiceUndo(t *testing.T) {
	undone := false
	mock := &MockTaskService{
		UndoFunc: func() error {
			undone = true
			return nil
		},
	}

	err := mock.Undo()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !undone {
		t.Error("Expected undo to be called")
	}
}

func TestMockTaskServiceEdit(t *testing.T) {
	edited := false
	mock := &MockTaskService{
		EditFunc: func(uuid string) error {
			if uuid == "abc-123" {
				edited = true
				return nil
			}
			return errors.New("task not found")
		},
	}

	err := mock.Edit("abc-123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !edited {
		t.Error("Expected edit to be called")
	}
}
