package api

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/models"
)

func TestFetchSymptomsBackfillsMissingBuiltinSymptoms(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "lume-symptom-backfill.db")
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
		Email:               "symptoms@example.com",
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

	oldBuiltin := models.DefaultBuiltinSymptoms()[:7]
	records := make([]models.SymptomType, 0, len(oldBuiltin))
	for _, symptom := range oldBuiltin {
		records = append(records, models.SymptomType{
			UserID:    user.ID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}
	if err := database.Create(&records).Error; err != nil {
		t.Fatalf("seed old builtin symptoms: %v", err)
	}

	handler := &Handler{db: database}
	symptoms, err := handler.fetchSymptoms(user.ID)
	if err != nil {
		t.Fatalf("fetch symptoms: %v", err)
	}

	expected := models.DefaultBuiltinSymptoms()
	if len(symptoms) != len(expected) {
		t.Fatalf("expected %d symptoms after backfill, got %d", len(expected), len(symptoms))
	}

	for index, symptom := range expected {
		if symptoms[index].Name != symptom.Name {
			t.Fatalf("expected symptom %q at index %d, got %q", symptom.Name, index, symptoms[index].Name)
		}
		if !symptoms[index].IsBuiltin {
			t.Fatalf("expected symptom %q to be builtin", symptom.Name)
		}
	}
}

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
