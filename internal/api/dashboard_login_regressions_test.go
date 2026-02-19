package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
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
	if got := redirectURL.Query().Get("error"); got != "invalid credentials" {
		t.Fatalf("expected error query %q, got %q", "invalid credentials", got)
	}
	if got := redirectURL.Query().Get("email"); got != user.Email {
		t.Fatalf("expected email query %q, got %q", user.Email, got)
	}

	followRequest := httptest.NewRequest(http.MethodGet, location, nil)
	followRequest.Header.Set("Accept-Language", "en")
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
	if !strings.Contains(renderedEnglish, `applyHTMLLanguage(readCookie("lume_lang")`) {
		t.Fatalf("expected language sync script in base template")
	}
	if !strings.Contains(renderedEnglish, `data-required-message="Please fill out this field."`) {
		t.Fatalf("expected english required validation message in login form")
	}
	if !strings.Contains(renderedEnglish, `data-email-message="Please enter a valid email address."`) {
		t.Fatalf("expected english email validation message in login form")
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

func responseCookieValue(cookies []*http.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}
