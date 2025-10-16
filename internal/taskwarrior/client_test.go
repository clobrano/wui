package taskwarrior

import (
	"testing"

	"github.com/clobrano/wui/internal/core"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		taskBin     string
		taskrcPath  string
		shouldError bool
	}{
		{
			name:        "valid client",
			taskBin:     "/usr/bin/task",
			taskrcPath:  "/home/user/.taskrc",
			shouldError: false,
		},
		{
			name:        "empty task binary",
			taskBin:     "",
			taskrcPath:  "/home/user/.taskrc",
			shouldError: true,
		},
		{
			name:        "empty taskrc path",
			taskBin:     "/usr/bin/task",
			taskrcPath:  "",
			shouldError: false, // taskrc is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.taskBin, tt.taskrcPath)
			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if client == nil {
				t.Error("Expected client, got nil")
			}
			if client.taskBin != tt.taskBin {
				t.Errorf("Expected taskBin %s, got %s", tt.taskBin, client.taskBin)
			}
			if client.taskrcPath != tt.taskrcPath {
				t.Errorf("Expected taskrcPath %s, got %s", tt.taskrcPath, client.taskrcPath)
			}
		})
	}
}

func TestClientImplementsTaskService(t *testing.T) {
	var _ core.TaskService = (*Client)(nil)
}

func TestClientExport(t *testing.T) {
	// This test requires a mock or test environment with taskwarrior
	// For now, we'll just verify the method exists and has the correct signature
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	// Test with empty filter - should not panic
	_, err := client.Export("")
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Export returned no error (unexpected in test environment)")
	}
}

func TestClientModify(t *testing.T) {
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	// Test with mock UUID and modifications
	err := client.Modify("abc-123", "priority:H")
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Modify returned no error (unexpected in test environment)")
	}
}

func TestClientAnnotate(t *testing.T) {
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	err := client.Annotate("abc-123", "Test annotation")
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Annotate returned no error (unexpected in test environment)")
	}
}

func TestClientDone(t *testing.T) {
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	err := client.Done("abc-123")
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Done returned no error (unexpected in test environment)")
	}
}

func TestClientDelete(t *testing.T) {
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	err := client.Delete("abc-123")
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Delete returned no error (unexpected in test environment)")
	}
}

func TestClientAdd(t *testing.T) {
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	_, err := client.Add("New task +work")
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Add returned no error (unexpected in test environment)")
	}
}

func TestClientUndo(t *testing.T) {
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	err := client.Undo()
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Undo returned no error (unexpected in test environment)")
	}
}

func TestClientEdit(t *testing.T) {
	client := &Client{
		taskBin:    "/usr/bin/task",
		taskrcPath: "/home/user/.taskrc",
	}

	err := client.Edit("abc-123")
	// We expect an error since we're not actually running taskwarrior
	if err == nil {
		t.Log("Edit returned no error (unexpected in test environment)")
	}
}
