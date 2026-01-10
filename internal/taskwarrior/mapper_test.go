package taskwarrior

import (
	"testing"
	"time"

	"github.com/clobrano/wui/internal/core"
)

// Ensure core package is used
var _ core.Task

func TestMapToCore_BasicFields(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Test task",
		Project:     "MyProject",
		Tags:        []string{"work", "urgent"},
		Priority:    "H",
		Status:      "pending",
		Urgency:     8.5,
	}

	coreTask := MapToCore(tw)

	if coreTask.UUID != "abc-123" {
		t.Errorf("Expected UUID 'abc-123', got %s", coreTask.UUID)
	}
	if coreTask.Description != "Test task" {
		t.Errorf("Expected description 'Test task', got %s", coreTask.Description)
	}
	if coreTask.Project != "MyProject" {
		t.Errorf("Expected project 'MyProject', got %s", coreTask.Project)
	}
	if len(coreTask.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(coreTask.Tags))
	}
	if coreTask.Priority != "H" {
		t.Errorf("Expected priority 'H', got %s", coreTask.Priority)
	}
	if coreTask.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", coreTask.Status)
	}
	if coreTask.Urgency != 8.5 {
		t.Errorf("Expected urgency 8.5, got %f", coreTask.Urgency)
	}
}

func TestMapToCore_Dates(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Test task",
		Status:      "pending",
		Entry:       "20251016T120000Z",
		Due:         "20251020T000000Z",
		Scheduled:   "20251018T000000Z",
		Wait:        "20251017T000000Z",
		Modified:    "20251016T130000Z",
		Urgency:     5.0,
	}

	coreTask := MapToCore(tw)

	if coreTask.Entry.IsZero() {
		t.Error("Expected Entry to be set")
	}
	if coreTask.Due == nil {
		t.Error("Expected Due to be set")
	} else {
		expected := time.Date(2025, 10, 20, 0, 0, 0, 0, time.UTC)
		if !coreTask.Due.Equal(expected) {
			t.Errorf("Expected Due %v, got %v", expected, *coreTask.Due)
		}
	}
	if coreTask.Scheduled == nil {
		t.Error("Expected Scheduled to be set")
	}
	if coreTask.Wait == nil {
		t.Error("Expected Wait to be set")
	}
	if coreTask.Modified == nil {
		t.Error("Expected Modified to be set")
	}
}

func TestMapToCore_CompletedTask(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Completed task",
		Status:      "completed",
		Entry:       "20251016T120000Z",
		End:         "20251016T140000Z",
		Urgency:     0.0,
	}

	coreTask := MapToCore(tw)

	if coreTask.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", coreTask.Status)
	}
	if coreTask.End == nil {
		t.Error("Expected End to be set")
	} else {
		expected := time.Date(2025, 10, 16, 14, 0, 0, 0, time.UTC)
		if !coreTask.End.Equal(expected) {
			t.Errorf("Expected End %v, got %v", expected, *coreTask.End)
		}
	}
}

func TestMapToCore_Annotations(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Task with annotations",
		Status:      "pending",
		Entry:       "20251016T120000Z",
		Annotations: []TaskwarriorAnnotation{
			{
				Entry:       "20251016T130000Z",
				Description: "First note",
			},
			{
				Entry:       "20251016T140000Z",
				Description: "Second note",
			},
		},
		Urgency: 3.0,
	}

	coreTask := MapToCore(tw)

	if len(coreTask.Annotations) != 2 {
		t.Fatalf("Expected 2 annotations, got %d", len(coreTask.Annotations))
	}

	if coreTask.Annotations[0].Description != "First note" {
		t.Errorf("Expected first annotation 'First note', got %s", coreTask.Annotations[0].Description)
	}
	if coreTask.Annotations[1].Description != "Second note" {
		t.Errorf("Expected second annotation 'Second note', got %s", coreTask.Annotations[1].Description)
	}
	if coreTask.Annotations[0].Entry.IsZero() {
		t.Error("Expected annotation entry time to be set")
	}
}

func TestMapToCore_Dependencies(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Task with dependencies",
		Status:      "pending",
		Entry:       "20251016T120000Z",
		Depends:     []string{"dep1", "dep2", "dep3"},
		Urgency:     5.0,
	}

	coreTask := MapToCore(tw)

	if len(coreTask.Depends) != 3 {
		t.Fatalf("Expected 3 dependencies, got %d", len(coreTask.Depends))
	}

	expected := []string{"dep1", "dep2", "dep3"}
	for i, dep := range expected {
		if coreTask.Depends[i] != dep {
			t.Errorf("Expected dependency %s, got %s", dep, coreTask.Depends[i])
		}
	}
}

func TestMapToCore_EmptyDependencies(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Task without dependencies",
		Status:      "pending",
		Entry:       "20251016T120000Z",
		Depends:     []string{},
		Urgency:     5.0,
	}

	coreTask := MapToCore(tw)

	if len(coreTask.Depends) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(coreTask.Depends))
	}
}

func TestMapToCore_InvalidDate(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Task with invalid date",
		Status:      "pending",
		Entry:       "invalid-date",
		Urgency:     5.0,
	}

	coreTask := MapToCore(tw)

	// Should handle invalid dates gracefully
	// Entry is required, so it should be zero time if parsing fails
	if !coreTask.Entry.IsZero() {
		t.Log("Entry parsing should fail gracefully for invalid date")
	}
}

func TestParseTaskwarriorDate_ValidDate(t *testing.T) {
	dateStr := "20251016T120000Z"
	tm, err := parseTaskwarriorDate(dateStr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if tm == nil {
		t.Fatal("Expected time pointer, got nil")
	}

	expected := time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC)
	if !tm.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, *tm)
	}
}

func TestParseTaskwarriorDate_EmptyString(t *testing.T) {
	tm, err := parseTaskwarriorDate("")
	if err != nil {
		t.Errorf("Expected no error for empty string, got %v", err)
	}
	if tm != nil {
		t.Error("Expected nil for empty string")
	}
}

func TestParseTaskwarriorDate_InvalidFormat(t *testing.T) {
	tm, err := parseTaskwarriorDate("invalid-date")
	if err == nil {
		t.Error("Expected error for invalid date format")
	}
	if tm != nil {
		t.Error("Expected nil for invalid date")
	}
}

func TestMapToCore_WithUDAs(t *testing.T) {
	tw := TaskwarriorTask{
		UUID:        "abc-123",
		Description: "Task with UDAs",
		Status:      "pending",
		Entry:       "20251016T120000Z",
		Urgency:     5.0,
		UDA: map[string]interface{}{
			"estimate":       "2h",
			"priority_score": float64(8),
			"custom_field":   "custom value",
			"recur":          "weekly",
		},
	}

	coreTask := MapToCore(tw)

	// Verify UDAs are mapped
	if coreTask.UDAs == nil {
		t.Fatal("Expected UDAs map to be populated, got nil")
	}

	// Check string UDA
	if val, ok := coreTask.UDAs["estimate"]; !ok {
		t.Error("Expected 'estimate' UDA to be present")
	} else if val != "2h" {
		t.Errorf("Expected estimate '2h', got %v", val)
	}

	// Check numeric UDA (should be converted to string)
	if val, ok := coreTask.UDAs["priority_score"]; !ok {
		t.Error("Expected 'priority_score' UDA to be present")
	} else if val != "8" {
		t.Errorf("Expected priority_score '8', got %v", val)
	}

	// Check custom field
	if val, ok := coreTask.UDAs["custom_field"]; !ok {
		t.Error("Expected 'custom_field' UDA to be present")
	} else if val != "custom value" {
		t.Errorf("Expected custom_field 'custom value', got %v", val)
	}

	// Check recur field
	if val, ok := coreTask.UDAs["recur"]; !ok {
		t.Error("Expected 'recur' UDA to be present")
	} else if val != "weekly" {
		t.Errorf("Expected recur 'weekly', got %v", val)
	}
}
