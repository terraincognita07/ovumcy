package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

const (
	authCookieName          = "ovumcy_auth"
	languageCookieName      = "ovumcy_lang"
	flashCookieName         = "ovumcy_flash"
	recoveryCodeCookieName  = "ovumcy_recovery_code"
	resetPasswordCookieName = "ovumcy_reset_password"
	contextUserKey          = "current_user"
	contextLanguageKey      = "current_language"
	contextMessagesKey      = "current_messages"
)

func currentUser(c *fiber.Ctx) (*models.User, bool) {
	user, ok := c.Locals(contextUserKey).(*models.User)
	return user, ok
}
