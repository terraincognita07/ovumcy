package api

import (
	"testing"
	"time"
)

func TestLocalizedDashboardDate_Russian(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	got := localizedDashboardDate("ru", value)
	want := "18 февраля 2026, среда"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalizedDashboardDate_English(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	got := localizedDashboardDate("en", value)
	want := "February 18, 2026, Wednesday"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLocalizedMonthYear(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	if got := localizedMonthYear("ru", value); got != "Февраль 2026" {
		t.Fatalf("expected russian month-year, got %q", got)
	}
	if got := localizedMonthYear("en", value); got != "February 2026" {
		t.Fatalf("expected english month-year, got %q", got)
	}
	if got := localizedMonthYear("de", value); got != "February 2026" {
		t.Fatalf("expected fallback month-year, got %q", got)
	}
}

func TestLocalizedDateLabel(t *testing.T) {
	value := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	if got := localizedDateLabel("ru", value); got != "Ср, Фев 18" {
		t.Fatalf("expected russian date label, got %q", got)
	}
	if got := localizedDateLabel("en", value); got != "Wed, Feb 18" {
		t.Fatalf("expected english date label, got %q", got)
	}
	if got := localizedDateLabel("de", value); got != "Wed, Feb 18" {
		t.Fatalf("expected fallback date label, got %q", got)
	}
}
