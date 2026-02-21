package api

import (
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestBuildLoginPageDataUsesFlashPriorityAndSetupFlag(t *testing.T) {
	t.Parallel()

	query := url.Values{
		"error": {"weak password"},
		"email": {"query@example.com"},
	}
	flash := FlashPayload{
		AuthError:  "invalid credentials",
		LoginEmail: " Flash@Example.com ",
	}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(buildLoginPageData(c, map[string]string{}, flash, true))
	})

	if payload["ErrorKey"] != "auth.error.invalid_credentials" {
		t.Fatalf("expected flash error key, got %#v", payload["ErrorKey"])
	}
	if payload["Email"] != "flash@example.com" {
		t.Fatalf("expected normalized flash email, got %#v", payload["Email"])
	}
	if payload["IsFirstLaunch"] != true {
		t.Fatalf("expected IsFirstLaunch=true, got %#v", payload["IsFirstLaunch"])
	}
}
