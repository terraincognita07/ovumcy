package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestBuildStatsTrendViewPartnerHasNoBaseline(t *testing.T) {
	t.Parallel()

	handler := &Handler{location: time.UTC}
	now := time.Date(2026, time.February, 21, 12, 0, 0, 0, time.UTC)
	user := &models.User{Role: models.RolePartner, CycleLength: 28}

	chartPayload, baseline, trendCount := handler.buildStatsTrendView(user, []models.DailyLog{}, now, map[string]string{})
	if baseline != 0 {
		t.Fatalf("expected partner baseline 0, got %d", baseline)
	}
	if trendCount != 0 {
		t.Fatalf("expected zero trend count for empty logs, got %d", trendCount)
	}
	if _, ok := chartPayload["baseline"]; ok {
		t.Fatal("did not expect baseline field for partner chart payload")
	}
	if _, ok := chartPayload["values"].([]int); !ok {
		t.Fatalf("expected values slice in chart payload, got %T", chartPayload["values"])
	}
}
