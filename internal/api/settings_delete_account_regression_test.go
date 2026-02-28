package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestSettingsDeleteAccountRejectsMissingPassword(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-delete-missing@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodDelete, "/api/settings/delete-account", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("delete-account request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
	if got := readAPIError(t, response.Body); got != "invalid password" {
		t.Fatalf("expected invalid password error, got %q", got)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 1 {
		t.Fatalf("expected user to stay in database, got count=%d", usersCount)
	}
}

func TestSettingsDeleteAccountRejectsInvalidPassword(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-delete-invalid@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodDelete, "/api/settings/delete-account", strings.NewReader(`{"password":"WrongPass1"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("delete-account request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.StatusCode)
	}
	if got := readAPIError(t, response.Body); got != "invalid password" {
		t.Fatalf("expected invalid password error, got %q", got)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 1 {
		t.Fatalf("expected user to stay in database, got count=%d", usersCount)
	}
}

func TestSettingsDeleteAccountDeletesUserAndClearsAuthCookie(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-delete-success@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodDelete, "/api/settings/delete-account", strings.NewReader(`{"password":"StrongPass1"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("delete-account request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	var usersCount int64
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount != 0 {
		t.Fatalf("expected user to be deleted, got count=%d", usersCount)
	}

	authCookieAfterDelete := responseCookie(response.Cookies(), authCookieName)
	if authCookieAfterDelete == nil {
		t.Fatalf("expected auth cookie to be cleared on delete-account success")
	}
	if authCookieAfterDelete.Value != "" {
		t.Fatalf("expected cleared auth cookie value, got %q", authCookieAfterDelete.Value)
	}
}
