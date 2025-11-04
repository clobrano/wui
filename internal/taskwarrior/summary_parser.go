package taskwarrior

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/clobrano/wui/internal/core"
)

// ParseSummaryOutput parses the output of "task summary" command
// Expected format with indentation-based hierarchy:
// Project      Remaining  Avg age  Complete  0%                        100%
// M8s                60     12w       51%  ===============
//   helm              2    1.0y        0%
//   SNR              11     3mo       47%  ==============
//     RHWA12          2     6mo        0%
func ParseSummaryOutput(output []byte) ([]core.ProjectSummary, error) {
	var summaries []core.ProjectSummary

	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header lines
	lineNum := 0
	headerFound := false

	// Regex to extract percentage (handles various formats like "75%", " 75%", "100%")
	percentRegex := regexp.MustCompile(`\s+(\d+)%`)

	// Stack to track parent projects at each indentation level
	// Index represents indentation level (0 = no indent, 1 = 2 spaces, 2 = 4 spaces, etc.)
	parentStack := make([]string, 0)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Identify and skip the header line (contains "Project" and "Complete")
		if strings.Contains(line, "Project") && strings.Contains(line, "Complete") {
			headerFound = true
			slog.Debug("Found summary header", "line", lineNum)
			continue
		}

		// Skip lines before header is found
		if !headerFound {
			continue
		}

		// Skip separator lines (dashes, equals, etc.)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "===") {
			continue
		}

		// Skip summary line at the end (e.g., "30 projects")
		if strings.HasSuffix(trimmed, "projects") || strings.HasSuffix(trimmed, "project") {
			continue
		}

		// Calculate indentation level (2 spaces per level)
		indentLevel := 0
		for i := 0; i < len(line) && line[i] == ' '; i += 2 {
			indentLevel++
		}

		// Parse data line
		// Split by whitespace to get fields
		fields := strings.Fields(line)
		if len(fields) < 1 {
			// Not enough fields, skip
			slog.Debug("Skipping line with insufficient fields", "line", line, "fields", len(fields))
			continue
		}

		// First field is project name (just the segment, not full path)
		projectSegment := fields[0]

		// Skip "(none)" entries
		if projectSegment == "(none)" {
			continue
		}

		// Find the percentage in the line using regex
		matches := percentRegex.FindStringSubmatch(line)
		percentage := 0
		if len(matches) >= 2 {
			percentStr := matches[1]
			p, err := strconv.Atoi(percentStr)
			if err == nil {
				percentage = p
			}
		}

		// Reconstruct full project name from parent stack
		// Adjust parent stack to current indentation level
		if indentLevel < len(parentStack) {
			// Going back up in hierarchy, trim the stack
			parentStack = parentStack[:indentLevel]
		}

		// Build full project name
		var fullProjectName string
		if len(parentStack) == 0 {
			// Top-level project
			fullProjectName = projectSegment
		} else {
			// Nested project: join parent stack with current segment
			fullProjectName = strings.Join(append(parentStack, projectSegment), ".")
		}

		// Add to summaries
		summaries = append(summaries, core.ProjectSummary{
			Name:       fullProjectName,
			Percentage: percentage,
		})

		slog.Debug("Parsed project summary",
			"project", fullProjectName,
			"segment", projectSegment,
			"indentLevel", indentLevel,
			"percentage", percentage,
			"line", lineNum)

		// Update parent stack for next iteration
		// If we're at this level, we become the parent for the next deeper level
		if indentLevel >= len(parentStack) {
			parentStack = append(parentStack, projectSegment)
		} else {
			// Replace the segment at this level
			parentStack = parentStack[:indentLevel]
			parentStack = append(parentStack, projectSegment)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading summary output: %w", err)
	}

	if len(summaries) == 0 {
		slog.Debug("No projects found in summary output")
	}

	return summaries, nil
}
