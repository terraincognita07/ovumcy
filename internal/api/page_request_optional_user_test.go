package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestOptionalAuthenticatedUserWithoutCookieReturnsNil(t *testing.T) {
	t.Parallel()

	handler := &Handler{secretKey: []byte("secret")}
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		user := handler.optionalAuthenticatedUser(c)
		return c.JSON(fiber.Map{"has_user": user != nil})
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	payload := map[string]any{}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload failed: %v", err)
	}
	if payload["has_user"] != false {
		t.Fatalf("expected has_user=false, got %#v", payload["has_user"])
	}
}
