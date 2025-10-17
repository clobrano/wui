package core

import (
	"sort"
	"strings"
)

// TaskGroup represents a group of tasks (by project or tag)
type TaskGroup struct {
	Name  string
	Count int
	Tasks []Task
}

// ExtractMainProject extracts the main project from a nested project name
// For example: "Work.company1.backend" -> "Work"
func ExtractMainProject(project string) string {
	if project == "" {
		return ""
	}

	parts := strings.Split(project, ".")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}

	return parts[0]
}

// GroupByProject groups tasks by their main project
// Nested projects (e.g., "Work.company1") are grouped under the main project ("Work")
func GroupByProject(tasks []Task) []TaskGroup {
	if len(tasks) == 0 {
		return []TaskGroup{}
	}

	// Map of project name to tasks
	projectMap := make(map[string][]Task)

	for _, task := range tasks {
		mainProject := ExtractMainProject(task.Project)
		if mainProject == "" {
			mainProject = "(none)"
		}
		projectMap[mainProject] = append(projectMap[mainProject], task)
	}

	// Convert map to sorted slice
	var groups []TaskGroup
	for projectName, projectTasks := range projectMap {
		groups = append(groups, TaskGroup{
			Name:  projectName,
			Count: len(projectTasks),
			Tasks: projectTasks,
		})
	}

	// Sort by name, but put "(none)" at the end
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Name == "(none)" {
			return false
		}
		if groups[j].Name == "(none)" {
			return true
		}
		return groups[i].Name < groups[j].Name
	})

	return groups
}

// GroupByTag groups tasks by their tags
// Tasks with multiple tags will appear in multiple groups
func GroupByTag(tasks []Task) []TaskGroup {
	if len(tasks) == 0 {
		return []TaskGroup{}
	}

	// Map of tag name to tasks
	tagMap := make(map[string][]Task)

	for _, task := range tasks {
		if len(task.Tags) == 0 {
			// Task has no tags
			tagMap["(none)"] = append(tagMap["(none)"], task)
		} else {
			// Add task to each tag group
			for _, tag := range task.Tags {
				tagMap[tag] = append(tagMap[tag], task)
			}
		}
	}

	// Convert map to sorted slice
	var groups []TaskGroup
	for tagName, tagTasks := range tagMap {
		groups = append(groups, TaskGroup{
			Name:  tagName,
			Count: len(tagTasks),
			Tasks: tagTasks,
		})
	}

	// Sort by name, but put "(none)" at the end
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Name == "(none)" {
			return false
		}
		if groups[j].Name == "(none)" {
			return true
		}
		return groups[i].Name < groups[j].Name
	})

	return groups
}
