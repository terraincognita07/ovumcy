package api

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func parseDayPayload(c *fiber.Ctx) (dayPayload, error) {
	payload := dayPayload{Flow: models.FlowNone, SymptomIDs: []uint{}}
	contentType := strings.ToLower(c.Get("Content-Type"))

	if strings.Contains(contentType, "application/json") {
		if err := c.BodyParser(&payload); err != nil {
			return payload, err
		}
	} else {
		payload.IsPeriod = parseBoolValue(c.FormValue("is_period"))
		payload.Flow = strings.ToLower(strings.TrimSpace(c.FormValue("flow")))
		payload.Notes = strings.TrimSpace(c.FormValue("notes"))

		symptomRaw := c.Context().PostArgs().PeekMulti("symptom_ids")
		for _, value := range symptomRaw {
			parsed, err := strconv.ParseUint(string(value), 10, 64)
			if err == nil {
				payload.SymptomIDs = append(payload.SymptomIDs, uint(parsed))
			}
		}
	}

	payload.Flow = strings.ToLower(strings.TrimSpace(payload.Flow))
	if payload.Flow == "" {
		payload.Flow = models.FlowNone
	}
	payload.Notes = strings.TrimSpace(payload.Notes)

	return payload, nil
}

func parseBoolValue(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "1" || normalized == "true" || normalized == "on" || normalized == "yes"
}
