package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const maxFilterHistory = 20

// FilterHistory persists recently used Taskwarrior filter strings.
type FilterHistory struct {
	path string
}

// NewFilterHistory creates a FilterHistory backed by <configDir>/filter_history.json.
func NewFilterHistory(configDir string) *FilterHistory {
	return &FilterHistory{path: filepath.Join(configDir, "filter_history.json")}
}

// Load reads the filter history from disk. Returns an empty slice if the file doesn't exist.
func (fh *FilterHistory) Load() ([]string, error) {
	data, err := os.ReadFile(fh.path)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []string
	if err := json.Unmarshal(data, &entries); err != nil {
		return []string{}, nil
	}
	return entries, nil
}

// Prepend adds filter to the front of the history (deduplicating and trimming to 20).
func (fh *FilterHistory) Prepend(filter string) error {
	entries, err := fh.Load()
	if err != nil {
		entries = []string{}
	}

	// Deduplicate: remove any existing occurrence of filter.
	deduped := make([]string, 0, len(entries))
	for _, e := range entries {
		if e != filter {
			deduped = append(deduped, e)
		}
	}

	// Prepend and trim.
	combined := append([]string{filter}, deduped...)
	if len(combined) > maxFilterHistory {
		combined = combined[:maxFilterHistory]
	}

	return fh.save(combined)
}

// Delete removes a single entry from the history.
func (fh *FilterHistory) Delete(filter string) error {
	entries, err := fh.Load()
	if err != nil {
		return err
	}
	result := make([]string, 0, len(entries))
	for _, e := range entries {
		if e != filter {
			result = append(result, e)
		}
	}
	return fh.save(result)
}

// Clear removes all filter history entries.
func (fh *FilterHistory) Clear() error {
	return fh.save([]string{})
}

func (fh *FilterHistory) save(entries []string) error {
	if err := os.MkdirAll(filepath.Dir(fh.path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(fh.path, data, 0644)
}
