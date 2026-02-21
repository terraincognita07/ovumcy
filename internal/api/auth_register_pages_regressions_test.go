package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRegisterPageKeepsEmailAfterPasswordValidationError(t *testing.T) {
	app, _ := newOnboardingTestApp(t)
	email := "persist-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"12345678"},
		"confirm_password": {"12345678"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/register" {
		t.Fatalf("expected redirect to /register, got %q", location)
	}

	flashValue := responseCookieValue(response.Cookies(), flashCookieName)
	if flashValue == "" {
		t.Fatalf("expected flash cookie after register validation error")
	}

	registerRequest := httptest.NewRequest(http.MethodGet, "/register", nil)
	registerRequest.Header.Set("Cookie", flashCookieName+"="+flashValue)
	registerResponse, err := app.Test(registerRequest, -1)
	if err != nil {
		t.Fatalf("register page request failed: %v", err)
	}
	defer registerResponse.Body.Close()

	if registerResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", registerResponse.StatusCode)
	}

	body, err := io.ReadAll(registerResponse.Body)
	if err != nil {
		t.Fatalf("read register body: %v", err)
	}
	if !strings.Contains(string(body), `id="register-email" type="email" name="email" value="`+email+`"`) {
		t.Fatalf("expected register page to keep submitted email after validation error")
	}
}

func TestRecoveryCodePageUsesStandardCheckboxLayout(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	form := url.Values{
		"email":            {"recovery-checkbox@example.com"},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept-Language", "en")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read recovery code page body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `class="remember-checkbox"`) {
		t.Fatalf("expected recovery page checkbox to use standard remember-checkbox style")
	}
	if strings.Contains(rendered, `class="choice-input"`) {
		t.Fatalf("did not expect recovery page checkbox to use chip toggle markup")
	}
}

func TestRegisterPageShowsFirstLaunchSubtitleWhenNoUsers(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/register", nil)
	request.Header.Set("Accept-Language", "en")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register page request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), firstLaunchRegisterSubtitle) {
		t.Fatalf("expected first launch subtitle to be present")
	}
}

func TestRegisterPageHidesFirstLaunchSubtitleWhenUserExists(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	createOnboardingTestUser(t, database, "owner@example.com", "StrongPass1", true)

	request := httptest.NewRequest(http.MethodGet, "/register", nil)
	request.Header.Set("Accept-Language", "en")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register page request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if strings.Contains(string(body), firstLaunchRegisterSubtitle) {
		t.Fatalf("expected first launch subtitle to be hidden when users already exist")
	}
}
