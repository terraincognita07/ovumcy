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

	userLastPeriod := mustParseBaselineDay("2026-02-07")
	user := &models.User{
		Role:            models.RoleOwner,
		CycleLength:     29,
		PeriodLength:    6,
		LastPeriodStart: &userLastPeriod,
	}

	logs := []models.DailyLog{
		{Date: mustParseBaselineDay("2026-02-07"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2026-02-16"), IsPeriod: true, Flow: models.FlowMedium},
	}

	now := mustParseBaselineDay("2026-02-17")
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

func TestApplyUserCycleBaseline_DoesNotOverrideReliableCycleData(t *testing.T) {
	handler := &Handler{
		location:        time.UTC,
		lutealPhaseDays: 14,
	}

	userLastPeriod := mustParseBaselineDay("2025-03-27")
	user := &models.User{
		Role:            models.RoleOwner,
		CycleLength:     29,
		PeriodLength:    6,
		LastPeriodStart: &userLastPeriod,
	}

	logs := []models.DailyLog{
		{Date: mustParseBaselineDay("2025-01-01"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-01-02"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-01-03"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-01-29"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-01-30"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-01-31"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-02-26"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-02-27"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay("2025-02-28"), IsPeriod: true, Flow: models.FlowMedium},
	}

	now := mustParseBaselineDay("2025-03-05")
	stats := services.BuildCycleStats(logs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, logs, stats, now)

	if stats.AverageCycleLength != 28 {
		t.Fatalf("expected reliable average cycle length 28, got %.2f", stats.AverageCycleLength)
	}
	if stats.MedianCycleLength != 28 {
		t.Fatalf("expected reliable median cycle length 28, got %d", stats.MedianCycleLength)
	}
	if stats.LastPeriodStart.Format("2006-01-02") != "2025-02-26" {
		t.Fatalf("expected reliable last period start 2025-02-26, got %s", stats.LastPeriodStart.Format("2006-01-02"))
	}
}

func mustParseBaselineDay(raw string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02", raw, time.UTC)
	if err != nil {
		panic(err)
	}
	return parsed
}
