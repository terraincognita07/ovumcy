package services

import (
	"errors"
	"strings"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSettingsPasswordMissing = errors.New("settings password missing")
	ErrSettingsPasswordInvalid = errors.New("settings password invalid")
)

type SettingsUserRepository interface {
	UpdateDisplayName(userID uint, displayName string) error
	UpdateRecoveryCodeHash(userID uint, recoveryHash string) error
	UpdatePassword(userID uint, passwordHash string, mustChangePassword bool) error
	UpdateByID(userID uint, updates map[string]any) error
	LoadSettingsByID(userID uint) (models.User, error)
	ClearAllDataAndResetSettings(userID uint) error
	DeleteAccountAndRelatedData(userID uint) error
}

type CycleSettingsUpdate struct {
	CycleLength        int
	PeriodLength       int
	AutoPeriodFill     bool
	LastPeriodStartSet bool
	LastPeriodStart    *time.Time
}

type SettingsService struct {
	users SettingsUserRepository
}

func NewSettingsService(users SettingsUserRepository) *SettingsService {
	return &SettingsService{users: users}
}

func (service *SettingsService) UpdateDisplayName(userID uint, displayName string) error {
	return service.users.UpdateDisplayName(userID, displayName)
}

func (service *SettingsService) UpdateRecoveryCodeHash(userID uint, recoveryHash string) error {
	return service.users.UpdateRecoveryCodeHash(userID, recoveryHash)
}

func (service *SettingsService) UpdatePassword(userID uint, passwordHash string, mustChangePassword bool) error {
	return service.users.UpdatePassword(userID, passwordHash, mustChangePassword)
}

func (service *SettingsService) ValidateDeleteAccountPassword(passwordHash string, rawPassword string) error {
	password := strings.TrimSpace(rawPassword)
	if password == "" {
		return ErrSettingsPasswordMissing
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return ErrSettingsPasswordInvalid
	}
	return nil
}

func (service *SettingsService) SaveCycleSettings(userID uint, settings CycleSettingsUpdate) error {
	updates := map[string]any{
		"cycle_length":     settings.CycleLength,
		"period_length":    settings.PeriodLength,
		"auto_period_fill": settings.AutoPeriodFill,
	}
	if settings.LastPeriodStartSet {
		if settings.LastPeriodStart == nil {
			updates["last_period_start"] = nil
		} else {
			updates["last_period_start"] = *settings.LastPeriodStart
		}
	}
	return service.users.UpdateByID(userID, updates)
}

func (service *SettingsService) LoadSettings(userID uint) (models.User, error) {
	return service.users.LoadSettingsByID(userID)
}

func (service *SettingsService) ClearAllData(userID uint) error {
	return service.users.ClearAllDataAndResetSettings(userID)
}

func (service *SettingsService) DeleteAccount(userID uint) error {
	return service.users.DeleteAccountAndRelatedData(userID)
}
