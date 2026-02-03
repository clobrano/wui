package tui

import (
	"github.com/clobrano/wui/internal/core"
)

// ResourceType indicates whether a resource is a URL or a file path
type ResourceType int

const (
	ResourceTypeURL ResourceType = iota
	ResourceTypeFile
)

// ResourceMatch represents a resource (URL or file path) found in an annotation
type ResourceMatch struct {
	Resource   string       // The actual URL or file path
	Annotation string       // The full annotation text containing this resource
	Type       ResourceType // Whether this is a URL or file path
}

// FormatForDisplay formats a ResourceMatch for display in the picker
func (r ResourceMatch) FormatForDisplay() string {
	return r.Annotation
}

// ExtractResourcesFromAnnotations extracts all URLs and file paths from a task's annotations
// Returns a combined, deduplicated list of resources
func ExtractResourcesFromAnnotations(task *core.Task) []ResourceMatch {
	if task == nil || len(task.Annotations) == 0 {
		return nil
	}

	var resources []ResourceMatch
	seen := make(map[string]bool) // Deduplicate by resource value

	// Extract URLs first
	urls := ExtractURLsFromAnnotations(task)
	for _, u := range urls {
		if !seen[u.URL] {
			seen[u.URL] = true
			resources = append(resources, ResourceMatch{
				Resource:   u.URL,
				Annotation: u.Annotation,
				Type:       ResourceTypeURL,
			})
		}
	}

	// Extract file paths
	filePaths := ExtractFilePathsFromAnnotations(task)
	for _, f := range filePaths {
		if !seen[f.Path] {
			seen[f.Path] = true
			resources = append(resources, ResourceMatch{
				Resource:   f.Path,
				Annotation: f.Annotation,
				Type:       ResourceTypeFile,
			})
		}
	}

	return resources
}
