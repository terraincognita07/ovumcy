package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestValidateSymptomIDsDeduplicatesAndSorts(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "validate-symptoms@example.com")

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "A", Icon: "A", Color: "#111111"},
		{UserID: user.ID, Name: "B", Icon: "B", Color: "#222222"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	ids, err := handler.validateSymptomIDs(user.ID, []uint{symptoms[1].ID, symptoms[0].ID, symptoms[1].ID})
	if err != nil {
		t.Fatalf("validateSymptomIDs returned error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
	if ids[0] != symptoms[0].ID || ids[1] != symptoms[1].ID {
		t.Fatalf("expected sorted unique ids [%d %d], got %#v", symptoms[0].ID, symptoms[1].ID, ids)
	}

	if _, err := handler.validateSymptomIDs(user.ID, []uint{symptoms[0].ID, 999999}); err == nil {
		t.Fatal("expected error for invalid symptom id")
	}
}

func TestRemoveSymptomFromLogsUpdatesEntries(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "remove-symptom@example.com")

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "A", Icon: "A", Color: "#111111"},
		{UserID: user.ID, Name: "B", Icon: "B", Color: "#222222"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	day := time.Date(2026, time.February, 17, 0, 0, 0, 0, time.UTC)
	entry := models.DailyLog{
		UserID:     user.ID,
		Date:       day,
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptoms[0].ID, symptoms[1].ID, symptoms[0].ID},
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create log: %v", err)
	}

	if err := handler.removeSymptomFromLogs(user.ID, symptoms[0].ID); err != nil {
		t.Fatalf("removeSymptomFromLogs returned error: %v", err)
	}

	updated := models.DailyLog{}
	if err := database.First(&updated, entry.ID).Error; err != nil {
		t.Fatalf("load updated log: %v", err)
	}
	if len(updated.SymptomIDs) != 1 || updated.SymptomIDs[0] != symptoms[1].ID {
		t.Fatalf("expected only remaining symptom id %d, got %#v", symptoms[1].ID, updated.SymptomIDs)
	}
}
