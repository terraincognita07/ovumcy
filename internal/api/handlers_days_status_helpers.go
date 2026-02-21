package api

import (
	"fmt"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) sendDaySaveStatus(c *fiber.Ctx) error {
	timestamp := time.Now().In(handler.location).Format("15:04")
	pattern := translateMessage(currentMessages(c), "common.saved_at")
	if pattern == "common.saved_at" {
		pattern = "Saved at %s"
	}
	message := fmt.Sprintf(pattern, timestamp)
	return c.SendString(fmt.Sprintf("<div class=\"status-ok status-transient\">%s</div>", template.HTMLEscapeString(message)))
}
