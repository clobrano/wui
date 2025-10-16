package taskwarrior

import (
	"encoding/json"
	"fmt"
)

// TaskwarriorTask represents a task as returned by Taskwarrior JSON export
type TaskwarriorTask struct {
	UUID        string                   `json:"uuid"`
	Description string                   `json:"description"`
	Project     string                   `json:"project,omitempty"`
	Tags        []string                 `json:"tags,omitempty"`
	Priority    string                   `json:"priority,omitempty"`
	Status      string                   `json:"status"`
	Due         string                   `json:"due,omitempty"`
	Scheduled   string                   `json:"scheduled,omitempty"`
	Wait        string                   `json:"wait,omitempty"`
	Entry       string                   `json:"entry"`
	Modified    string                   `json:"modified,omitempty"`
	End         string                   `json:"end,omitempty"`
	Depends     string                   `json:"depends,omitempty"`
	Annotations []TaskwarriorAnnotation  `json:"annotations,omitempty"`
	UDA         map[string]interface{}   `json:"-"` // Populated from unmapped JSON fields
	Urgency     float64                  `json:"urgency"`
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
