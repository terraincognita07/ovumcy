package api

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) NotFound(c *fiber.Ctx) error {
	if strings.HasPrefix(c.Path(), "/api/") || acceptsJSON(c) {
		return apiError(c, fiber.StatusNotFound, "not found")
	}

	if isHTMX(c) {
		message := translateMessage(currentMessages(c), "not_found.title")
		if message == "not_found.title" {
			message = "Page not found"
		}
		c.Status(fiber.StatusNotFound)
		return c.SendString(fmt.Sprintf("<div class=\"status-error\">%s</div>", template.HTMLEscapeString(message)))
	}

	currentUser := handler.optionalAuthenticatedUser(c)
	if currentUser != nil {
		c.Locals(contextUserKey, currentUser)
	}

	primaryPath := "/login"
	primaryLabelKey := "not_found.action_login"
	if currentUser != nil {
		primaryPath = "/dashboard"
		primaryLabelKey = "not_found.action_dashboard"
	}

	c.Status(fiber.StatusNotFound)
	return handler.render(c, "not_found", fiber.Map{
		"Title":           localizedPageTitle(currentMessages(c), "meta.title.not_found", "Ovumcy | Page Not Found"),
		"CurrentUser":     currentUser,
		"PrimaryPath":     primaryPath,
		"PrimaryLabelKey": primaryLabelKey,
	})
}
