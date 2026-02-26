package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/services"
)

var errOnboardingStepsRequired = services.ErrOnboardingStepsRequired

func (handler *Handler) completeOnboardingForUser(userID uint) (time.Time, error) {
	handler.ensureDependencies()
	return handler.onboardingSvc.CompleteOnboardingForUser(userID, handler.location)
}
