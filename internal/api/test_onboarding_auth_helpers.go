package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func loginAndExtractAuthCookie(t *testing.T, app *fiber.App, email string, password string) string {
	t.Helper()

	form := url.Values{
		"email":    {email},
		"password": {password},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected login status 303, got %d", response.StatusCode)
	}

	for _, cookie := range response.Cookies() {
		if cookie.Name == "lume_auth" && cookie.Value != "" {
			return cookie.Name + "=" + cookie.Value
		}
	}

	t.Fatal("auth cookie is missing in login response")
	return ""
}
