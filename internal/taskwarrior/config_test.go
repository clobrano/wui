package taskwarrior

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTaskrc_FileNotExist(t *testing.T) {
	cfg, err := ParseTaskrc("/nonexistent/.taskrc")
	if err != nil {
		t.Errorf("Expected no error for missing file, got %v", err)
	}
	if cfg == nil {
		t.Error("Expected empty config, got nil")
	}
}

func TestParseTaskrc_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	taskrcPath := filepath.Join(tmpDir, ".taskrc")

	err := os.WriteFile(taskrcPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test taskrc: %v", err)
	}

	cfg, err := ParseTaskrc(taskrcPath)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cfg == nil {
		t.Error("Expected config, got nil")
	}
}

func TestParseTaskrc_BasicConfig(t *testing.T) {
	tmpDir := t.TempDir()
	taskrcPath := filepath.Join(tmpDir, ".taskrc")

	content := `# Taskwarrior configuration
data.location=/home/user/.task
default.command=list
`

	err := os.WriteFile(taskrcPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test taskrc: %v", err)
	}

	cfg, err := ParseTaskrc(taskrcPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.DataLocation != "/home/user/.task" {
		t.Errorf("Expected DataLocation '/home/user/.task', got %s", cfg.DataLocation)
	}
	if cfg.DefaultCommand != "list" {
		t.Errorf("Expected DefaultCommand 'list', got %s", cfg.DefaultCommand)
	}
}

func TestParseTaskrc_UDAs(t *testing.T) {
	tmpDir := t.TempDir()
	taskrcPath := filepath.Join(tmpDir, ".taskrc")

	content := `# UDA definitions
uda.estimate.type=string
uda.estimate.label=Estimate
uda.estimate.values=S,M,L,XL
uda.priority.type=string
uda.priority.label=Priority Level
uda.reviewed.type=date
uda.reviewed.label=Reviewed Date
`

	err := os.WriteFile(taskrcPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test taskrc: %v", err)
	}

	cfg, err := ParseTaskrc(taskrcPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(cfg.UDAs) != 3 {
		t.Fatalf("Expected 3 UDAs, got %d", len(cfg.UDAs))
	}

	// Check estimate UDA
	estimate, exists := cfg.UDAs["estimate"]
	if !exists {
		t.Error("Expected 'estimate' UDA to exist")
	}
	if estimate.Type != "string" {
		t.Errorf("Expected estimate type 'string', got %s", estimate.Type)
	}
	if estimate.Label != "Estimate" {
		t.Errorf("Expected estimate label 'Estimate', got %s", estimate.Label)
	}
	if estimate.Values != "S,M,L,XL" {
		t.Errorf("Expected estimate values 'S,M,L,XL', got %s", estimate.Values)
	}

	// Check reviewed UDA (date type)
	reviewed, exists := cfg.UDAs["reviewed"]
	if !exists {
		t.Error("Expected 'reviewed' UDA to exist")
	}
	if reviewed.Type != "date" {
		t.Errorf("Expected reviewed type 'date', got %s", reviewed.Type)
	}
}

func TestParseTaskrc_DefaultFilter(t *testing.T) {
	tmpDir := t.TempDir()
	taskrcPath := filepath.Join(tmpDir, ".taskrc")

	content := `report.next.filter=status:pending -WAITING limit:page
report.list.filter=status:pending
`

	err := os.WriteFile(taskrcPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test taskrc: %v", err)
	}

	cfg, err := ParseTaskrc(taskrcPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(cfg.Reports) != 2 {
		t.Fatalf("Expected 2 reports, got %d", len(cfg.Reports))
	}

	next, exists := cfg.Reports["next"]
	if !exists {
		t.Error("Expected 'next' report to exist")
	}
	if next.Filter != "status:pending -WAITING limit:page" {
		t.Errorf("Expected next filter 'status:pending -WAITING limit:page', got %s", next.Filter)
	}
}

func TestParseTaskrc_Comments(t *testing.T) {
	tmpDir := t.TempDir()
	taskrcPath := filepath.Join(tmpDir, ".taskrc")

	content := `# This is a comment
data.location=/home/user/.task
# Another comment
# uda.test=should be ignored

default.command=list
`

	err := os.WriteFile(taskrcPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test taskrc: %v", err)
	}

	cfg, err := ParseTaskrc(taskrcPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.DataLocation != "/home/user/.task" {
		t.Errorf("Expected DataLocation '/home/user/.task', got %s", cfg.DataLocation)
	}
	// Commented UDA should not be parsed
	if len(cfg.UDAs) > 0 {
		t.Errorf("Expected 0 UDAs (commented), got %d", len(cfg.UDAs))
	}
}

func TestParseTaskrc_ComplexConfig(t *testing.T) {
	tmpDir := t.TempDir()
	taskrcPath := filepath.Join(tmpDir, ".taskrc")

	content := `# Complex taskrc
data.location=/home/user/.task
default.command=next

# UDAs
uda.estimate.type=string
uda.estimate.label=Estimate
uda.reviewed.type=date
uda.reviewed.label=Last Reviewed

# Reports
report.next.filter=status:pending -WAITING
report.next.labels=ID,Project,Description,Due
report.list.filter=status:pending

# Colors and settings
color=on
color.due=red
`

	err := os.WriteFile(taskrcPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test taskrc: %v", err)
	}

	cfg, err := ParseTaskrc(taskrcPath)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check basic settings
	if cfg.DataLocation != "/home/user/.task" {
		t.Errorf("Expected DataLocation '/home/user/.task', got %s", cfg.DataLocation)
	}
	if cfg.DefaultCommand != "next" {
		t.Errorf("Expected DefaultCommand 'next', got %s", cfg.DefaultCommand)
	}

	// Check UDAs
	if len(cfg.UDAs) != 2 {
		t.Errorf("Expected 2 UDAs, got %d", len(cfg.UDAs))
	}

	// Check reports
	if len(cfg.Reports) != 2 {
		t.Errorf("Expected 2 reports (next.filter and list.filter), got %d", len(cfg.Reports))
	}
}

func TestTaskrcConfigStruct(t *testing.T) {
	cfg := &TaskrcConfig{
		DataLocation:   "/home/user/.task",
		DefaultCommand: "next",
		UDAs: map[string]UDA{
			"estimate": {
				Type:   "string",
				Label:  "Estimate",
				Values: "S,M,L,XL",
			},
		},
		Reports: map[string]Report{
			"next": {
				Filter: "status:pending",
			},
		},
	}

	if cfg.DataLocation != "/home/user/.task" {
		t.Errorf("Expected DataLocation '/home/user/.task', got %s", cfg.DataLocation)
	}
	if len(cfg.UDAs) != 1 {
		t.Errorf("Expected 1 UDA, got %d", len(cfg.UDAs))
	}
	if len(cfg.Reports) != 1 {
		t.Errorf("Expected 1 report, got %d", len(cfg.Reports))
	}
}

func TestUDAStruct(t *testing.T) {
	uda := UDA{
		Type:   "string",
		Label:  "Estimate",
		Values: "S,M,L,XL",
	}

	if uda.Type != "string" {
		t.Errorf("Expected type 'string', got %s", uda.Type)
	}
	if uda.Label != "Estimate" {
		t.Errorf("Expected label 'Estimate', got %s", uda.Label)
	}
	if uda.Values != "S,M,L,XL" {
		t.Errorf("Expected values 'S,M,L,XL', got %s", uda.Values)
	}
}

func TestReportStruct(t *testing.T) {
	report := Report{
		Filter:  "status:pending",
		Labels:  "ID,Project,Description",
		Columns: "id,project,description",
		Sort:    "urgency-",
	}

	if report.Filter != "status:pending" {
		t.Errorf("Expected filter 'status:pending', got %s", report.Filter)
	}
	if report.Labels != "ID,Project,Description" {
		t.Errorf("Expected labels 'ID,Project,Description', got %s", report.Labels)
	}
}
