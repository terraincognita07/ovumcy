package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func TestApplyUserCycleBaseline_UsesOnboardingValuesWhenDataIsSparse(t *testing.T) {
	handler := &Handler{
		location:        time.UTC,
		lutealPhaseDays: 14,
	}

	userLastPeriod := mustParseBaselineDay(t, "2026-02-07")
	user := &models.User{
		Role:            models.RoleOwner,
		CycleLength:     29,
		PeriodLength:    6,
		LastPeriodStart: &userLastPeriod,
	}

	logs := []models.DailyLog{
		{Date: mustParseBaselineDay(t, "2026-02-07"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2026-02-16"), IsPeriod: true, Flow: models.FlowMedium},
	}

	now := mustParseBaselineDay(t, "2026-02-17")
	stats := services.BuildCycleStats(logs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, logs, stats, now)

	if stats.AverageCycleLength != 29 {
		t.Fatalf("expected average cycle length 29, got %.2f", stats.AverageCycleLength)
	}
	if stats.MedianCycleLength != 29 {
		t.Fatalf("expected median cycle length 29, got %d", stats.MedianCycleLength)
	}
	if stats.AveragePeriodLength != 6 {
		t.Fatalf("expected average period length 6, got %.2f", stats.AveragePeriodLength)
	}
	if stats.LastPeriodStart.Format("2006-01-02") != "2026-02-16" {
		t.Fatalf("expected last period start 2026-02-16, got %s", stats.LastPeriodStart.Format("2006-01-02"))
	}
	if stats.NextPeriodStart.Format("2006-01-02") != "2026-03-17" {
		t.Fatalf("expected next period start 2026-03-17, got %s", stats.NextPeriodStart.Format("2006-01-02"))
	}
	if stats.CurrentCycleDay != 2 {
		t.Fatalf("expected current cycle day 2, got %d", stats.CurrentCycleDay)
	}
	if stats.CurrentPhase != "menstrual" {
		t.Fatalf("expected menstrual phase, got %s", stats.CurrentPhase)
	}
}

func TestApplyUserCycleBaseline_MarksIncompatibleCycleAsUncalculable(t *testing.T) {
	handler := &Handler{
		location:        time.UTC,
		lutealPhaseDays: 14,
	}

	userLastPeriod := mustParseBaselineDay(t, "2026-02-10")
	user := &models.User{
		Role:            models.RoleOwner,
		CycleLength:     15,
		PeriodLength:    10,
		LastPeriodStart: &userLastPeriod,
	}

	logs := []models.DailyLog{
		{Date: mustParseBaselineDay(t, "2026-02-10"), IsPeriod: true, Flow: models.FlowMedium},
	}

	now := mustParseBaselineDay(t, "2026-02-12")
	stats := services.BuildCycleStats(logs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, logs, stats, now)

	if !stats.OvulationDate.IsZero() {
		t.Fatalf("expected empty ovulation date for incompatible values, got %s", stats.OvulationDate.Format("2006-01-02"))
	}
	if !stats.FertilityWindowStart.IsZero() || !stats.FertilityWindowEnd.IsZero() {
		t.Fatalf("expected empty fertility window for incompatible values")
	}
	if stats.OvulationExact {
		t.Fatalf("expected exact=false for incompatible values")
	}
	if !stats.OvulationImpossible {
		t.Fatalf("expected impossible=true for incompatible values")
	}
}
