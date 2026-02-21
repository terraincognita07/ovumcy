package api

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/models"
)

func TestFetchSymptoms_NormalizesLegacyFatigueTypo(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "lume-symptom-legacy-fatigue.db")
	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	user := models.User{
		Email:               "legacy-fatigue@example.com",
		PasswordHash:        "hash",
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

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

func TestFetchSymptoms_NormalizesLegacyFatigueTypoForCustomSymptom(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "lume-symptom-legacy-fatigue-custom.db")
	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	user := models.User{
		Email:               "legacy-fatigue-custom@example.com",
		PasswordHash:        "hash",
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

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
