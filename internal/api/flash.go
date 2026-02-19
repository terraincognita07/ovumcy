package api

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetFlashCookie(c *fiber.Ctx, payload FlashPayload) {
	setFlashCookie(c, payload)
}

func setFlashCookie(c *fiber.Ctx, payload FlashPayload) {
	payload.AuthError = strings.TrimSpace(payload.AuthError)
	payload.SettingsError = strings.TrimSpace(payload.SettingsError)
	payload.SettingsSuccess = strings.TrimSpace(payload.SettingsSuccess)
	payload.LoginEmail = normalizeLoginEmail(payload.LoginEmail)

	if payload.AuthError == "" &&
		payload.SettingsError == "" &&
		payload.SettingsSuccess == "" &&
		payload.LoginEmail == "" {
		clearFlashCookie(c)
		return
	}

	serialized, err := json.Marshal(payload)
	if err != nil {
		return
	}
	encoded := base64.RawURLEncoding.EncodeToString(serialized)

	c.Cookie(&fiber.Cookie{
		Name:     flashCookieName,
		Value:    encoded,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(5 * time.Minute),
	})
}

func popFlashCookie(c *fiber.Ctx) FlashPayload {
	raw := strings.TrimSpace(c.Cookies(flashCookieName))
	if raw == "" {
		return FlashPayload{}
	}
	clearFlashCookie(c)

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return FlashPayload{}
	}

	payload := FlashPayload{}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return FlashPayload{}
	}
	payload.AuthError = strings.TrimSpace(payload.AuthError)
	payload.SettingsError = strings.TrimSpace(payload.SettingsError)
	payload.SettingsSuccess = strings.TrimSpace(payload.SettingsSuccess)
	payload.LoginEmail = normalizeLoginEmail(payload.LoginEmail)
	return payload
}

func clearFlashCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}
