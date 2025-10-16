package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

// Run starts the TUI application
// Returns an error if the TUI fails to start or exits with an error
func Run(service core.TaskService, cfg *config.Config) error {
	// Create the model
	model := NewModel(service, cfg)

	// Create the program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	_, err := p.Run()
	return err
}
