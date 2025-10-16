package core

// Filter represents a saved filter with optional activation state
type Filter struct {
	Text   string
	Active bool
}

// Section represents a view section with a name, filter, and description
type Section struct {
	Name        string
	Filter      string
	Description string
}

// Bookmark represents a saved filter bookmark
type Bookmark struct {
	Name   string
	Filter string
}

// DefaultSections returns the default sections for the task list view
func DefaultSections() []Section {
	return []Section{
		{
			Name:        "Next",
			Filter:      "status:pending -WAITING",
			Description: "Next tasks to work on (pending, not waiting)",
		},
		{
			Name:        "Waiting",
			Filter:      "status:waiting",
			Description: "Tasks waiting on something",
		},
		{
			Name:        "Projects",
			Filter:      "status:pending",
			Description: "Tasks grouped by project",
		},
		{
			Name:        "Tags",
			Filter:      "status:pending",
			Description: "Tasks grouped by tags",
		},
		{
			Name:        "All",
			Filter:      "status:pending",
			Description: "All pending tasks",
		},
	}
}
