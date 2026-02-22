package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalendarPageKeepsSelectedDayFromQueryAndBootstrapsEditor(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-selected-day-query@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/calendar?month=2026-02&day=2026-02-17", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("calendar request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read calendar body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `selectedDate: "2026-02-17"`) {
		t.Fatalf("expected selected day in alpine state from day query")
	}
	if !strings.Contains(rendered, `hx-get="/calendar/day/2026-02-17"`) || !strings.Contains(rendered, `hx-trigger="load"`) {
		t.Fatalf("expected day editor bootstrap request for selected day")
	}
	if !strings.Contains(rendered, `next=%2Fcalendar%3Fmonth%3D2026-02%26day%3D2026-02-17`) {
		t.Fatalf("expected language switch links to preserve selected day in next param")
	}
	if !strings.Contains(rendered, `<script defer src="/static/js/app.js?v=`) {
		t.Fatalf("expected shared app script to keep language links in sync")
	}
}
