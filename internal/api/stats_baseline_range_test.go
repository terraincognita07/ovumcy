package api

import (
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestBuildCycleStatsForRange_AppliesOnboardingBaseline(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "stats-range@example.com")

	lastPeriodStart := mustParseBaselineDay(t, "2026-02-16")
	user.LastPeriodStart = &lastPeriodStart
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Update("last_period_start", lastPeriodStart).Error; err != nil {
		t.Fatalf("update user last period start: %v", err)
	}

	entry := models.DailyLog{
		UserID:   user.ID,
		Date:     lastPeriodStart,
		IsPeriod: true,
		Flow:     models.FlowMedium,
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	now := mustParseBaselineDay(t, "2026-02-20")
	stats, logs, err := handler.buildCycleStatsForRange(&user, now.AddDate(0, 0, -10), now, now)
	if err != nil {
		t.Fatalf("buildCycleStatsForRange returned error: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("expected one log in range, got %d", len(logs))
	}
	if stats.AverageCycleLength != 28 {
		t.Fatalf("expected baseline average cycle length 28, got %.2f", stats.AverageCycleLength)
	}
	if stats.AveragePeriodLength != 5 {
		t.Fatalf("expected baseline average period length 5, got %.2f", stats.AveragePeriodLength)
	}
}
