package tui

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// FilePathMatch represents a file path found in an annotation
type FilePathMatch struct {
	Path       string // The actual file path (expanded and normalized)
	Annotation string // The full annotation text containing this path
}

// ExtractFilePathsFromAnnotations extracts all file paths from a task's annotations
// Supports:
// - Absolute paths: /home/user/file.txt, /tmp/data
// - Paths with spaces (quoted): "/home/user/my documents/file.txt"
// - Paths with escaped spaces: /home/user/my\ documents/file.txt
// - Home directory paths: ~/documents/file.txt
// - Relative paths: ./file.txt, ../parent/file
// - file:// URLs: file:///home/user/file.txt (with %20 for spaces)
func ExtractFilePathsFromAnnotations(task *core.Task) []FilePathMatch {
	if task == nil || len(task.Annotations) == 0 {
		return nil
	}

	var paths []FilePathMatch
	seen := make(map[string]bool) // Deduplicate by normalized path

	for _, annotation := range task.Annotations {
		extractedPaths := extractFilePathsFromText(annotation.Description)
		for _, path := range extractedPaths {
			// Normalize the path for deduplication
			normalizedPath := normalizePath(path)
			if !seen[normalizedPath] {
				seen[normalizedPath] = true
				paths = append(paths, FilePathMatch{
					Path:       normalizedPath,
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

	addPath := func(path string) {
		normalized := normalizePath(path)
		if normalized != "" && !seen[normalized] && isValidFilePath(normalized) {
			seen[normalized] = true
			paths = append(paths, normalized)
		}
	}

	// Pattern 1: file:// URLs (may contain %20 for spaces)
	fileURLRegex := regexp.MustCompile(`file://(/[^\s<>\[\]]+)`)
	fileURLMatches := fileURLRegex.FindAllStringSubmatch(text, -1)
	for _, match := range fileURLMatches {
		if len(match) >= 2 {
			path := match[1]
			// Decode URL-encoded characters (like %20 for spaces)
			if decoded, err := url.PathUnescape(path); err == nil {
				path = decoded
			}
			path = cleanPathTrailing(path)
			addPath(path)
		}
	}

	// Pattern 2: Double-quoted paths (supports spaces)
	doubleQuotedRegex := regexp.MustCompile(`"((?:/|~/|\.\.?/)[^"]+)"`)
	doubleQuotedMatches := doubleQuotedRegex.FindAllStringSubmatch(text, -1)
	for _, match := range doubleQuotedMatches {
		if len(match) >= 2 {
			path := match[1]
			addPath(path)
		}
	}

	// Pattern 3: Single-quoted paths (supports spaces)
	singleQuotedRegex := regexp.MustCompile(`'((?:/|~/|\.\.?/)[^']+)'`)
	singleQuotedMatches := singleQuotedRegex.FindAllStringSubmatch(text, -1)
	for _, match := range singleQuotedMatches {
		if len(match) >= 2 {
			path := match[1]
			addPath(path)
		}
	}

	// Pattern 4: Paths with escaped spaces (backslash before space)
	// Match paths that contain \  (backslash-space) sequences
	escapedSpaceRegex := regexp.MustCompile(`(?:^|[^a-zA-Z0-9"'])((?:/|~/|\.\.?/)(?:[^\s"'<>\[\]]|\\ )+)`)
	escapedMatches := escapedSpaceRegex.FindAllStringSubmatch(text, -1)
	for _, match := range escapedMatches {
		if len(match) >= 2 {
			path := match[1]
			// Only process if it actually contains escaped spaces
			if strings.Contains(path, `\ `) {
				// Unescape the spaces
				path = strings.ReplaceAll(path, `\ `, " ")
				path = cleanPathTrailing(path)
				addPath(path)
			}
		}
	}

	// Pattern 5: Absolute Unix paths without spaces (starting with /)
	// Must start with / followed by a word character to avoid matching standalone /
	absolutePathRegex := regexp.MustCompile(`(?:^|[^a-zA-Z0-9"'/\\])(/[a-zA-Z0-9_][^\s<>\[\]"']*[^\s<>\[\]"'.,;:!?)])`)
	absoluteMatches := absolutePathRegex.FindAllStringSubmatch(text, -1)
	for _, match := range absoluteMatches {
		if len(match) >= 2 {
			path := cleanPathTrailing(match[1])
			// Skip if this looks like it was already captured as a quoted path
			if !strings.Contains(path, `\ `) {
				addPath(path)
			}
		}
	}

	// Pattern 6: Home directory paths without spaces (starting with ~/)
	homePathRegex := regexp.MustCompile(`(?:^|[^a-zA-Z0-9"'\\])(~/[^\s<>\[\]"']+)`)
	homeMatches := homePathRegex.FindAllStringSubmatch(text, -1)
	for _, match := range homeMatches {
		if len(match) >= 2 {
			path := cleanPathTrailing(match[1])
			addPath(path)
		}
	}

	// Pattern 7: Relative paths without spaces (starting with ./ or ../)
	relativePathRegex := regexp.MustCompile(`(?:^|[^a-zA-Z0-9"'\\])(\.\.?/[^\s<>\[\]"']+)`)
	relativeMatches := relativePathRegex.FindAllStringSubmatch(text, -1)
	for _, match := range relativeMatches {
		if len(match) >= 2 {
			path := cleanPathTrailing(match[1])
			addPath(path)
		}
	}

	return paths
}

// normalizePath normalizes a path by expanding home directory and converting to absolute
func normalizePath(path string) string {
	if path == "" {
		return ""
	}

	// Expand home directory
	if strings.HasPrefix(path, "~/") {
		path = expandHomePath(path)
	}

	// Convert relative paths to absolute
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		absPath, err := filepath.Abs(path)
		if err == nil {
			path = absPath
		}
	}

	// Clean the path (removes redundant slashes, etc.)
	path = filepath.Clean(path)

	return path
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
