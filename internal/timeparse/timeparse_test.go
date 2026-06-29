package timeparse

import (
	"testing"
	"time"
)

func TestParseDateEmpty(t *testing.T) {
	got, err := ParseDate("")
	if err != nil || got != "" {
		t.Fatalf("ParseDate(\"\") = %q, %v; want \"\", nil", got, err)
	}
}

func TestParseDateAbsolute(t *testing.T) {
	got, err := ParseDate("2026-06-01")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "2026-06-01" {
		t.Errorf("got %q, want 2026-06-01", got)
	}
}

func TestParseDateRelativeDays(t *testing.T) {
	got, err := ParseDate("7d")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02")
	if got != want {
		t.Errorf("ParseDate(7d) = %q, want %q", got, want)
	}
}

func TestParseDateNow(t *testing.T) {
	got, err := ParseDate("now")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := time.Now().UTC().Format("2006-01-02")
	if got != want {
		t.Errorf("ParseDate(now) = %q, want %q", got, want)
	}
}

func TestParseDateHoursCollapseToDay(t *testing.T) {
	// Sub-day values collapse to a calendar day (Postmark filters are date-only).
	got, err := ParseDate("2h")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != len("2006-01-02") {
		t.Errorf("expected a YYYY-MM-DD date, got %q", got)
	}
}

func TestParseDateInvalid(t *testing.T) {
	if _, err := ParseDate("not-a-date"); err == nil {
		t.Error("expected error for invalid input")
	}
}
