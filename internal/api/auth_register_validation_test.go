package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestRegisterRejectsWeakNumericPassword(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	email := "weak-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"12345678"},
		"confirm_password": {"12345678"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "weak password" {
		t.Fatalf("expected weak password error, got %q", errorValue)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("email = ?", email).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 0 {
		t.Fatalf("expected user not to be created, found %d records", usersCount)
	}
}

func TestRegisterSuccessSetsAuthCookieAndShowsRecoveryStep(t *testing.T) {
	app, _ := newOnboardingTestApp(t)
	email := "autologin-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept-Language", "en")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register success request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	authCookie := responseCookieValue(response.Cookies(), authCookieName)
	if authCookie == "" {
		t.Fatalf("expected auth cookie in register response for auto-login")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "Save your recovery code") {
		t.Fatalf("expected recovery code screen after register")
	}
	if !strings.Contains(rendered, "Continue to app") {
		t.Fatalf("expected continue button after register")
	}
}

func TestRegisterRejectsPasswordMismatch(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	email := "mismatch-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass2"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register mismatch request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "password mismatch" {
		t.Fatalf("expected password mismatch error, got %q", errorValue)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("email = ?", email).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 0 {
		t.Fatalf("expected user not to be created, found %d records", usersCount)
	}
}

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
