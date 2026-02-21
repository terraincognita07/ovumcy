package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestBuildStatsSymptomCountsNonOwnerSkipsDataAccess(t *testing.T) {
	t.Parallel()

	handler := &Handler{}
	user := &models.User{Role: models.RolePartner}

	counts, errorMessage, err := handler.buildStatsSymptomCounts(user, "en")
	if err != nil {
		t.Fatalf("buildStatsSymptomCounts returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if len(counts) != 0 {
		t.Fatalf("expected empty counts for non-owner, got %#v", counts)
	}
}

func TestBuildStatsSymptomCountsOwnerReturnsLocalizedCounts(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "stats-symptom-owner@example.com")

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Headache",
		Icon:   "🤕",
		Color:  "#CC8844",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC),
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptom.ID},
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	counts, errorMessage, err := handler.buildStatsSymptomCounts(&user, "en")
	if err != nil {
		t.Fatalf("buildStatsSymptomCounts returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if len(counts) != 1 {
		t.Fatalf("expected one symptom count entry, got %d", len(counts))
	}
	if counts[0].Count != 1 || counts[0].TotalDays != 1 {
		t.Fatalf("unexpected count payload: %#v", counts[0])
	}
	if counts[0].FrequencySummary == "" {
		t.Fatalf("expected localized frequency summary, got empty value")
	}
}
