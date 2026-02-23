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

	futureDateForm := url.Values{
		"cycle_length":      {"28"},
		"period_length":     {"6"},
		"last_period_start": {"2999-01-01"},
	}
	futureDateRequest := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(futureDateForm.Encode()))
	futureDateRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	futureDateRequest.Header.Set("HX-Request", "true")
	futureDateRequest.Header.Set("Cookie", authCookie)

	futureDateResponse, err := app.Test(futureDateRequest, -1)
	if err != nil {
		t.Fatalf("future-date settings request failed: %v", err)
	}
	defer futureDateResponse.Body.Close()
	if futureDateResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected HTMX status 200 for future last_period_start, got %d", futureDateResponse.StatusCode)
	}
	futureDateBody, err := io.ReadAll(futureDateResponse.Body)
	if err != nil {
		t.Fatalf("read future-date response body: %v", err)
	}
	if !strings.Contains(string(futureDateBody), "status-error") {
		t.Fatalf("expected status-error markup for future last_period_start, got %q", string(futureDateBody))
	}

	tooOldDateForm := url.Values{
		"cycle_length":      {"28"},
		"period_length":     {"6"},
		"last_period_start": {"1969-12-31"},
	}
	tooOldDateRequest := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(tooOldDateForm.Encode()))
	tooOldDateRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tooOldDateRequest.Header.Set("HX-Request", "true")
	tooOldDateRequest.Header.Set("Cookie", authCookie)

	tooOldDateResponse, err := app.Test(tooOldDateRequest, -1)
	if err != nil {
		t.Fatalf("too-old-date settings request failed: %v", err)
	}
	defer tooOldDateResponse.Body.Close()
	if tooOldDateResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected HTMX status 200 for too-old last_period_start, got %d", tooOldDateResponse.StatusCode)
	}
	tooOldDateBody, err := io.ReadAll(tooOldDateResponse.Body)
	if err != nil {
		t.Fatalf("read too-old-date response body: %v", err)
	}
	if !strings.Contains(string(tooOldDateBody), "status-error") {
		t.Fatalf("expected status-error markup for too-old last_period_start, got %q", string(tooOldDateBody))
	}
}
