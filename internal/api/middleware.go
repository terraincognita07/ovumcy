package api

import (
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

func currentUser(c *fiber.Ctx) (*models.User, bool) {
	user, ok := c.Locals(contextUserKey).(*models.User)
	return user, ok
}
