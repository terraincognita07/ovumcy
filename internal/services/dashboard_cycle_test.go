package services

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestDashboardCycleReferenceLengthPrefersUserValue(t *testing.T) {
	user := &models.User{CycleLength: 29}
	stats := CycleStats{MedianCycleLength: 28, AverageCycleLength: 27}
	if got := DashboardCycleReferenceLength(user, stats); got != 29 {
		t.Fatalf("expected 29, got %d", got)
	}
}

func TestDashboardCycleStaleAnchorPrefersUserBaseline(t *testing.T) {
	userBaseline := time.Date(2026, time.January, 1, 15, 30, 0, 0, time.UTC)
	statsBaseline := time.Date(2026, time.February, 20, 7, 0, 0, 0, time.UTC)
	user := &models.User{LastPeriodStart: &userBaseline}
	stats := CycleStats{LastPeriodStart: statsBaseline}

	anchor := DashboardCycleStaleAnchor(user, stats, time.UTC)
	if anchor.Format("2006-01-02") != "2026-01-01" {
		t.Fatalf("expected user baseline date, got %s", anchor.Format("2006-01-02"))
	}
}

func TestCompletedCycleTrendLengths(t *testing.T) {
	logs := []models.DailyLog{
		{Date: mustParseDashboardDay(t, "2026-01-01"), IsPeriod: true},
		{Date: mustParseDashboardDay(t, "2026-01-29"), IsPeriod: true},
		{Date: mustParseDashboardDay(t, "2026-02-26"), IsPeriod: true},
	}
	now := mustParseDashboardDay(t, "2026-03-10")

	got := CompletedCycleTrendLengths(logs, now, time.UTC)
	if len(got) != 2 || got[0] != 28 || got[1] != 28 {
		t.Fatalf("expected [28 28], got %#v", got)
	}
}

func TestBuildDashboardCycleContext(t *testing.T) {
	userStart := mustParseDashboardDay(t, "2026-02-10")
	user := &models.User{
		CycleLength:     28,
		PeriodLength:    5,
		LastPeriodStart: &userStart,
	}
	stats := CycleStats{
		CurrentCycleDay:     36,
		LastPeriodStart:     mustParseDashboardDay(t, "2026-02-10"),
		MedianCycleLength:   28,
		AveragePeriodLength: 5,
		NextPeriodStart:     mustParseDashboardDay(t, "2026-03-10"),
		OvulationDate:       mustParseDashboardDay(t, "2026-02-24"),
	}
	today := mustParseDashboardDay(t, "2026-03-14")

	context := BuildDashboardCycleContext(user, stats, today, time.UTC)
	if context.CycleDayReference != 28 {
		t.Fatalf("expected cycle day reference 28, got %d", context.CycleDayReference)
	}
	if !context.CycleDayWarning {
		t.Fatalf("expected cycle day warning for long cycle day")
	}
	if !context.CycleDataStale {
		t.Fatalf("expected stale cycle data flag")
	}
	if context.DisplayNextPeriodStart.IsZero() {
		t.Fatalf("expected next period start prediction")
	}
}

func mustParseDashboardDay(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02", raw, time.UTC)
	if err != nil {
		t.Fatalf("parse day %q: %v", raw, err)
	}
	return parsed
}
