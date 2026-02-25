package api

import (
	"fmt"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
)

func htmxDismissibleSuccessStatusMarkup(messages map[string]string, message string) string {
	closeLabel := translateMessage(messages, "common.close")
	if closeLabel == "" || closeLabel == "common.close" {
		closeLabel = "Close"
	}

	return fmt.Sprintf(
		"<div class=\"status-ok\"><div class=\"toast-body\"><span class=\"toast-message\">%s</span><button type=\"button\" class=\"toast-close\" data-dismiss-status aria-label=\"%s\">×</button></div></div>",
		template.HTMLEscapeString(message),
		template.HTMLEscapeString(closeLabel),
	)
}

func (handler *Handler) sendDaySaveStatus(c *fiber.Ctx) error {
	timestamp := time.Now().In(handler.location).Format("15:04")
	pattern := translateMessage(currentMessages(c), "common.saved_at")
	if pattern == "common.saved_at" {
		pattern = "Saved at %s"
	}
	message := fmt.Sprintf(pattern, timestamp)
	return c.SendString(htmxDismissibleSuccessStatusMarkup(currentMessages(c), message))
}
