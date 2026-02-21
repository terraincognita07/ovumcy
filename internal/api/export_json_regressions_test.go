package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestExportJSONNormalizesFlowAndMapsSymptoms(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "export-json@example.com", "StrongPass1", true)

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "Mood swings", Icon: "A", Color: "#111111"},
		{UserID: user.ID, Name: "My Custom", Icon: "B", Color: "#222222"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 19, 0, 0, 0, 0, time.UTC),
		IsPeriod:   false,
		Flow:       "unexpected-flow",
		SymptomIDs: []uint{symptoms[0].ID, symptoms[1].ID},
		Notes:      "json-note",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	request := httptest.NewRequest(http.MethodGet, "/api/export/json", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("export json request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected application/json content type, got %q", got)
	}
	if got := response.Header.Get("Content-Disposition"); !strings.Contains(got, "attachment; filename=lume-export-") {
		t.Fatalf("expected attachment filename header, got %q", got)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	payload := struct {
		ExportedAt string            `json:"exported_at"`
		Entries    []exportJSONEntry `json:"entries"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode json payload: %v", err)
	}
	if payload.ExportedAt == "" {
		t.Fatalf("expected exported_at in payload")
	}
	if _, err := time.Parse(time.RFC3339, payload.ExportedAt); err != nil {
		t.Fatalf("expected RFC3339 exported_at, got %q", payload.ExportedAt)
	}
	if len(payload.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(payload.Entries))
	}

	entry := payload.Entries[0]
	if entry.Flow != models.FlowNone {
		t.Fatalf("expected unknown flow normalized to %q, got %q", models.FlowNone, entry.Flow)
	}
	if !entry.Symptoms.Mood {
		t.Fatalf("expected mood flag to be true")
	}
	if len(entry.OtherSymptoms) != 1 || entry.OtherSymptoms[0] != "My Custom" {
		t.Fatalf("expected custom symptom in other list, got %#v", entry.OtherSymptoms)
	}
	if entry.Notes != "json-note" {
		t.Fatalf("expected notes to be preserved, got %q", entry.Notes)
	}
}
