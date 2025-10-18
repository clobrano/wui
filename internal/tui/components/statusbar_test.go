package components

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewStatusBar(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	if sb.width != 80 {
		t.Errorf("Expected width 80, got %d", sb.width)
	}
	if !sb.showKeybindings {
		t.Error("Expected showKeybindings to be true")
	}
	if sb.currentMessage != nil {
		t.Error("Expected no current message")
	}
}

func TestStatusBar_SetMessage(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	cmd := sb.SetMessage("Test message", MessageInfo, 2*time.Second)

	if sb.currentMessage == nil {
		t.Fatal("Expected current message to be set")
	}
	if sb.currentMessage.Content != "Test message" {
		t.Errorf("Expected content 'Test message', got '%s'", sb.currentMessage.Content)
	}
	if sb.currentMessage.Type != MessageInfo {
		t.Errorf("Expected type MessageInfo, got %v", sb.currentMessage.Type)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for timed message")
	}
}

func TestStatusBar_SetMessageNoDuration(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	cmd := sb.SetMessage("Permanent message", MessageError, 0)

	if sb.currentMessage == nil {
		t.Fatal("Expected current message to be set")
	}
	if cmd != nil {
		t.Error("Expected no command for permanent message")
	}
}

func TestStatusBar_ClearMessage(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	sb.SetMessage("Test", MessageInfo, 0)
	if sb.currentMessage == nil {
		t.Fatal("Expected message to be set")
	}

	sb.ClearMessage()
	if sb.currentMessage != nil {
		t.Error("Expected message to be cleared")
	}
}

func TestStatusBar_Update_ClearStatusMsg(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	sb.SetMessage("Test", MessageInfo, 0)

	updatedSb, _ := sb.Update(ClearStatusMsg{})
	if updatedSb.currentMessage != nil {
		t.Error("Expected message to be cleared after ClearStatusMsg")
	}
}

func TestStatusBar_Update_WindowSizeMsg(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	updatedSb, _ := sb.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if updatedSb.width != 100 {
		t.Errorf("Expected width 100, got %d", updatedSb.width)
	}
}

func TestStatusBar_View_Empty(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, false, styles)

	view := sb.View()
	if view == "" {
		t.Error("Expected non-empty view (padding)")
	}
}

func TestStatusBar_View_WithMessage(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	sb.SetMessage("Task completed", MessageSuccess, 0)
	view := sb.View()

	if !strings.Contains(view, "Task completed") {
		t.Error("Expected view to contain message content")
	}
}

func TestStatusBar_View_WithKeybindings(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	sb.SetKeybindings("q: quit | ?: help")
	view := sb.View()

	if !strings.Contains(view, "quit") {
		t.Error("Expected view to contain keybindings")
	}
}

func TestStatusBar_View_MessageOverridesKeybindings(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	sb.SetKeybindings("q: quit")
	sb.SetMessage("Error occurred", MessageError, 0)

	view := sb.View()

	if !strings.Contains(view, "Error occurred") {
		t.Error("Expected view to contain message")
	}
	// Message should take priority, keybindings may not be shown
}

func TestStatusBar_SetWidth(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	sb.SetWidth(120)
	if sb.width != 120 {
		t.Errorf("Expected width 120, got %d", sb.width)
	}
}

func TestStatusBar_MessageTypes(t *testing.T) {
	styles := DefaultStatusBarStyles()
	sb := NewStatusBar(80, true, styles)

	tests := []struct {
		name    string
		msgType MessageType
		content string
	}{
		{"Info", MessageInfo, "Info message"},
		{"Success", MessageSuccess, "Success message"},
		{"Error", MessageError, "Error message"},
		{"Warning", MessageWarning, "Warning message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetMessage(tt.content, tt.msgType, 0)
			view := sb.View()

			if !strings.Contains(view, tt.content) {
				t.Errorf("Expected view to contain '%s'", tt.content)
			}
		})
	}
}
