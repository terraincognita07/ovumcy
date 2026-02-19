package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func TestUpsertDayNormalizesFlowWhenNotPeriod(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-normalize@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	payload := map[string]any{
		"is_period":   false,
		"flow":        models.FlowHeavy,
		"symptom_ids": []uint{},
		"notes":       "note",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-19", bytes.NewReader(body))
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("upsert request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	day, err := parseDayParam("2026-02-19", time.UTC)
	if err != nil {
		t.Fatalf("parse day for assertion: %v", err)
	}
	entry, err := (&Handler{db: database, location: time.UTC}).fetchLogByDate(user.ID, day)
	if err != nil {
		t.Fatalf("load stored log: %v", err)
	}
	if entry.Flow != models.FlowNone {
		t.Fatalf("expected non-period flow normalized to %q, got %q", models.FlowNone, entry.Flow)
	}
}

func TestUpsertDayRejectsPeriodWithoutFlow(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-invalid-flow@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	payload := map[string]any{
		"is_period":   true,
		"flow":        models.FlowNone,
		"symptom_ids": []uint{},
		"notes":       "note",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-19", bytes.NewReader(body))
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("Accept", fiber.MIMEApplicationJSON)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("upsert request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "period flow is required" {
		t.Fatalf("expected period flow validation error, got %q", errorValue)
	}
}

func TestDeleteSymptomRemovesIDFromLogs(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "delete-symptom-cleanup@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Custom",
		Icon:   "A",
		Color:  "#111111",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC),
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptom.ID, symptom.ID},
		Notes:      "",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	request := httptest.NewRequest(http.MethodDelete, "/api/symptoms/"+strconv.FormatUint(uint64(symptom.ID), 10), nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("delete symptom request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected status 200, got %d: %s", response.StatusCode, string(body))
	}

	var deletedCount int64
	if err := database.Model(&models.SymptomType{}).Where("id = ?", symptom.ID).Count(&deletedCount).Error; err != nil {
		t.Fatalf("count deleted symptom: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("expected symptom to be deleted, count=%d", deletedCount)
	}

	updated := models.DailyLog{}
	if err := database.First(&updated, logEntry.ID).Error; err != nil {
		t.Fatalf("load updated log: %v", err)
	}
	if len(updated.SymptomIDs) != 0 {
		t.Fatalf("expected symptom IDs to be removed from log, got %#v", updated.SymptomIDs)
	}
}
