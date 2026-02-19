package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

const firstLaunchRegisterSubtitle = "First launch detected. Create the main account to begin."
const weakPasswordErrorText = "Use at least 8 characters with uppercase, lowercase, and a number."

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

func TestChangePasswordRejectsWeakNumericPassword(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "change-password@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"current_password": {"StrongPass1"},
		"new_password":     {"12345678"},
		"confirm_password": {"12345678"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/change-password", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("change password request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "weak password" {
		t.Fatalf("expected weak password error, got %q", errorValue)
	}

	var updatedUser models.User
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("StrongPass1")) != nil {
		t.Fatalf("expected old password hash to stay unchanged")
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("12345678")) == nil {
		t.Fatalf("expected weak password not to be applied")
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

func TestChangePasswordRejectsPasswordMismatch(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "change-password-mismatch@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"current_password": {"StrongPass1"},
		"new_password":     {"EvenStronger2"},
		"confirm_password": {"DifferentPass3"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/change-password", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("change password mismatch request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "password mismatch" {
		t.Fatalf("expected password mismatch error, got %q", errorValue)
	}

	var updatedUser models.User
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("StrongPass1")) != nil {
		t.Fatalf("expected old password hash to stay unchanged")
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

func TestSettingsPageUsesFlashSuccessOnRedirect(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		form         url.Values
		successLabel string
	}{
		{
			name: "cycle settings success",
			path: "/settings/cycle",
			form: url.Values{
				"cycle_length":  {"29"},
				"period_length": {"6"},
			},
			successLabel: "Cycle settings updated successfully.",
		},
		{
			name: "password status success",
			path: "/api/settings/change-password",
			form: url.Values{
				"current_password": {"StrongPass1"},
				"new_password":     {"EvenStronger2"},
				"confirm_password": {"EvenStronger2"},
			},
			successLabel: "Password changed successfully.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app, database := newOnboardingTestApp(t)
			user := createOnboardingTestUser(t, database, "settings-user@example.com", "StrongPass1", true)
			authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

			request := httptest.NewRequest(http.MethodPost, testCase.path, strings.NewReader(testCase.form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Cookie", authCookie)

			response, err := app.Test(request, -1)
			if err != nil {
				t.Fatalf("settings action request failed: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusSeeOther {
				t.Fatalf("expected status 303, got %d", response.StatusCode)
			}
			if location := response.Header.Get("Location"); location != "/settings" {
				t.Fatalf("expected redirect %q, got %q", "/settings", location)
			}

			flashValue := responseCookieValue(response.Cookies(), flashCookieName)
			if flashValue == "" {
				t.Fatalf("expected flash cookie for settings success message")
			}

			followRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
			followRequest.Header.Set("Accept-Language", "en")
			followRequest.Header.Set("Cookie", authCookie+"; "+flashCookieName+"="+flashValue)

			followResponse, err := app.Test(followRequest, -1)
			if err != nil {
				t.Fatalf("follow-up settings request failed: %v", err)
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
			if !strings.Contains(rendered, testCase.successLabel) {
				t.Fatalf("expected success label %q in settings page", testCase.successLabel)
			}
			if strings.Contains(rendered, weakPasswordErrorText) {
				t.Fatalf("did not expect weak password error on success page")
			}

			afterFlashRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
			afterFlashRequest.Header.Set("Accept-Language", "en")
			afterFlashRequest.Header.Set("Cookie", authCookie)

			afterFlashResponse, err := app.Test(afterFlashRequest, -1)
			if err != nil {
				t.Fatalf("settings request after flash consumption failed: %v", err)
			}
			defer afterFlashResponse.Body.Close()

			afterFlashBody, err := io.ReadAll(afterFlashResponse.Body)
			if err != nil {
				t.Fatalf("read settings body after flash consumption: %v", err)
			}
			if strings.Contains(string(afterFlashBody), testCase.successLabel) {
				t.Fatalf("did not expect success label after flash is consumed")
			}
		})
	}
}

func TestSettingsPageRendersPersistedCycleValues(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-values@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":  29,
		"period_length": 6,
	}).Error; err != nil {
		t.Fatalf("update cycle values: %v", err)
	}
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `x-data="{ cycleLength: 29, periodLength: 6 }"`) {
		t.Fatalf("expected settings x-data to include persisted values")
	}

	cycleInputPattern := regexp.MustCompile(`(?s)name="cycle_length".*?value="29"`)
	if !cycleInputPattern.MatchString(rendered) {
		t.Fatalf("expected cycle slider value attribute to be rendered from DB")
	}
	periodInputPattern := regexp.MustCompile(`(?s)name="period_length".*?value="6"`)
	if !periodInputPattern.MatchString(rendered) {
		t.Fatalf("expected period slider value attribute to be rendered from DB")
	}
}

func TestCalendarDayPanelShowsDeleteForLegacyTimestampedLog(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-legacy@example.com", "StrongPass1", true)

	legacyTimestamp := "2026-01-15T15:30:00Z"
	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		legacyTimestamp,
		true,
		models.FlowMedium,
		"[]",
		"legacy entry",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert legacy log: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-01-15", nil)
	panelRequest.Header.Set("Cookie", authCookie)
	panelResponse, err := app.Test(panelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer panelResponse.Body.Close()

	if panelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", panelResponse.StatusCode)
	}

	panelBody, err := io.ReadAll(panelResponse.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	if !strings.Contains(string(panelBody), "/api/log/delete?date=2026-01-15&source=calendar") {
		t.Fatalf("expected delete button for legacy timestamped log")
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/log/delete?date=2026-01-15&source=calendar", nil)
	deleteRequest.Header.Set("Cookie", authCookie)
	deleteResponse, err := app.Test(deleteRequest, -1)
	if err != nil {
		t.Fatalf("delete day request failed: %v", err)
	}
	defer deleteResponse.Body.Close()

	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", deleteResponse.StatusCode)
	}

	var remaining int64
	dayStart := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)
	if err := database.Model(&models.DailyLog{}).
		Where("user_id = ? AND date >= ? AND date < ?", user.ID, dayStart, dayEnd).
		Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining legacy logs: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected legacy day logs to be deleted, got %d", remaining)
	}
}

func TestCalendarDayExistsAndDeleteButtonUseAnyLegacyRowInDayRange(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-day-range@example.com", "StrongPass1", true)

	now := time.Now().UTC()
	dayWithData := "2026-02-17T08:30:00Z"
	dayWithoutData := "2026-02-17T20:15:00Z"

	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		dayWithData,
		true,
		models.FlowMedium,
		"[]",
		"has data",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert legacy data row: %v", err)
	}

	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		dayWithoutData,
		false,
		models.FlowNone,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert legacy empty row: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	existsRequest := httptest.NewRequest(http.MethodGet, "/api/days/2026-02-17/exists", nil)
	existsRequest.Header.Set("Accept", "application/json")
	existsRequest.Header.Set("Cookie", authCookie)
	existsResponse, err := app.Test(existsRequest, -1)
	if err != nil {
		t.Fatalf("exists request failed: %v", err)
	}
	defer existsResponse.Body.Close()

	if existsResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected exists status 200, got %d", existsResponse.StatusCode)
	}

	existsPayload := map[string]bool{}
	existsBody, err := io.ReadAll(existsResponse.Body)
	if err != nil {
		t.Fatalf("read exists body: %v", err)
	}
	if err := json.Unmarshal(existsBody, &existsPayload); err != nil {
		t.Fatalf("decode exists body: %v", err)
	}
	if !existsPayload["exists"] {
		t.Fatalf("expected exists=true when any row in day range has data")
	}

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	panelRequest.Header.Set("Cookie", authCookie)
	panelResponse, err := app.Test(panelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer panelResponse.Body.Close()

	if panelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected panel status 200, got %d", panelResponse.StatusCode)
	}

	panelBody, err := io.ReadAll(panelResponse.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	if !strings.Contains(string(panelBody), "/api/log/delete?date=2026-02-17&source=calendar") {
		t.Fatalf("expected delete button when any legacy row in day range has data")
	}
}

func readAPIError(t *testing.T, body io.Reader) string {
	t.Helper()

	payload := map[string]string{}
	bytes, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	if err := json.Unmarshal(bytes, &payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return payload["error"]
}
