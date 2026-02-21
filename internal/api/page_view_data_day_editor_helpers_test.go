package api

import (
	"testing"
	"time"
)

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
