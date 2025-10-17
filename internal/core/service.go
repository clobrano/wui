package core

// TaskService defines the interface for task operations
// This interface abstracts the underlying task management system (Taskwarrior, etc.)
// and allows different UI implementations (TUI, Web, GUI) to share the same business logic
type TaskService interface {
	// Export retrieves tasks matching the given filter
	// Filter uses Taskwarrior filter syntax (e.g., "status:pending +work")
	// Returns a slice of tasks or an error if the operation fails
	Export(filter string) ([]Task, error)

	// Modify updates a task with the given modifications
	// Modifications use Taskwarrior modify syntax (e.g., "priority:H due:tomorrow")
	// Returns an error if the task is not found or the modification fails
	Modify(uuid, modifications string) error

	// Annotate adds an annotation (note) to a task
	// Returns an error if the task is not found or the annotation fails
	Annotate(uuid, text string) error

	// Done marks a task as completed
	// Returns an error if the task is not found or already completed
	Done(uuid string) error

	// Delete removes a task
	// Returns an error if the task is not found or deletion fails
	Delete(uuid string) error

	// Add creates a new task with the given description
	// Description can include task attributes (e.g., "Buy milk project:home +shopping")
	// Returns the UUID of the newly created task or an error if creation fails
	Add(description string) (string, error)

	// Undo reverts the last task operation
	// Returns an error if there is nothing to undo or the operation fails
	Undo() error

	// Edit opens the task in an external editor for manual editing
	// This typically suspends the TUI and launches the configured editor
	// Returns an error if the task is not found or the edit fails
	Edit(uuid string) error

	// Start marks a task as started (active)
	// Returns an error if the task is not found or already started
	Start(uuid string) error

	// Stop marks a task as stopped (pending)
	// Returns an error if the task is not found or not started
	Stop(uuid string) error
}
