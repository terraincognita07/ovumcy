package api

import (
	"testing"
	"time"
)

func TestResolveCalendarMonthAndSelectedDateInvalidMonth(t *testing.T) {
	t.Parallel()

	_, _, err := resolveCalendarMonthAndSelectedDate("2026-99", "", time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC), time.UTC)
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

func TestResolveCalendarMonthAndSelectedDateUsesSelectedDayMonthWhenMonthMissing(t *testing.T) {
	t.Parallel()

	month, selectedDate, err := resolveCalendarMonthAndSelectedDate("", "2026-02-17", time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC), time.UTC)
	if err != nil {
		t.Fatalf("resolveCalendarMonthAndSelectedDate returned error: %v", err)
	}
	if selectedDate != "2026-02-17" {
		t.Fatalf("expected selected date 2026-02-17, got %q", selectedDate)
	}
	if month.Format("2006-01") != "2026-02" {
		t.Fatalf("expected month 2026-02, got %s", month.Format("2006-01"))
	}
}

func TestResolveCalendarMonthAndSelectedDateKeepsMonthQueryWhenPresent(t *testing.T) {
	t.Parallel()

	month, selectedDate, err := resolveCalendarMonthAndSelectedDate("2026-03", "2026-02-17", time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC), time.UTC)
	if err != nil {
		t.Fatalf("resolveCalendarMonthAndSelectedDate returned error: %v", err)
	}
	if selectedDate != "2026-02-17" {
		t.Fatalf("expected selected date 2026-02-17, got %q", selectedDate)
	}
	if month.Format("2006-01") != "2026-03" {
		t.Fatalf("expected month 2026-03, got %s", month.Format("2006-01"))
	}
}

func TestResolveCalendarMonthAndSelectedDateIgnoresInvalidSelectedDay(t *testing.T) {
	t.Parallel()

	month, selectedDate, err := resolveCalendarMonthAndSelectedDate("2026-03", "invalid-day", time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC), time.UTC)
	if err != nil {
		t.Fatalf("resolveCalendarMonthAndSelectedDate returned error: %v", err)
	}
	if selectedDate != "" {
		t.Fatalf("expected empty selected date for invalid day, got %q", selectedDate)
	}
	if month.Format("2006-01") != "2026-03" {
		t.Fatalf("expected month 2026-03, got %s", month.Format("2006-01"))
	}
}

func TestResolveCalendarMonthAndSelectedDateDefaultsToTodayWhenEmpty(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.February, 21, 10, 30, 0, 0, time.UTC)
	month, selectedDate, err := resolveCalendarMonthAndSelectedDate("", "", now, time.UTC)
	if err != nil {
		t.Fatalf("resolveCalendarMonthAndSelectedDate returned error: %v", err)
	}
	if selectedDate != "2026-02-21" {
		t.Fatalf("expected selected date 2026-02-21, got %q", selectedDate)
	}
	if month.Format("2006-01") != "2026-02" {
		t.Fatalf("expected month 2026-02, got %s", month.Format("2006-01"))
	}
}

func TestCalendarHelpersRangeAndAdjacentMonths(t *testing.T) {
	t.Parallel()

	monthStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	from, to := calendarLogRange(monthStart)
	if from.Format("2006-01-02") != "2025-11-23" {
		t.Fatalf("expected range start 2025-11-23, got %s", from.Format("2006-01-02"))
	}
	if to.Format("2006-01-02") != "2026-05-09" {
		t.Fatalf("expected range end 2026-05-09, got %s", to.Format("2006-01-02"))
	}

	prev, next := calendarAdjacentMonthValues(monthStart)
	if prev != "2026-01" {
		t.Fatalf("expected prev month 2026-01, got %q", prev)
	}
	if next != "2026-03" {
		t.Fatalf("expected next month 2026-03, got %q", next)
	}
}
