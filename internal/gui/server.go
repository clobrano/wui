package gui

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Server is the web GUI HTTP server.
type Server struct {
	svc           core.TaskService
	cfg           *config.Config
	filterHistory *FilterHistory
	httpServer    *http.Server
	tmpl          *template.Template
}

// NewServer creates a new GUI server.
func NewServer(svc core.TaskService, cfg *config.Config, fh *FilterHistory) *Server {
	s := &Server{
		svc:           svc,
		cfg:           cfg,
		filterHistory: fh,
	}
	s.tmpl = s.buildTemplates()
	return s
}

func (s *Server) buildTemplates() *template.Template {
	funcs := template.FuncMap{
		"lower":         strings.ToLower,
		"shortRelDate":  ShortRelDate,
		"longRelDate":   func(t time.Time) string { return LongRelDate(&t) },
		"longRelDatePtr": LongRelDate,
		"absDate":       func(t time.Time) string { return FormatAbsDate(&t) },
		"absDatePtr":    FormatAbsDate,
		"dueCSSClass":   DueCSSClass,
		"joinTags": func(tags []string) string {
			return strings.Join(tags, " ")
		},
		"urlify":          urlify,
		"formatDateInput": formatDateInput,
		"not": func(b bool) bool { return !b },
	}

	tmpl := template.New("").Funcs(funcs)
	tmpl = template.Must(tmpl.ParseFS(templateFS, "templates/*.html"))
	return tmpl
}

// Start begins listening and serving requests. It blocks until the server stops.
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("GUI server listening", "addr", addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("GUI server error: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Static assets
	mux.Handle("GET /static/", http.FileServer(http.FS(staticFS)))

	// GUI pages
	mux.HandleFunc("GET /", s.handleTaskList)
	mux.HandleFunc("GET /tasks", s.handleTaskList)
	mux.HandleFunc("GET /tasks/new", s.handleNewTaskForm)
	mux.HandleFunc("POST /tasks/new", s.handleCreateTask)
	mux.HandleFunc("GET /tasks/{uuid}/edit", s.handleEditTaskForm)
	mux.HandleFunc("POST /tasks/{uuid}/edit", s.handleUpdateTask)
	mux.HandleFunc("GET /tasks/{uuid}", s.handleTaskDetail)

	// HTMX partials
	mux.HandleFunc("GET /api/gui/tasks", s.handleTaskListPartial)

	// Filter history API (called from JS)
	mux.HandleFunc("POST /api/gui/filter-history", s.handleFilterHistoryAdd)
	mux.HandleFunc("DELETE /api/gui/filter-history", s.handleFilterHistoryDelete)
	mux.HandleFunc("POST /api/gui/filter-history/clear", s.handleFilterHistoryClear)
}

// urlify converts URLs in text into <a> tags.
func urlify(text string) template.HTML {
	escaped := template.HTMLEscapeString(text)
	// Simple URL detection: replace http(s):// links.
	words := strings.Fields(escaped)
	for i, w := range words {
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			words[i] = fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener">%s</a>`, w, w)
		}
	}
	return template.HTML(strings.Join(words, " "))
}

// formatDateInput formats a *time.Time for an HTML date input (YYYY-MM-DD).
func formatDateInput(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Local().Format("2006-01-02")
}
