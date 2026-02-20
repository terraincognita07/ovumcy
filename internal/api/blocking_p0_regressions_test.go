package api

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestLoginPageUsesLoginHeadingOnFirstLaunch(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/login", nil)
	request.Header.Set("Accept-Language", "ru")
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read login body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "Войти в аккаунт") {
		t.Fatalf("expected login heading in russian")
	}
	if strings.Contains(rendered, "Создайте аккаунт") {
		t.Fatalf("did not expect registration heading on login page")
	}
}

func TestOnboardingStep1ShowsLocalizedErrorWhenDateMissing(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "missing-date@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"last_period_start": {""},
	}
	request := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Accept-Language", "ru")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("step1 request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected step1 status 400, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read step1 body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "status-error") {
		t.Fatalf("expected status-error markup in response")
	}
	if !strings.Contains(rendered, "Пожалуйста, выберите дату") {
		t.Fatalf("expected localized missing-date error in russian, got response: %q", rendered)
	}
}

func TestOnboardingStep1NextButtonIsNotDisabledWithoutDate(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "onboarding-ui@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	request.Header.Set("Cookie", authCookie)
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("onboarding request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read onboarding body: %v", err)
	}
	if strings.Contains(string(body), `:disabled="!selectedDate"`) {
		t.Fatalf("did not expect client-side disabled next button binding on step1")
	}
}

func TestStatsChartExcludesCycleEndingToday(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "stats-trend@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	today := dateAtLocation(time.Now().In(time.UTC), time.UTC)
	previousStart := today.AddDate(0, 0, -10)

	logs := []models.DailyLog{
		{
			UserID:     user.ID,
			Date:       previousStart,
			IsPeriod:   true,
			Flow:       models.FlowMedium,
			SymptomIDs: []uint{},
		},
		{
			UserID:     user.ID,
			Date:       today,
			IsPeriod:   true,
			Flow:       models.FlowMedium,
			SymptomIDs: []uint{},
		},
	}
	if err := database.Create(&logs).Error; err != nil {
		t.Fatalf("create period logs: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/stats", nil)
	request.Header.Set("Cookie", authCookie)
	request.Header.Set("Accept-Language", "en")
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("stats request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected stats status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read stats body: %v", err)
	}

	chartPayload, err := extractStatsChartPayload(string(body))
	if err != nil {
		t.Fatalf("extract chart payload: %v", err)
	}
	if len(chartPayload.Values) != 0 {
		t.Fatalf("expected no completed cycle points when latest cycle starts today, got %v", chartPayload.Values)
	}
}

var statsChartDataPattern = regexp.MustCompile(`data-chart='([^']+)'`)

type statsChartPayload struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
}

func extractStatsChartPayload(rendered string) (statsChartPayload, error) {
	matches := statsChartDataPattern.FindStringSubmatch(rendered)
	if len(matches) != 2 {
		return statsChartPayload{}, fmt.Errorf("data-chart attribute not found")
	}

	rawJSON := html.UnescapeString(matches[1])
	payload := statsChartPayload{}
	if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
		return statsChartPayload{}, err
	}
	return payload, nil
}
