package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/clobrano/wui/internal/calendar"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/taskwarrior"
	"github.com/clobrano/wui/internal/tui"
	"github.com/clobrano/wui/internal/version"
	"github.com/spf13/cobra"
)

var (
	// CLI flags
	configPath   string
	taskrcPath   string
	taskBinPath  string
	logLevel     string
	logFormat    string
	searchFilter string
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

var (
	syncCalendarName string
	syncTaskFilter   string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync Taskwarrior tasks to Google Calendar",
	Long: `Synchronize Taskwarrior tasks to Google Calendar.

This command syncs tasks from Taskwarrior to a specified Google Calendar using
settings from your config file (~/.config/wui/config.yaml). You can override
these settings with command-line flags.

Before syncing, you need to:
1. Create a Google Cloud project and enable the Google Calendar API
2. Download the credentials.json file from Google Cloud Console
3. Place it in ~/.config/wui/credentials.json
4. Configure calendar_name and task_filter in config.yaml

On first run, you'll be prompted to authorize the app in your browser.

Examples:
  wui sync                                    # Use config.yaml settings
  wui sync --calendar "Work"                  # Override calendar
  wui sync --filter "+urgent"                 # Override filter
  wui sync --calendar "Tasks" --filter "due:today"  # Override both`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSync(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(syncCmd)

	// Sync command flags (optional - override config file values)
	syncCmd.Flags().StringVar(&syncCalendarName, "calendar", "", "Google Calendar name (overrides config)")
	syncCmd.Flags().StringVar(&syncTaskFilter, "filter", "", "Taskwarrior filter for tasks to sync (overrides config)")

	// Persistent flags available to all commands
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (default: ~/.config/wui/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&taskrcPath, "taskrc", "", "taskrc file path (default: ~/.taskrc)")
	rootCmd.PersistentFlags().StringVar(&taskBinPath, "task-bin", "", "task binary path (default: /usr/local/bin/task)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "log format (text, json)")
	rootCmd.PersistentFlags().StringVar(&searchFilter, "search", "", "open in Search tab with the specified filter")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runTUI initializes and runs the TUI application
func runTUI() error {
	// Initialize logging
	initLogging()

	slog.Info("Starting wui", "version", version.GetVersion())

	// Resolve config path
	cfgPath := config.ResolveConfigPath(configPath)
	slog.Debug("Using config path", "path", cfgPath)

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		slog.Error("Failed to load config", "error", err, "path", cfgPath)
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with CLI flags if provided
	if taskBinPath != "" {
		slog.Debug("Overriding task binary path", "path", taskBinPath)
		cfg.TaskBin = taskBinPath
	}
	if taskrcPath != "" {
		slog.Debug("Overriding taskrc path", "path", taskrcPath)
		cfg.TaskrcPath = taskrcPath
	}
	if searchFilter != "" {
		slog.Debug("Setting initial search filter", "filter", searchFilter)
		cfg.InitialSearchFilter = searchFilter
	}

	slog.Info("Configuration loaded",
		"task_bin", cfg.TaskBin,
		"taskrc_path", cfg.TaskrcPath)

	// Check if task binary exists and is executable
	if err := checkTaskBinary(cfg.TaskBin); err != nil {
		slog.Error("Task binary not found or not executable", "error", err, "path", cfg.TaskBin)
		return err
	}

	// Create Taskwarrior client
	client, err := taskwarrior.NewClient(cfg.TaskBin, cfg.TaskrcPath)
	if err != nil {
		slog.Error("Failed to create taskwarrior client", "error", err)
		return fmt.Errorf("failed to create taskwarrior client: %w", err)
	}

	slog.Debug("Taskwarrior client created successfully")

	// Run the TUI
	return tui.Run(client, cfg)
}

// initLogging initializes the logging system based on CLI flags
func initLogging() {
	// Parse log level
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// For TUI mode, log to file instead of stderr to avoid interfering with display
	logFile := os.Getenv("WUI_LOG_FILE")
	if logFile == "" {
		// Default to temp file
		logFile = "/tmp/wui.log"
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fallback to discarding logs if file can't be opened
		f = nil
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	output := os.Stderr
	if f != nil {
		output = f
	}

	if logFormat == "json" {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	// Set default logger
	slog.SetDefault(slog.New(handler))
}

// checkTaskBinary verifies that the task binary exists and is executable
func checkTaskBinary(taskBin string) error {
	// Use exec.LookPath to check if the binary is in PATH or at the specified location
	path, err := exec.LookPath(taskBin)
	if err != nil {
		return fmt.Errorf(`task binary not found: %w

Taskwarrior is required to run wui. Please install it:

  • Ubuntu/Debian:    sudo apt install taskwarrior
  • Fedora/RHEL:      sudo dnf install task
  • macOS (Homebrew): brew install task
  • Arch Linux:       sudo pacman -S task

Or specify a custom path with: wui --task-bin /path/to/task

Visit https://taskwarrior.org for more information.`, err)
	}

	slog.Debug("Task binary found", "path", path)
	return nil
}

// runSync performs the Google Calendar sync operation
func runSync() error {
	// Initialize logging
	initLogging()

	slog.Info("Starting Google Calendar sync", "version", version.GetVersion())

	// Resolve config path
	cfgPath := config.ResolveConfigPath(configPath)
	slog.Debug("Using config path", "path", cfgPath)

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		slog.Error("Failed to load config", "error", err, "path", cfgPath)
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with CLI flags if provided
	if taskBinPath != "" {
		cfg.TaskBin = taskBinPath
	}
	if taskrcPath != "" {
		cfg.TaskrcPath = taskrcPath
	}

	// Get calendar name and filter from config, allow flags to override
	calendarName := cfg.CalendarSync.CalendarName
	taskFilter := cfg.CalendarSync.TaskFilter

	if syncCalendarName != "" {
		calendarName = syncCalendarName
	}
	if syncTaskFilter != "" {
		taskFilter = syncTaskFilter
	}

	// Validate required fields
	if calendarName == "" {
		return fmt.Errorf("calendar name is required (set in config.yaml or use --calendar flag)")
	}
	if taskFilter == "" {
		return fmt.Errorf("task filter is required (set in config.yaml or use --filter flag)")
	}

	slog.Info("Sync configuration",
		"calendar", calendarName,
		"filter", taskFilter,
		"task_bin", cfg.TaskBin,
		"taskrc_path", cfg.TaskrcPath)

	// Check if task binary exists
	if err := checkTaskBinary(cfg.TaskBin); err != nil {
		return err
	}

	// Create Taskwarrior client
	taskClient, err := taskwarrior.NewClient(cfg.TaskBin, cfg.TaskrcPath)
	if err != nil {
		slog.Error("Failed to create taskwarrior client", "error", err)
		return fmt.Errorf("failed to create taskwarrior client: %w", err)
	}

	// Get credentials and token paths from config
	credentialsPath := cfg.CalendarSync.CredentialsPath
	tokenPath := cfg.CalendarSync.TokenPath

	slog.Info("Using credentials", "path", credentialsPath, "token_path", tokenPath)

	// Create sync client
	ctx := context.Background()
	syncClient, err := calendar.NewSyncClient(ctx, taskClient, credentialsPath, tokenPath, calendarName, taskFilter)
	if err != nil {
		slog.Error("Failed to create sync client", "error", err)
		return fmt.Errorf("failed to create sync client: %w", err)
	}

	// Perform sync
	if err := syncClient.Sync(ctx); err != nil {
		slog.Error("Sync failed", "error", err)
		return fmt.Errorf("sync failed: %w", err)
	}

	slog.Info("Sync completed successfully")
	return nil
}
