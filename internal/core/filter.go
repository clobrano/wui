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

// Tab represents a tab/section configuration
type Tab struct {
	Name   string
	Filter string
}

// TabsToSections converts Tab configs to Section objects
func TabsToSections(tabs []Tab) []Section {
	sections := make([]Section, 0, len(tabs))

	for _, tab := range tabs {
		sections = append(sections, Section{
			Name:        tab.Name,
			Filter:      tab.Filter,
			Description: tab.Name + " tasks",
		})
	}

	return sections
}

// Deprecated: Use TabsToSections instead
// Bookmark represents a saved filter bookmark (deprecated, use Tab)
type Bookmark struct {
	Name   string
	Filter string
}

// Deprecated: Use TabsToSections with config.DefaultTabs() instead
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

// Deprecated: Use TabsToSections instead
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
