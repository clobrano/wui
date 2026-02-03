package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// FilePathMatch represents a file path found in an annotation
type FilePathMatch struct {
	Path       string // The actual file path (expanded)
	Annotation string // The full annotation text containing this path
}

// ExtractFilePathsFromAnnotations extracts all file paths from a task's annotations
// Supports:
// - Absolute paths: /home/user/file.txt, /tmp/data
// - Home directory paths: ~/documents/file.txt
// - Relative paths: ./file.txt, ../parent/file
// - file:// URLs: file:///home/user/file.txt
func ExtractFilePathsFromAnnotations(task *core.Task) []FilePathMatch {
	if task == nil || len(task.Annotations) == 0 {
		return nil
	}

	var paths []FilePathMatch
	seen := make(map[string]bool) // Deduplicate paths

	for _, annotation := range task.Annotations {
		extractedPaths := extractFilePathsFromText(annotation.Description)
		for _, path := range extractedPaths {
			if !seen[path] {
				seen[path] = true
				paths = append(paths, FilePathMatch{
					Path:       path,
					Annotation: annotation.Description,
				})
			}
		}
	}

	return paths
}

// extractFilePathsFromText extracts file paths from a text string
// Returns a slice of path strings (deduplicated and expanded)
func extractFilePathsFromText(text string) []string {
	var paths []string
	seen := make(map[string]bool)

	// Pattern 1: file:// URLs
	fileURLRegex := regexp.MustCompile(`file://(/[^\s<>\[\]"']+)`)
	fileURLMatches := fileURLRegex.FindAllStringSubmatch(text, -1)
	for _, match := range fileURLMatches {
		if len(match) >= 2 {
			path := cleanPathTrailing(match[1])
			if !seen[path] && isValidFilePath(path) {
				seen[path] = true
				paths = append(paths, path)
			}
		}
	}

	// Pattern 2: Absolute Unix paths (starting with /)
	// Must start with / followed by a word character to avoid matching standalone /
	absolutePathRegex := regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(/[a-zA-Z0-9_][^\s<>\[\]"']*[^\s<>\[\]"'.,;:!?)])`)
	absoluteMatches := absolutePathRegex.FindAllStringSubmatch(text, -1)
	for _, match := range absoluteMatches {
		if len(match) >= 2 {
			path := cleanPathTrailing(match[1])
			if !seen[path] && isValidFilePath(path) {
				seen[path] = true
				paths = append(paths, path)
			}
		}
	}

	// Pattern 3: Home directory paths (starting with ~/)
	homePathRegex := regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(~/[^\s<>\[\]"']+)`)
	homeMatches := homePathRegex.FindAllStringSubmatch(text, -1)
	for _, match := range homeMatches {
		if len(match) >= 2 {
			path := cleanPathTrailing(match[1])
			expandedPath := expandHomePath(path)
			if !seen[expandedPath] && isValidFilePath(expandedPath) {
				seen[expandedPath] = true
				paths = append(paths, expandedPath)
			}
		}
	}

	// Pattern 4: Relative paths (starting with ./ or ../)
	relativePathRegex := regexp.MustCompile(`(?:^|[^a-zA-Z0-9])(\.\.?/[^\s<>\[\]"']+)`)
	relativeMatches := relativePathRegex.FindAllStringSubmatch(text, -1)
	for _, match := range relativeMatches {
		if len(match) >= 2 {
			path := cleanPathTrailing(match[1])
			// Convert to absolute path
			absPath, err := filepath.Abs(path)
			if err == nil && !seen[absPath] && isValidFilePath(absPath) {
				seen[absPath] = true
				paths = append(paths, absPath)
			}
		}
	}

	return paths
}

// expandHomePath expands ~ to the user's home directory
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// cleanPathTrailing removes trailing punctuation that's likely not part of the path
func cleanPathTrailing(path string) string {
	// Balance parentheses - if there are more closing than opening, trim them
	openParens := strings.Count(path, "(")
	closeParens := strings.Count(path, ")")
	for closeParens > openParens && strings.HasSuffix(path, ")") {
		path = strings.TrimSuffix(path, ")")
		closeParens--
	}

	// Remove common trailing punctuation
	for {
		trimmed := false
		for _, suffix := range []string{".", ",", ";", ":", "!", "?"} {
			if strings.HasSuffix(path, suffix) {
				path = strings.TrimSuffix(path, suffix)
				trimmed = true
			}
		}
		if !trimmed {
			break
		}
	}

	return path
}

// isValidFilePath checks if a path looks like a valid file path
// This is a basic check to filter out false positives
func isValidFilePath(path string) bool {
	// Must not be empty
	if path == "" {
		return false
	}

	// Must start with / for absolute paths
	if !strings.HasPrefix(path, "/") {
		return false
	}

	// Should not contain certain characters that are unlikely in file paths
	invalidChars := []string{"<", ">", "|", "\x00"}
	for _, ch := range invalidChars {
		if strings.Contains(path, ch) {
			return false
		}
	}

	// Path should have at least 2 components (not just /)
	if path == "/" {
		return false
	}

	return true
}

// FormatForDisplay formats a FilePathMatch for display in the picker
// Shows the full annotation text so users can recognize which path they want
func (f FilePathMatch) FormatForDisplay() string {
	return f.Annotation
}
