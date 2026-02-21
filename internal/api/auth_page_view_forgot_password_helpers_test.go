package api

import (
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestBuildForgotPasswordPageDataPrefersFlashError(t *testing.T) {
	t.Parallel()

	query := url.Values{
		"error": {"weak password"},
	}
	flash := FlashPayload{AuthError: "invalid credentials"}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(buildForgotPasswordPageData(c, map[string]string{}, flash))
	})

	if payload["ErrorKey"] != "auth.error.invalid_credentials" {
		t.Fatalf("expected flash error key, got %#v", payload["ErrorKey"])
	}
}
