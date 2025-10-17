package taskwarrior

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
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
	// Split filter into separate arguments for proper parsing
	// Taskwarrior syntax: task [filter] [command]
	filterArgs := strings.Fields(filter)
	args := append(filterArgs, "export")
	args = c.buildArgs(args...)

	slog.Debug("Exporting tasks", "filter", filter)

	output, err := c.runCommand(args...)
	if err != nil {
		slog.Error("Failed to export tasks", "error", err, "filter", filter)
		return nil, fmt.Errorf("failed to export tasks: %w", err)
	}

	// Log raw output for debugging JSON parsing issues
	slog.Debug("Raw taskwarrior output", "output_preview", string(output[:min(500, len(output))]))

	// Parse JSON output
	tasks, err := ParseTaskJSON(output)
	if err != nil {
		slog.Error("Failed to parse task JSON",
			"error", err,
			"output_preview", string(output[:min(500, len(output))]))
		return nil, fmt.Errorf("failed to parse task JSON: %w", err)
	}

	slog.Info("Successfully exported tasks", "count", len(tasks))

	// Map to core.Task
	coreTasks := make([]core.Task, len(tasks))
	for i, t := range tasks {
		coreTasks[i] = MapToCore(t)
	}

	return coreTasks, nil
}

// Modify updates a task with the given modifications
func (c *Client) Modify(uuid, modifications string) error {
	// Split modifications into separate arguments so taskwarrior parses them correctly
	// e.g., "project:home +duties" becomes ["project:home", "+duties"]
	modArgs := strings.Fields(modifications)
	args := append([]string{uuid, "modify"}, modArgs...)
	args = c.buildArgs(args...)
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
	// Split description into separate arguments so taskwarrior parses them correctly
	// e.g., "Buy milk project:home +shopping" becomes ["Buy", "milk", "project:home", "+shopping"]
	descArgs := strings.Fields(description)
	args := append([]string{"add"}, descArgs...)
	args = c.buildArgs(args...)
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
		cmd.Env = append(os.Environ(), fmt.Sprintf("TASKRC=%s", c.taskrcPath))
	}

	// Edit requires interactive terminal access
	// This will be handled by the TUI suspension mechanism
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to edit task %s: %w", uuid, err)
	}
	return nil
}

// Start marks a task as started (active)
func (c *Client) Start(uuid string) error {
	args := c.buildArgs(uuid, "start")
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to start task %s: %w", uuid, err)
	}
	return nil
}

// Stop marks a task as stopped (pending)
func (c *Client) Stop(uuid string) error {
	args := c.buildArgs(uuid, "stop")
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to stop task %s: %w", uuid, err)
	}
	return nil
}

// buildArgs constructs command-line arguments for taskwarrior
// It handles the taskrc path configuration
func (c *Client) buildArgs(args ...string) []string {
	// Simply return the arguments as-is
	// The taskrc path is handled via TASKRC environment variable in runCommand
	return args
}

// runCommand executes a taskwarrior command and returns the output
func (c *Client) runCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(c.taskBin, args...)

	// Set TASKRC environment variable if taskrcPath is specified
	if c.taskrcPath != "" {
		// Preserve existing environment and add TASKRC
		cmd.Env = append(os.Environ(), fmt.Sprintf("TASKRC=%s", c.taskrcPath))
	}

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Log the command being executed
	slog.Debug("Executing taskwarrior command",
		"bin", c.taskBin,
		"args", args,
		"taskrc", c.taskrcPath)

	err := cmd.Run()

	// Log stderr if present (informational messages from taskwarrior)
	if stderr.Len() > 0 {
		slog.Debug("Taskwarrior stderr output",
			"stderr", strings.TrimSpace(stderr.String()))
	}

	if err != nil {
		// Log error with full context
		slog.Error("Taskwarrior command failed",
			"error", err,
			"stderr", strings.TrimSpace(stderr.String()),
			"stdout_preview", strings.TrimSpace(stdout.String()[:min(200, stdout.Len())]))
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	// Log successful execution
	slog.Debug("Taskwarrior command succeeded",
		"output_size", stdout.Len())

	return stdout.Bytes(), nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
