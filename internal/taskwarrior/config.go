package taskwarrior

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TaskrcConfig represents parsed .taskrc configuration
type TaskrcConfig struct {
	DataLocation   string            // data.location
	DefaultCommand string            // default.command
	UDAs           map[string]UDA    // User Defined Attributes
	Reports        map[string]Report // Report configurations
}

// UDA represents a User Defined Attribute
type UDA struct {
	Type   string // string, numeric, date, duration
	Label  string // Display label
	Values string // Comma-separated allowed values (optional)
}

// Report represents a report configuration
type Report struct {
	Filter  string // Report filter
	Labels  string // Column labels
	Columns string // Column names
	Sort    string // Sort order
}

// ParseTaskrc reads and parses a .taskrc file
// Returns an empty config if file doesn't exist (not an error)
func ParseTaskrc(path string) (*TaskrcConfig, error) {
	cfg := &TaskrcConfig{
		UDAs:    make(map[string]UDA),
		Reports: make(map[string]Report),
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		// File doesn't exist - return empty config
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open taskrc: %w", err)
	}
	defer file.Close()

	// Parse line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse based on key pattern
		parseConfigLine(cfg, key, value)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading taskrc: %w", err)
	}

	return cfg, nil
}

// parseConfigLine parses a single configuration line
func parseConfigLine(cfg *TaskrcConfig, key, value string) {
	// Basic settings
	if key == "data.location" {
		cfg.DataLocation = value
		return
	}
	if key == "default.command" {
		cfg.DefaultCommand = value
		return
	}

	// UDA patterns: uda.<name>.(type|label|values)
	udaRegex := regexp.MustCompile(`^uda\.([^.]+)\.(type|label|values)$`)
	if matches := udaRegex.FindStringSubmatch(key); matches != nil {
		udaName := matches[1]
		udaField := matches[2]

		// Get or create UDA
		uda, exists := cfg.UDAs[udaName]
		if !exists {
			uda = UDA{}
		}

		// Set field
		switch udaField {
		case "type":
			uda.Type = value
		case "label":
			uda.Label = value
		case "values":
			uda.Values = value
		}

		cfg.UDAs[udaName] = uda
		return
	}

	// Report patterns: report.<name>.(filter|labels|columns|sort)
	reportRegex := regexp.MustCompile(`^report\.([^.]+)\.(filter|labels|columns|sort)$`)
	if matches := reportRegex.FindStringSubmatch(key); matches != nil {
		reportName := matches[1]
		reportField := matches[2]

		// Get or create report
		report, exists := cfg.Reports[reportName]
		if !exists {
			report = Report{}
		}

		// Set field
		switch reportField {
		case "filter":
			report.Filter = value
		case "labels":
			report.Labels = value
		case "columns":
			report.Columns = value
		case "sort":
			report.Sort = value
		}

		cfg.Reports[reportName] = report
		return
	}

	// Ignore other settings (colors, etc.)
}
