package api

import (
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) ensureBuiltinSymptoms(userID uint) error {
	handler.ensureDependencies()
	return handler.symptomService.EnsureBuiltinSymptoms(userID)
}

func missingBuiltinSymptomsForUser(userID uint, existingByName map[string]struct{}) []models.SymptomType {
	return services.MissingBuiltinSymptomsForUser(userID, existingByName)
}
