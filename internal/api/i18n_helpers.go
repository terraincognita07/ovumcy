package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func translateMessage(messages map[string]string, key string) string {
	if key == "" {
		return ""
	}
	if messages != nil {
		if value, ok := messages[key]; ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return key
}

func currentLanguage(c *fiber.Ctx) string {
	language, ok := c.Locals(contextLanguageKey).(string)
	if !ok || strings.TrimSpace(language) == "" {
		return ""
	}
	return language
}

func currentMessages(c *fiber.Ctx) map[string]string {
	messages, ok := c.Locals(contextMessagesKey).(map[string]string)
	if !ok || messages == nil {
		return map[string]string{}
	}
	return messages
}

func (handler *Handler) withTemplateDefaults(c *fiber.Ctx, data fiber.Map) fiber.Map {
	if data == nil {
		data = fiber.Map{}
	}

	messages := currentMessages(c)
	if _, ok := data["Messages"]; !ok {
		data["Messages"] = messages
	}

	if _, ok := data["Lang"]; !ok {
		language := currentLanguage(c)
		if language == "" {
			language = handler.i18n.DefaultLanguage()
		}
		data["Lang"] = language
	}

	if _, ok := data["CurrentPath"]; !ok {
		data["CurrentPath"] = currentPathWithQuery(c)
	}

	if _, ok := data["CSRFToken"]; !ok {
		data["CSRFToken"] = csrfToken(c)
	}

	if _, ok := data["NoDataLabel"]; !ok {
		noData := translateMessage(messages, "common.not_available")
		if noData == "common.not_available" {
			noData = "-"
		}
		data["NoDataLabel"] = noData
	}

	return data
}

func currentPathWithQuery(c *fiber.Ctx) string {
	path := string(c.Request().URI().RequestURI())
	if path == "" {
		return c.Path()
	}
	return path
}
