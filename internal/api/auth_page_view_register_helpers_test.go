package api

import (
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestBuildRegisterPageDataFallsBackToQueryValues(t *testing.T) {
	t.Parallel()

	query := url.Values{
		"error": {"weak password"},
		"email": {"Query@Example.com"},
	}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(buildRegisterPageData(c, map[string]string{}, FlashPayload{}, false))
	})

	if payload["ErrorKey"] != "auth.error.weak_password" {
		t.Fatalf("expected query error key, got %#v", payload["ErrorKey"])
	}
	if payload["Email"] != "query@example.com" {
		t.Fatalf("expected normalized query email, got %#v", payload["Email"])
	}
	if payload["IsFirstLaunch"] != false {
		t.Fatalf("expected IsFirstLaunch=false, got %#v", payload["IsFirstLaunch"])
	}
}
