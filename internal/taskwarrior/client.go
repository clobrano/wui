package taskwarrior

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// Client implements core.TaskService for Taskwarrior
type Client struct {
	taskBin    string
	taskrcPath string
}

// NewClient creates a new Taskwarrior client
// taskBin is the path to the task binary (required)
// taskrcPath is the path to the .taskrc file (optional)
func NewClient(taskBin, taskrcPath string) (*Client, error) {
	if taskBin == "" {
		return nil, errors.New("task binary path cannot be empty")
	}

	return &Client{
		taskBin:    taskBin,
		taskrcPath: taskrcPath,
	}, nil
}

// Export retrieves tasks matching the given filter
func (c *Client) Export(filter string) ([]core.Task, error) {
	args := c.buildArgs("export", filter)
	output, err := c.runCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to export tasks: %w", err)
	}

	// Parse JSON output
	tasks, err := ParseTaskJSON(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task JSON: %w", err)
	}

	// Map to core.Task
	coreTasks := make([]core.Task, len(tasks))
	for i, t := range tasks {
		coreTasks[i] = MapToCore(t)
	}

	return coreTasks, nil
}

// Modify updates a task with the given modifications
func (c *Client) Modify(uuid, modifications string) error {
	args := c.buildArgs(uuid, "modify", modifications)
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to modify task %s: %w", uuid, err)
	}
	return nil
}

// Annotate adds an annotation to a task
func (c *Client) Annotate(uuid, text string) error {
	args := c.buildArgs(uuid, "annotate", text)
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to annotate task %s: %w", uuid, err)
	}
	return nil
}

// Done marks a task as completed
func (c *Client) Done(uuid string) error {
	args := c.buildArgs(uuid, "done")
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to mark task %s as done: %w", uuid, err)
	}
	return nil
}

// Delete removes a task
func (c *Client) Delete(uuid string) error {
	args := c.buildArgs(uuid, "delete", "rc.confirmation=off")
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to delete task %s: %w", uuid, err)
	}
	return nil
}

// Add creates a new task
func (c *Client) Add(description string) (string, error) {
	args := c.buildArgs("add", description)
	_, err := c.runCommand(args...)
	if err != nil {
		return "", fmt.Errorf("failed to add task: %w", err)
	}

	// Extract UUID from output
	// Taskwarrior outputs: "Created task <id>."
	// We need to export the task to get the UUID
	// For now, we'll parse it from the last task added
	// A better approach would be to export with specific filter

	// Run export to get the newly created task
	// Use a heuristic: get the most recent task
	tasks, err := c.Export("status:pending limit:1")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve new task UUID: %w", err)
	}

	if len(tasks) == 0 {
		return "", errors.New("no task found after add")
	}

	return tasks[0].UUID, nil
}

// Undo reverts the last task operation
func (c *Client) Undo() error {
	args := c.buildArgs("undo", "rc.confirmation=off")
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to undo: %w", err)
	}
	return nil
}

// Edit opens the task in an external editor
func (c *Client) Edit(uuid string) error {
	args := c.buildArgs(uuid, "edit")
	cmd := exec.Command(c.taskBin, args...)
	if c.taskrcPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TASKRC=%s", c.taskrcPath))
	}

	// Edit requires interactive terminal access
	// This will be handled by the TUI suspension mechanism
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to edit task %s: %w", uuid, err)
	}
	return nil
}

// buildArgs constructs command-line arguments for taskwarrior
// It handles the taskrc path configuration
func (c *Client) buildArgs(args ...string) []string {
	result := make([]string, 0, len(args)+1)

	// Add rc.confirmation=off for non-interactive operations
	// This is added to specific commands in their respective methods

	// Add taskrc path if specified
	if c.taskrcPath != "" {
		result = append(result, fmt.Sprintf("rc:%s", c.taskrcPath))
	}

	// Add the actual command arguments
	result = append(result, args...)

	return result
}

// runCommand executes a taskwarrior command and returns the output
func (c *Client) runCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(c.taskBin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include command output in error for debugging
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

// Placeholder functions for parser and mapper
// These will be implemented in separate files

// TaskwarriorTask represents a task as returned by Taskwarrior JSON export
type TaskwarriorTask struct {
	UUID        string                 `json:"uuid"`
	Description string                 `json:"description"`
	Project     string                 `json:"project,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Priority    string                 `json:"priority,omitempty"`
	Status      string                 `json:"status"`
	Due         string                 `json:"due,omitempty"`
	Scheduled   string                 `json:"scheduled,omitempty"`
	Wait        string                 `json:"wait,omitempty"`
	Entry       string                 `json:"entry"`
	Modified    string                 `json:"modified,omitempty"`
	End         string                 `json:"end,omitempty"`
	Depends     string                 `json:"depends,omitempty"`
	Annotations []TaskwarriorAnnotation `json:"annotations,omitempty"`
	UDA         map[string]interface{} `json:"-"` // Populated from unmapped JSON fields
	Urgency     float64                `json:"urgency"`
}

// TaskwarriorAnnotation represents an annotation in Taskwarrior format
type TaskwarriorAnnotation struct {
	Entry       string `json:"entry"`
	Description string `json:"description"`
}

// ParseTaskJSON parses Taskwarrior JSON export output
// This is a placeholder - will be fully implemented in parser.go
func ParseTaskJSON(jsonBytes []byte) ([]TaskwarriorTask, error) {
	var tasks []TaskwarriorTask
	err := json.Unmarshal(jsonBytes, &tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task JSON: %w", err)
	}
	return tasks, nil
}

// MapToCore converts a TaskwarriorTask to a core.Task
// This is a placeholder - will be fully implemented in mapper.go
func MapToCore(t TaskwarriorTask) core.Task {
	// Basic mapping for now
	// Full implementation will handle date parsing, depends, etc.
	return core.Task{
		UUID:        t.UUID,
		Description: t.Description,
		Project:     t.Project,
		Tags:        t.Tags,
		Priority:    t.Priority,
		Status:      t.Status,
		Urgency:     t.Urgency,
	}
}
