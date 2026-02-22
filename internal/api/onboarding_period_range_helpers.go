package api

import (
	"errors"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

func (handler *Handler) upsertOnboardingPeriodRange(tx *gorm.DB, userID uint, startDay time.Time, endDay time.Time) error {
	if endDay.Before(startDay) {
		return errors.New("invalid onboarding range")
	}

	for cursor := startDay; !cursor.After(endDay); cursor = cursor.AddDate(0, 0, 1) {
		day := dateAtLocation(cursor, handler.location)
		if err := handler.upsertOnboardingPeriodDay(tx, userID, day); err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) upsertOnboardingPeriodDay(tx *gorm.DB, userID uint, day time.Time) error {
	dayKey := dayStorageKey(day, handler.location)

	var entry models.DailyLog
	result := tx.
		Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).
		Order("date DESC, id DESC").
		First(&entry)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entry = models.DailyLog{
			UserID:     userID,
			Date:       day,
			IsPeriod:   true,
			Flow:       models.FlowNone,
			SymptomIDs: []uint{},
		}
		return tx.Create(&entry).Error
	}
	if result.Error != nil {
		return result.Error
	}

	return tx.Model(&entry).Updates(map[string]any{
		"is_period": true,
		"flow":      models.FlowNone,
	}).Error
}
