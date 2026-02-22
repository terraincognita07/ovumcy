package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestOnboardingPageUsesFailOpenFlowFallback(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "onboarding-fallback@example.com", "StrongPass1", false)
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
	rendered := string(body)

	if !strings.Contains(rendered, `x-data='typeof onboardingFlow === "function" ? onboardingFlow(`) {
		t.Fatalf("expected fail-open onboarding flow fallback in x-data")
	}
	if !strings.Contains(rendered, `clearStepStatuses: function ()`) {
		t.Fatalf("expected onboarding flow to include status cleanup helper")
	}
	if !strings.Contains(rendered, `this.clearStepStatuses()`) {
		t.Fatalf("expected onboarding step navigation to clear stale status messages")
	}

	rootSectionPattern := regexp.MustCompile(`(?s)<section\s+class="mx-auto max-w-4xl"[^>]*>`)
	rootSection := rootSectionPattern.FindString(rendered)
	if rootSection == "" {
		t.Fatalf("expected onboarding root section tag")
	}
	if strings.Contains(rootSection, "x-cloak") {
		t.Fatalf("did not expect x-cloak on onboarding root section")
	}
}
