package taskwarrior

import (
	"encoding/json"
	"fmt"
)

// TaskwarriorTask represents a task as returned by Taskwarrior JSON export
type TaskwarriorTask struct {
	ID          int                      `json:"id,omitempty"` // Sequential ID (only for pending tasks)
	UUID        string                   `json:"uuid"`
	Description string                   `json:"description"`
	Project     string                   `json:"project,omitempty"`
	Tags        []string                 `json:"tags,omitempty"`
	Priority    string                   `json:"priority,omitempty"`
	Status      string                   `json:"status"`
	Due         string                   `json:"due,omitempty"`
	Scheduled   string                   `json:"scheduled,omitempty"`
	Wait        string                   `json:"wait,omitempty"`
	Start       string                   `json:"start,omitempty"`
	Entry       string                   `json:"entry"`
	Modified    string                   `json:"modified,omitempty"`
	End         string                   `json:"end,omitempty"`
	Depends     []string                 `json:"depends,omitempty"`
	Annotations []TaskwarriorAnnotation  `json:"annotations,omitempty"`
	UDA         map[string]interface{}   `json:"-"` // Populated from unmapped JSON fields
	Urgency     float64                  `json:"urgency"`
}

// UnmarshalJSON implements custom JSON unmarshaling to capture UDA fields
func (t *TaskwarriorTask) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a map to get all fields
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	// Define known fields that should NOT go into UDA
	// Only exclude fields that are explicitly mapped to struct fields
	knownFields := map[string]bool{
		"id":          true,
		"uuid":        true,
		"description": true,
		"project":     true,
		"tags":        true,
		"priority":    true,
		"status":      true,
		"due":         true,
		"scheduled":   true,
		"wait":        true,
		"start":       true,
		"entry":       true,
		"modified":    true,
		"end":         true,
		"depends":     true,
		"annotations": true,
		"urgency":     true,
		// Internal taskwarrior fields that should be hidden
		"mask":  true,
		"imask": true,
	}

	// Create a temporary struct for standard unmarshaling
	type Alias TaskwarriorTask
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	// Unmarshal into the struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Extract UDA fields (anything not in knownFields)
	t.UDA = make(map[string]interface{})
	for key, value := range rawMap {
		if !knownFields[key] {
			t.UDA[key] = value
		}
	}

	return nil
}

// TaskwarriorAnnotation represents an annotation in Taskwarrior format
type TaskwarriorAnnotation struct {
	Entry       string `json:"entry"`
	Description string `json:"description"`
}

// ParseTaskJSON parses Taskwarrior JSON export output
// Returns a slice of TaskwarriorTask or an error if parsing fails
func ParseTaskJSON(jsonBytes []byte) ([]TaskwarriorTask, error) {
	// Handle empty input
	if len(jsonBytes) == 0 {
		return []TaskwarriorTask{}, nil
	}

	var tasks []TaskwarriorTask
	err := json.Unmarshal(jsonBytes, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task JSON: %w", err)
	}

	return tasks, nil
}
