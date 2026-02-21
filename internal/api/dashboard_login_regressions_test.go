package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginInvalidCredentialsRedirectPreservesEmail(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "login-email@example.com", "StrongPass1", true)

	form := url.Values{
		"email":    {user.Email},
		"password": {"WrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}

	location := strings.TrimSpace(response.Header.Get("Location"))
	if location == "" {
		t.Fatalf("expected redirect location")
	}
	redirectURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if redirectURL.Path != "/login" {
		t.Fatalf("expected redirect path /login, got %q", redirectURL.Path)
	}
	if query := strings.TrimSpace(redirectURL.RawQuery); query != "" {
		t.Fatalf("expected clean redirect without query params, got %q", query)
	}

	flashValue := responseCookieValue(response.Cookies(), flashCookieName)
	if flashValue == "" {
		t.Fatalf("expected flash cookie in login redirect response")
	}

	followRequest := httptest.NewRequest(http.MethodGet, location, nil)
	followRequest.Header.Set("Accept-Language", "en")
	followRequest.Header.Set("Cookie", flashCookieName+"="+flashValue)
	followResponse, err := app.Test(followRequest, -1)
	if err != nil {
		t.Fatalf("follow-up login request failed: %v", err)
	}
	defer followResponse.Body.Close()

	if followResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected follow-up status 200, got %d", followResponse.StatusCode)
	}

	body, err := io.ReadAll(followResponse.Body)
	if err != nil {
		t.Fatalf("read follow-up body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `id="login-email"`) {
		t.Fatalf("expected login email input in page")
	}
	if !strings.Contains(rendered, `value="login-email@example.com"`) {
		t.Fatalf("expected login email input to keep previous value")
	}
	if !strings.Contains(rendered, "Invalid email or password.") {
		t.Fatalf("expected localized login error message from flash")
	}

	afterFlashRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	afterFlashRequest.Header.Set("Accept-Language", "en")
	afterFlashResponse, err := app.Test(afterFlashRequest, -1)
	if err != nil {
		t.Fatalf("login request after flash consumption failed: %v", err)
	}
	defer afterFlashResponse.Body.Close()

	afterFlashBody, err := io.ReadAll(afterFlashResponse.Body)
	if err != nil {
		t.Fatalf("read body after flash consumption: %v", err)
	}
	if strings.Contains(string(afterFlashBody), `value="login-email@example.com"`) {
		t.Fatalf("did not expect login email to persist after flash is consumed")
	}
}

func TestLoginRememberMeControlsCookiePersistence(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "remember-session@example.com", "StrongPass1", true)

	sessionForm := url.Values{
		"email":    {user.Email},
		"password": {"StrongPass1"},
	}
	sessionRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(sessionForm.Encode()))
	sessionRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	sessionResponse, err := app.Test(sessionRequest, -1)
	if err != nil {
		t.Fatalf("session login request failed: %v", err)
	}
	defer sessionResponse.Body.Close()

	if sessionResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", sessionResponse.StatusCode)
	}

	sessionCookie := responseCookie(sessionResponse.Cookies(), authCookieName)
	if sessionCookie == nil {
		t.Fatalf("expected auth cookie for default session login")
	}
	if !sessionCookie.Expires.IsZero() {
		t.Fatalf("expected session cookie without Expires when remember_me is disabled")
	}

	rememberForm := url.Values{
		"email":       {user.Email},
		"password":    {"StrongPass1"},
		"remember_me": {"1"},
	}
	rememberRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(rememberForm.Encode()))
	rememberRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rememberResponse, err := app.Test(rememberRequest, -1)
	if err != nil {
		t.Fatalf("remember-me login request failed: %v", err)
	}
	defer rememberResponse.Body.Close()

	if rememberResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", rememberResponse.StatusCode)
	}

	rememberCookie := responseCookie(rememberResponse.Cookies(), authCookieName)
	if rememberCookie == nil {
		t.Fatalf("expected auth cookie for remember-me login")
	}
	if rememberCookie.Expires.IsZero() {
		t.Fatalf("expected persistent auth cookie when remember_me is enabled")
	}
	if rememberCookie.Expires.Before(time.Now().Add(20 * 24 * time.Hour)) {
		t.Fatalf("expected remember-me cookie to expire in ~30 days, got %s", rememberCookie.Expires)
	}
}

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

func TestDashboardLogoutFormsRequireConfirmation(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "logout-confirm@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	rendered := string(body)
	if strings.Count(rendered, `action="/api/auth/logout"`) < 2 {
		t.Fatalf("expected desktop and mobile logout forms")
	}
	if strings.Count(rendered, `data-confirm="Log out of your account now?"`) < 2 {
		t.Fatalf("expected confirmation attribute for both logout forms")
	}
}

func TestDashboardNavigationShowsCurrentUserIdentity(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "identity-owner@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `aria-label="Current user"`) {
		t.Fatalf("expected current user label in navigation")
	}
	if !strings.Contains(rendered, "identity-owner@example.com") {
		t.Fatalf("expected email identity in navigation when display name is empty")
	}
}
