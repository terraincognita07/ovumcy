package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOnboardingStep2IncludesClientSideCrossValidationHooks(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "onboarding-step2-client-validation@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("onboarding request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read onboarding body: %v", err)
	}
	rendered := string(body)

	if !strings.Contains(rendered, `validateStepTwoBeforeSubmit($event)`) {
		t.Fatalf("expected onboarding step2 form to call client cross-validation before submit")
	}
	if !strings.Contains(rendered, `max="14"`) {
		t.Fatalf("expected onboarding period slider max=14")
	}
	if !strings.Contains(rendered, `periodExceedsCycleMessage:`) {
		t.Fatalf("expected onboarding flow config to provide localized period/cycle validation message")
	}
	if !strings.Contains(rendered, `x-show="(cycleLength - periodLength) < 8"`) {
		t.Fatalf("expected onboarding to include hard-validation state for incompatible cycle values")
	}
	if !strings.Contains(rendered, `'btn--disabled': (cycleLength - periodLength) < 8`) {
		t.Fatalf("expected onboarding next button to include disabled visual state class binding")
	}
}
