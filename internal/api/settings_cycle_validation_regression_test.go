package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSettingsCycleRejectsOutOfRangeAndIncompatibleValues(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-cycle-validation@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	outOfRangeForm := url.Values{
		"cycle_length":  {"28"},
		"period_length": {"15"},
	}
	outOfRangeRequest := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(outOfRangeForm.Encode()))
	outOfRangeRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	outOfRangeRequest.Header.Set("HX-Request", "true")
	outOfRangeRequest.Header.Set("Cookie", authCookie)

	outOfRangeResponse, err := app.Test(outOfRangeRequest, -1)
	if err != nil {
		t.Fatalf("out-of-range settings request failed: %v", err)
	}
	defer outOfRangeResponse.Body.Close()
	if outOfRangeResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected HTMX status 200 for period_length=15, got %d", outOfRangeResponse.StatusCode)
	}
	outOfRangeBody, err := io.ReadAll(outOfRangeResponse.Body)
	if err != nil {
		t.Fatalf("read out-of-range response body: %v", err)
	}
	if !strings.Contains(string(outOfRangeBody), "status-error") {
		t.Fatalf("expected status-error markup for period_length=15, got %q", string(outOfRangeBody))
	}

	incompatibleForm := url.Values{
		"cycle_length":  {"21"},
		"period_length": {"14"},
	}
	incompatibleRequest := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(incompatibleForm.Encode()))
	incompatibleRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	incompatibleRequest.Header.Set("HX-Request", "true")
	incompatibleRequest.Header.Set("Cookie", authCookie)

	incompatibleResponse, err := app.Test(incompatibleRequest, -1)
	if err != nil {
		t.Fatalf("incompatible settings request failed: %v", err)
	}
	defer incompatibleResponse.Body.Close()
	if incompatibleResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected HTMX status 200 for incompatible cycle/period values, got %d", incompatibleResponse.StatusCode)
	}
	incompatibleBody, err := io.ReadAll(incompatibleResponse.Body)
	if err != nil {
		t.Fatalf("read incompatible response body: %v", err)
	}
	if !strings.Contains(string(incompatibleBody), "status-error") {
		t.Fatalf("expected status-error markup for incompatible cycle/period values, got %q", string(incompatibleBody))
	}
}
