package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestCurrentPageViewContextUsesLocalsAndHandlerLocation(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("UTC+5", 5*60*60)
	handler := &Handler{location: location}
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals(contextLanguageKey, "en")
		c.Locals(contextMessagesKey, map[string]string{"sample.key": "value"})

		language, messages, now := handler.currentPageViewContext(c)
		return c.JSON(fiber.Map{
			"language":    language,
			"has_message": messages["sample.key"] == "value",
			"location":    now.Location().String(),
		})
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
	if payload["language"] != "en" {
		t.Fatalf("expected language en, got %#v", payload["language"])
	}
	if payload["has_message"] != true {
		t.Fatalf("expected has_message=true, got %#v", payload["has_message"])
	}
	if payload["location"] != "UTC+5" {
		t.Fatalf("expected location UTC+5, got %#v", payload["location"])
	}
}
