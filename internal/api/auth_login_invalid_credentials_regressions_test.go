package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
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
	if !strings.Contains(rendered, `data-login-has-error="true"`) {
		t.Fatalf("expected login form to mark error state for password draft restore")
	}
	if !strings.Contains(rendered, `data-password-draft-key="ovumcy_login_password_draft"`) {
		t.Fatalf("expected login form to include password draft storage key")
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
	if !strings.Contains(string(afterFlashBody), `data-login-has-error="false"`) {
		t.Fatalf("expected clean login page without error state marker after flash is consumed")
	}
}
