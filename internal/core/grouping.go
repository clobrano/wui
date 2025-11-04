package core

import (
	"sort"
	"strings"
)

// ProjectSummary represents project completion data from task summary
type ProjectSummary struct {
	Name       string // Full project name (e.g., "Home.Cleaning")
	Percentage int    // Completion percentage (0-100)
}

// TaskGroup represents a group of tasks (by project or tag)
type TaskGroup struct {
	Name       string
	Count      int
	Tasks      []Task
	Percentage int  // Completion percentage (0-100), -1 if not applicable
	IsSubitem  bool // True if this is a subproject (indented)
	Depth      int  // Nesting depth (0 = main project, 1 = first level sub, etc.)
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

// GroupProjectsByHierarchy creates a hierarchical project list from summary data
// Shows all projects at all nesting levels with proper indentation
func GroupProjectsByHierarchy(summaries []ProjectSummary, tasks []Task) []TaskGroup {
	if len(summaries) == 0 {
		return []TaskGroup{}
	}

	// Build a map of project name to tasks for counting
	taskMap := make(map[string][]Task)
	for _, task := range tasks {
		if task.Project != "" {
			taskMap[task.Project] = append(taskMap[task.Project], task)
		}
	}

	// Create groups with hierarchy information
	var groups []TaskGroup

	for _, summary := range summaries {
		// Calculate depth based on number of dots
		depth := strings.Count(summary.Name, ".")
		isSubproject := depth > 0

		// Get tasks for this exact project
		projectTasks := taskMap[summary.Name]
		count := len(projectTasks)

		groups = append(groups, TaskGroup{
			Name:       summary.Name,
			Count:      count,
			Tasks:      projectTasks,
			Percentage: summary.Percentage,
			IsSubitem:  isSubproject,
			Depth:      depth,
		})
	}

	// Sort hierarchically: maintain parent-child relationships
	// This is a stable sort that keeps the order from task summary output
	// which already has the correct hierarchical structure
	return groups
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
