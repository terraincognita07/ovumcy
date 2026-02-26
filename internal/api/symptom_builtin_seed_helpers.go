package api

import (
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) seedBuiltinSymptoms(userID uint) error {
	handler.ensureDependencies()
	return handler.symptomService.SeedBuiltinSymptoms(userID)
}

func builtinSymptomRecordsForUser(userID uint) []models.SymptomType {
	return services.BuiltinSymptomRecordsForUser(userID)
}
