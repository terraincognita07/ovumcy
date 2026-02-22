package api

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) respondAuthError(c *fiber.Ctx, status int, message string) error {
	if strings.HasPrefix(c.Path(), "/api/auth/") && !acceptsJSON(c) && !isHTMX(c) {
		flash := FlashPayload{AuthError: message}
		switch c.Path() {
		case "/api/auth/register":
			email := normalizeLoginEmail(c.FormValue("email"))
			flash.RegisterEmail = email
			handler.setFlashCookie(c, flash)
			redirectValues := url.Values{}
			redirectValues.Set("error", strings.TrimSpace(message))
			if email != "" {
				redirectValues.Set("email", email)
			}
			return c.Redirect("/register?"+redirectValues.Encode(), fiber.StatusSeeOther)
		case "/api/auth/login":
			flash.LoginEmail = normalizeLoginEmail(c.FormValue("email"))
			handler.setFlashCookie(c, flash)
			return c.Redirect("/login", fiber.StatusSeeOther)
		case "/api/auth/forgot-password":
			handler.setFlashCookie(c, flash)
			return c.Redirect("/forgot-password", fiber.StatusSeeOther)
		case "/api/auth/reset-password":
			token := strings.TrimSpace(c.FormValue("token"))
			if token == "" {
				token = strings.TrimSpace(c.Query("token"))
			}
			handler.setFlashCookie(c, flash)
			if token == "" {
				return c.Redirect("/reset-password", fiber.StatusSeeOther)
			}
			return c.Redirect("/reset-password?token="+url.QueryEscape(token), fiber.StatusSeeOther)
		default:
			handler.setFlashCookie(c, flash)
			return c.Redirect("/login", fiber.StatusSeeOther)
		}
	}
	return apiError(c, status, message)
}

func (handler *Handler) respondSettingsError(c *fiber.Ctx, status int, message string) error {
	if isHTMX(c) {
		rendered := message
		if key := authErrorTranslationKey(message); key != "" {
			if localized := translateMessage(currentMessages(c), key); localized != key {
				rendered = localized
			}
		}
		return c.Status(fiber.StatusOK).SendString(fmt.Sprintf("<div class=\"status-error\">%s</div>", template.HTMLEscapeString(rendered)))
	}
	if (strings.HasPrefix(c.Path(), "/api/settings/") || strings.HasPrefix(c.Path(), "/settings/")) && !acceptsJSON(c) {
		handler.setFlashCookie(c, FlashPayload{SettingsError: message})
		return c.Redirect("/settings", fiber.StatusSeeOther)
	}
	return apiError(c, status, message)
}
