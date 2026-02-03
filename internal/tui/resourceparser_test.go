package tui

import (
	"testing"

	"github.com/clobrano/wui/internal/core"
)

func TestExtractResourcesFromAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		task        *core.Task
		expectedLen int
		checkFirst  func(t *testing.T, r ResourceMatch)
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
			name: "task with URL annotation",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "See https://example.com"},
				},
			},
			expectedLen: 1,
			checkFirst: func(t *testing.T, r ResourceMatch) {
				if r.Type != ResourceTypeURL {
					t.Errorf("expected URL type, got %v", r.Type)
				}
				if r.Resource != "https://example.com" {
					t.Errorf("expected https://example.com, got %s", r.Resource)
				}
			},
		},
		{
			name: "task with file path annotation",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "See /home/user/file.txt"},
				},
			},
			expectedLen: 1,
			checkFirst: func(t *testing.T, r ResourceMatch) {
				if r.Type != ResourceTypeFile {
					t.Errorf("expected File type, got %v", r.Type)
				}
				if r.Resource != "/home/user/file.txt" {
					t.Errorf("expected /home/user/file.txt, got %s", r.Resource)
				}
			},
		},
		{
			name: "task with both URL and file path",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "URL: https://example.com"},
					{Description: "File: /home/user/data.txt"},
				},
			},
			expectedLen: 2,
			checkFirst: func(t *testing.T, r ResourceMatch) {
				// First should be the URL (URLs are extracted first)
				if r.Type != ResourceTypeURL {
					t.Errorf("expected URL type for first resource, got %v", r.Type)
				}
			},
		},
		{
			name: "task with duplicate resources",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "Link: https://example.com"},
					{Description: "Same link: https://example.com"},
					{Description: "File: /home/user/file.txt"},
					{Description: "Same file: /home/user/file.txt"},
				},
			},
			expectedLen: 2, // One URL, one file (duplicates removed)
		},
		{
			name: "task with mixed content in same annotation",
			task: &core.Task{
				Description: "Test task",
				Annotations: []core.Annotation{
					{Description: "See https://docs.example.com and /home/user/local-docs.txt"},
				},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractResourcesFromAnnotations(tt.task)
			if len(result) != tt.expectedLen {
				t.Errorf("expected %d resources, got %d", tt.expectedLen, len(result))
				for i, r := range result {
					t.Errorf("  result[%d]: type=%v resource=%q", i, r.Type, r.Resource)
				}
				return
			}
			if tt.checkFirst != nil && len(result) > 0 {
				tt.checkFirst(t, result[0])
			}
		})
	}
}

func TestResourceMatchFormatForDisplay(t *testing.T) {
	r := ResourceMatch{
		Resource:   "https://example.com",
		Annotation: "Check out https://example.com for details",
		Type:       ResourceTypeURL,
	}

	expected := "Check out https://example.com for details"
	if r.FormatForDisplay() != expected {
		t.Errorf("expected %q, got %q", expected, r.FormatForDisplay())
	}
}

func TestResourceTypeConstants(t *testing.T) {
	// Ensure the constants are distinct
	if ResourceTypeURL == ResourceTypeFile {
		t.Error("ResourceTypeURL and ResourceTypeFile should be different")
	}
}
