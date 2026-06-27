package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

// taskListData is the template data for the task list page and partial.
type taskListData struct {
	Tabs          []config.Tab
	ActiveTab     string
	Tasks         []core.Task
	FilterHistory []string
	ActiveFilter  string
	ShowFilter    bool
}

// taskDetailData is the template data for the task detail page.
type taskDetailData struct {
	Task         core.Task
	ActiveTab    string
	DependsTasks []core.Task
	Recur        *recurInfo
	UDAs         map[string]string
}

type recurInfo struct {
	Interval string
	Type     string
	Until    string
}

// taskFormData is the template data for the create/edit form.
type taskFormData struct {
	Task         taskFormTask
	IsNew        bool
	Action       string
	CancelURL    string
	ActiveTab    string
	Projects     []core.ProjectSummary
	DependsTasks []core.Task
	Error        string
	RawValue     string
}

// taskFormTask holds editable fields used by the form template.
type taskFormTask struct {
	UUID        string
	Description string
	Project     string
	Tags        []string
	Priority    string
	Due         *time.Time
	Scheduled   *time.Time
	Wait        *time.Time
	Until       *time.Time
	Recur       string
}

func (s *Server) activeTab(r *http.Request) string {
	tab := r.URL.Query().Get("tab")
	if tab == "" && len(s.cfg.TUI.Tabs) > 0 {
		tab = s.cfg.TUI.Tabs[0].Name
	}
	return tab
}

func (s *Server) filterForTab(tabName string) string {
	for _, t := range s.cfg.TUI.Tabs {
		if t.Name == tabName {
			return t.Filter
		}
	}
	return ""
}

func (s *Server) sortForTab(tabName string) (method string, reverse bool) {
	for _, t := range s.cfg.TUI.Tabs {
		if t.Name == tabName {
			return t.Sort, t.Reverse
		}
	}
	return "", false
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data any) {
	tmpl, ok := s.templates[name]
	if !ok {
		http.Error(w, "no template: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		// Log to stderr so it's visible regardless of log-level config.
		fmt.Fprintf(os.Stderr, "gui: template %s error: %v\n", name, err)
	}
}

// handleTaskList serves GET / and GET /tasks (full page).
func (s *Server) handleTaskList(w http.ResponseWriter, r *http.Request) {
	tab := s.activeTab(r)
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = s.filterForTab(tab)
	}

	tasks, err := s.svc.Export(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sortMethod, reverse := s.sortForTab(tab)
	tasks = SortTasks(tasks, sortMethod, reverse)

	history, _ := s.filterHistory.Load()

	data := taskListData{
		Tabs:          s.cfg.TUI.Tabs,
		ActiveTab:     tab,
		Tasks:         tasks,
		FilterHistory: history,
		ActiveFilter:  r.URL.Query().Get("filter"),
	}
	s.renderTemplate(w, "tasklist.html", data)
}

// handleTaskListPartial serves GET /api/gui/tasks (HTMX partial).
func (s *Server) handleTaskListPartial(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	filter := r.URL.Query().Get("filter")
	q := r.URL.Query().Get("q")

	tabFilter := s.filterForTab(tab)
	switch {
	case q != "" && tabFilter != "":
		filter = "( " + tabFilter + " ) description.contains:" + q
	case q != "":
		filter = "description.contains:" + q
	case filter == "":
		filter = tabFilter
	}

	tasks, err := s.svc.Export(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sortMethod, reverse := s.sortForTab(tab)
	tasks = SortTasks(tasks, sortMethod, reverse)

	data := taskListData{
		Tabs:         s.cfg.TUI.Tabs,
		ActiveTab:    tab,
		Tasks:        tasks,
		ActiveFilter: r.URL.Query().Get("filter"),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// task-rows is defined inside tasklist.html, so use that template set.
	if err := s.templates["tasklist.html"].ExecuteTemplate(w, "task-rows", data); err != nil {
		fmt.Fprintf(os.Stderr, "gui: template task-rows error: %v\n", err)
	}
}

// handleTaskDetail serves GET /tasks/{uuid}.
func (s *Server) handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	tab := r.URL.Query().Get("tab")

	tasks, err := s.svc.Export(fmt.Sprintf("uuid:%s", uuid))
	if err != nil || len(tasks) == 0 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	task := tasks[0]

	// Resolve dependency descriptions.
	var depTasks []core.Task
	for _, depUUID := range task.Depends {
		deps, err := s.svc.Export(fmt.Sprintf("uuid:%s", depUUID))
		if err == nil && len(deps) > 0 {
			depTasks = append(depTasks, deps[0])
		}
	}

	// Build recur info from UDAs if present.
	var ri *recurInfo
	if task.UDAs != nil {
		interval := task.UDAs["recur"]
		if interval != "" {
			ri = &recurInfo{
				Interval: interval,
				Type:     task.UDAs["rtype"],
				Until:    task.UDAs["until"],
			}
		}
	}

	// Separate UDAs from known recur fields.
	udas := map[string]string{}
	for k, v := range task.UDAs {
		if k != "recur" && k != "rtype" && k != "until" {
			udas[k] = v
		}
	}
	if len(udas) == 0 {
		udas = nil
	}

	data := taskDetailData{
		Task:         task,
		ActiveTab:    tab,
		DependsTasks: depTasks,
		Recur:        ri,
		UDAs:         udas,
	}
	s.renderTemplate(w, "taskdetail.html", data)
}

// handleNewTaskForm serves GET /tasks/new.
func (s *Server) handleNewTaskForm(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	projects, _ := s.svc.GetProjectSummary()

	data := taskFormData{
		IsNew:     true,
		Action:    "/tasks/new",
		CancelURL: fmt.Sprintf("/?tab=%s", tab),
		ActiveTab: tab,
		Projects:  projects,
	}
	s.renderTemplate(w, "taskform.html", data)
}

// handleCreateTask handles POST /tasks/new.
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	tab := r.FormValue("tab")

	description, args := s.buildTaskArgs(r, true)
	if description == "" {
		s.renderFormWithError(w, r, tab, true, "Description is required")
		return
	}

	raw := r.FormValue("raw")
	var taskSpec string
	if raw != "" {
		taskSpec = raw
	} else {
		taskSpec = description
		if args != "" {
			taskSpec += " " + args
		}
	}

	uuid, err := s.svc.Add(taskSpec)
	if err != nil {
		s.renderFormWithError(w, r, tab, true, err.Error())
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/tasks/%s?tab=%s", uuid, tab), http.StatusSeeOther)
}

// handleEditTaskForm serves GET /tasks/{uuid}/edit.
func (s *Server) handleEditTaskForm(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	tab := r.URL.Query().Get("tab")

	tasks, err := s.svc.Export(fmt.Sprintf("uuid:%s", uuid))
	if err != nil || len(tasks) == 0 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	task := tasks[0]
	projects, _ := s.svc.GetProjectSummary()

	// Resolve dependency tasks for chips.
	var depTasks []core.Task
	for _, depUUID := range task.Depends {
		deps, _ := s.svc.Export(fmt.Sprintf("uuid:%s", depUUID))
		if len(deps) > 0 {
			depTasks = append(depTasks, deps[0])
		}
	}

	data := taskFormData{
		Task: taskFormTask{
			UUID:        task.UUID,
			Description: task.Description,
			Project:     task.Project,
			Tags:        task.Tags,
			Priority:    task.Priority,
			Due:         task.Due,
			Scheduled:   task.Scheduled,
			Wait:        task.Wait,
		},
		IsNew:        false,
		Action:       fmt.Sprintf("/tasks/%s/edit", uuid),
		CancelURL:    fmt.Sprintf("/tasks/%s?tab=%s", uuid, tab),
		ActiveTab:    tab,
		Projects:     projects,
		DependsTasks: depTasks,
	}
	s.renderTemplate(w, "taskform.html", data)
}

// handleUpdateTask handles POST /tasks/{uuid}/edit.
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	tab := r.FormValue("tab")

	raw := r.FormValue("raw")
	var mods string
	if raw != "" {
		mods = raw
	} else {
		_, mods = s.buildTaskArgs(r, false)
	}

	if mods != "" {
		if err := s.svc.Modify(uuid, mods); err != nil {
			s.renderFormWithError(w, r, tab, false, err.Error())
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/tasks/%s?tab=%s", uuid, tab), http.StatusSeeOther)
}

// buildTaskArgs assembles form values into a description + modification string.
func (s *Server) buildTaskArgs(r *http.Request, _ bool) (description, mods string) {
	description = strings.TrimSpace(r.FormValue("description"))

	var parts []string
	if project := r.FormValue("project"); project != "" {
		parts = append(parts, "project:"+project)
	}
	if priority := r.FormValue("priority"); priority != "" {
		parts = append(parts, "priority:"+priority)
	}
	if due := r.FormValue("due"); due != "" {
		parts = append(parts, "due:"+due)
	}
	if sched := r.FormValue("scheduled"); sched != "" {
		parts = append(parts, "scheduled:"+sched)
	}
	if wait := r.FormValue("wait"); wait != "" {
		parts = append(parts, "wait:"+wait)
	}
	if until := r.FormValue("until"); until != "" {
		parts = append(parts, "until:"+until)
	}
	if recur := r.FormValue("recur"); recur != "" {
		parts = append(parts, "recur:"+recur)
	}
	for _, tag := range r.Form["tags"] {
		if tag != "" {
			parts = append(parts, "+"+tag)
		}
	}
	for _, dep := range r.Form["depends"] {
		if dep != "" {
			parts = append(parts, "depends:"+dep)
		}
	}

	mods = strings.Join(parts, " ")
	return
}

func (s *Server) renderFormWithError(w http.ResponseWriter, r *http.Request, tab string, isNew bool, errMsg string) {
	projects, _ := s.svc.GetProjectSummary()
	action := "/tasks/new"
	cancelURL := fmt.Sprintf("/?tab=%s", tab)
	if !isNew {
		uuid := r.PathValue("uuid")
		action = fmt.Sprintf("/tasks/%s/edit", uuid)
		cancelURL = fmt.Sprintf("/tasks/%s?tab=%s", uuid, tab)
	}
	data := taskFormData{
		IsNew:     isNew,
		Action:    action,
		CancelURL: cancelURL,
		ActiveTab: tab,
		Projects:  projects,
		Error:     errMsg,
	}
	w.WriteHeader(http.StatusUnprocessableEntity)
	s.renderTemplate(w, "taskform.html", data)
}

// handleFilterHistoryAdd handles POST /api/gui/filter-history.
func (s *Server) handleFilterHistoryAdd(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Filter string `json:"filter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Filter == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	_ = s.filterHistory.Prepend(body.Filter)
	w.WriteHeader(http.StatusNoContent)
}

// handleFilterHistoryDelete handles DELETE /api/gui/filter-history.
func (s *Server) handleFilterHistoryDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Filter string `json:"filter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Filter == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	_ = s.filterHistory.Delete(body.Filter)
	w.WriteHeader(http.StatusNoContent)
}

// handleFilterHistoryClear handles POST /api/gui/filter-history/clear.
func (s *Server) handleFilterHistoryClear(w http.ResponseWriter, r *http.Request) {
	_ = s.filterHistory.Clear()
	w.WriteHeader(http.StatusNoContent)
}

// UntilPtr returns the Until field (used by templates).
func (t *taskFormTask) UntilPtr() *time.Time { return t.Until }
