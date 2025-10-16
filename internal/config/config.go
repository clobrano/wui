package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the wui configuration
type Config struct {
	TaskBin    string     `yaml:"task_bin"`
	TaskrcPath string     `yaml:"taskrc_path"`
	TUI        *TUIConfig `yaml:"tui"`
}

// LoadConfig loads configuration from a YAML file
// If the file doesn't exist, returns default configuration
// Merges loaded config with defaults to ensure all fields are set
func LoadConfig(path string) (*Config, error) {
	// Get defaults first
	cfg := DefaultConfig()

	// Try to read the file
	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, return defaults
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var loaded Config
	err = yaml.Unmarshal(data, &loaded)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Merge with defaults
	cfg = mergeWithDefaults(cfg, &loaded)

	return cfg, nil
}

// SaveConfig writes configuration to a YAML file
func SaveConfig(cfg *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ResolveConfigPath resolves the config file path
// If empty, returns default path: ~/.config/wui/config.yaml
// If starts with ~, expands home directory
// Otherwise returns the path as-is
func ResolveConfigPath(path string) string {
	// Empty path - use default
	if path == "" {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".config", "wui", "config.yaml")
	}

	// Expand tilde
	if strings.HasPrefix(path, "~") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[1:])
	}

	// Return absolute path
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err == nil {
			return absPath
		}
	}

	return path
}

// mergeWithDefaults merges loaded config with defaults
// Fields that are not set in loaded config will use default values
func mergeWithDefaults(defaults, loaded *Config) *Config {
	result := DefaultConfig()

	// Merge top-level fields
	if loaded.TaskBin != "" {
		result.TaskBin = loaded.TaskBin
	}
	if loaded.TaskrcPath != "" {
		result.TaskrcPath = loaded.TaskrcPath
	}

	// Merge TUI config
	if loaded.TUI != nil {
		if loaded.TUI.SidebarWidth > 0 {
			result.TUI.SidebarWidth = loaded.TUI.SidebarWidth
		}
		if len(loaded.TUI.Bookmarks) > 0 {
			result.TUI.Bookmarks = loaded.TUI.Bookmarks
		}
		if len(loaded.TUI.Columns) > 0 {
			result.TUI.Columns = loaded.TUI.Columns
		}
		if len(loaded.TUI.Keybindings) > 0 {
			// Merge keybindings (loaded overrides defaults)
			for k, v := range loaded.TUI.Keybindings {
				result.TUI.Keybindings[k] = v
			}
		}
		if loaded.TUI.Theme != nil {
			result.TUI.Theme = mergeThem(result.TUI.Theme, loaded.TUI.Theme)
		}
	}

	return result
}

// mergeTheme merges loaded theme with default theme
func mergeThem(defaultTheme, loaded *Theme) *Theme {
	result := DefaultTheme()

	if loaded.PriorityHigh != "" {
		result.PriorityHigh = loaded.PriorityHigh
	}
	if loaded.PriorityMedium != "" {
		result.PriorityMedium = loaded.PriorityMedium
	}
	if loaded.PriorityLow != "" {
		result.PriorityLow = loaded.PriorityLow
	}
	if loaded.Overdue != "" {
		result.Overdue = loaded.Overdue
	}
	if loaded.DueToday != "" {
		result.DueToday = loaded.DueToday
	}
	if loaded.DueSoon != "" {
		result.DueSoon = loaded.DueSoon
	}
	if loaded.Selected != "" {
		result.Selected = loaded.Selected
	}
	if loaded.Border != "" {
		result.Border = loaded.Border
	}

	return result
}
