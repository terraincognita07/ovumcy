package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
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
