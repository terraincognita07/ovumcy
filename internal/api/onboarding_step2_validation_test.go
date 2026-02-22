package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestOnboardingStep2RejectsOutOfRangeValues(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "step2-validation@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	invalidCycleForm := url.Values{
		"cycle_length":  {"14"},
		"period_length": {"5"},
	}
	invalidCycleRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(invalidCycleForm.Encode()))
	invalidCycleRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidCycleRequest.Header.Set("HX-Request", "true")
	invalidCycleRequest.Header.Set("Cookie", authCookie)

	invalidCycleResponse, err := app.Test(invalidCycleRequest, -1)
	if err != nil {
		t.Fatalf("invalid cycle request failed: %v", err)
	}
	defer invalidCycleResponse.Body.Close()
	if invalidCycleResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected invalid cycle status 400, got %d", invalidCycleResponse.StatusCode)
	}

	invalidPeriodForm := url.Values{
		"cycle_length":  {"29"},
		"period_length": {"15"},
	}
	invalidPeriodRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(invalidPeriodForm.Encode()))
	invalidPeriodRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidPeriodRequest.Header.Set("HX-Request", "true")
	invalidPeriodRequest.Header.Set("Cookie", authCookie)

	invalidPeriodResponse, err := app.Test(invalidPeriodRequest, -1)
	if err != nil {
		t.Fatalf("invalid period request failed: %v", err)
	}
	defer invalidPeriodResponse.Body.Close()
	if invalidPeriodResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected invalid period status 400, got %d", invalidPeriodResponse.StatusCode)
	}

	incompatibleForm := url.Values{
		"cycle_length":  {"21"},
		"period_length": {"14"},
	}
	incompatibleRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(incompatibleForm.Encode()))
	incompatibleRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	incompatibleRequest.Header.Set("HX-Request", "true")
	incompatibleRequest.Header.Set("Cookie", authCookie)

	incompatibleResponse, err := app.Test(incompatibleRequest, -1)
	if err != nil {
		t.Fatalf("incompatible request failed: %v", err)
	}
	defer incompatibleResponse.Body.Close()
	if incompatibleResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected incompatible values status 400, got %d", incompatibleResponse.StatusCode)
	}
}
