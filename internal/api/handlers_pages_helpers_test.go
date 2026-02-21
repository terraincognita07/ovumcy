package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type authHelperPayload struct {
	ErrorKey string `json:"error_key"`
	Email    string `json:"email"`
}

func evaluateAuthHelpers(t *testing.T, rawQuery string, flashError string, flashEmail string) authHelperPayload {
	t.Helper()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"error_key": authErrorKeyFromFlashOrQuery(c, flashError),
			"email":     loginEmailFromFlashOrQuery(c, flashEmail),
		})
	})

	request := httptest.NewRequest(http.MethodGet, "/?"+rawQuery, nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", response.StatusCode)
	}

	var payload authHelperPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}
	return payload
}

func TestAuthErrorKeyFromFlashOrQueryPrefersFlashValue(t *testing.T) {
	payload := evaluateAuthHelpers(t, "error=weak+password", "  invalid credentials ", "")
	if payload.ErrorKey != "auth.error.invalid_credentials" {
		t.Fatalf("expected flash error key, got %q", payload.ErrorKey)
	}
}

func TestAuthErrorKeyFromFlashOrQueryFallsBackToQuery(t *testing.T) {
	payload := evaluateAuthHelpers(t, "error=weak+password", "", "")
	if payload.ErrorKey != "auth.error.weak_password" {
		t.Fatalf("expected query-based error key, got %q", payload.ErrorKey)
	}
}

func TestLoginEmailFromFlashOrQueryPrefersFlashValue(t *testing.T) {
	payload := evaluateAuthHelpers(t, "email=query@example.com", "", "  Flash@Example.com ")
	if payload.Email != "flash@example.com" {
		t.Fatalf("expected flash email, got %q", payload.Email)
	}
}

func TestLoginEmailFromFlashOrQueryFallsBackToQuery(t *testing.T) {
	payload := evaluateAuthHelpers(t, "email=Query@Example.com", "", "")
	if payload.Email != "query@example.com" {
		t.Fatalf("expected query email fallback, got %q", payload.Email)
	}
}
