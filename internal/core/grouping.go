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

// GroupProjectsByHierarchy creates a two-level project hierarchy from summary data
// Shows only main projects (no dots) and first-level subprojects (exactly one dot)
// Filters out deeper nested projects (e.g., "Activity1.Sub1.SubA")
func GroupProjectsByHierarchy(summaries []ProjectSummary, tasks []Task) []TaskGroup {
	if len(summaries) == 0 {
		return []TaskGroup{}
	}

	// Filter to two-level hierarchy
	var filtered []ProjectSummary
	for _, summary := range summaries {
		dotCount := strings.Count(summary.Name, ".")
		if dotCount <= 1 {
			filtered = append(filtered, summary)
		}
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
	mainProjects := make(map[string]bool)

	// First pass: identify main projects and create groups
	for _, summary := range filtered {
		isSubproject := strings.Contains(summary.Name, ".")

		if !isSubproject {
			mainProjects[summary.Name] = true
		}

		// Get tasks for this exact project
		projectTasks := taskMap[summary.Name]
		count := len(projectTasks)

		groups = append(groups, TaskGroup{
			Name:       summary.Name,
			Count:      count,
			Tasks:      projectTasks,
			Percentage: summary.Percentage,
			IsSubitem:  isSubproject,
		})
	}

	// Sort: main projects first (alphabetically), then their subprojects
	sort.Slice(groups, func(i, j int) bool {
		// Both are main projects or both are subprojects
		if groups[i].IsSubitem == groups[j].IsSubitem {
			return groups[i].Name < groups[j].Name
		}

		// Main project comes before subproject
		if !groups[i].IsSubitem {
			return true
		}
		return false
	})

	// Better sorting: group subprojects under their parent
	var sorted []TaskGroup
	for _, group := range groups {
		if !group.IsSubitem {
			// Add main project
			sorted = append(sorted, group)

			// Add all subprojects of this main project
			for _, subgroup := range groups {
				if subgroup.IsSubitem && strings.HasPrefix(subgroup.Name, group.Name+".") {
					sorted = append(sorted, subgroup)
				}
			}
		}
	}

	return sorted
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
