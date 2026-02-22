package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLanguageSwitchSetsEnglishCookieAndRendersEnglishLogin(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	switchToEnglish := httptest.NewRequest(http.MethodGet, "/lang/en?next=/login", nil)
	switchToEnglish.Header.Set("Accept-Language", "ru")
	englishResponse, err := app.Test(switchToEnglish, -1)
	if err != nil {
		t.Fatalf("switch language request failed: %v", err)
	}
	defer englishResponse.Body.Close()

	if englishResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", englishResponse.StatusCode)
	}
	if location := englishResponse.Header.Get("Location"); location != "/login" {
		t.Fatalf("expected redirect to /login, got %q", location)
	}

	englishCookie := responseCookieValue(englishResponse.Cookies(), "lume_lang")
	if englishCookie != "en" {
		t.Fatalf("expected lume_lang cookie value %q, got %q", "en", englishCookie)
	}

	englishLoginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	englishLoginRequest.Header.Set("Cookie", "lume_lang="+englishCookie)
	englishLoginResponse, err := app.Test(englishLoginRequest, -1)
	if err != nil {
		t.Fatalf("english login request failed: %v", err)
	}
	defer englishLoginResponse.Body.Close()

	englishBody, err := io.ReadAll(englishLoginResponse.Body)
	if err != nil {
		t.Fatalf("read english login body: %v", err)
	}
	renderedEnglish := string(englishBody)
	if !strings.Contains(renderedEnglish, `<html lang="en"`) {
		t.Fatalf("expected login page html lang to be en")
	}
	if !strings.Contains(renderedEnglish, `<script defer src="/static/js/app.js?v=20260221-2"></script>`) {
		t.Fatalf("expected shared app script in base template")
	}
	if !strings.Contains(renderedEnglish, `data-required-message="Please fill out this field."`) {
		t.Fatalf("expected english required validation message in login form")
	}
	if !strings.Contains(renderedEnglish, `data-email-message="Please enter a valid email address."`) {
		t.Fatalf("expected english email validation message in login form")
	}
	if !strings.Contains(renderedEnglish, "Stay signed in for 30 days") {
		t.Fatalf("expected remember-me control on login form in english")
	}
	if !strings.Contains(renderedEnglish, "only until you close the browser") {
		t.Fatalf("expected remember-me helper text in english")
	}
}
