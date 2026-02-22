package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestDashboardPeriodToggleSyncsToCalendarDayPanel(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-calendar-sync@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	today := dateAtLocation(time.Now().In(time.UTC), time.UTC).Format("2006-01-02")
	form := url.Values{
		"is_period": {"true"},
		"flow":      {"none"},
		"notes":     {"sync check"},
	}
	saveRequest := httptest.NewRequest(http.MethodPost, "/api/days/"+today, strings.NewReader(form.Encode()))
	saveRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveRequest.Header.Set("HX-Request", "true")
	saveRequest.Header.Set("Accept-Language", "en")
	saveRequest.Header.Set("Cookie", authCookie)

	saveResponse, err := app.Test(saveRequest, -1)
	if err != nil {
		t.Fatalf("dashboard save request failed: %v", err)
	}
	defer saveResponse.Body.Close()

	if saveResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected save status 200, got %d", saveResponse.StatusCode)
	}

	dayPanelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/"+today, nil)
	dayPanelRequest.Header.Set("Accept-Language", "en")
	dayPanelRequest.Header.Set("Cookie", authCookie)

	dayPanelResponse, err := app.Test(dayPanelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer dayPanelResponse.Body.Close()

	if dayPanelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected day panel status 200, got %d", dayPanelResponse.StatusCode)
	}

	body, err := io.ReadAll(dayPanelResponse.Body)
	if err != nil {
		t.Fatalf("read day panel body: %v", err)
	}
	rendered := string(body)
	periodCheckedPattern := regexp.MustCompile(`(?s)name="is_period"[^>]*checked`)
	if !periodCheckedPattern.MatchString(rendered) {
		t.Fatalf("expected calendar day panel period toggle to remain checked after dashboard save")
	}
	if !strings.Contains(rendered, "sync check") {
		t.Fatalf("expected calendar day panel notes to include saved dashboard note")
	}
}

func TestDashboardPeriodToggleSyncsToCalendarDayPanelInLocalTimezone(t *testing.T) {
	app, database, location := newOnboardingTestAppWithLocation(t, time.Local)
	user := createOnboardingTestUser(t, database, "dashboard-calendar-sync-local@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	today := dateAtLocation(time.Now().In(location), location).Format("2006-01-02")
	form := url.Values{
		"is_period": {"true"},
		"flow":      {"none"},
		"notes":     {"sync check local"},
	}
	saveRequest := httptest.NewRequest(http.MethodPost, "/api/days/"+today, strings.NewReader(form.Encode()))
	saveRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	saveRequest.Header.Set("HX-Request", "true")
	saveRequest.Header.Set("Accept-Language", "en")
	saveRequest.Header.Set("Cookie", authCookie)

	saveResponse, err := app.Test(saveRequest, -1)
	if err != nil {
		t.Fatalf("dashboard save request failed: %v", err)
	}
	defer saveResponse.Body.Close()

	if saveResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected save status 200, got %d", saveResponse.StatusCode)
	}

	dayPanelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/"+today, nil)
	dayPanelRequest.Header.Set("Accept-Language", "en")
	dayPanelRequest.Header.Set("Cookie", authCookie)

	dayPanelResponse, err := app.Test(dayPanelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer dayPanelResponse.Body.Close()

	if dayPanelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected day panel status 200, got %d", dayPanelResponse.StatusCode)
	}

	body, err := io.ReadAll(dayPanelResponse.Body)
	if err != nil {
		t.Fatalf("read day panel body: %v", err)
	}
	rendered := string(body)
	periodCheckedPattern := regexp.MustCompile(`(?s)name="is_period"[^>]*checked`)
	if !periodCheckedPattern.MatchString(rendered) {
		t.Fatalf("expected calendar day panel period toggle to remain checked after dashboard save in local timezone")
	}
	if !strings.Contains(rendered, "sync check local") {
		t.Fatalf("expected calendar day panel notes to include saved dashboard note")
	}
}
