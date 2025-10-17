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
			Filter:      "( status:pending or status:active ) -WAITING",
			Description: "Next tasks to work on (pending or active, not waiting)",
		},
		{
			Name:        "Waiting",
			Filter:      "status:waiting",
			Description: "Tasks waiting on something",
		},
		{
			Name:        "Projects",
			Filter:      "status:pending or status:active",
			Description: "Tasks grouped by project",
		},
		{
			Name:        "Tags",
			Filter:      "status:pending or status:active",
			Description: "Tasks grouped by tags",
		},
		{
			Name:        "All",
			Filter:      "status:pending or status:active",
			Description: "All pending and active tasks",
		},
	}
}

// SectionsWithBookmarks returns sections including bookmarks
func SectionsWithBookmarks(bookmarks []Bookmark) []Section {
	sections := DefaultSections()

	// Add bookmarks as sections
	for _, bookmark := range bookmarks {
		sections = append(sections, Section{
			Name:        bookmark.Name,
			Filter:      bookmark.Filter,
			Description: "Saved filter",
		})
	}

	return sections
}
