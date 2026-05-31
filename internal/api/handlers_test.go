package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clobrano/wui/internal/core"
)

// newTestServer wires a MockTaskService into a real mux, matching production routing.
func newTestServer(svc core.TaskService) http.Handler {
	h := &handlers{svc: svc}
	mux := http.NewServeMux()
	registerRoutes(mux, h)
	return mux
}

func TestListTags(t *testing.T) {
	tests := []struct {
		name       string
		mockFunc   func() ([]string, error)
		wantStatus int
		wantTags   []string
	}{
		{
			name: "returns tags",
			mockFunc: func() ([]string, error) {
				return []string{"home", "work", "urgent"}, nil
			},
			wantStatus: http.StatusOK,
			wantTags:   []string{"home", "work", "urgent"},
		},
		{
			name: "empty tag list",
			mockFunc: func() ([]string, error) {
				return []string{}, nil
			},
			wantStatus: http.StatusOK,
			wantTags:   []string{},
		},
		{
			name: "service error",
			mockFunc: func() ([]string, error) {
				return nil, errors.New("taskwarrior failed")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &core.MockTaskService{GetTagsFunc: tt.mockFunc}
			srv := newTestServer(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantTags != nil {
				var got []string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(got) != len(tt.wantTags) {
					t.Errorf("tags len = %d, want %d", len(got), len(tt.wantTags))
				}
				for i, tag := range tt.wantTags {
					if got[i] != tag {
						t.Errorf("tags[%d] = %q, want %q", i, got[i], tag)
					}
				}
			}
		})
	}
}

func TestListUdas(t *testing.T) {
	tests := []struct {
		name       string
		mockFunc   func() ([]string, error)
		wantStatus int
		wantUdas   []string
	}{
		{
			name: "returns UDAs",
			mockFunc: func() ([]string, error) {
				return []string{"estimate", "reviewed"}, nil
			},
			wantStatus: http.StatusOK,
			wantUdas:   []string{"estimate", "reviewed"},
		},
		{
			name: "no UDAs configured",
			mockFunc: func() ([]string, error) {
				return []string{}, nil
			},
			wantStatus: http.StatusOK,
			wantUdas:   []string{},
		},
		{
			name: "service error",
			mockFunc: func() ([]string, error) {
				return nil, errors.New("taskwarrior failed")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &core.MockTaskService{GetUdasFunc: tt.mockFunc}
			srv := newTestServer(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/udas", nil)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantUdas != nil {
				var got []string
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(got) != len(tt.wantUdas) {
					t.Errorf("udas len = %d, want %d", len(got), len(tt.wantUdas))
				}
				for i, uda := range tt.wantUdas {
					if got[i] != uda {
						t.Errorf("udas[%d] = %q, want %q", i, got[i], uda)
					}
				}
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name            string
		mockFunc        func() (string, error)
		wantStatus      int
		wantTaskVersion string
	}{
		{
			name: "returns version",
			mockFunc: func() (string, error) {
				return "3.0.2", nil
			},
			wantStatus:      http.StatusOK,
			wantTaskVersion: "3.0.2",
		},
		{
			name: "service error",
			mockFunc: func() (string, error) {
				return "", errors.New("task binary not found")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &core.MockTaskService{GetVersionFunc: tt.mockFunc}
			srv := newTestServer(svc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantTaskVersion != "" {
				var got VersionResponse
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if got.TaskVersion != tt.wantTaskVersion {
					t.Errorf("task_version = %q, want %q", got.TaskVersion, tt.wantTaskVersion)
				}
				if got.WuiVersion == "" {
					t.Error("wui_version should not be empty")
				}
			}
		})
	}
}

func TestDenotateTask(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		body       any
		mockFunc   func(uuid, description string) error
		wantStatus int
	}{
		{
			name: "removes annotation",
			uuid: "abc-123",
			body: DenotateRequest{Description: "fix the bug"},
			mockFunc: func(uuid, description string) error {
				if uuid == "abc-123" && description == "fix the bug" {
					return nil
				}
				return errors.New("not found")
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing uuid",
			uuid:       "",
			body:       DenotateRequest{Description: "fix the bug"},
			wantStatus: http.StatusMovedPermanently, // mux cleans double-slash path, redirects
		},
		{
			name:       "invalid JSON body",
			uuid:       "abc-123",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty description",
			uuid:       "abc-123",
			body:       DenotateRequest{Description: ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			uuid: "abc-123",
			body: DenotateRequest{Description: "fix the bug"},
			mockFunc: func(uuid, description string) error {
				return errors.New("taskwarrior failed")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &core.MockTaskService{DenotateFunc: tt.mockFunc}
			srv := newTestServer(svc)

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				if err := json.NewEncoder(&buf).Encode(tt.body); err != nil {
					t.Fatalf("encode body: %v", err)
				}
			}

			path := "/api/v1/tasks/" + tt.uuid + "/annotate"
			req := httptest.NewRequest(http.MethodDelete, path, &buf)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
