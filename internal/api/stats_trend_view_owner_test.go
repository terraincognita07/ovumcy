package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestBuildStatsTrendViewTrimsPointsAndAppliesOwnerBaseline(t *testing.T) {
	t.Parallel()

	handler := &Handler{location: time.UTC}
	now := time.Date(2026, time.February, 21, 12, 0, 0, 0, time.UTC)
	today := dateAtLocation(now, time.UTC)

	start := today.AddDate(0, 0, -(14*28 + 2))
	logs := make([]models.DailyLog, 0, 15)
	for index := 0; index < 15; index++ {
		logs = append(logs, models.DailyLog{
			ID:       uint(index + 1),
			Date:     start.AddDate(0, 0, index*28),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		})
	}

	user := &models.User{Role: models.RoleOwner, CycleLength: 28}
	chartPayload, baseline, trendCount := handler.buildStatsTrendView(user, logs, now, map[string]string{"stats.cycle_label": "Cycle %d"})

	if baseline != 28 {
		t.Fatalf("expected baseline 28, got %d", baseline)
	}
	if trendCount != maxStatsTrendPoints {
		t.Fatalf("expected trimmed trend point count %d, got %d", maxStatsTrendPoints, trendCount)
	}

	labels, labelsOK := chartPayload["labels"].([]string)
	values, valuesOK := chartPayload["values"].([]int)
	if !labelsOK || !valuesOK {
		t.Fatalf("expected labels/values slices in chart payload, got labels=%T values=%T", chartPayload["labels"], chartPayload["values"])
	}
	if len(labels) != maxStatsTrendPoints || len(values) != maxStatsTrendPoints {
		t.Fatalf("expected %d labels/values, got %d/%d", maxStatsTrendPoints, len(labels), len(values))
	}
	if _, ok := chartPayload["baseline"]; !ok {
		t.Fatal("expected baseline field in owner chart payload")
	}
}
