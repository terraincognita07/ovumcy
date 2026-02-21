package api

import (
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

const maxDayNotesLength = 2000

var (
	errInvalidFlowValue   = errors.New("invalid flow value")
	errPeriodFlowRequired = errors.New("period flow is required")
)

func normalizeDayPayload(payload dayPayload) (dayPayload, error) {
	if !isValidFlow(payload.Flow) {
		return payload, errInvalidFlowValue
	}
	if payload.IsPeriod && payload.Flow == models.FlowNone {
		return payload, errPeriodFlowRequired
	}
	if !payload.IsPeriod {
		payload.Flow = models.FlowNone
	}
	payload.Notes = trimDayNotes(payload.Notes)
	return payload, nil
}

func trimDayNotes(value string) string {
	if len(value) > maxDayNotesLength {
		return value[:maxDayNotesLength]
	}
	return value
}

func (handler *Handler) loadDayAutoFillSettings(userID uint) (int, bool, error) {
	periodLength := 5
	settings := struct {
		PeriodLength   int
		AutoPeriodFill bool
	}{}

	if err := handler.db.Model(&models.User{}).
		Select("period_length", "auto_period_fill").
		First(&settings, userID).Error; err != nil {
		return periodLength, false, err
	}
	if isValidOnboardingPeriodLength(settings.PeriodLength) {
		periodLength = settings.PeriodLength
	}
	return periodLength, settings.AutoPeriodFill, nil
}

func (handler *Handler) shouldAutoFillPeriodDays(userID uint, dayStart time.Time, wasPeriod bool, autoPeriodFillEnabled bool, periodLength int) (bool, error) {
	if !autoPeriodFillEnabled || periodLength <= 1 || wasPeriod {
		return false, nil
	}

	previousDay := dayStart.AddDate(0, 0, -1)
	previousDayEntry, err := handler.fetchLogByDate(userID, previousDay)
	if err != nil {
		return false, err
	}

	hasRecentPeriod, err := handler.hasPeriodInRecentDays(userID, dayStart, 3)
	if err != nil {
		return false, err
	}

	return !previousDayEntry.IsPeriod && !hasRecentPeriod, nil
}

func (handler *Handler) sendDaySaveStatus(c *fiber.Ctx) error {
	timestamp := time.Now().In(handler.location).Format("15:04")
	pattern := translateMessage(currentMessages(c), "common.saved_at")
	if pattern == "common.saved_at" {
		pattern = "Saved at %s"
	}
	message := fmt.Sprintf(pattern, timestamp)
	return c.SendString(fmt.Sprintf("<div class=\"status-ok status-transient\">%s</div>", template.HTMLEscapeString(message)))
}
