package gui_test

import (
	"testing"
	"time"

	"github.com/clobrano/wui/internal/gui"
)

func tp(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func TestShortRelDate(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		date *time.Time
		want string
	}{
		{tp("2026-06-25"), "today"},
		{tp("2026-06-26"), "tmrw"},
		{tp("2026-06-24"), "yest"},
		{tp("2026-06-28"), "3d"},
		{tp("2026-07-09"), "2w"},
		{tp("2026-08-25"), "2m"},
		{tp("2027-06-25"), "1y"},
		{tp("2026-06-22"), "3d ago"},
		{tp("2026-06-11"), "2w ago"},
		{tp("2026-04-25"), "2m ago"},
		{tp("2025-06-25"), "1y ago"},
		{nil, ""},
	}

	for _, tc := range cases {
		got := gui.ShortRelDateFrom(tc.date, now)
		if got != tc.want {
			t.Errorf("ShortRelDate(%v) = %q, want %q", tc.date, got, tc.want)
		}
	}
}

func TestLongRelDate(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		date *time.Time
		want string
	}{
		{tp("2026-06-25"), "today"},
		{tp("2026-06-26"), "tomorrow"},
		{tp("2026-06-24"), "yesterday"},
		{tp("2026-06-28"), "in 3 days"},
		{tp("2026-07-09"), "in 2 weeks"},
		{tp("2026-08-25"), "in 2 months"},
		{tp("2027-06-25"), "in 1 year"},
		{tp("2026-06-22"), "3 days ago"},
		{tp("2026-06-11"), "2 weeks ago"},
		{tp("2026-04-25"), "2 months ago"},
		{tp("2025-06-25"), "1 year ago"},
		{nil, ""},
	}

	for _, tc := range cases {
		got := gui.LongRelDateFrom(tc.date, now)
		if got != tc.want {
			t.Errorf("LongRelDate(%v) = %q, want %q", tc.date, got, tc.want)
		}
	}
}

func TestDueCSSClass(t *testing.T) {
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	overdue := now.AddDate(0, 0, -1)
	// today at midnight — before "now" clock-wise but same calendar day
	today := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)
	soon := now.AddDate(0, 0, 3)
	far := now.AddDate(0, 0, 30)

	cases := []struct {
		t    *time.Time
		want string
	}{
		{&overdue, "due-overdue"},
		{&today, "due-today"},
		{&soon, "due-soon"},
		{&far, ""},
		{nil, ""},
	}

	for _, tc := range cases {
		got := gui.DueCSSClassFrom(tc.t, now)
		if got != tc.want {
			t.Errorf("DueCSSClassFrom(%v) = %q, want %q", tc.t, got, tc.want)
		}
	}
}
