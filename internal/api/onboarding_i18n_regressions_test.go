package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestOnboardingDateInputUsesCurrentLanguage(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "onboarding-lang@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	request.Header.Set("Cookie", authCookie+"; lume_lang=en")
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("onboarding request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read onboarding body: %v", err)
	}

	pattern := regexp.MustCompile(`(?s)<input[^>]*id="last-period-start"[^>]*lang="en"`)
	if !pattern.Match(body) {
		t.Fatalf("expected date input #last-period-start to render with lang=en")
	}
	if !regexp.MustCompile(`(?s)<input[^>]*id="last-period-start"[^>]*placeholder="dd\.mm\.yyyy"`).Match(body) {
		t.Fatalf("expected english onboarding date placeholder")
	}
}

func TestOnboardingDateInputUsesRussianPlaceholder(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "onboarding-lang-ru@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	request.Header.Set("Cookie", authCookie+"; lume_lang=ru")
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("onboarding request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read onboarding body: %v", err)
	}

	if !regexp.MustCompile(`(?s)<input[^>]*id="last-period-start"[^>]*placeholder="дд\.мм\.гггг"`).Match(body) {
		t.Fatalf("expected russian onboarding date placeholder")
	}
}
