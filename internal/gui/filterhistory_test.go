package gui_test

import (
	"testing"

	"github.com/clobrano/wui/internal/gui"
)

func TestFilterHistory_PrependAndLoad(t *testing.T) {
	fh := gui.NewFilterHistory(t.TempDir())

	if err := fh.Prepend("status:pending"); err != nil {
		t.Fatalf("Prepend: %v", err)
	}
	if err := fh.Prepend("+work"); err != nil {
		t.Fatalf("Prepend: %v", err)
	}

	entries, err := fh.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0] != "+work" {
		t.Errorf("expected newest entry first, got %q", entries[0])
	}
}

func TestFilterHistory_Dedup(t *testing.T) {
	fh := gui.NewFilterHistory(t.TempDir())

	_ = fh.Prepend("status:pending")
	_ = fh.Prepend("+work")
	_ = fh.Prepend("status:pending") // duplicate – should move to front

	entries, _ := fh.Load()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after dedup, got %d", len(entries))
	}
	if entries[0] != "status:pending" {
		t.Errorf("expected deduped entry at front, got %q", entries[0])
	}
}

func TestFilterHistory_Trim(t *testing.T) {
	fh := gui.NewFilterHistory(t.TempDir())

	for i := 0; i < 25; i++ {
		_ = fh.Prepend(string(rune('a' + i)))
	}

	entries, _ := fh.Load()
	if len(entries) != 20 {
		t.Errorf("expected 20 entries (trim), got %d", len(entries))
	}
}

func TestFilterHistory_Delete(t *testing.T) {
	fh := gui.NewFilterHistory(t.TempDir())

	_ = fh.Prepend("a")
	_ = fh.Prepend("b")
	_ = fh.Delete("a")

	entries, _ := fh.Load()
	if len(entries) != 1 || entries[0] != "b" {
		t.Errorf("unexpected entries after delete: %v", entries)
	}
}

func TestFilterHistory_Clear(t *testing.T) {
	fh := gui.NewFilterHistory(t.TempDir())

	_ = fh.Prepend("a")
	_ = fh.Prepend("b")
	_ = fh.Clear()

	entries, _ := fh.Load()
	if len(entries) != 0 {
		t.Errorf("expected empty after clear, got %v", entries)
	}
}

func TestFilterHistory_EmptyFile(t *testing.T) {
	fh := gui.NewFilterHistory(t.TempDir())
	entries, err := fh.Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice, got %v", entries)
	}
}
