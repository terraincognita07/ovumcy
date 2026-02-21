package api

import (
	"net/url"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestBuildResetPasswordPageDataValidTokenAndForcedFlag(t *testing.T) {
	t.Parallel()

	handler := &Handler{secretKey: []byte("test-reset-secret")}
	token, err := handler.buildPasswordResetToken(42, 30*time.Minute)
	if err != nil {
		t.Fatalf("buildPasswordResetToken returned error: %v", err)
	}

	query := url.Values{
		"token":  {token},
		"forced": {"1"},
		"error":  {"weak password"},
	}
	flash := FlashPayload{AuthError: "invalid credentials"}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(handler.buildResetPasswordPageData(c, map[string]string{}, flash))
	})

	if payload["Token"] != token {
		t.Fatalf("expected token in payload, got %#v", payload["Token"])
	}
	if payload["InvalidToken"] != false {
		t.Fatalf("expected InvalidToken=false, got %#v", payload["InvalidToken"])
	}
	if payload["ForcedReset"] != true {
		t.Fatalf("expected ForcedReset=true, got %#v", payload["ForcedReset"])
	}
	if payload["ErrorKey"] != "auth.error.invalid_credentials" {
		t.Fatalf("expected flash error key, got %#v", payload["ErrorKey"])
	}
}

func TestBuildResetPasswordPageDataMarksInvalidToken(t *testing.T) {
	t.Parallel()

	handler := &Handler{secretKey: []byte("test-reset-secret")}
	query := url.Values{
		"token": {"invalid-token"},
	}

	payload := evaluateAuthPageBuilder(t, query, func(c *fiber.Ctx) error {
		return c.JSON(handler.buildResetPasswordPageData(c, map[string]string{}, FlashPayload{}))
	})

	if payload["InvalidToken"] != true {
		t.Fatalf("expected InvalidToken=true, got %#v", payload["InvalidToken"])
	}
}
