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

func TestRedirectAuthenticatedUserIfPresentRedirectsAuthenticatedRequest(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	handler.secretKey = []byte("test-secret")
	user := createDataAccessTestUser(t, database, "redirect-helper@example.com")

	token, err := handler.buildToken(&user, time.Hour)
	if err != nil {
		t.Fatalf("buildToken returned error: %v", err)
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		redirected, redirectErr := handler.redirectAuthenticatedUserIfPresent(c)
		if redirectErr != nil {
			return redirectErr
		}
		if redirected {
			return nil
		}
		return c.SendStatus(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: authCookieName, Value: token})

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if response.Header.Get("Location") != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %q", response.Header.Get("Location"))
	}
}
