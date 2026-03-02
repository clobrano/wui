package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/clobrano/wui/internal/calendar"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/taskwarrior"
	"github.com/clobrano/wui/internal/tui"
	"github.com/clobrano/wui/internal/version"
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

var (
	syncCalendarName string
	syncTaskFilter   string
)

func usage() {
	fmt.Fprintf(os.Stderr, `wui (Warrior UI) - A modern TUI for Taskwarrior

Usage:
  wui [flags]         Run the TUI
  wui version         Print the version number of wui
  wui sync [flags]    Sync Taskwarrior tasks to Google Calendar
  wui help            Show this help

Flags:
  --config string      config file path (default: ~/.config/wui/config.yaml)
  --taskrc string      taskrc file path (default: ~/.taskrc)
  --task-bin string    task binary path (default: /usr/local/bin/task)
  --log-level string   log level: debug, info, warn, error (default: error)
  --log-format string  log format: text, json (default: text)
  --search string      open in Search tab with the specified filter

Sync flags:
  --calendar string    Google Calendar name (overrides config)
  --filter string      Taskwarrior filter for tasks to sync (overrides config)

Before syncing you need to:
  1. Create a Google Cloud project and enable the Google Calendar API
  2. Download the credentials.json file from Google Cloud Console
  3. Place it in ~/.config/wui/credentials.json
  4. Configure calendar_name and task_filter in config.yaml
`)
}

func addCommonFlags(fs *flag.FlagSet) {
	fs.StringVar(&configPath, "config", "", "config file path (default: ~/.config/wui/config.yaml)")
	fs.StringVar(&taskrcPath, "taskrc", "", "taskrc file path (default: ~/.taskrc)")
	fs.StringVar(&taskBinPath, "task-bin", "", "task binary path (default: /usr/local/bin/task)")
	fs.StringVar(&logLevel, "log-level", "error", "log level (debug, info, warn, error)")
	fs.StringVar(&logFormat, "log-format", "text", "log format (text, json)")
	fs.StringVar(&searchFilter, "search", "", "open in Search tab with the specified filter")
}

func main() {
	args := os.Args
	// Guard against empty os.Args (can occur on Android/Termux arm64)
	if len(args) == 0 {
		args = []string{"wui"}
	}

	if len(args) >= 2 {
		switch args[1] {
		case "version":
			fmt.Printf("wui version %s\n", version.GetVersion())
			return
		case "sync":
			if err := runSyncWithFlags(args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "help":
			usage()
			return
		}
	}

	if err := runRootWithFlags(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runRootWithFlags(args []string) error {
	fs := flag.NewFlagSet("wui", flag.ExitOnError)
	addCommonFlags(fs)
	fs.Usage = usage
	if err := fs.Parse(args); err != nil {
		return err
	}
	return runTUI()
}

func runSyncWithFlags(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	addCommonFlags(fs)
	fs.StringVar(&syncCalendarName, "calendar", "", "Google Calendar name (overrides config)")
	fs.StringVar(&syncTaskFilter, "filter", "", "Taskwarrior filter for tasks to sync (overrides config)")
	fs.Usage = usage
	if err := fs.Parse(args); err != nil {
		return err
	}
	return runSync()
}

// runTUI initializes and runs the TUI application
func runTUI() error {
	// Resolve config path
	cfgPath := config.ResolveConfigPath(configPath)

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
		logFile = "/tmp/wui.log"
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		f = nil
	}

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

	slog.SetDefault(slog.New(handler))
}

// checkTaskBinary verifies that the task binary exists and is executable
func checkTaskBinary(taskBin string) error {
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
	cfgPath := config.ResolveConfigPath(configPath)

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		initLogging(nil)
		slog.Error("Failed to load config", "error", err, "path", cfgPath)
		return fmt.Errorf("failed to load config: %w", err)
	}

	initLogging(cfg)

	slog.Info("Starting Google Calendar sync", "version", version.GetVersion())
	slog.Debug("Using config path", "path", cfgPath)

	if taskBinPath != "" {
		cfg.TaskBin = taskBinPath
	}
	if taskrcPath != "" {
		cfg.TaskrcPath = taskrcPath
	}

	calendarName := cfg.CalendarSync.CalendarName
	taskFilter := cfg.CalendarSync.TaskFilter

	if syncCalendarName != "" {
		calendarName = syncCalendarName
	}
	if syncTaskFilter != "" {
		taskFilter = syncTaskFilter
	}

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

	if err := checkTaskBinary(cfg.TaskBin); err != nil {
		return err
	}

	taskClient, err := taskwarrior.NewClient(cfg.TaskBin, cfg.TaskrcPath)
	if err != nil {
		slog.Error("Failed to create taskwarrior client", "error", err)
		return fmt.Errorf("failed to create taskwarrior client: %w", err)
	}

	credentialsPath := cfg.CalendarSync.CredentialsPath
	tokenPath := cfg.CalendarSync.TokenPath

	slog.Info("Using credentials", "path", credentialsPath, "token_path", tokenPath)

	ctx := context.Background()
	syncClient, err := calendar.NewSyncClient(ctx, taskClient, credentialsPath, tokenPath, calendarName, taskFilter)
	if err != nil {
		slog.Error("Failed to create sync client", "error", err)
		return fmt.Errorf("failed to create sync client: %w", err)
	}

	result, err := syncClient.Sync(ctx)
	if err != nil {
		slog.Error("Sync failed", "error", err)
		return fmt.Errorf("sync failed: %w", err)
	}

	slog.Info("Sync completed successfully", "created", result.Created, "updated", result.Updated)

	if len(result.Warnings) > 0 {
		fmt.Println("\n========================================")
		for _, warning := range result.Warnings {
			fmt.Printf("⚠️  WARNING: %s\n", warning)
		}
		fmt.Println("========================================")
	}

	return nil
}
