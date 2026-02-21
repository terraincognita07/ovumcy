package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestOnboardingStep1RejectsFutureAndTooOldDates(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "step1-validation@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	futureDate := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, 1).Format("2006-01-02")
	futureForm := url.Values{
		"last_period_start": {futureDate},
		"period_status":     {onboardingPeriodStatusOngoing},
	}
	futureRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(futureForm.Encode()))
	futureRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	futureRequest.Header.Set("HX-Request", "true")
	futureRequest.Header.Set("Cookie", authCookie)

	futureResponse, err := app.Test(futureRequest, -1)
	if err != nil {
		t.Fatalf("future date request failed: %v", err)
	}
	defer futureResponse.Body.Close()
	if futureResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected future date status 400, got %d", futureResponse.StatusCode)
	}

	oldDate := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, -61).Format("2006-01-02")
	oldForm := url.Values{
		"last_period_start": {oldDate},
		"period_status":     {onboardingPeriodStatusOngoing},
	}
	oldRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(oldForm.Encode()))
	oldRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	oldRequest.Header.Set("HX-Request", "true")
	oldRequest.Header.Set("Cookie", authCookie)

	oldResponse, err := app.Test(oldRequest, -1)
	if err != nil {
		t.Fatalf("old date request failed: %v", err)
	}
	defer oldResponse.Body.Close()
	if oldResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected old date status 400, got %d", oldResponse.StatusCode)
	}
}

func TestOnboardingStep1RequiresPeriodEndForFinishedStatus(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "step1-finished-validation@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	stepDate := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, -4).Format("2006-01-02")
	form := url.Values{
		"last_period_start": {stepDate},
		"period_status":     {onboardingPeriodStatusFinished},
		"period_end":        {""},
	}
	request := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("step1 finished-without-end request failed: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}
