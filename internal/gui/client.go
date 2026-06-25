// Package gui provides the web-based GUI for wui, including an HTTP server,
// HTML/CSS/JS assets, and an HTTP client adapter that implements core.TaskService
// by proxying calls to the wui serve REST API.
package gui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/clobrano/wui/internal/core"
)

// APIClient implements core.TaskService by forwarding calls to a running wui serve instance.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates an APIClient that targets the given base URL (e.g. "http://localhost:7007/api/v1").
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// apiError is returned when the server responds with a 4xx/5xx status.
type apiError struct {
	Status  int
	Message string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Status, e.Message)
}

// do performs an HTTP request and decodes the JSON response body into dest (may be nil).
// On 4xx/5xx it tries to decode {"error":"..."} and returns an *apiError.
func (c *APIClient) do(method, path string, body any, dest any) error {
	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		msg := errResp.Error
		if msg == "" {
			msg = resp.Status
		}
		return &apiError{Status: resp.StatusCode, Message: msg}
	}

	if dest != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// taskDTOToCore converts an API TaskDTO into a core.Task.
func taskDTOToCore(dto taskDTO) core.Task {
	t := core.Task{
		ID:          dto.ID,
		UUID:        dto.UUID,
		Description: dto.Description,
		Project:     dto.Project,
		Tags:        dto.Tags,
		Priority:    dto.Priority,
		Status:      dto.Status,
		Depends:     dto.Depends,
		UDAs:        dto.UDAs,
		Urgency:     dto.Urgency,
	}
	if dto.Entry != "" {
		if ts, err := time.Parse(time.RFC3339, dto.Entry); err == nil {
			t.Entry = ts
		}
	}
	t.Due = parseTimePtr(dto.Due)
	t.Scheduled = parseTimePtr(dto.Scheduled)
	t.Wait = parseTimePtr(dto.Wait)
	t.Start = parseTimePtr(dto.Start)
	t.Modified = parseTimePtr(dto.Modified)
	t.End = parseTimePtr(dto.End)

	if len(dto.Annotations) > 0 {
		t.Annotations = make([]core.Annotation, len(dto.Annotations))
		for i, a := range dto.Annotations {
			var ts time.Time
			if a.Entry != "" {
				ts, _ = time.Parse(time.RFC3339, a.Entry)
			}
			t.Annotations[i] = core.Annotation{Entry: ts, Description: a.Description}
		}
	}
	return t
}

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}

// local DTO types mirror api.TaskDTO / api.AnnotationDTO to avoid cross-package coupling.
type taskDTO struct {
	ID          int               `json:"id"`
	UUID        string            `json:"uuid"`
	Description string            `json:"description"`
	Project     string            `json:"project,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Priority    string            `json:"priority,omitempty"`
	Status      string            `json:"status"`
	Due         *string           `json:"due,omitempty"`
	Scheduled   *string           `json:"scheduled,omitempty"`
	Wait        *string           `json:"wait,omitempty"`
	Start       *string           `json:"start,omitempty"`
	Entry       string            `json:"entry"`
	Modified    *string           `json:"modified,omitempty"`
	End         *string           `json:"end,omitempty"`
	Depends     []string          `json:"depends,omitempty"`
	Annotations []annotationDTO   `json:"annotations,omitempty"`
	UDAs        map[string]string `json:"udas,omitempty"`
	Urgency     float64           `json:"urgency"`
}

type annotationDTO struct {
	Entry       string `json:"entry"`
	Description string `json:"description"`
}

// Export implements core.TaskService.
func (c *APIClient) Export(filter string) ([]core.Task, error) {
	path := "/tasks"
	if filter != "" {
		path += "?filter=" + url.QueryEscape(filter)
	}
	var dtos []taskDTO
	if err := c.do(http.MethodGet, path, nil, &dtos); err != nil {
		return nil, err
	}
	tasks := make([]core.Task, len(dtos))
	for i, d := range dtos {
		tasks[i] = taskDTOToCore(d)
	}
	return tasks, nil
}

// Add implements core.TaskService.
func (c *APIClient) Add(description string) (string, error) {
	var resp struct {
		UUID string `json:"uuid"`
	}
	err := c.do(http.MethodPost, "/tasks", map[string]string{"description": description}, &resp)
	return resp.UUID, err
}

// Modify implements core.TaskService.
func (c *APIClient) Modify(uuid, modifications string) error {
	return c.do(http.MethodPut, "/tasks/"+uuid, map[string]string{"modifications": modifications}, nil)
}

// Done implements core.TaskService.
func (c *APIClient) Done(uuid string) error {
	return c.do(http.MethodPost, "/tasks/"+uuid+"/done", nil, nil)
}

// Delete implements core.TaskService.
func (c *APIClient) Delete(uuid string) error {
	return c.do(http.MethodDelete, "/tasks/"+uuid, nil, nil)
}

// Start implements core.TaskService.
func (c *APIClient) Start(uuid string) error {
	return c.do(http.MethodPost, "/tasks/"+uuid+"/start", nil, nil)
}

// Stop implements core.TaskService.
func (c *APIClient) Stop(uuid string) error {
	return c.do(http.MethodPost, "/tasks/"+uuid+"/stop", nil, nil)
}

// Annotate implements core.TaskService.
func (c *APIClient) Annotate(uuid, text string) error {
	return c.do(http.MethodPost, "/tasks/"+uuid+"/annotate", map[string]string{"text": text}, nil)
}

// Denotate implements core.TaskService.
func (c *APIClient) Denotate(uuid, description string) error {
	return c.do(http.MethodDelete, "/tasks/"+uuid+"/annotate", map[string]string{"description": description}, nil)
}

// Undo implements core.TaskService.
func (c *APIClient) Undo() error {
	return c.do(http.MethodPost, "/undo", nil, nil)
}

// Edit is not supported via the HTTP API (TUI-only operation).
func (c *APIClient) Edit(_ string) error {
	return fmt.Errorf("Edit is not supported in the web GUI")
}

// GetProjectSummary implements core.TaskService.
func (c *APIClient) GetProjectSummary() ([]core.ProjectSummary, error) {
	var dtos []struct {
		Name       string `json:"name"`
		Percentage int    `json:"percentage"`
	}
	if err := c.do(http.MethodGet, "/projects", nil, &dtos); err != nil {
		return nil, err
	}
	result := make([]core.ProjectSummary, len(dtos))
	for i, d := range dtos {
		result[i] = core.ProjectSummary{Name: d.Name, Percentage: d.Percentage}
	}
	return result, nil
}

// GetTags implements core.TaskService.
func (c *APIClient) GetTags() ([]string, error) {
	var tags []string
	if err := c.do(http.MethodGet, "/tags", nil, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetUdas implements core.TaskService.
func (c *APIClient) GetUdas() ([]string, error) {
	var udas []string
	if err := c.do(http.MethodGet, "/udas", nil, &udas); err != nil {
		return nil, err
	}
	return udas, nil
}

// GetVersion implements core.TaskService.
func (c *APIClient) GetVersion() (string, error) {
	var resp struct {
		WuiVersion  string `json:"wui_version"`
		TaskVersion string `json:"task_version"`
	}
	if err := c.do(http.MethodGet, "/version", nil, &resp); err != nil {
		return "", err
	}
	return resp.TaskVersion, nil
}
