package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func TestApplyUserCycleBaseline_DoesNotOverrideReliableCycleData(t *testing.T) {
	handler := &Handler{
		location: time.UTC,
	}

	userLastPeriod := mustParseBaselineDay(t, "2025-03-27")
	user := &models.User{
		Role:            models.RoleOwner,
		CycleLength:     29,
		PeriodLength:    6,
		LastPeriodStart: &userLastPeriod,
	}

	logs := []models.DailyLog{
		{Date: mustParseBaselineDay(t, "2025-01-01"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-01-02"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-01-03"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-01-29"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-01-30"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-01-31"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-02-26"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-02-27"), IsPeriod: true, Flow: models.FlowMedium},
		{Date: mustParseBaselineDay(t, "2025-02-28"), IsPeriod: true, Flow: models.FlowMedium},
	}

	now := mustParseBaselineDay(t, "2025-03-05")
	stats := services.BuildCycleStats(logs, now)
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
