package db

import (
	"errors"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) *UserRepository {
	return &UserRepository{database: database}
}

func (repo *UserRepository) CountUsers() (int64, error) {
	var count int64
	if err := repo.database.Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *UserRepository) FindByID(userID uint) (models.User, error) {
	var user models.User
	if err := repo.database.First(&user, userID).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (repo *UserRepository) FindByNormalizedEmail(email string) (models.User, error) {
	var user models.User
	if err := repo.database.Where("lower(trim(email)) = ?", email).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (repo *UserRepository) ExistsByNormalizedEmail(email string) (bool, error) {
	var matched int64
	if err := repo.database.Model(&models.User{}).
		Where("lower(trim(email)) = ?", email).
		Count(&matched).Error; err != nil {
		return false, err
	}
	return matched > 0, nil
}

func (repo *UserRepository) Create(user *models.User) error {
	return repo.database.Create(user).Error
}

func (repo *UserRepository) Save(user *models.User) error {
	return repo.database.Save(user).Error
}

func (repo *UserRepository) UpdateDisplayName(userID uint, displayName string) error {
	return repo.database.Model(&models.User{}).Where("id = ?", userID).Update("display_name", displayName).Error
}

func (repo *UserRepository) UpdateRecoveryCodeHash(userID uint, recoveryHash string) error {
	return repo.database.Model(&models.User{}).Where("id = ?", userID).Update("recovery_code_hash", recoveryHash).Error
}

func (repo *UserRepository) UpdatePassword(userID uint, passwordHash string, mustChangePassword bool) error {
	return repo.database.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
		"password_hash":        passwordHash,
		"must_change_password": mustChangePassword,
	}).Error
}

func (repo *UserRepository) UpdateByID(userID uint, updates map[string]any) error {
	return repo.database.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (repo *UserRepository) LoadSettingsByID(userID uint) (models.User, error) {
	var user models.User
	if err := repo.database.
		Select("cycle_length", "period_length", "auto_period_fill", "last_period_start").
		First(&user, userID).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (repo *UserRepository) ListWithRecoveryCodeHash() ([]models.User, error) {
	users := make([]models.User, 0)
	if err := repo.database.Where("recovery_code_hash <> ''").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *UserRepository) SaveOnboardingStep1(userID uint, start time.Time) error {
	return repo.database.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
		"last_period_start": start,
	}).Error
}

func (repo *UserRepository) SaveOnboardingStep2(userID uint, cycleLength int, periodLength int, autoPeriodFill bool) error {
	return repo.database.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
		"cycle_length":     cycleLength,
		"period_length":    periodLength,
		"auto_period_fill": autoPeriodFill,
	}).Error
}

func (repo *UserRepository) ClearAllDataAndResetSettings(userID uint) error {
	return repo.database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.DailyLog{}).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
			"cycle_length":      models.DefaultCycleLength,
			"period_length":     models.DefaultPeriodLength,
			"auto_period_fill":  true,
			"last_period_start": nil,
		}).Error
	})
}

func (repo *UserRepository) DeleteAccountAndRelatedData(userID uint) error {
	return repo.database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.DailyLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.SymptomType{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.User{}, userID).Error
	})
}

func (repo *UserRepository) CompleteOnboarding(userID uint, startDay time.Time, periodLength int) error {
	if periodLength <= 0 {
		return errors.New("invalid period length")
	}
	endDay := startDay.AddDate(0, 0, periodLength-1)
	if endDay.Before(startDay) {
		return errors.New("invalid onboarding range")
	}

	return repo.database.Transaction(func(tx *gorm.DB) error {
		for cursor := startDay; !cursor.After(endDay); cursor = cursor.AddDate(0, 0, 1) {
			dayStart := cursor
			dayEnd := dayStart.AddDate(0, 0, 1)

			var entry models.DailyLog
			result := tx.
				Where("user_id = ? AND date >= ? AND date < ?", userID, dayStart, dayEnd).
				Order("date DESC, id DESC").
				First(&entry)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				entry = models.DailyLog{
					UserID:     userID,
					Date:       dayStart,
					IsPeriod:   true,
					Flow:       models.FlowNone,
					SymptomIDs: []uint{},
				}
				if err := tx.Create(&entry).Error; err != nil {
					return err
				}
				continue
			}
			if result.Error != nil {
				return result.Error
			}

			if err := tx.Model(&entry).Updates(map[string]any{
				"is_period": true,
				"flow":      models.FlowNone,
			}).Error; err != nil {
				return err
			}
		}

		return tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
			"last_period_start":    startDay,
			"onboarding_completed": true,
		}).Error
	})
}
