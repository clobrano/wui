package taskwarrior

import (
	"testing"
)

func TestParseTaskJSON_Empty(t *testing.T) {
	json := []byte("[]")
	tasks, err := ParseTaskJSON(json)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestParseTaskJSON_SingleTask(t *testing.T) {
	json := []byte(`[
		{
			"uuid": "abc-123-def",
			"description": "Test task",
			"status": "pending",
			"entry": "20251016T120000Z",
			"urgency": 5.2
		}
	]`)

	tasks, err := ParseTaskJSON(json)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.UUID != "abc-123-def" {
		t.Errorf("Expected UUID 'abc-123-def', got %s", task.UUID)
	}
	if task.Description != "Test task" {
		t.Errorf("Expected description 'Test task', got %s", task.Description)
	}
	if task.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", task.Status)
	}
	if task.Urgency != 5.2 {
		t.Errorf("Expected urgency 5.2, got %f", task.Urgency)
	}
}

func TestParseTaskJSON_MultipleTasks(t *testing.T) {
	json := []byte(`[
		{
			"uuid": "task-1",
			"description": "First task",
			"status": "pending",
			"entry": "20251016T120000Z",
			"urgency": 1.0
		},
		{
			"uuid": "task-2",
			"description": "Second task",
			"status": "completed",
			"entry": "20251016T130000Z",
			"end": "20251016T140000Z",
			"urgency": 0.0
		}
	]`)

	tasks, err := ParseTaskJSON(json)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}

	if tasks[0].UUID != "task-1" {
		t.Errorf("Expected first UUID 'task-1', got %s", tasks[0].UUID)
	}
	if tasks[1].UUID != "task-2" {
		t.Errorf("Expected second UUID 'task-2', got %s", tasks[1].UUID)
	}
}

func TestParseTaskJSON_WithOptionalFields(t *testing.T) {
	json := []byte(`[
		{
			"uuid": "abc-123",
			"description": "Task with all fields",
			"project": "MyProject",
			"tags": ["work", "urgent"],
			"priority": "H",
			"status": "pending",
			"due": "20251020T000000Z",
			"scheduled": "20251018T000000Z",
			"wait": "20251017T000000Z",
			"entry": "20251016T120000Z",
			"modified": "20251016T130000Z",
			"depends": "other-task-uuid",
			"urgency": 8.5
		}
	]`)

	tasks, err := ParseTaskJSON(json)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.Project != "MyProject" {
		t.Errorf("Expected project 'MyProject', got %s", task.Project)
	}
	if len(task.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(task.Tags))
	}
	if task.Priority != "H" {
		t.Errorf("Expected priority 'H', got %s", task.Priority)
	}
	if task.Due != "20251020T000000Z" {
		t.Errorf("Expected due '20251020T000000Z', got %s", task.Due)
	}
	if task.Depends != "other-task-uuid" {
		t.Errorf("Expected depends 'other-task-uuid', got %s", task.Depends)
	}
}

func TestParseTaskJSON_WithAnnotations(t *testing.T) {
	json := []byte(`[
		{
			"uuid": "abc-123",
			"description": "Task with annotations",
			"status": "pending",
			"entry": "20251016T120000Z",
			"annotations": [
				{
					"entry": "20251016T130000Z",
					"description": "First note"
				},
				{
					"entry": "20251016T140000Z",
					"description": "Second note"
				}
			],
			"urgency": 3.0
		}
	]`)

	tasks, err := ParseTaskJSON(json)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if len(task.Annotations) != 2 {
		t.Fatalf("Expected 2 annotations, got %d", len(task.Annotations))
	}

	if task.Annotations[0].Description != "First note" {
		t.Errorf("Expected first annotation 'First note', got %s", task.Annotations[0].Description)
	}
	if task.Annotations[1].Description != "Second note" {
		t.Errorf("Expected second annotation 'Second note', got %s", task.Annotations[1].Description)
	}
}

func TestParseTaskJSON_InvalidJSON(t *testing.T) {
	json := []byte(`{invalid json}`)
	_, err := ParseTaskJSON(json)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseTaskJSON_NotArray(t *testing.T) {
	json := []byte(`{"uuid": "abc-123", "description": "Single object"}`)
	_, err := ParseTaskJSON(json)
	if err == nil {
		t.Error("Expected error for non-array JSON, got nil")
	}
}
