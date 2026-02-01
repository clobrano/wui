package tui

import (
	"regexp"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// URLMatch represents a URL found in an annotation with optional display text
type URLMatch struct {
	URL        string // The actual URL
	Annotation string // The full annotation text containing this URL
}

// ExtractURLsFromAnnotations extracts all URLs from a task's description and annotations
// Supports:
// - Plain URLs (http://, https://, ftp://)
// - Markdown links: [text](url)
// - URLs with complex query strings and fragments
func ExtractURLsFromAnnotations(task *core.Task) []URLMatch {
	if task == nil {
		return nil
	}

	var urls []URLMatch
	seen := make(map[string]bool) // Deduplicate URLs

	// First, extract URLs from task description
	if task.Description != "" {
		extractedURLs := extractURLsFromText(task.Description)
		for _, url := range extractedURLs {
			if !seen[url] {
				seen[url] = true
				urls = append(urls, URLMatch{
					URL:        url,
					Annotation: task.Description,
				})
			}
		}
	}

	// Then extract URLs from annotations
	for _, annotation := range task.Annotations {
		extractedURLs := extractURLsFromText(annotation.Description)
		for _, url := range extractedURLs {
			if !seen[url] {
				seen[url] = true
				urls = append(urls, URLMatch{
					URL:        url,
					Annotation: annotation.Description,
				})
			}
		}
	}

	return urls
}

// extractURLsFromText extracts URLs from a text string
// Returns a slice of URL strings (deduplicated)
func extractURLsFromText(text string) []string {
	var urls []string
	seen := make(map[string]bool)

	// First, extract markdown links: [text](url)
	// This regex matches [any text](url) where url starts with http/https/ftp
	markdownLinkRegex := regexp.MustCompile(`\[([^\]]+)\]\(((?:https?|ftp)://[^\s\)]+)\)`)
	markdownMatches := markdownLinkRegex.FindAllStringSubmatch(text, -1)

	for _, match := range markdownMatches {
		if len(match) >= 3 {
			url := match[2]
			if !seen[url] {
				seen[url] = true
				urls = append(urls, url)
			}
		}
	}

	// Then extract plain URLs that are not part of markdown links
	// This regex matches URLs starting with http://, https://, or ftp://
	// It handles:
	// - Query strings with special characters
	// - Fragments (#section)
	// - Parentheses in URLs (but tries to balance them)
	// - Various TLDs and paths
	plainURLRegex := regexp.MustCompile(`(?:https?|ftp)://[^\s<>\[\]]+`)
	plainMatches := plainURLRegex.FindAllString(text, -1)

	for _, url := range plainMatches {
		// Clean up trailing punctuation that's likely not part of the URL
		url = cleanURLTrailingPunctuation(url)

		// Skip if this URL was already captured
		if seen[url] {
			continue
		}

		seen[url] = true
		urls = append(urls, url)
	}

	return urls
}

// cleanURLTrailingPunctuation removes trailing punctuation that's likely not part of the URL
func cleanURLTrailingPunctuation(url string) string {
	// Remove trailing punctuation that's commonly added after URLs in text
	// but keep valid URL characters like / and query strings

	// Balance parentheses - if there are more closing than opening, trim them
	openParens := strings.Count(url, "(")
	closeParens := strings.Count(url, ")")
	for closeParens > openParens && strings.HasSuffix(url, ")") {
		url = strings.TrimSuffix(url, ")")
		closeParens--
	}

	// Remove common trailing punctuation
	for {
		trimmed := false
		for _, suffix := range []string{".", ",", ";", ":", "!", "?"} {
			if strings.HasSuffix(url, suffix) {
				url = strings.TrimSuffix(url, suffix)
				trimmed = true
			}
		}
		if !trimmed {
			break
		}
	}

	return url
}

// FormatForDisplay formats a URLMatch for display in the picker
// Shows the full annotation text so users can recognize which URL they want
func (u URLMatch) FormatForDisplay() string {
	return u.Annotation
}
