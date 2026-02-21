package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

const (
	authCookieName     = "lume_auth"
	languageCookieName = "lume_lang"
	flashCookieName    = "lume_flash"
	contextUserKey     = "current_user"
	contextLanguageKey = "current_language"
	contextMessagesKey = "current_messages"
)

func (handler *Handler) AuthRequired(c *fiber.Ctx) error {
	user, err := handler.authenticateRequest(c)
	if err != nil {
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Redirect("/login", fiber.StatusSeeOther)
	}

	c.Locals(contextUserKey, user)
	if requiresOnboarding(user) && !isOnboardingPath(c.Path()) {
		if strings.HasPrefix(c.Path(), "/api/") {
			if c.Path() == "/api/auth/logout" {
				return c.Next()
			}
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "onboarding required"})
		}
		return c.Redirect("/onboarding", fiber.StatusSeeOther)
	}

	return c.Next()
}

func (handler *Handler) OwnerOnly(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if user.Role != models.RoleOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "owner access required"})
	}
	return c.Next()
}

func currentUser(c *fiber.Ctx) (*models.User, bool) {
	user, ok := c.Locals(contextUserKey).(*models.User)
	return user, ok
}

func isOnboardingPath(path string) bool {
	cleanPath := strings.TrimSpace(path)
	return cleanPath == "/onboarding" || strings.HasPrefix(cleanPath, "/onboarding/")
}
