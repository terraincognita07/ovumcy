package api

import (
	"github.com/terraincognita07/ovumcy/internal/db"
	"github.com/terraincognita07/ovumcy/internal/services"
	"gorm.io/gorm"
)

func (handler *Handler) withDependencies(database *gorm.DB) *Handler {
	handler.repositories = db.NewRepositories(database)
	handler.authService = services.NewAuthService(handler.repositories.Users)
	handler.dayService = services.NewDayService(handler.repositories.DailyLogs, handler.repositories.Users)
	handler.symptomService = services.NewSymptomService(handler.repositories.Symptoms, handler.repositories.DailyLogs)
	handler.settingsService = services.NewSettingsService(handler.repositories.Users)
	handler.notificationService = services.NewNotificationService()
	handler.onboardingSvc = services.NewOnboardingService(handler.repositories.Users)
	handler.setupService = services.NewSetupService(handler.repositories.Users)
	return handler
}

func (handler *Handler) ensureDependencies() {
	if handler.repositories == nil {
		if handler.db == nil {
			return
		}
		handler.repositories = db.NewRepositories(handler.db)
	}

	if handler.authService == nil {
		handler.authService = services.NewAuthService(handler.repositories.Users)
	}
	if handler.dayService == nil {
		handler.dayService = services.NewDayService(handler.repositories.DailyLogs, handler.repositories.Users)
	}
	if handler.symptomService == nil {
		handler.symptomService = services.NewSymptomService(handler.repositories.Symptoms, handler.repositories.DailyLogs)
	}
	if handler.settingsService == nil {
		handler.settingsService = services.NewSettingsService(handler.repositories.Users)
	}
	if handler.notificationService == nil {
		handler.notificationService = services.NewNotificationService()
	}
	if handler.onboardingSvc == nil {
		handler.onboardingSvc = services.NewOnboardingService(handler.repositories.Users)
	}
	if handler.setupService == nil {
		handler.setupService = services.NewSetupService(handler.repositories.Users)
	}
}
