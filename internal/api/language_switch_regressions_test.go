package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLanguageSwitchSetsCookieAndRendersMatchingHTMLLang(t *testing.T) {
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
	if !strings.Contains(renderedEnglish, `<script defer src="/static/js/app.js"></script>`) {
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

	switchToRussian := httptest.NewRequest(http.MethodGet, "/lang/ru?next=/login", nil)
	switchToRussian.Header.Set("Cookie", "lume_lang="+englishCookie)
	russianResponse, err := app.Test(switchToRussian, -1)
	if err != nil {
		t.Fatalf("switch back language request failed: %v", err)
	}
	defer russianResponse.Body.Close()

	russianCookie := responseCookieValue(russianResponse.Cookies(), "lume_lang")
	if russianCookie != "ru" {
		t.Fatalf("expected lume_lang cookie value %q, got %q", "ru", russianCookie)
	}

	russianLoginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	russianLoginRequest.Header.Set("Cookie", "lume_lang="+russianCookie)
	russianLoginResponse, err := app.Test(russianLoginRequest, -1)
	if err != nil {
		t.Fatalf("russian login request failed: %v", err)
	}
	defer russianLoginResponse.Body.Close()

	russianBody, err := io.ReadAll(russianLoginResponse.Body)
	if err != nil {
		t.Fatalf("read russian login body: %v", err)
	}
	if !strings.Contains(string(russianBody), `<html lang="ru"`) {
		t.Fatalf("expected login page html lang to be ru")
	}
	if !strings.Contains(string(russianBody), `data-required-message="Заполните это поле."`) {
		t.Fatalf("expected russian required validation message in login form")
	}
	if !strings.Contains(string(russianBody), `data-email-message="Введите корректный email адрес."`) {
		t.Fatalf("expected russian email validation message in login form")
	}
	if !strings.Contains(string(russianBody), "Оставаться в системе 30 дней") {
		t.Fatalf("expected remember-me control on login form in russian")
	}
	if !strings.Contains(string(russianBody), "только до закрытия браузера") {
		t.Fatalf("expected remember-me helper text in russian")
	}
}
