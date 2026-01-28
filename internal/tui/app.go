package tui

import (
	"fmt"

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
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Print warnings after TUI has exited (if any)
	if m, ok := finalModel.(Model); ok {
		// Print shortcut override warnings (unless silenced in config)
		silenceWarnings := m.config != nil && m.config.TUI != nil && m.config.TUI.SilenceShortcutOverrideWarnings
		if len(m.shortcutWarnings) > 0 && !silenceWarnings {
			fmt.Println("\n========================================")
			fmt.Printf("Shortcut override warnings (%d):\n", len(m.shortcutWarnings))
			fmt.Println("========================================")
			for _, warning := range m.shortcutWarnings {
				fmt.Printf("⚠️  %s\n", warning)
			}
			fmt.Println("========================================")
		}

		// Print sync warnings
		if m.syncWarnings != nil && len(m.syncWarnings.Warnings) > 0 {
			fmt.Println("\n========================================")
			fmt.Printf("Calendar sync completed with %d warnings:\n", len(m.syncWarnings.Warnings))
			fmt.Println("========================================")
			for _, warning := range m.syncWarnings.Warnings {
				fmt.Printf("⚠️  %s\n", warning)
			}
			fmt.Println("========================================\n")
		}
	}

	return nil
}
