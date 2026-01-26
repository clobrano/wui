package tui

import (
	"regexp"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// URLMatch represents a URL found in an annotation with optional display text
type URLMatch struct {
	URL         string // The actual URL
	DisplayText string // Display text (for markdown links) or the URL itself
}

// ExtractURLsFromAnnotations extracts all URLs from a task's annotations
// Supports:
// - Plain URLs (http://, https://, ftp://)
// - Markdown links: [text](url)
// - URLs with complex query strings and fragments
func ExtractURLsFromAnnotations(task *core.Task) []URLMatch {
	if task == nil || len(task.Annotations) == 0 {
		return nil
	}

	var urls []URLMatch
	seen := make(map[string]bool) // Deduplicate URLs

	for _, annotation := range task.Annotations {
		matches := extractURLsFromText(annotation.Description)
		for _, match := range matches {
			if !seen[match.URL] {
				seen[match.URL] = true
				urls = append(urls, match)
			}
		}
	}

	return urls
}

// extractURLsFromText extracts URLs from a text string
func extractURLsFromText(text string) []URLMatch {
	var urls []URLMatch

	// First, extract markdown links: [text](url)
	// This regex matches [any text](url) where url starts with http/https/ftp
	markdownLinkRegex := regexp.MustCompile(`\[([^\]]+)\]\(((?:https?|ftp)://[^\s\)]+)\)`)
	markdownMatches := markdownLinkRegex.FindAllStringSubmatch(text, -1)

	// Track positions of markdown links to avoid double-matching their URLs
	markdownURLs := make(map[string]bool)
	for _, match := range markdownMatches {
		if len(match) >= 3 {
			displayText := match[1]
			url := match[2]
			markdownURLs[url] = true
			urls = append(urls, URLMatch{
				URL:         url,
				DisplayText: displayText,
			})
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

		// Skip if this URL was already captured as part of a markdown link
		if markdownURLs[url] {
			continue
		}

		urls = append(urls, URLMatch{
			URL:         url,
			DisplayText: url,
		})
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

// FormatURLForDisplay formats a URLMatch for display in the picker
// Shows "Display Text (url)" for markdown links, or just the URL for plain links
func (u URLMatch) FormatForDisplay() string {
	if u.DisplayText != u.URL {
		return u.DisplayText + " (" + truncateURL(u.URL, 50) + ")"
	}
	return u.URL
}

// truncateURL truncates a URL to maxLen characters, adding "..." if truncated
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
