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

func TestProfileUpdatePersistsDisplayNameAndShowsItInNavigation(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "profile-owner@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"display_name": {"Maya"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/profile", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("profile update request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/settings" {
		t.Fatalf("expected redirect to /settings, got %q", location)
	}

	flashValue := responseCookieValue(response.Cookies(), flashCookieName)
	if flashValue == "" {
		t.Fatalf("expected flash cookie for profile update")
	}

	updatedUser := models.User{}
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updatedUser.DisplayName != "Maya" {
		t.Fatalf("expected display name to be persisted, got %q", updatedUser.DisplayName)
	}

	settingsRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
	settingsRequest.Header.Set("Accept-Language", "en")
	settingsRequest.Header.Set("Cookie", authCookie+"; "+flashCookieName+"="+flashValue)
	settingsResponse, err := app.Test(settingsRequest, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer settingsResponse.Body.Close()

	settingsBody, err := io.ReadAll(settingsResponse.Body)
	if err != nil {
		t.Fatalf("read settings body: %v", err)
	}
	if !strings.Contains(string(settingsBody), "Profile updated successfully.") {
		t.Fatalf("expected profile update success flash message")
	}
	if !strings.Contains(string(settingsBody), `value="Maya"`) {
		t.Fatalf("expected profile display name input to show persisted value")
	}

	dashboardRequest := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashboardRequest.Header.Set("Accept-Language", "en")
	dashboardRequest.Header.Set("Cookie", authCookie)
	dashboardResponse, err := app.Test(dashboardRequest, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer dashboardResponse.Body.Close()

	dashboardBody, err := io.ReadAll(dashboardResponse.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	if !strings.Contains(string(dashboardBody), `aria-label="Current user"`) {
		t.Fatalf("expected current user identity chip after profile update")
	}
	if !strings.Contains(string(dashboardBody), ">Maya</span>") {
		t.Fatalf("expected display name in navigation after profile update")
	}
}

func TestProfileUpdateShowsMessageWhenDisplayNameCleared(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "profile-clear@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Update("display_name", "Maya").Error; err != nil {
		t.Fatalf("seed display name: %v", err)
	}
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"display_name": {""},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/profile", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("profile clear request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}

	flashValue := responseCookieValue(response.Cookies(), flashCookieName)
	if flashValue == "" {
		t.Fatalf("expected flash cookie for profile clear")
	}

	updatedUser := models.User{}
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if updatedUser.DisplayName != "" {
		t.Fatalf("expected display name to be cleared, got %q", updatedUser.DisplayName)
	}

	settingsRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
	settingsRequest.Header.Set("Accept-Language", "en")
	settingsRequest.Header.Set("Cookie", authCookie+"; "+flashCookieName+"="+flashValue)
	settingsResponse, err := app.Test(settingsRequest, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer settingsResponse.Body.Close()

	settingsBody, err := io.ReadAll(settingsResponse.Body)
	if err != nil {
		t.Fatalf("read settings body: %v", err)
	}
	if !strings.Contains(string(settingsBody), "Profile name removed.") {
		t.Fatalf("expected profile cleared success message")
	}

	dashboardRequest := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashboardRequest.Header.Set("Accept-Language", "en")
	dashboardRequest.Header.Set("Cookie", authCookie)
	dashboardResponse, err := app.Test(dashboardRequest, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer dashboardResponse.Body.Close()

	dashboardBody, err := io.ReadAll(dashboardResponse.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	if strings.Contains(string(dashboardBody), ">Maya</span>") {
		t.Fatalf("did not expect stale display name in navigation after clear")
	}
	if !strings.Contains(string(dashboardBody), "profile-clear@example.com") {
		t.Fatalf("expected navigation fallback to email after display name clear")
	}
}

func TestSettingsChangePasswordFormUsesHTMXInlineFeedback(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-password-htmx@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read settings body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `hx-post="/api/settings/change-password"`) {
		t.Fatalf("expected change password form to submit via htmx")
	}
	if !strings.Contains(rendered, `hx-target="#settings-change-password-status"`) {
		t.Fatalf("expected inline feedback target for change password form")
	}
	if !strings.Contains(rendered, `id="settings-change-password-status"`) {
		t.Fatalf("expected dedicated change password feedback container")
	}
}

func TestAuthPagesIncludeSwitchTransitionHooks(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	loginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRequest.Header.Set("Accept-Language", "en")
	loginResponse, err := app.Test(loginRequest, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer loginResponse.Body.Close()

	loginBody, err := io.ReadAll(loginResponse.Body)
	if err != nil {
		t.Fatalf("read login body: %v", err)
	}
	loginRendered := string(loginBody)
	if !strings.Contains(loginRendered, `data-auth-panel`) {
		t.Fatalf("expected auth panel transition hook on login page")
	}
	if !strings.Contains(loginRendered, `data-auth-switch`) {
		t.Fatalf("expected auth switch transition hook on login page")
	}
	if !strings.Contains(loginRendered, `<script defer src="/static/js/app.js"></script>`) {
		t.Fatalf("expected shared app script for auth panel transitions")
	}

	registerRequest := httptest.NewRequest(http.MethodGet, "/register", nil)
	registerRequest.Header.Set("Accept-Language", "en")
	registerResponse, err := app.Test(registerRequest, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer registerResponse.Body.Close()

	registerBody, err := io.ReadAll(registerResponse.Body)
	if err != nil {
		t.Fatalf("read register body: %v", err)
	}
	registerRendered := string(registerBody)
	if !strings.Contains(registerRendered, `data-auth-panel`) {
		t.Fatalf("expected auth panel transition hook on register page")
	}
	if !strings.Contains(registerRendered, `data-auth-switch`) {
		t.Fatalf("expected auth switch transition hook on register page")
	}
}

func TestRecoveryCodePageIncludesDownloadFeedbackMessage(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	form := url.Values{
		"email":            {"recovery-feedback@example.com"},
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
		t.Fatalf("read recovery page body: %v", err)
	}
	if !strings.Contains(string(body), "Recovery code downloaded.") {
		t.Fatalf("expected recovery code download feedback message")
	}
}
