package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) parseExportRange(c *fiber.Ctx) (*time.Time, *time.Time, string) {
	rawFrom := strings.TrimSpace(c.Query("from"))
	rawTo := strings.TrimSpace(c.Query("to"))

	var from *time.Time
	if rawFrom != "" {
		parsedFrom, err := parseDayParam(rawFrom, handler.location)
		if err != nil {
			return nil, nil, "invalid from date"
		}
		from = &parsedFrom
	}

	var to *time.Time
	if rawTo != "" {
		parsedTo, err := parseDayParam(rawTo, handler.location)
		if err != nil {
			return nil, nil, "invalid to date"
		}
		to = &parsedTo
	}

	if from != nil && to != nil && to.Before(*from) {
		return nil, nil, "invalid range"
	}

	return from, to, ""
}

func (handler *Handler) exportUserAndRange(c *fiber.Ctx) (*models.User, *time.Time, *time.Time, int, string) {
	user, ok := currentUser(c)
	if !ok || user == nil {
		return nil, nil, nil, fiber.StatusUnauthorized, "unauthorized"
	}

	from, to, rangeError := handler.parseExportRange(c)
	if rangeError != "" {
		return nil, nil, nil, fiber.StatusBadRequest, rangeError
	}

	return user, from, to, 0, ""
}

func buildExportFilename(now time.Time, extension string) string {
	return fmt.Sprintf("lume-export-%s.%s", now.Format("2006-01-02"), extension)
}

func setExportAttachmentHeaders(c *fiber.Ctx, contentType string, filename string) {
	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
}
