package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
	"golang.org/x/crypto/bcrypt"
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

func TestCalendarDayPanelDeleteEntryUsesConfirmForm(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-confirm@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	logEntry := models.DailyLog{
		UserID:   user.ID,
		Date:     time.Date(2026, time.February, 17, 0, 0, 0, 0, time.UTC),
		IsPeriod: true,
		Flow:     models.FlowMedium,
		Notes:    "entry",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `hx-delete="/api/log/delete?date=2026-02-17&source=calendar"`) {
		t.Fatalf("expected delete endpoint in day panel")
	}
	if !strings.Contains(rendered, `data-confirm="Are you sure you want to delete this entry?"`) {
		t.Fatalf("expected confirm prompt on calendar delete entry action")
	}
	if !strings.Contains(rendered, `data-confirm-accept="Yes, delete"`) {
		t.Fatalf("expected confirm accept label on calendar delete entry action")
	}
}

func TestDashboardSymptomsNotesPanelUsesSavedSymptomsAndNotesState(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-journal@example.com", "StrongPass1", true)

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "Custom cramps", Icon: "A", Color: "#FF7755"},
		{UserID: user.ID, Name: "Custom headache", Icon: "B", Color: "#55AAFF"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	today := dateAtLocation(time.Now().In(time.UTC), time.UTC)
	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       today,
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptoms[0].ID, symptoms[1].ID},
		Notes:      "Remember to hydrate",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

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
	if !strings.Contains(rendered, "Symptoms and notes") {
		t.Fatalf("expected symptoms and notes panel header")
	}
	if !strings.Contains(rendered, `x-for="(symptom, index) in activeSymptoms"`) {
		t.Fatalf("expected dynamic symptoms list renderer in panel")
	}
	if !strings.Contains(rendered, `data-symptom-label="Custom cramps"`) {
		t.Fatalf("expected symptom checkbox metadata for dashboard preview")
	}
	if !strings.Contains(rendered, `data-symptom-label="Custom headache"`) {
		t.Fatalf("expected second symptom checkbox metadata for dashboard preview")
	}
	if !strings.Contains(rendered, `x-model="notesPreview"`) {
		t.Fatalf("expected notes field binding for dashboard preview")
	}
	if !strings.Contains(rendered, "Remember to hydrate") {
		t.Fatalf("expected saved note to be rendered on dashboard")
	}
}

func TestCalendarDayPanelFlowControlsDependOnPeriodToggle(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-flow-toggle@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	panelRequest.Header.Set("Accept-Language", "en")
	panelRequest.Header.Set("Cookie", authCookie)

	panelResponse, err := app.Test(panelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer panelResponse.Body.Close()

	if panelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", panelResponse.StatusCode)
	}

	body, err := io.ReadAll(panelResponse.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	rendered := string(body)

	if !strings.Contains(rendered, `x-data="{ isPeriod: false }"`) {
		t.Fatalf("expected calendar panel form to initialize period state")
	}
	if !strings.Contains(rendered, `x-model="isPeriod"`) {
		t.Fatalf("expected period toggle to drive alpine state")
	}
	if !strings.Contains(rendered, `x-cloak x-show="isPeriod" :disabled="!isPeriod"`) {
		t.Fatalf("expected flow fieldset to be shown/enabled only when period is selected")
	}
	if strings.Count(rendered, `:disabled="!isPeriod"`) != 1 {
		t.Fatalf("expected only flow controls to depend on period toggle")
	}
	if !strings.Contains(rendered, `name="symptom_ids"`) {
		t.Fatalf("expected symptoms controls to stay available regardless of period toggle")
	}
}

func TestCalendarDayPanelUsesLanguageSpecificSymptomLabelClass(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-symptom-label-class@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Breast tenderness",
		Icon:   "💗",
		Color:  "#D98395",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	enRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	enRequest.Header.Set("Accept-Language", "en")
	enRequest.Header.Set("Cookie", authCookie)
	enResponse, err := app.Test(enRequest, -1)
	if err != nil {
		t.Fatalf("english panel request failed: %v", err)
	}
	defer enResponse.Body.Close()

	enBody, err := io.ReadAll(enResponse.Body)
	if err != nil {
		t.Fatalf("read english panel body: %v", err)
	}
	if !strings.Contains(string(enBody), `symptom-label symptom-label-nowrap`) {
		t.Fatalf("expected nowrap symptom class for english labels")
	}

	ruRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	ruRequest.Header.Set("Accept-Language", "ru")
	ruRequest.Header.Set("Cookie", authCookie)
	ruResponse, err := app.Test(ruRequest, -1)
	if err != nil {
		t.Fatalf("russian panel request failed: %v", err)
	}
	defer ruResponse.Body.Close()

	ruBody, err := io.ReadAll(ruResponse.Body)
	if err != nil {
		t.Fatalf("read russian panel body: %v", err)
	}
	if strings.Contains(string(ruBody), `symptom-label symptom-label-nowrap`) {
		t.Fatalf("did not expect nowrap class for russian labels")
	}
	if !strings.Contains(string(ruBody), `class="symptom-label">`) {
		t.Fatalf("expected default symptom label class for russian locale")
	}
}

func TestCalendarPageKeepsSelectedDayFromQueryAndBootstrapsEditor(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-selected-day-query@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/calendar?month=2026-02&day=2026-02-17", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("calendar request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read calendar body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `selectedDate: "2026-02-17"`) {
		t.Fatalf("expected selected day in alpine state from day query")
	}
	if !strings.Contains(rendered, `hx-get="/calendar/day/2026-02-17"`) || !strings.Contains(rendered, `hx-trigger="load"`) {
		t.Fatalf("expected day editor bootstrap request for selected day")
	}
	if !strings.Contains(rendered, `next=%2Fcalendar%3Fmonth%3D2026-02%26day%3D2026-02-17`) {
		t.Fatalf("expected language switch links to preserve selected day in next param")
	}
	if !strings.Contains(rendered, `<script defer src="/static/js/app.js"></script>`) {
		t.Fatalf("expected shared app script to keep language links in sync")
	}
}

func TestBuildCalendarDaysUsesLatestLogPerDateDeterministically(t *testing.T) {
	handler := &Handler{location: time.UTC}
	monthStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC)

	logs := []models.DailyLog{
		{
			ID:       20,
			Date:     time.Date(2026, time.February, 17, 20, 0, 0, 0, time.UTC),
			IsPeriod: false,
			Flow:     models.FlowNone,
		},
		{
			ID:       10,
			Date:     time.Date(2026, time.February, 17, 8, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		},
		{
			ID:       30,
			Date:     time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		},
		{
			ID:       31,
			Date:     time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC),
			IsPeriod: false,
			Flow:     models.FlowNone,
		},
	}

	days := handler.buildCalendarDays(monthStart, logs, services.CycleStats{}, now)

	day17 := findCalendarDayByDateString(t, days, "2026-02-17")
	if day17.IsPeriod {
		t.Fatalf("expected 2026-02-17 period=false from latest log, got true")
	}

	day18 := findCalendarDayByDateString(t, days, "2026-02-18")
	if day18.IsPeriod {
		t.Fatalf("expected 2026-02-18 period=false from highest id tie-breaker, got true")
	}
}

func TestFetchLogByDateFindsZuluStoredRowForLocalCalendarDay(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}

	apiDir := filepath.Dir(testFile)
	internalDir := filepath.Dir(apiDir)
	templatesDir := filepath.Join(internalDir, "templates")
	localesDir := filepath.Join(internalDir, "i18n", "locales")
	databasePath := filepath.Join(t.TempDir(), "lume-zulu-fetch.db")

	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := models.User{
		Email:               "zulu-fetch@example.com",
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		"2026-02-17T00:00:00Z",
		true,
		models.FlowLight,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert zulu row: %v", err)
	}

	i18nManager, err := i18n.NewManager("en", localesDir)
	if err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	moscow := time.FixedZone("UTC+3", 3*60*60)
	handler, err := NewHandler(database, "test-secret-key", templatesDir, moscow, i18nManager, false)
	if err != nil {
		t.Fatalf("init handler: %v", err)
	}

	day, err := parseDayParam("2026-02-17", moscow)
	if err != nil {
		t.Fatalf("parse day: %v", err)
	}

	entry, err := handler.fetchLogByDate(user.ID, day)
	if err != nil {
		t.Fatalf("fetchLogByDate: %v", err)
	}

	if !entry.IsPeriod {
		t.Fatalf("expected is_period=true for local day 2026-02-17")
	}
	if entry.Flow != models.FlowLight {
		t.Fatalf("expected flow %q, got %q", models.FlowLight, entry.Flow)
	}
}

func responseCookieValue(cookies []*http.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func responseCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func findCalendarDayByDateString(t *testing.T, days []CalendarDay, date string) CalendarDay {
	t.Helper()
	for _, day := range days {
		if day.DateString == date {
			return day
		}
	}
	t.Fatalf("calendar day %s not found", date)
	return CalendarDay{}
}
