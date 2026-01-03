package tui

import "github.com/clobrano/wui/internal/core"

// TasksLoadedMsg is sent when tasks have been loaded from the service
type TasksLoadedMsg struct {
	Tasks []core.Task
	Err   error
}

// TaskModifiedMsg is sent when a task has been modified
type TaskModifiedMsg struct {
	UUID string
	Err  error
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err error
}

// RefreshMsg is sent to trigger a task list refresh
type RefreshMsg struct{}

// StatusMsg is sent to display a status message to the user
type StatusMsg struct {
	Message string
	IsError bool
}

// ProjectSummaryLoadedMsg is sent when project summaries have been loaded
type ProjectSummaryLoadedMsg struct {
	Summaries []core.ProjectSummary
	Err       error
}

// CalendarSyncCompletedMsg is sent when calendar sync completes
type CalendarSyncCompletedMsg struct {
	Err error
}
