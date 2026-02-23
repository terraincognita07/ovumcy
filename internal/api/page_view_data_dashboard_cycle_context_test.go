package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func TestBuildDashboardViewDataFlagsLongCycleAndPastPredictions(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "dashboard-cycle-context@example.com")

	now := time.Date(2026, time.February, 22, 0, 0, 0, 0, time.UTC)
	lastPeriodStart := now.AddDate(0, 0, -60)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":      28,
		"period_length":     5,
		"last_period_start": lastPeriodStart,
	}).Error; err != nil {
		t.Fatalf("update user cycle context: %v", err)
	}

	user.CycleLength = 28
	user.PeriodLength = 5
	user.LastPeriodStart = &lastPeriodStart

	data, errorMessage, err := handler.buildDashboardViewData(&user, "en", map[string]string{
		"meta.title.dashboard": "Dashboard",
	}, now)
	if err != nil {
		t.Fatalf("buildDashboardViewData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}

	if reference, ok := data["CycleDayReference"].(int); !ok || reference != 28 {
		t.Fatalf("expected cycle day reference 28, got %#v", data["CycleDayReference"])
	}
	if warning, ok := data["CycleDayWarning"].(bool); !ok || !warning {
		t.Fatalf("expected long cycle day warning=true, got %#v", data["CycleDayWarning"])
	}
	if past, ok := data["NextPeriodInPast"].(bool); !ok || !past {
		t.Fatalf("expected next period to be marked as past, got %#v", data["NextPeriodInPast"])
	}
	if past, ok := data["OvulationInPast"].(bool); !ok || !past {
		t.Fatalf("expected ovulation to be marked as past, got %#v", data["OvulationInPast"])
	}
}

func TestBuildDashboardViewDataMarksOvulationAsUncalculableForIncompatibleCycle(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "dashboard-ovulation-impossible@example.com")

	lastPeriodStart := time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":      15,
		"period_length":     10,
		"last_period_start": lastPeriodStart,
	}).Error; err != nil {
		t.Fatalf("update user cycle context: %v", err)
	}
	if err := database.Create(&models.DailyLog{
		UserID:   user.ID,
		Date:     lastPeriodStart,
		IsPeriod: true,
		Flow:     models.FlowMedium,
	}).Error; err != nil {
		t.Fatalf("create period log: %v", err)
	}

	user.CycleLength = 15
	user.PeriodLength = 10
	user.LastPeriodStart = &lastPeriodStart

	now := time.Date(2026, time.February, 22, 0, 0, 0, 0, time.UTC)
	data, errorMessage, err := handler.buildDashboardViewData(&user, "en", map[string]string{
		"meta.title.dashboard": "Dashboard",
	}, now)
	if err != nil {
		t.Fatalf("buildDashboardViewData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}

	stats, ok := data["Stats"].(services.CycleStats)
	if !ok {
		t.Fatalf("expected cycle stats in view data, got %#v", data["Stats"])
	}
	if !stats.OvulationImpossible {
		t.Fatalf("expected ovulation impossible flag=true for incompatible cycle values")
	}
	if !stats.OvulationDate.IsZero() {
		t.Fatalf("expected no ovulation date for incompatible cycle values")
	}
}
