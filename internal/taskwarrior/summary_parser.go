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
// Expected format:
// Project      Remaining  Avg age  Complete  0%                        100%
// Project1            10  30 days       75%  =====================
// Project1.Sub1        5  15 days       90%  ==========================
func ParseSummaryOutput(output []byte) ([]core.ProjectSummary, error) {
	var summaries []core.ProjectSummary

	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header lines
	lineNum := 0
	headerFound := false

	// Regex to extract percentage (handles various formats like "75%", " 75%", "100%")
	percentRegex := regexp.MustCompile(`\s+(\d+)%`)

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
		if strings.HasPrefix(strings.TrimSpace(line), "---") ||
		   strings.HasPrefix(strings.TrimSpace(line), "===") {
			continue
		}

		// Parse data line
		// Expected format: "ProjectName      <numbers>  <age>  XX%  <bar>"
		// We need to extract ProjectName and XX%

		// Split by whitespace to get fields
		fields := strings.Fields(line)
		if len(fields) < 4 {
			// Not enough fields, skip
			slog.Debug("Skipping line with insufficient fields", "line", line, "fields", len(fields))
			continue
		}

		// First field is project name
		projectName := fields[0]

		// Find the percentage in the line using regex
		matches := percentRegex.FindStringSubmatch(line)
		if len(matches) < 2 {
			slog.Debug("No percentage found in line", "line", line)
			continue
		}

		percentStr := matches[1]
		percentage, err := strconv.Atoi(percentStr)
		if err != nil {
			slog.Debug("Failed to parse percentage", "value", percentStr, "error", err)
			continue
		}

		summaries = append(summaries, core.ProjectSummary{
			Name:       projectName,
			Percentage: percentage,
		})

		slog.Debug("Parsed project summary",
			"project", projectName,
			"percentage", percentage,
			"line", lineNum)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading summary output: %w", err)
	}

	if len(summaries) == 0 {
		slog.Debug("No projects found in summary output")
	}

	return summaries, nil
}
