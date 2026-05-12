package api

import (
	"time"

	"github.com/clobrano/wui/internal/core"
)

// TaskDTO is the JSON representation of a core.Task for the REST API.
// Time fields use RFC3339 strings so any client (Flutter, web, etc.) can parse them easily.
type TaskDTO struct {
	ID          int              `json:"id"`
	UUID        string           `json:"uuid"`
	Description string           `json:"description"`
	Project     string           `json:"project,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Priority    string           `json:"priority,omitempty"`
	Status      string           `json:"status"`
	Due         *string          `json:"due,omitempty"`
	Scheduled   *string          `json:"scheduled,omitempty"`
	Wait        *string          `json:"wait,omitempty"`
	Start       *string          `json:"start,omitempty"`
	Entry       string           `json:"entry"`
	Modified    *string          `json:"modified,omitempty"`
	End         *string          `json:"end,omitempty"`
	Depends     []string         `json:"depends,omitempty"`
	Annotations []AnnotationDTO  `json:"annotations,omitempty"`
	UDAs        map[string]string `json:"udas,omitempty"`
	Urgency     float64          `json:"urgency"`
}

// AnnotationDTO is the JSON representation of a core.Annotation.
type AnnotationDTO struct {
	Entry       string `json:"entry"`
	Description string `json:"description"`
}

// ProjectSummaryDTO is the JSON representation of a core.ProjectSummary.
type ProjectSummaryDTO struct {
	Name       string `json:"name"`
	Percentage int    `json:"percentage"`
}

// AddTaskRequest is the body for POST /api/v1/tasks.
type AddTaskRequest struct {
	Description string `json:"description"`
}

// AddTaskResponse is the response for POST /api/v1/tasks.
type AddTaskResponse struct {
	UUID string `json:"uuid"`
}

// ModifyTaskRequest is the body for PUT /api/v1/tasks/{uuid}.
// Modifications use Taskwarrior syntax, e.g. "priority:H due:tomorrow +urgent".
type ModifyTaskRequest struct {
	Modifications string `json:"modifications"`
}

// AnnotateTaskRequest is the body for POST /api/v1/tasks/{uuid}/annotate.
type AnnotateTaskRequest struct {
	Text string `json:"text"`
}

// ErrorResponse is returned for all error cases.
type ErrorResponse struct {
	Error string `json:"error"`
}

// taskToDTO converts a core.Task to a TaskDTO.
func taskToDTO(t core.Task) TaskDTO {
	dto := TaskDTO{
		ID:          t.ID,
		UUID:        t.UUID,
		Description: t.Description,
		Project:     t.Project,
		Tags:        t.Tags,
		Priority:    t.Priority,
		Status:      t.Status,
		Depends:     t.Depends,
		UDAs:        t.UDAs,
		Urgency:     t.Urgency,
		Entry:       t.Entry.UTC().Format(time.RFC3339),
	}

	dto.Due = formatTimePtr(t.Due)
	dto.Scheduled = formatTimePtr(t.Scheduled)
	dto.Wait = formatTimePtr(t.Wait)
	dto.Start = formatTimePtr(t.Start)
	dto.Modified = formatTimePtr(t.Modified)
	dto.End = formatTimePtr(t.End)

	if len(t.Annotations) > 0 {
		dto.Annotations = make([]AnnotationDTO, len(t.Annotations))
		for i, a := range t.Annotations {
			dto.Annotations[i] = AnnotationDTO{
				Entry:       a.Entry.UTC().Format(time.RFC3339),
				Description: a.Description,
			}
		}
	}

	return dto
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

// projectSummaryToDTO converts a core.ProjectSummary to a ProjectSummaryDTO.
func projectSummaryToDTO(p core.ProjectSummary) ProjectSummaryDTO {
	return ProjectSummaryDTO{
		Name:       p.Name,
		Percentage: p.Percentage,
	}
}
