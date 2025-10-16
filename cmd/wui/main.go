package main

import (
	"fmt"
	"os"

	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/taskwarrior"
	"github.com/clobrano/wui/internal/tui"
	"github.com/clobrano/wui/internal/version"
	"github.com/spf13/cobra"
)

var (
	// CLI flags
	configPath  string
	taskrcPath  string
	taskBinPath string
)

var rootCmd = &cobra.Command{
	Use:   "wui",
	Short: "Warrior UI - A modern TUI for Taskwarrior",
	Long: `wui (Warrior UI) is a Terminal User Interface for Taskwarrior built with Go and bubbletea.
It provides an intuitive, keyboard-driven interface for managing your tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of wui",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("wui version %s\n", version.GetVersion())
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(versionCmd)

	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (default: ~/.config/wui/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&taskrcPath, "taskrc", "", "taskrc file path (default: ~/.taskrc)")
	rootCmd.PersistentFlags().StringVar(&taskBinPath, "task-bin", "", "task binary path (default: /usr/local/bin/task)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runTUI initializes and runs the TUI application
func runTUI() error {
	// Resolve config path
	cfgPath := config.ResolveConfigPath(configPath)

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with CLI flags if provided
	if taskBinPath != "" {
		cfg.TaskBin = taskBinPath
	}
	if taskrcPath != "" {
		cfg.TaskrcPath = taskrcPath
	}

	// Create Taskwarrior client
	client, err := taskwarrior.NewClient(cfg.TaskBin, cfg.TaskrcPath)
	if err != nil {
		return fmt.Errorf("failed to create taskwarrior client: %w", err)
	}

	// Run the TUI
	return tui.Run(client, cfg)
}
