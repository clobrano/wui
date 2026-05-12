package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// handlers holds the TaskService and provides HTTP handler methods.
type handlers struct {
	svc core.TaskService
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// listTasks handles GET /api/v1/tasks?filter=<filter>
func (h *handlers) listTasks(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	tasks, err := h.svc.Export(filter)
	if err != nil {
		slog.Error("Export failed", "filter", filter, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]TaskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = taskToDTO(t)
	}
	writeJSON(w, http.StatusOK, dtos)
}

// addTask handles POST /api/v1/tasks
func (h *handlers) addTask(w http.ResponseWriter, r *http.Request) {
	var req AddTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Description) == "" {
		writeError(w, http.StatusBadRequest, "description is required")
		return
	}

	uuid, err := h.svc.Add(req.Description)
	if err != nil {
		slog.Error("Add task failed", "description", req.Description, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, AddTaskResponse{UUID: uuid})
}

// modifyTask handles PUT /api/v1/tasks/{uuid}
func (h *handlers) modifyTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	var req ModifyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Modifications) == "" {
		writeError(w, http.StatusBadRequest, "modifications is required")
		return
	}

	if err := h.svc.Modify(uuid, req.Modifications); err != nil {
		slog.Error("Modify task failed", "uuid", uuid, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// annotateTask handles POST /api/v1/tasks/{uuid}/annotate
func (h *handlers) annotateTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	var req AnnotateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}

	if err := h.svc.Annotate(uuid, req.Text); err != nil {
		slog.Error("Annotate task failed", "uuid", uuid, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// doneTask handles POST /api/v1/tasks/{uuid}/done
func (h *handlers) doneTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	if err := h.svc.Done(uuid); err != nil {
		slog.Error("Done task failed", "uuid", uuid, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteTask handles DELETE /api/v1/tasks/{uuid}
func (h *handlers) deleteTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	if err := h.svc.Delete(uuid); err != nil {
		slog.Error("Delete task failed", "uuid", uuid, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// startTask handles POST /api/v1/tasks/{uuid}/start
func (h *handlers) startTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	if err := h.svc.Start(uuid); err != nil {
		slog.Error("Start task failed", "uuid", uuid, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// stopTask handles POST /api/v1/tasks/{uuid}/stop
func (h *handlers) stopTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	if err := h.svc.Stop(uuid); err != nil {
		slog.Error("Stop task failed", "uuid", uuid, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// undoLast handles POST /api/v1/undo
func (h *handlers) undoLast(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Undo(); err != nil {
		slog.Error("Undo failed", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listProjects handles GET /api/v1/projects
func (h *handlers) listProjects(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.svc.GetProjectSummary()
	if err != nil {
		slog.Error("GetProjectSummary failed", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]ProjectSummaryDTO, len(summaries))
	for i, s := range summaries {
		dtos[i] = projectSummaryToDTO(s)
	}
	writeJSON(w, http.StatusOK, dtos)
}
