package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}

	// Check task binary defaults
	if cfg.TaskBin == "" {
		t.Error("Expected TaskBin to be set")
	}

	// Check taskrc defaults
	if cfg.TaskrcPath == "" {
		t.Error("Expected TaskrcPath to be set")
	}

	// Check TUI config exists
	if cfg.TUI == nil {
		t.Error("Expected TUI config to be set")
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		TaskBin:    "/usr/bin/task",
		TaskrcPath: "/home/user/.taskrc",
		TUI: &TUIConfig{
			SidebarWidth: 50,
		},
	}

	if cfg.TaskBin != "/usr/bin/task" {
		t.Errorf("Expected TaskBin '/usr/bin/task', got %s", cfg.TaskBin)
	}
	if cfg.TaskrcPath != "/home/user/.taskrc" {
		t.Errorf("Expected TaskrcPath '/home/user/.taskrc', got %s", cfg.TaskrcPath)
	}
	if cfg.TUI.SidebarWidth != 50 {
		t.Errorf("Expected SidebarWidth 50, got %d", cfg.TUI.SidebarWidth)
	}
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Load config from non-existent file (not explicit) - should create default
	cfg, err := LoadConfig(configPath, false)
	if err != nil {
		t.Errorf("Expected no error for missing config, got %v", err)
	}
	if cfg == nil {
		t.Error("Expected default config when file doesn't exist")
	}
	// Should return default config
	if cfg.TaskBin == "" {
		t.Error("Expected default TaskBin")
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected default config file to be created")
	}

	// Verify file contains valid YAML
	loadedCfg, err := LoadConfig(configPath, false)
	if err != nil {
		t.Fatalf("Failed to load created config: %v", err)
	}
	if loadedCfg.TaskBin != cfg.TaskBin {
		t.Errorf("Expected TaskBin %s, got %s", cfg.TaskBin, loadedCfg.TaskBin)
	}
}

func TestLoadConfig_ExplicitPathNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	// When path is explicitly specified and file doesn't exist, expect error
	_, err := LoadConfig(configPath, true)
	if err == nil {
		t.Error("Expected error when explicitly specified config file doesn't exist")
	}
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `task_bin: /custom/bin/task
taskrc_path: /custom/.taskrc
tui:
  sidebar_width: 60
  tabs:
    - name: "Work Tasks"
      filter: "+work status:pending"
    - name: "Urgent"
      filter: "priority:H"
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.TaskBin != "/custom/bin/task" {
		t.Errorf("Expected TaskBin '/custom/bin/task', got %s", cfg.TaskBin)
	}
	if cfg.TaskrcPath != "/custom/.taskrc" {
		t.Errorf("Expected TaskrcPath '/custom/.taskrc', got %s", cfg.TaskrcPath)
	}
	if cfg.TUI.SidebarWidth != 60 {
		t.Errorf("Expected SidebarWidth 60, got %d", cfg.TUI.SidebarWidth)
	}
	if len(cfg.TUI.Tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(cfg.TUI.Tabs))
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `task_bin: /usr/bin/task
invalid yaml content {{{
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	_, err = LoadConfig(configPath, false)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		TaskBin:    "/usr/local/bin/task",
		TaskrcPath: "/home/user/.taskrc",
		TUI: &TUIConfig{
			SidebarWidth: 45,
			Tabs: []Tab{
				{Name: "Test", Filter: "status:pending"},
			},
		},
	}

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load it back and verify
	loadedCfg, err := LoadConfig(configPath, false)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedCfg.TaskBin != cfg.TaskBin {
		t.Errorf("Expected TaskBin %s, got %s", cfg.TaskBin, loadedCfg.TaskBin)
	}
	if loadedCfg.TUI.SidebarWidth != cfg.TUI.SidebarWidth {
		t.Errorf("Expected SidebarWidth %d, got %d", cfg.TUI.SidebarWidth, loadedCfg.TUI.SidebarWidth)
	}
}

func TestResolveConfigPath(t *testing.T) {
	// Ensure WUI_CONFIG is not set for these tests
	t.Setenv("WUI_CONFIG", "")

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "empty path uses default",
			input:    "",
			contains: ".config/wui/config.yaml",
		},
		{
			name:     "absolute path unchanged",
			input:    "/custom/path/config.yaml",
			contains: "/custom/path/config.yaml",
		},
		{
			name:     "tilde expansion",
			input:    "~/myconfig.yaml",
			contains: "myconfig.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveConfigPath(tt.input)
			if !filepath.IsAbs(result) {
				t.Errorf("Expected absolute path, got %s", result)
			}
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("Expected path to contain %q, got %s", tt.contains, result)
			}
		})
	}
}

func TestResolveConfigPath_EnvVar(t *testing.T) {
	t.Setenv("WUI_CONFIG", "/env/path/config.yaml")

	// Empty flag should fall back to env var
	result := ResolveConfigPath("")
	if result != "/env/path/config.yaml" {
		t.Errorf("Expected WUI_CONFIG path, got %s", result)
	}

	// Explicit flag should take precedence over env var
	result = ResolveConfigPath("/flag/path/config.yaml")
	if result != "/flag/path/config.yaml" {
		t.Errorf("Expected flag path to override WUI_CONFIG, got %s", result)
	}
}


func TestConfigMergeWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config with only some fields set
	partialYAML := `task_bin: /custom/bin/task
`

	err := os.WriteFile(configPath, []byte(partialYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Custom field should be set
	if cfg.TaskBin != "/custom/bin/task" {
		t.Errorf("Expected custom TaskBin, got %s", cfg.TaskBin)
	}

	// Default fields should be set
	if cfg.TaskrcPath == "" {
		t.Error("Expected default TaskrcPath to be set")
	}
	if cfg.TUI == nil {
		t.Error("Expected default TUI config")
	}
	if cfg.TUI.SidebarWidth == 0 {
		t.Error("Expected default SidebarWidth")
	}

	// NarrowViewFields should default to due and tags
	if len(cfg.TUI.NarrowViewFields) != 2 {
		t.Errorf("Expected 2 default NarrowViewFields (due, tags), got %d fields", len(cfg.TUI.NarrowViewFields))
	}
	if len(cfg.TUI.NarrowViewFields) > 0 && cfg.TUI.NarrowViewFields[0].Name != "due" {
		t.Errorf("Expected first default NarrowViewField to be 'due', got %s", cfg.TUI.NarrowViewFields[0].Name)
	}
	if len(cfg.TUI.NarrowViewFields) > 1 && cfg.TUI.NarrowViewFields[1].Name != "tags" {
		t.Errorf("Expected second default NarrowViewField to be 'tags', got %s", cfg.TUI.NarrowViewFields[1].Name)
	}
}

func TestConfigMergeNarrowViewFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config with custom narrow_view_fields
	customYAML := `tui:
  narrow_view_fields:
    - name: "project"
      label: "Project"
    - name: "priority"
      label: "Priority"
`

	err := os.WriteFile(configPath, []byte(customYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Custom narrow_view_fields should be set
	if len(cfg.TUI.NarrowViewFields) != 2 {
		t.Errorf("Expected 2 narrow view fields, got %d", len(cfg.TUI.NarrowViewFields))
	}
	if len(cfg.TUI.NarrowViewFields) > 0 && cfg.TUI.NarrowViewFields[0].Name != "project" {
		t.Errorf("Expected first field to be 'project', got %s", cfg.TUI.NarrowViewFields[0].Name)
	}
	if len(cfg.TUI.NarrowViewFields) > 1 && cfg.TUI.NarrowViewFields[1].Name != "priority" {
		t.Errorf("Expected second field to be 'priority', got %s", cfg.TUI.NarrowViewFields[1].Name)
	}
}
