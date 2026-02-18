package api

import (
	"path/filepath"
	"testing"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/models"
)

func TestCalculateSymptomFrequencies_IncludesTotalDaysContext(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "lume-symptom-frequency.db")
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
		Email:        "symptom-frequency@example.com",
		PasswordHash: "hash",
		Role:         models.RoleOwner,
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "Custom A", Icon: "A", Color: "#A66F5A"},
		{UserID: user.ID, Name: "Custom B", Icon: "B", Color: "#5A74A6"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	logs := []models.DailyLog{
		{SymptomIDs: []uint{symptoms[0].ID}},
		{SymptomIDs: []uint{symptoms[0].ID, symptoms[1].ID}},
		{SymptomIDs: []uint{}},
	}

	handler := &Handler{db: database}
	counts, err := handler.calculateSymptomFrequencies(user.ID, logs)
	if err != nil {
		t.Fatalf("calculate symptom frequencies: %v", err)
	}

	if len(counts) != 2 {
		t.Fatalf("expected 2 symptom counts, got %d", len(counts))
	}
	if counts[0].Name != "Custom A" || counts[0].Count != 2 {
		t.Fatalf("expected first symptom to be Custom A with count 2, got %#v", counts[0])
	}
	for _, count := range counts {
		if count.TotalDays != len(logs) {
			t.Fatalf("expected total days %d, got %d for %s", len(logs), count.TotalDays, count.Name)
		}
	}
}
