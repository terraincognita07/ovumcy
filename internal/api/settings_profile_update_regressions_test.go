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
