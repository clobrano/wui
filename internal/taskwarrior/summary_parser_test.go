package taskwarrior

import (
	"testing"
)

func TestParseSummaryOutput(t *testing.T) {
	// Sample output from task summary with indentation-based hierarchy
	input := `
Project              Remaining Avg age Complete 0%                        100%
-------------------- --------- ------- -------- ------------------------------
(none)                      97      7w      39% ===========
CI                           3    1.0y      25% =======
M8s                         60     12w      51% ===============
  helm                       2    1.0y       0%
  SNR                       11     3mo      47% ==============
    RHWA12                   2     6mo       0%
  NMO                        2    11mo      33% ==========
WUI                          8      2w       0%
dty                          2     12h      87% ==========================
  condominio                 2     12h      87% ==========================

10 projects
`

	summaries, err := ParseSummaryOutput([]byte(input))
	if err != nil {
		t.Fatalf("ParseSummaryOutput failed: %v", err)
	}

	// Expected projects (excluding "(none)")
	expected := map[string]int{
		"CI":            25,
		"M8s":           51,
		"M8s.helm":      0,
		"M8s.SNR":       47,
		"M8s.SNR.RHWA12": 0,
		"M8s.NMO":       33,
		"WUI":           0,
		"dty":           87,
		"dty.condominio": 87,
	}

	// Check count
	if len(summaries) != len(expected) {
		t.Errorf("Expected %d projects, got %d", len(expected), len(summaries))
	}

	// Build map of actual results
	actual := make(map[string]int)
	for _, s := range summaries {
		actual[s.Name] = s.Percentage
	}

	// Verify each expected project
	for name, expectedPct := range expected {
		actualPct, found := actual[name]
		if !found {
			t.Errorf("Expected project %s not found in results", name)
			continue
		}
		if actualPct != expectedPct {
			t.Errorf("Project %s: expected %d%%, got %d%%", name, expectedPct, actualPct)
		}
	}
}

func TestParseSummaryOutputEmpty(t *testing.T) {
	input := `
Project              Remaining Avg age Complete 0%                        100%
-------------------- --------- ------- -------- ------------------------------

0 projects
`

	summaries, err := ParseSummaryOutput([]byte(input))
	if err != nil {
		t.Fatalf("ParseSummaryOutput failed: %v", err)
	}

	if len(summaries) != 0 {
		t.Errorf("Expected 0 projects, got %d", len(summaries))
	}
}
