package core

import "errors"

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
