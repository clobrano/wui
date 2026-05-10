package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/clobrano/wui/internal/api"
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

var serveAddr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Long: `Start the wui REST/JSON API server.

The server exposes the Taskwarrior backend over HTTP so that alternative UIs
(Flutter, web, CLI scripts) can drive the same business logic as the TUI.

API base: http://<addr>/api/v1

Endpoints:
  GET    /api/v1/tasks               list tasks (optional ?filter= query param)
  POST   /api/v1/tasks               create a task  {"description":"..."}
  PUT    /api/v1/tasks/{uuid}         modify a task  {"modifications":"priority:H"}
  DELETE /api/v1/tasks/{uuid}         delete a task
  POST   /api/v1/tasks/{uuid}/done    mark done
  POST   /api/v1/tasks/{uuid}/start   start
  POST   /api/v1/tasks/{uuid}/stop    stop
  POST   /api/v1/tasks/{uuid}/annotate add annotation  {"text":"..."}
  POST   /api/v1/undo                undo last operation
  GET    /api/v1/projects            list project summaries

Examples:
  wui serve                    # Listen on localhost:7007
  wui serve --addr :7007       # Listen on all interfaces`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runServe(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

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
	rootCmd.AddCommand(serveCmd)

	// Serve command flags
	serveCmd.Flags().StringVar(&serveAddr, "addr", "localhost:7007", "address to listen on (host:port)")

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
	// Resolve config path
	cfgPath := config.ResolveConfigPath(configPath)

	// If the user explicitly passed --config, the file must exist
	if err := config.ValidateExplicitConfigPath(configPath, cfgPath); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		// Use basic logging before config is loaded
		initLogging(nil)
		slog.Error("Failed to load config", "error", err, "path", cfgPath)
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logging with config (priority: flag > env > config)
	initLogging(cfg)

	slog.Info("Starting wui", "version", version.GetVersion())
	slog.Debug("Using config path", "path", cfgPath)

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

	// Check if taskrc file exists
	if err := config.ValidateTaskrcPath(cfg.TaskrcPath); err != nil {
		slog.Error("Taskrc file not found", "error", err, "path", cfg.TaskrcPath)
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

// initLogging initializes the logging system with priority: flag > env > config
func initLogging(cfg *config.Config) {
	// Determine log level with priority: CLI flag > env variable > config file
	effectiveLogLevel := "error" // fallback default

	// Start with config value if available
	if cfg != nil && cfg.LogLevel != "" {
		effectiveLogLevel = cfg.LogLevel
	}

	// Override with environment variable if set
	if envLogLevel := os.Getenv("WUI_LOG_LEVEL"); envLogLevel != "" {
		effectiveLogLevel = envLogLevel
	}

	// Override with CLI flag if provided (non-default)
	// Note: We check if it's different from the default value
	if logLevel != "" {
		effectiveLogLevel = logLevel
	}

	// Parse log level
	var level slog.Level
	switch effectiveLogLevel {
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

// runServe starts the REST API server backed by the local Taskwarrior installation.
func runServe() error {
	cfgPath := config.ResolveConfigPath(configPath)
	if err := config.ValidateExplicitConfigPath(configPath, cfgPath); err != nil {
		return err
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		initLogging(nil)
		slog.Error("Failed to load config", "error", err, "path", cfgPath)
		return fmt.Errorf("failed to load config: %w", err)
	}

	initLogging(cfg)
	slog.Info("Starting wui API server", "version", version.GetVersion())

	if taskBinPath != "" {
		cfg.TaskBin = taskBinPath
	}
	if taskrcPath != "" {
		cfg.TaskrcPath = taskrcPath
	}

	if err := checkTaskBinary(cfg.TaskBin); err != nil {
		return err
	}
	if err := config.ValidateTaskrcPath(cfg.TaskrcPath); err != nil {
		return err
	}

	client, err := taskwarrior.NewClient(cfg.TaskBin, cfg.TaskrcPath)
	if err != nil {
		return fmt.Errorf("failed to create taskwarrior client: %w", err)
	}

	srv := api.NewServer(client, serveAddr)

	// Graceful shutdown on SIGINT / SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()

	select {
	case err := <-errCh:
		return err
	case <-stop:
		fmt.Println("\nShutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}

// runSync performs the Google Calendar sync operation
func runSync() error {
	// Resolve config path
	cfgPath := config.ResolveConfigPath(configPath)

	// If the user explicitly passed --config, the file must exist
	if err := config.ValidateExplicitConfigPath(configPath, cfgPath); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		// Use basic logging before config is loaded
		initLogging(nil)
		slog.Error("Failed to load config", "error", err, "path", cfgPath)
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logging with config (priority: flag > env > config)
	initLogging(cfg)

	slog.Info("Starting Google Calendar sync", "version", version.GetVersion())
	slog.Debug("Using config path", "path", cfgPath)

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

	// Check if taskrc file exists
	if err := config.ValidateTaskrcPath(cfg.TaskrcPath); err != nil {
		slog.Error("Taskrc file not found", "error", err, "path", cfg.TaskrcPath)
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
	result, err := syncClient.Sync(ctx)
	if err != nil {
		slog.Error("Sync failed", "error", err)
		return fmt.Errorf("sync failed: %w", err)
	}

	slog.Info("Sync completed successfully", "created", result.Created, "updated", result.Updated)

	// Print warnings if any (for TUI mode when output might be lost)
	if len(result.Warnings) > 0 {
		fmt.Println("\n========================================")
		for _, warning := range result.Warnings {
			fmt.Printf("⚠️  WARNING: %s\n", warning)
		}
		fmt.Println("========================================")
	}

	return nil
}
