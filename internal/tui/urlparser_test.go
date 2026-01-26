package tui

import (
	"testing"

	"github.com/clobrano/wui/internal/core"
)

func TestExtractURLsFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "plain http URL",
			text:     "Check this link: http://example.com/path",
			expected: []string{"http://example.com/path"},
		},
		{
			name:     "plain https URL",
			text:     "Visit https://github.com/user/repo",
			expected: []string{"https://github.com/user/repo"},
		},
		{
			name:     "markdown link",
			text:     "Click [here](https://example.com/page) for more info",
			expected: []string{"https://example.com/page"},
		},
		{
			name:     "multiple URLs",
			text:     "Links: https://first.com and https://second.com",
			expected: []string{"https://first.com", "https://second.com"},
		},
		{
			name:     "URL with query string",
			text:     "Search: https://google.com/search?q=test&lang=en",
			expected: []string{"https://google.com/search?q=test&lang=en"},
		},
		{
			name:     "URL with fragment",
			text:     "Section: https://docs.example.com/page#section-1",
			expected: []string{"https://docs.example.com/page#section-1"},
		},
		{
			name:     "markdown link with complex URL",
			text:     "[GitHub Issue](https://github.com/owner/repo/issues/123?label=bug)",
			expected: []string{"https://github.com/owner/repo/issues/123?label=bug"},
		},
		{
			name:     "URL ending with punctuation",
			text:     "Check this URL: https://example.com/page.",
			expected: []string{"https://example.com/page"},
		},
		{
			name:     "mixed markdown and plain URLs",
			text:     "See [docs](https://docs.example.com) and also https://api.example.com",
			expected: []string{"https://docs.example.com", "https://api.example.com"},
		},
		{
			name:     "no URLs",
			text:     "This is just plain text without any links",
			expected: []string{},
		},
		{
			name:     "URL with parentheses (Wikipedia style)",
			text:     "See https://en.wikipedia.org/wiki/Example_(disambiguation)",
			expected: []string{"https://en.wikipedia.org/wiki/Example_(disambiguation)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractURLsFromText(tt.text)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d URLs, got %d", len(tt.expected), len(result))
				for i, r := range result {
					t.Errorf("  result[%d]: %q", i, r)
				}
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("URL mismatch at index %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestExtractURLsFromAnnotations(t *testing.T) {
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
			name: "task with annotation containing URL",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "See https://example.com"},
				},
			},
			expectedLen:        1,
			expectedAnnotation: "See https://example.com",
		},
		{
			name: "task with multiple annotations containing URLs",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "First link: https://first.com"},
					{Description: "Second link: https://second.com"},
				},
			},
			expectedLen:        2,
			expectedAnnotation: "First link: https://first.com",
		},
		{
			name: "task with duplicate URLs across annotations",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "Link: https://example.com"},
					{Description: "Same link: https://example.com"},
				},
			},
			expectedLen:        1, // duplicates should be removed
			expectedAnnotation: "Link: https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractURLsFromAnnotations(tt.task)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d URLs, got %d", tt.expectedLen, len(result))
			}
			if tt.expectedLen > 0 && len(result) > 0 {
				if result[0].Annotation != tt.expectedAnnotation {
					t.Errorf("expected annotation %q, got %q", tt.expectedAnnotation, result[0].Annotation)
				}
			}
		})
	}
}

func TestURLMatchFormatForDisplay(t *testing.T) {
	tests := []struct {
		name     string
		match    URLMatch
		expected string
	}{
		{
			name: "shows full annotation",
			match: URLMatch{
				URL:        "https://example.com",
				Annotation: "Check out https://example.com for details",
			},
			expected: "Check out https://example.com for details",
		},
		{
			name: "markdown link annotation",
			match: URLMatch{
				URL:        "https://github.com/owner/repo",
				Annotation: "See [the repo](https://github.com/owner/repo) for more info",
			},
			expected: "See [the repo](https://github.com/owner/repo) for more info",
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
