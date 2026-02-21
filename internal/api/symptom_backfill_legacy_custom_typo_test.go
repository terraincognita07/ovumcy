package api

import (
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestFetchSymptoms_NormalizesLegacyFatigueTypoForCustomSymptom(t *testing.T) {
	database := newLegacySymptomTestDB(t, "lume-symptom-legacy-fatigue-custom.db")
	user := createLegacySymptomTestUser(t, database, "legacy-fatigue-custom@example.com")

	customLegacy := models.SymptomType{
		UserID:    user.ID,
		Name:      "Fatique",
		Icon:      "🧪",
		Color:     "#8A6EA9",
		IsBuiltin: false,
	}
	if err := database.Create(&customLegacy).Error; err != nil {
		t.Fatalf("seed custom legacy fatigue typo: %v", err)
	}

	handler := &Handler{db: database}
	symptoms, err := handler.fetchSymptoms(user.ID)
	if err != nil {
		t.Fatalf("fetch symptoms: %v", err)
	}

	foundNormalizedCustom := false
	for _, symptom := range symptoms {
		if symptom.Name == "Fatique" {
			t.Fatalf("unexpected legacy typo in fetched symptoms: %#v", symptom)
		}
		if symptom.Icon == customLegacy.Icon && symptom.Name == "Fatigue" {
			foundNormalizedCustom = true
		}
	}

	if !foundNormalizedCustom {
		t.Fatalf("expected custom symptom %q to be normalized to Fatigue", customLegacy.Icon)
	}
}
