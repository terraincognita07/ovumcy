package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
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

func TestBuildDayEditorPartialDataSetsFutureFlagAndNoDataLabel(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "day-editor-view-data@example.com")

	now := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	futureDay := now.AddDate(0, 0, 1)

	messages := map[string]string{
		"common.not_available": "N/A",
	}
	payload, errorMessage, err := handler.buildDayEditorPartialData(&user, "en", messages, futureDay, now)
	if err != nil {
		t.Fatalf("buildDayEditorPartialData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}

	if isFutureDate, ok := payload["IsFutureDate"].(bool); !ok || !isFutureDate {
		t.Fatalf("expected IsFutureDate=true, got %#v", payload["IsFutureDate"])
	}
	if noDataLabel, ok := payload["NoDataLabel"].(string); !ok || noDataLabel != "N/A" {
		t.Fatalf("expected NoDataLabel=N/A, got %#v", payload["NoDataLabel"])
	}
	if hasDayData, ok := payload["HasDayData"].(bool); !ok || hasDayData {
		t.Fatalf("expected HasDayData=false for empty future day, got %#v", payload["HasDayData"])
	}
}
