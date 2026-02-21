package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestSanitizeLogForViewer_PartnerSanitizesPrivateFields(t *testing.T) {
	partner := &models.User{Role: models.RolePartner}
	entry := models.DailyLog{
		Date:       time.Date(2026, time.February, 19, 0, 0, 0, 0, time.UTC),
		IsPeriod:   true,
		Flow:       models.FlowMedium,
		Notes:      "private",
		SymptomIDs: []uint{1, 2},
	}

	got := sanitizeLogForViewer(partner, entry)
	if got.Notes != "" {
		t.Fatalf("expected partner notes to be hidden, got %q", got.Notes)
	}
	if len(got.SymptomIDs) != 0 {
		t.Fatalf("expected partner symptom ids to be hidden, got %#v", got.SymptomIDs)
	}
}

func TestSanitizeLogForViewer_OwnerKeepsFields(t *testing.T) {
	owner := &models.User{Role: models.RoleOwner}
	entry := models.DailyLog{
		Notes:      "keep",
		SymptomIDs: []uint{3, 4},
	}

	got := sanitizeLogForViewer(owner, entry)
	if got.Notes != "keep" {
		t.Fatalf("expected owner notes preserved, got %q", got.Notes)
	}
	if len(got.SymptomIDs) != 2 {
		t.Fatalf("expected owner symptom ids preserved, got %#v", got.SymptomIDs)
	}
}

func TestSanitizeLogsForViewer_PartnerSanitizesAll(t *testing.T) {
	partner := &models.User{Role: models.RolePartner}
	logs := []models.DailyLog{
		{Notes: "a", SymptomIDs: []uint{1}},
		{Notes: "b", SymptomIDs: []uint{2, 3}},
	}

	sanitizeLogsForViewer(partner, logs)
	for index := range logs {
		if logs[index].Notes != "" {
			t.Fatalf("expected notes to be hidden for entry %d, got %q", index, logs[index].Notes)
		}
		if len(logs[index].SymptomIDs) != 0 {
			t.Fatalf("expected symptom ids to be hidden for entry %d, got %#v", index, logs[index].SymptomIDs)
		}
	}
}

func TestFetchSymptomsForViewer_NonOwnerReturnsEmpty(t *testing.T) {
	handler := &Handler{}
	partner := &models.User{Role: models.RolePartner}

	symptoms, err := handler.fetchSymptomsForViewer(partner)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(symptoms) != 0 {
		t.Fatalf("expected empty symptoms for non-owner, got %#v", symptoms)
	}
}

func TestFetchSymptomsForViewer_NilUserReturnsEmpty(t *testing.T) {
	handler := &Handler{}

	symptoms, err := handler.fetchSymptomsForViewer(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(symptoms) != 0 {
		t.Fatalf("expected empty symptoms for nil user, got %#v", symptoms)
	}
}

func TestFetchDayLogForViewer_OwnerKeepsPrivateFieldsAndLoadsSymptoms(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	owner := createDataAccessTestUser(t, database, "viewer-owner@example.com")
	day := time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC)

	symptom := models.SymptomType{
		UserID: owner.ID,
		Name:   "Headache",
		Icon:   "🤕",
		Color:  "#FFAA66",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	entry := models.DailyLog{
		UserID:     owner.ID,
		Date:       day,
		IsPeriod:   true,
		Flow:       models.FlowLight,
		Notes:      "private owner note",
		SymptomIDs: []uint{symptom.ID},
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	logEntry, symptoms, err := handler.fetchDayLogForViewer(&owner, day)
	if err != nil {
		t.Fatalf("fetchDayLogForViewer returned error: %v", err)
	}
	if logEntry.Notes != "private owner note" {
		t.Fatalf("expected owner notes preserved, got %q", logEntry.Notes)
	}
	if len(logEntry.SymptomIDs) != 1 || logEntry.SymptomIDs[0] != symptom.ID {
		t.Fatalf("expected owner symptom ids preserved, got %#v", logEntry.SymptomIDs)
	}
	if len(symptoms) == 0 {
		t.Fatal("expected owner symptom list to be loaded")
	}
}

func TestFetchDayLogForViewer_PartnerSanitizesAndSkipsSymptoms(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	partner := models.User{
		Email:               "viewer-partner@example.com",
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
		Notes:      "partner-private-note",
		SymptomIDs: []uint{1, 2, 3},
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	logEntry, symptoms, err := handler.fetchDayLogForViewer(&partner, day)
	if err != nil {
		t.Fatalf("fetchDayLogForViewer returned error: %v", err)
	}
	if logEntry.Notes != "" {
		t.Fatalf("expected partner notes hidden, got %q", logEntry.Notes)
	}
	if len(logEntry.SymptomIDs) != 0 {
		t.Fatalf("expected partner symptom ids hidden, got %#v", logEntry.SymptomIDs)
	}
	if len(symptoms) != 0 {
		t.Fatalf("expected partner symptoms to be skipped, got %#v", symptoms)
	}
}
