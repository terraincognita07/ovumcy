package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLanguageSwitchSetsRussianCookieAndRendersRussianLogin(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	switchToEnglish := httptest.NewRequest(http.MethodGet, "/lang/en?next=/login", nil)
	switchToEnglish.Header.Set("Accept-Language", "ru")
	englishResponse, err := app.Test(switchToEnglish, -1)
	if err != nil {
		t.Fatalf("switch to english request failed: %v", err)
	}
	defer englishResponse.Body.Close()

	englishCookie := responseCookieValue(englishResponse.Cookies(), "lume_lang")
	if englishCookie != "en" {
		t.Fatalf("expected english cookie value before russian switch, got %q", englishCookie)
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
	renderedRussian := string(russianBody)
	if !strings.Contains(renderedRussian, `<html lang="ru"`) {
		t.Fatalf("expected login page html lang to be ru")
	}
	if !strings.Contains(renderedRussian, `data-required-message="Заполните это поле."`) {
		t.Fatalf("expected russian required validation message in login form")
	}
	if !strings.Contains(renderedRussian, `data-email-message="Введите корректный email адрес."`) {
		t.Fatalf("expected russian email validation message in login form")
	}
	if !strings.Contains(renderedRussian, "Оставаться в системе 30 дней") {
		t.Fatalf("expected remember-me control on login form in russian")
	}
	if !strings.Contains(renderedRussian, "только до закрытия браузера") {
		t.Fatalf("expected remember-me helper text in russian")
	}
}
