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
