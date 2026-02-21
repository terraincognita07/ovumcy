package api

import (
	"errors"
	"fmt"
	"html/template"
	"net/mail"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func postLoginRedirectPath(user *models.User) string {
	if requiresOnboarding(user) {
		return "/onboarding"
	}
	return "/dashboard"
}

func requiresOnboarding(user *models.User) bool {
	if user == nil {
		return false
	}
	return user.Role == models.RoleOwner && !user.OnboardingCompleted
}

func (handler *Handler) respondAuthError(c *fiber.Ctx, status int, message string) error {
	if strings.HasPrefix(c.Path(), "/api/auth/") && !acceptsJSON(c) && !isHTMX(c) {
		flash := FlashPayload{AuthError: message}
		switch c.Path() {
		case "/api/auth/register":
			flash.RegisterEmail = normalizeLoginEmail(c.FormValue("email"))
			handler.setFlashCookie(c, flash)
			return c.Redirect("/register", fiber.StatusSeeOther)
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

func normalizeLoginEmail(raw string) string {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return ""
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ""
	}
	return email
}

func (handler *Handler) findUserByRecoveryCode(code string) (*models.User, error) {
	users := make([]models.User, 0)
	if err := handler.db.Where("recovery_code_hash <> ''").Find(&users).Error; err != nil {
		return nil, err
	}

	for index := range users {
		hash := strings.TrimSpace(users[index].RecoveryCodeHash)
		if hash == "" {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil {
			return &users[index], nil
		}
	}
	return nil, errors.New("recovery code not found")
}
