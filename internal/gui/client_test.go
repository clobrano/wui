package gui_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clobrano/wui/internal/gui"
)

func newTestServer(t *testing.T, handler http.Handler) (*httptest.Server, *gui.APIClient) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := gui.NewAPIClient(srv.URL + "/api/v1")
	return srv, client
}

func TestAPIClient_Export(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tasks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"uuid":"abc","description":"Test","status":"pending","entry":"2024-01-01T00:00:00Z","urgency":1.5}]`))
	}))

	tasks, err := client.Export("")
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].UUID != "abc" {
		t.Errorf("expected UUID abc, got %s", tasks[0].UUID)
	}
}

func TestAPIClient_Export_WithFilter(t *testing.T) {
	var gotQuery string
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))

	_, err := client.Export("status:pending")
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if gotQuery == "" {
		t.Error("expected query string with filter")
	}
}

func TestAPIClient_Add(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"uuid":"new-uuid"}`))
	}))

	uuid, err := client.Add("New task")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if uuid != "new-uuid" {
		t.Errorf("expected new-uuid, got %s", uuid)
	}
}

func TestAPIClient_Done(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tasks/some-uuid/done" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.Done("some-uuid"); err != nil {
		t.Fatalf("Done: %v", err)
	}
}

func TestAPIClient_HTTPError(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
	}))

	err := client.Delete("missing-uuid")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestAPIClient_GetVersion(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/version" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"wui_version":"0.1.0","task_version":"2.6.3"}`))
	}))

	v, err := client.GetVersion()
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	if v != "2.6.3" {
		t.Errorf("expected 2.6.3, got %s", v)
	}
}

func TestAPIClient_Undo(t *testing.T) {
	_, client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/undo" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	if err := client.Undo(); err != nil {
		t.Fatalf("Undo: %v", err)
	}
}
