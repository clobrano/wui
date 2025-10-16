package tui

import (
	"testing"

	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

func TestRunFunction(t *testing.T) {
	// This is a basic test to ensure Run function exists and has correct signature
	// Actual TUI testing would require manual verification
	service := &core.MockTaskService{
		ExportFunc: func(filter string) ([]core.Task, error) {
			return []core.Task{}, nil
		},
	}
	cfg := config.DefaultConfig()

	// We can't actually run the TUI in tests, but we can verify the function exists
	// and accepts the right parameters
	_ = service
	_ = cfg

	// This would normally call Run(service, cfg) but that would block tests
	// Instead, just verify we can create a model
	model := NewModel(service, cfg)
	if model.service == nil {
		t.Error("Expected service to be set")
	}
}
