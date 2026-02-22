package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestBuildDashboardViewDataPartnerSanitizesPrivateFields(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	partner := models.User{
		Email:               "dashboard-partner-view-data@example.com",
		PasswordHash:        "test-hash",
		Role:                models.RolePartner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&partner).Error; err != nil {
		t.Fatalf("create partner user: %v", err)
	}

	day := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	entry := models.DailyLog{
		UserID:     partner.ID,
		Date:       day,
		IsPeriod:   true,
		Flow:       models.FlowMedium,
		Notes:      "private note",
		SymptomIDs: []uint{1, 2},
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	data, errorMessage, err := handler.buildDashboardViewData(&partner, "en", map[string]string{}, day)
	if err != nil {
		t.Fatalf("buildDashboardViewData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}

	logEntry, ok := data["TodayLog"].(models.DailyLog)
	if !ok {
		t.Fatalf("expected TodayLog type models.DailyLog, got %T", data["TodayLog"])
	}
	if logEntry.Notes != "" {
		t.Fatalf("expected partner notes hidden, got %q", logEntry.Notes)
	}
	todayEntry, ok := data["TodayEntry"].(models.DailyLog)
	if !ok {
		t.Fatalf("expected TodayEntry type models.DailyLog, got %T", data["TodayEntry"])
	}
	if todayEntry.Notes != "" {
		t.Fatalf("expected partner TodayEntry notes hidden, got %q", todayEntry.Notes)
	}
	if len(logEntry.SymptomIDs) != 0 {
		t.Fatalf("expected partner symptom ids hidden, got %#v", logEntry.SymptomIDs)
	}

	symptoms, ok := data["Symptoms"].([]models.SymptomType)
	if !ok {
		t.Fatalf("expected Symptoms type []models.SymptomType, got %T", data["Symptoms"])
	}
	if len(symptoms) != 0 {
		t.Fatalf("expected empty partner symptoms list, got %#v", symptoms)
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || isOwner {
		t.Fatalf("expected IsOwner=false, got %#v", data["IsOwner"])
	}
}
