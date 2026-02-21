package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) LanguageMiddleware(c *fiber.Ctx) error {
	cookieLanguage := c.Cookies(languageCookieName)
	language := handler.i18n.DetectFromAcceptLanguage(c.Get("Accept-Language"))
	if cookieLanguage != "" {
		language = handler.i18n.NormalizeLanguage(cookieLanguage)
	}

	if cookieLanguage != language {
		handler.setLanguageCookie(c, language)
	}

	c.Locals(contextLanguageKey, language)
	c.Locals(contextMessagesKey, handler.i18n.Messages(language))
	return c.Next()
}

func (handler *Handler) setLanguageCookie(c *fiber.Ctx, language string) {
	c.Cookie(&fiber.Cookie{
		Name:     languageCookieName,
		Value:    handler.i18n.NormalizeLanguage(language),
		Path:     "/",
		HTTPOnly: false,
		Secure:   handler.cookieSecure,
		SameSite: "Lax",
		Expires:  time.Now().AddDate(1, 0, 0),
	})
}
