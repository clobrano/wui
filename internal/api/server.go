// Package api provides a REST/JSON HTTP server that exposes the core.TaskService
// interface over the network. This allows any HTTP client — Flutter, web, CLI tools —
// to drive the same business logic that the TUI uses locally.
//
// API base path: /api/v1
//
//	GET    /api/v1/tasks                    list tasks (optional ?filter= query param)
//	POST   /api/v1/tasks                    create a task
//	PUT    /api/v1/tasks/{uuid}              modify a task
//	DELETE /api/v1/tasks/{uuid}              delete a task
//	POST   /api/v1/tasks/{uuid}/done         mark a task done
//	POST   /api/v1/tasks/{uuid}/start        start a task
//	POST   /api/v1/tasks/{uuid}/stop         stop a task
//	POST   /api/v1/tasks/{uuid}/annotate     add an annotation
//	DELETE /api/v1/tasks/{uuid}/annotate     remove an annotation (denotate)
//	POST   /api/v1/undo                     undo last operation
//	GET    /api/v1/projects                 list project summaries
//	GET    /api/v1/tags                     list all tags in use
//	GET    /api/v1/udas                     list User Defined Attribute names
//	GET    /api/v1/version                  wui and taskwarrior version info
package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/clobrano/wui/internal/core"
)

// Server is the HTTP API server.
type Server struct {
	httpServer *http.Server
	tlsCert    string
	tlsKey     string
}

// NewServer creates a new Server that wraps the given TaskService and listens on addr.
// addr format: "host:port", e.g. "localhost:7007" or ":7007".
// When tlsCert and tlsKey are both non-empty the server uses TLS (HTTPS).
func NewServer(svc core.TaskService, addr, tlsCert, tlsKey string) *Server {
	h := &handlers{svc: svc}

	mux := http.NewServeMux()
	registerRoutes(mux, h)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      corsMiddleware(mux),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		tlsCert: tlsCert,
		tlsKey:  tlsKey,
	}
}

// corsMiddleware adds permissive CORS headers so browser-based clients
// (e.g. the Flutter web build) can reach the API from any origin.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// registerRoutes wires all API endpoints into the mux.
// Uses Go 1.22+ pattern matching: "METHOD /path/{param}".
func registerRoutes(mux *http.ServeMux, h *handlers) {
	mux.HandleFunc("GET /api/v1/tasks", h.listTasks)
	mux.HandleFunc("POST /api/v1/tasks", h.addTask)
	mux.HandleFunc("PUT /api/v1/tasks/{uuid}", h.modifyTask)
	mux.HandleFunc("DELETE /api/v1/tasks/{uuid}", h.deleteTask)
	mux.HandleFunc("POST /api/v1/tasks/{uuid}/done", h.doneTask)
	mux.HandleFunc("POST /api/v1/tasks/{uuid}/start", h.startTask)
	mux.HandleFunc("POST /api/v1/tasks/{uuid}/stop", h.stopTask)
	mux.HandleFunc("POST /api/v1/tasks/{uuid}/annotate", h.annotateTask)
	mux.HandleFunc("DELETE /api/v1/tasks/{uuid}/annotate", h.denotateTask)
	mux.HandleFunc("POST /api/v1/undo", h.undoLast)
	mux.HandleFunc("GET /api/v1/projects", h.listProjects)
	mux.HandleFunc("GET /api/v1/tags", h.listTags)
	mux.HandleFunc("GET /api/v1/udas", h.listUdas)
	mux.HandleFunc("GET /api/v1/version", h.getVersion)
}

// Start begins listening and serving requests. It blocks until the server stops.
// Uses TLS if both tlsCert and tlsKey were provided to NewServer.
func (s *Server) Start() error {
	scheme := "http"
	if s.tlsCert != "" && s.tlsKey != "" {
		scheme = "https"
	}
	slog.Info("API server listening", "addr", s.httpServer.Addr, "scheme", scheme)
	fmt.Printf("wui API server listening on %s://%s\n", scheme, s.httpServer.Addr)

	var err error
	if scheme == "https" {
		err = s.httpServer.ListenAndServeTLS(s.tlsCert, s.tlsKey)
	} else {
		err = s.httpServer.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server within the given timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down API server")
	return s.httpServer.Shutdown(ctx)
}
