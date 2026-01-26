package tui

import (
	"testing"

	"github.com/clobrano/wui/internal/core"
)

func TestExtractURLsFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []URLMatch
	}{
		{
			name: "plain http URL",
			text: "Check this link: http://example.com/path",
			expected: []URLMatch{
				{URL: "http://example.com/path", DisplayText: "http://example.com/path"},
			},
		},
		{
			name: "plain https URL",
			text: "Visit https://github.com/user/repo",
			expected: []URLMatch{
				{URL: "https://github.com/user/repo", DisplayText: "https://github.com/user/repo"},
			},
		},
		{
			name: "markdown link",
			text: "Click [here](https://example.com/page) for more info",
			expected: []URLMatch{
				{URL: "https://example.com/page", DisplayText: "here"},
			},
		},
		{
			name: "multiple URLs",
			text: "Links: https://first.com and https://second.com",
			expected: []URLMatch{
				{URL: "https://first.com", DisplayText: "https://first.com"},
				{URL: "https://second.com", DisplayText: "https://second.com"},
			},
		},
		{
			name: "URL with query string",
			text: "Search: https://google.com/search?q=test&lang=en",
			expected: []URLMatch{
				{URL: "https://google.com/search?q=test&lang=en", DisplayText: "https://google.com/search?q=test&lang=en"},
			},
		},
		{
			name: "URL with fragment",
			text: "Section: https://docs.example.com/page#section-1",
			expected: []URLMatch{
				{URL: "https://docs.example.com/page#section-1", DisplayText: "https://docs.example.com/page#section-1"},
			},
		},
		{
			name: "markdown link with complex URL",
			text: "[GitHub Issue](https://github.com/owner/repo/issues/123?label=bug)",
			expected: []URLMatch{
				{URL: "https://github.com/owner/repo/issues/123?label=bug", DisplayText: "GitHub Issue"},
			},
		},
		{
			name: "URL ending with punctuation",
			text: "Check this URL: https://example.com/page.",
			expected: []URLMatch{
				{URL: "https://example.com/page", DisplayText: "https://example.com/page"},
			},
		},
		{
			name: "mixed markdown and plain URLs",
			text: "See [docs](https://docs.example.com) and also https://api.example.com",
			expected: []URLMatch{
				{URL: "https://docs.example.com", DisplayText: "docs"},
				{URL: "https://api.example.com", DisplayText: "https://api.example.com"},
			},
		},
		{
			name: "no URLs",
			text: "This is just plain text without any links",
			expected: []URLMatch{},
		},
		{
			name: "URL with parentheses (Wikipedia style)",
			text: "See https://en.wikipedia.org/wiki/Example_(disambiguation)",
			expected: []URLMatch{
				{URL: "https://en.wikipedia.org/wiki/Example_(disambiguation)", DisplayText: "https://en.wikipedia.org/wiki/Example_(disambiguation)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractURLsFromText(tt.text)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d URLs, got %d", len(tt.expected), len(result))
				for i, r := range result {
					t.Errorf("  result[%d]: URL=%q DisplayText=%q", i, r.URL, r.DisplayText)
				}
				return
			}

			for i, expected := range tt.expected {
				if result[i].URL != expected.URL {
					t.Errorf("URL mismatch at index %d: expected %q, got %q", i, expected.URL, result[i].URL)
				}
				if result[i].DisplayText != expected.DisplayText {
					t.Errorf("DisplayText mismatch at index %d: expected %q, got %q", i, expected.DisplayText, result[i].DisplayText)
				}
			}
		})
	}
}

func TestExtractURLsFromAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		task        *core.Task
		expectedLen int
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
			expectedLen: 1,
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
			expectedLen: 2,
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
			expectedLen: 1, // duplicates should be removed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractURLsFromAnnotations(tt.task)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d URLs, got %d", tt.expectedLen, len(result))
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
			name: "plain URL",
			match: URLMatch{
				URL:         "https://example.com",
				DisplayText: "https://example.com",
			},
			expected: "https://example.com",
		},
		{
			name: "markdown link",
			match: URLMatch{
				URL:         "https://example.com",
				DisplayText: "Example Site",
			},
			expected: "Example Site (https://example.com)",
		},
		{
			name: "long URL truncated",
			match: URLMatch{
				URL:         "https://very-long-domain-name.example.com/with/a/very/long/path/that/exceeds/limit",
				DisplayText: "Long Link",
			},
			expected: "Long Link (https://very-long-domain-name.example.com/with...)",
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
