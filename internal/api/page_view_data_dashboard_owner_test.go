package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestBuildDashboardViewDataOwnerIncludesSymptomsAndNotes(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "dashboard-owner-view-data@example.com")

	day := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Headache",
		Icon:   "🤕",
		Color:  "#CC8844",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	entry := models.DailyLog{
		UserID:     user.ID,
		Date:       day,
		IsPeriod:   false,
		Flow:       models.FlowNone,
		Notes:      "owner note",
		SymptomIDs: []uint{symptom.ID},
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	messages := map[string]string{
		"meta.title.dashboard": "Dashboard",
	}
	data, errorMessage, err := handler.buildDashboardViewData(&user, "en", messages, day)
	if err != nil {
		t.Fatalf("buildDashboardViewData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}

	if got, ok := data["Today"].(string); !ok || got != "2026-02-21" {
		t.Fatalf("expected Today=2026-02-21, got %#v", data["Today"])
	}
	logEntry, ok := data["TodayLog"].(models.DailyLog)
	if !ok {
		t.Fatalf("expected TodayLog type models.DailyLog, got %T", data["TodayLog"])
	}
	if logEntry.Notes != "owner note" {
		t.Fatalf("expected owner notes preserved, got %q", logEntry.Notes)
	}
	todayEntry, ok := data["TodayEntry"].(models.DailyLog)
	if !ok {
		t.Fatalf("expected TodayEntry type models.DailyLog, got %T", data["TodayEntry"])
	}
	if todayEntry.Notes != "owner note" {
		t.Fatalf("expected owner TodayEntry notes preserved, got %q", todayEntry.Notes)
	}
	if len(logEntry.SymptomIDs) != 1 || logEntry.SymptomIDs[0] != symptom.ID {
		t.Fatalf("expected owner symptom ids preserved, got %#v", logEntry.SymptomIDs)
	}

	symptoms, ok := data["Symptoms"].([]models.SymptomType)
	if !ok {
		t.Fatalf("expected Symptoms type []models.SymptomType, got %T", data["Symptoms"])
	}
	if len(symptoms) == 0 {
		t.Fatal("expected owner symptoms list to be populated")
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || !isOwner {
		t.Fatalf("expected IsOwner=true, got %#v", data["IsOwner"])
	}
}
