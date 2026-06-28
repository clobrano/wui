package taskwarrior

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// Client implements core.TaskService for Taskwarrior
type Client struct {
	taskBin       string
	taskrcPath    string
	wuiConfigPath string // passed to "wui sync" subprocess; empty → default config
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

// SetWuiConfigPath tells the client which wui config file to pass when running
// the "wui sync" subprocess.
func (c *Client) SetWuiConfigPath(p string) { c.wuiConfigPath = p }

// Export retrieves tasks matching the given filter
func (c *Client) Export(filter string) ([]core.Task, error) {
	// Split filter into separate arguments for proper parsing
	// Taskwarrior syntax: task [filter] [command]
	filterArgs := strings.Fields(filter)
	args := append(filterArgs, "export")

	// TODO: Add sorting support
	// Allow changing sort order from wui (e.g., via keybinding or config)
	// Examples:
	//   args = append(args, "rc.report.export.sort=urgency-")
	//   args = append(args, "rc.report.export.sort=due+,priority-")
	//   args = append(args, "rc.report.export.sort=start-,urgency-")
	// Could be configurable per-section or globally via Settings

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
		return nil, fmt.Errorf("failed to parse task data (corrupted JSON): %w\nPlease check your Taskwarrior database integrity with 'task diagnostics'", err)
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
var createdTaskIDRe = regexp.MustCompile(`Created task (\d+)`)

func (c *Client) Add(description string) (string, error) {
	descArgs := strings.Fields(description)
	args := append([]string{"add"}, descArgs...)
	args = c.buildArgs(args...)
	output, err := c.runCommand(args...)
	if err != nil {
		return "", fmt.Errorf("failed to add task: %w", err)
	}

	// Parse the numeric task ID from "Created task N." in stdout, then export
	// that specific task to get its UUID — avoids the unreliable "most urgent" heuristic.
	if m := createdTaskIDRe.FindSubmatch(output); m != nil {
		taskID := string(m[1])
		tasks, err := c.Export(taskID)
		if err == nil && len(tasks) > 0 {
			return tasks[0].UUID, nil
		}
		slog.Warn("Could not export task by ID after add", "taskID", taskID, "err", err)
	} else {
		slog.Warn("Could not parse task ID from 'task add' output", "output", string(output))
	}

	return "", errors.New("could not determine UUID of the newly created task")
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

// GetTags returns all tags currently in use across tasks
func (c *Client) GetTags() ([]string, error) {
	args := c.buildArgs("_tags")
	output, err := c.runCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	return parseLines(output), nil
}

// GetUdas returns the names of all User Defined Attributes configured in taskwarrior
func (c *Client) GetUdas() ([]string, error) {
	args := c.buildArgs("_udas")
	output, err := c.runCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get UDAs: %w", err)
	}
	return parseLines(output), nil
}

// GetVersion returns the version string of the underlying taskwarrior installation
func (c *Client) GetVersion() (string, error) {
	cmd := exec.Command(c.taskBin, "--version")
	if c.taskrcPath != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("TASKRC=%s", c.taskrcPath))
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get taskwarrior version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Denotate removes an annotation from a task matching the given description
func (c *Client) Denotate(uuid, description string) error {
	args := c.buildArgs(uuid, "denotate", description)
	_, err := c.runCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to denotate task %s: %w", uuid, err)
	}
	return nil
}

// Sync runs "wui sync" as a subprocess to push tasks to Google Calendar.
// Errors are logged but not returned — a failed sync never blocks a task operation.
func (c *Client) Sync() error {
	exe, err := os.Executable()
	if err != nil {
		slog.Warn("sync: cannot locate wui binary", "error", err)
		return nil
	}
	args := []string{"sync"}
	if c.wuiConfigPath != "" {
		args = append(args, "--config", c.wuiConfigPath)
	}
	cmd := exec.Command(exe, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("wui sync failed", "error", err, "output", string(out))
	} else {
		slog.Info("wui sync completed", "output", string(out))
	}
	return nil
}

// parseLines splits newline-delimited output into a trimmed, non-empty slice
func parseLines(output []byte) []string {
	var result []string
	for _, line := range strings.Split(string(output), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}
	return result
}

// GetProjectSummary retrieves project completion data from task summary
func (c *Client) GetProjectSummary() ([]core.ProjectSummary, error) {
	args := c.buildArgs("summary")

	slog.Debug("Getting project summary")

	output, err := c.runCommand(args...)
	if err != nil {
		slog.Error("Failed to get project summary", "error", err)
		return nil, fmt.Errorf("failed to get project summary: %w", err)
	}

	// Parse the summary output
	summaries, err := ParseSummaryOutput(output)
	if err != nil {
		slog.Error("Failed to parse summary output", "error", err)
		return nil, fmt.Errorf("failed to parse summary output: %w", err)
	}

	slog.Info("Successfully retrieved project summary", "count", len(summaries))
	return summaries, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
