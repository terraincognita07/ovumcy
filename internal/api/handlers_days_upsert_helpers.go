package api

import (
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

var (
	errInvalidFlowValue     = errors.New("invalid flow value")
	errPeriodFlowRequired   = errors.New("period flow is required")
	errDayEntryLoadFailed   = errors.New("load day entry failed")
	errDayEntryCreateFailed = errors.New("create day entry failed")
	errDayEntryUpdateFailed = errors.New("update day entry failed")
	errDeleteDayFailed      = errors.New("delete day failed")
	errSyncLastPeriodFailed = errors.New("sync last period failed")
)

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

func (handler *Handler) upsertDayEntry(userID uint, dayStart time.Time, payload dayPayload, cleanIDs []uint) (models.DailyLog, bool, error) {
	dayKey := dayStorageKey(dayStart, handler.location)
	nextDayKey := nextDayStorageKey(dayStart, handler.location)

	wasPeriod := false
	var entry models.DailyLog
	result := handler.db.
		Where("user_id = ? AND date >= ? AND date < ?", userID, dayKey, nextDayKey).
		Order("date DESC, id DESC").
		First(&entry)
	if result.Error == nil {
		wasPeriod = entry.IsPeriod
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entry = models.DailyLog{
			UserID:   userID,
			Date:     dayStart,
			IsPeriod: payload.IsPeriod,
			Flow:     payload.Flow,
			Notes:    payload.Notes,
		}
		entry.SymptomIDs = cleanIDs
		if err := handler.db.Create(&entry).Error; err != nil {
			return models.DailyLog{}, false, errDayEntryCreateFailed
		}
		return entry, wasPeriod, nil
	}
	if result.Error != nil {
		return models.DailyLog{}, false, errDayEntryLoadFailed
	}

	entry.IsPeriod = payload.IsPeriod
	entry.Flow = payload.Flow
	entry.SymptomIDs = cleanIDs
	entry.Notes = payload.Notes
	if err := handler.db.Save(&entry).Error; err != nil {
		return models.DailyLog{}, false, errDayEntryUpdateFailed
	}
	return entry, wasPeriod, nil
}

func (handler *Handler) deleteDayAndRefreshLastPeriod(userID uint, day time.Time) error {
	if err := handler.deleteDailyLogByDate(userID, day); err != nil {
		return errDeleteDayFailed
	}
	if err := handler.refreshUserLastPeriodStart(userID); err != nil {
		return errSyncLastPeriodFailed
	}
	return nil
}
