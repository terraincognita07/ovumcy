package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestClearDataRemovesTrackedCalendarEntriesAndResetsCycleSettings(t *testing.T) {
	scenario := setupClearDataScenario(t)

	request := httptest.NewRequest(http.MethodPost, "/api/settings/clear-data", strings.NewReader(url.Values{}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", scenario.authCookie)

	response, err := scenario.app.Test(request, -1)
	if err != nil {
		t.Fatalf("clear data request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected clear data status 200, got %d", response.StatusCode)
	}

	assertClearDataPostconditions(t, scenario.database, scenario.user)
}
