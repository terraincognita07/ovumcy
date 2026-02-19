package api

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestExportCSVIncludesKnownAndOtherSymptoms(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "export-csv@example.com", "StrongPass1", true)

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "Cramps", Icon: "A", Color: "#111111"},
		{UserID: user.ID, Name: "Custom Symptom", Icon: "B", Color: "#222222"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC),
		IsPeriod:   true,
		Flow:       models.FlowLight,
		SymptomIDs: []uint{symptoms[0].ID, symptoms[1].ID},
		Notes:      "note",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	request := httptest.NewRequest(http.MethodGet, "/api/export/csv", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("export csv request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); !strings.Contains(got, "text/csv") {
		t.Fatalf("expected text/csv content type, got %q", got)
	}
	if got := response.Header.Get("Content-Disposition"); !strings.Contains(got, "attachment; filename=lume-export-") {
		t.Fatalf("expected attachment filename header, got %q", got)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	records, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row, got %d rows", len(records))
	}

	header := records[0]
	row := records[1]
	indexByName := make(map[string]int, len(header))
	for index, name := range header {
		indexByName[name] = index
	}

	if got := row[indexByName["Date"]]; got != "2026-02-18" {
		t.Fatalf("expected date 2026-02-18, got %q", got)
	}
	if got := row[indexByName["Period"]]; got != "Yes" {
		t.Fatalf("expected period Yes, got %q", got)
	}
	if got := row[indexByName["Flow"]]; got != "Light" {
		t.Fatalf("expected flow Light, got %q", got)
	}
	if got := row[indexByName["Cramps"]]; got != "Yes" {
		t.Fatalf("expected Cramps Yes, got %q", got)
	}
	if got := row[indexByName["Other"]]; got != "Custom Symptom" {
		t.Fatalf("expected Other to include custom symptom, got %q", got)
	}
	if got := row[indexByName["Notes"]]; got != "note" {
		t.Fatalf("expected notes to be exported, got %q", got)
	}
}

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
