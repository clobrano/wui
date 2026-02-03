package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/clobrano/wui/internal/core"
)

func TestExtractFilePathsFromText(t *testing.T) {
	// Get home directory for test expectations
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "absolute path",
			text:     "Check the file /home/user/documents/file.txt",
			expected: []string{"/home/user/documents/file.txt"},
		},
		{
			name:     "absolute path with extension",
			text:     "See /var/log/syslog for details",
			expected: []string{"/var/log/syslog"},
		},
		{
			name:     "file:// URL",
			text:     "Open file:///home/user/data.json",
			expected: []string{"/home/user/data.json"},
		},
		{
			name:     "file:// URL with encoded spaces",
			text:     "Open file:///home/user/my%20documents/file.txt",
			expected: []string{"/home/user/my documents/file.txt"},
		},
		{
			name:     "home directory path",
			text:     "Check ~/documents/report.pdf",
			expected: []string{filepath.Join(homeDir, "documents/report.pdf")},
		},
		{
			name:     "multiple paths",
			text:     "Files: /home/user/file1.txt and /tmp/file2.log",
			expected: []string{"/home/user/file1.txt", "/tmp/file2.log"},
		},
		{
			name:     "path ending with punctuation",
			text:     "See /home/user/readme.md.",
			expected: []string{"/home/user/readme.md"},
		},
		{
			name:     "path with spaces in text",
			text:     "The log file is at /var/log/app.log, check it",
			expected: []string{"/var/log/app.log"},
		},
		{
			name:     "no paths",
			text:     "This is just plain text without any file paths",
			expected: []string{},
		},
		{
			name:     "path with parentheses",
			text:     "See /home/user/dir(1)/file.txt for more",
			expected: []string{"/home/user/dir(1)/file.txt"},
		},
		{
			name:     "path with underscores and dashes",
			text:     "Check /home/user/my_project/some-file.txt",
			expected: []string{"/home/user/my_project/some-file.txt"},
		},
		{
			name:     "tmp directory path",
			text:     "Temp file at /tmp/cache/data.tmp",
			expected: []string{"/tmp/cache/data.tmp"},
		},
		{
			name:     "etc directory path",
			text:     "Config at /etc/app/config.yaml",
			expected: []string{"/etc/app/config.yaml"},
		},
		{
			name:     "should not match standalone slash",
			text:     "Use / for root or // for comments",
			expected: []string{},
		},
		{
			name:     "should not match URLs as file paths",
			text:     "Visit https://example.com/path/to/page",
			expected: []string{},
		},
		// Tests for paths with spaces
		{
			name:     "double-quoted path with spaces",
			text:     `Check the file "/home/user/my documents/file.txt"`,
			expected: []string{"/home/user/my documents/file.txt"},
		},
		{
			name:     "single-quoted path with spaces",
			text:     `Check the file '/home/user/my documents/file.txt'`,
			expected: []string{"/home/user/my documents/file.txt"},
		},
		{
			name:     "escaped spaces in path",
			text:     `Check the file /home/user/my\ documents/file.txt`,
			expected: []string{"/home/user/my documents/file.txt"},
		},
		{
			name:     "double-quoted home path with spaces",
			text:     `See "~/my documents/report.pdf"`,
			expected: []string{filepath.Join(homeDir, "my documents/report.pdf")},
		},
		{
			name:     "multiple spaces in quoted path",
			text:     `Open "/home/user/a b c/d e f.txt"`,
			expected: []string{"/home/user/a b c/d e f.txt"},
		},
		{
			name:     "path with spaces and special chars",
			text:     `File at "/home/user/my docs (2024)/report-v1.pdf"`,
			expected: []string{"/home/user/my docs (2024)/report-v1.pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFilePathsFromText(tt.text)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d paths, got %d", len(tt.expected), len(result))
				for i, r := range result {
					t.Errorf("  result[%d]: %q", i, r)
				}
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("path mismatch at index %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestExtractFilePathsFromAnnotations(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name               string
		task               *core.Task
		expectedLen        int
		expectedAnnotation string // Check first annotation text if expectedLen > 0
	}{
		{
			name:        "nil task",
			task:        nil,
			expectedLen: 0,
		},
		{
			name: "task with no annotations",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{},
			},
			expectedLen: 0,
		},
		{
			name: "task with annotation containing file path",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "See /home/user/data.txt"},
				},
			},
			expectedLen:        1,
			expectedAnnotation: "See /home/user/data.txt",
		},
		{
			name: "task with multiple annotations containing paths",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "First file: /home/user/file1.txt"},
					{Description: "Second file: /tmp/file2.txt"},
				},
			},
			expectedLen:        2,
			expectedAnnotation: "First file: /home/user/file1.txt",
		},
		{
			name: "task with duplicate paths across annotations",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "File: /home/user/data.txt"},
					{Description: "Same file: /home/user/data.txt"},
				},
			},
			expectedLen:        1, // duplicates should be removed
			expectedAnnotation: "File: /home/user/data.txt",
		},
		{
			name: "task with home directory path",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "Check ~/documents/report.pdf"},
				},
			},
			expectedLen:        1,
			expectedAnnotation: "Check ~/documents/report.pdf",
		},
		{
			name: "task with path with spaces",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: `Check "/home/user/my documents/file.txt"`},
				},
			},
			expectedLen:        1,
			expectedAnnotation: `Check "/home/user/my documents/file.txt"`,
		},
		{
			name: "task with same path in different formats - should deduplicate",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: `Check "/home/user/my documents/file.txt"`},
					{Description: `Also check /home/user/my\ documents/file.txt`},
					{Description: `And file:///home/user/my%20documents/file.txt`},
				},
			},
			expectedLen:        1, // All three refer to the same path
			expectedAnnotation: `Check "/home/user/my documents/file.txt"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFilePathsFromAnnotations(tt.task)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d paths, got %d", tt.expectedLen, len(result))
				for i, r := range result {
					t.Errorf("  result[%d]: path=%q annotation=%q", i, r.Path, r.Annotation)
				}
			}
			if tt.expectedLen > 0 && len(result) > 0 {
				if result[0].Annotation != tt.expectedAnnotation {
					t.Errorf("expected annotation %q, got %q", tt.expectedAnnotation, result[0].Annotation)
				}
			}
		})
	}

	// Test home directory expansion separately
	t.Run("home directory expansion", func(t *testing.T) {
		task := &core.Task{
			Description: "Test task",
			Annotations: []core.Annotation{
				{Description: "Check ~/documents/report.pdf"},
			},
		}
		result := ExtractFilePathsFromAnnotations(task)
		if len(result) != 1 {
			t.Errorf("expected 1 path, got %d", len(result))
			return
		}
		expectedPath := filepath.Join(homeDir, "documents/report.pdf")
		if result[0].Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, result[0].Path)
		}
	})
}

func TestFilePathMatchFormatForDisplay(t *testing.T) {
	tests := []struct {
		name     string
		match    FilePathMatch
		expected string
	}{
		{
			name: "shows full annotation",
			match: FilePathMatch{
				Path:       "/home/user/file.txt",
				Annotation: "Check out /home/user/file.txt for details",
			},
			expected: "Check out /home/user/file.txt for details",
		},
		{
			name: "file URL annotation",
			match: FilePathMatch{
				Path:       "/home/user/data.json",
				Annotation: "See file:///home/user/data.json for config",
			},
			expected: "See file:///home/user/data.json for config",
		},
		{
			name: "path with spaces annotation",
			match: FilePathMatch{
				Path:       "/home/user/my documents/file.txt",
				Annotation: `Open "/home/user/my documents/file.txt"`,
			},
			expected: `Open "/home/user/my documents/file.txt"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.match.FormatForDisplay()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExpandHomePath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "expands home directory",
			path:     "~/documents/file.txt",
			expected: filepath.Join(homeDir, "documents/file.txt"),
		},
		{
			name:     "does not expand absolute path",
			path:     "/home/user/file.txt",
			expected: "/home/user/file.txt",
		},
		{
			name:     "does not expand path without tilde",
			path:     "documents/file.txt",
			expected: "documents/file.txt",
		},
		{
			name:     "expands home directory with spaces",
			path:     "~/my documents/file.txt",
			expected: filepath.Join(homeDir, "my documents/file.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandHomePath(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCleanPathTrailing(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "removes trailing period",
			path:     "/home/user/file.txt.",
			expected: "/home/user/file.txt",
		},
		{
			name:     "removes trailing comma",
			path:     "/home/user/file.txt,",
			expected: "/home/user/file.txt",
		},
		{
			name:     "removes trailing semicolon",
			path:     "/home/user/file.txt;",
			expected: "/home/user/file.txt",
		},
		{
			name:     "removes multiple trailing punctuation",
			path:     "/home/user/file.txt.,;",
			expected: "/home/user/file.txt",
		},
		{
			name:     "balances parentheses",
			path:     "/home/user/file(1).txt)",
			expected: "/home/user/file(1).txt",
		},
		{
			name:     "keeps balanced parentheses",
			path:     "/home/user/file(1).txt",
			expected: "/home/user/file(1).txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPathTrailing(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsValidFilePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid absolute path",
			path:     "/home/user/file.txt",
			expected: true,
		},
		{
			name:     "valid tmp path",
			path:     "/tmp/data",
			expected: true,
		},
		{
			name:     "valid path with spaces",
			path:     "/home/user/my documents/file.txt",
			expected: true,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "just root",
			path:     "/",
			expected: false,
		},
		{
			name:     "relative path",
			path:     "relative/path",
			expected: false,
		},
		{
			name:     "path with invalid chars",
			path:     "/home/user/file<name>.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFilePath(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "absolute path unchanged",
			path:     "/home/user/file.txt",
			expected: "/home/user/file.txt",
		},
		{
			name:     "home path expanded",
			path:     "~/documents/file.txt",
			expected: filepath.Join(homeDir, "documents/file.txt"),
		},
		{
			name:     "path with redundant slashes cleaned",
			path:     "/home/user//documents///file.txt",
			expected: "/home/user/documents/file.txt",
		},
		{
			name:     "path with spaces preserved",
			path:     "/home/user/my documents/file.txt",
			expected: "/home/user/my documents/file.txt",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.path)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDeduplicationAcrossFormats(t *testing.T) {
	// Test that the same path written in different formats is deduplicated
	task := &core.Task{
		Description: "Test task",
		Annotations: []core.Annotation{
			{Description: `Path 1: "/home/user/my documents/file.txt"`},
			{Description: `Path 2: /home/user/my\ documents/file.txt`},
			{Description: `Path 3: file:///home/user/my%20documents/file.txt`},
		},
	}

	result := ExtractFilePathsFromAnnotations(task)

	if len(result) != 1 {
		t.Errorf("expected 1 deduplicated path, got %d", len(result))
		for i, r := range result {
			t.Errorf("  result[%d]: path=%q", i, r.Path)
		}
	}

	if len(result) > 0 {
		expectedPath := "/home/user/my documents/file.txt"
		if result[0].Path != expectedPath {
			t.Errorf("expected path %q, got %q", expectedPath, result[0].Path)
		}
	}
}
