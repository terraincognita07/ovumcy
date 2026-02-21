package api

import (
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestFetchSymptoms_NormalizesLegacyFatigueTypo(t *testing.T) {
	database := newLegacySymptomTestDB(t, "lume-symptom-legacy-fatigue.db")
	user := createLegacySymptomTestUser(t, database, "legacy-fatigue@example.com")

	if err := database.Create(&models.SymptomType{
		UserID:    user.ID,
		Name:      "Fatique",
		Icon:      "😴",
		Color:     "#95A5A6",
		IsBuiltin: true,
	}).Error; err != nil {
		t.Fatalf("seed legacy fatigue typo: %v", err)
	}

	handler := &Handler{db: database}
	symptoms, err := handler.fetchSymptoms(user.ID)
	if err != nil {
		t.Fatalf("fetch symptoms: %v", err)
	}

	fatigueCount := 0
	legacyCount := 0
	for _, symptom := range symptoms {
		switch symptom.Name {
		case "Fatigue":
			fatigueCount++
		case "Fatique":
			legacyCount++
		}
	}
	if fatigueCount != 1 {
		t.Fatalf("expected exactly one Fatigue symptom, got %d", fatigueCount)
	}
	if legacyCount != 0 {
		t.Fatalf("expected no legacy Fatique symptoms, got %d", legacyCount)
	}
}
