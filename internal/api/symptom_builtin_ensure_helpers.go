package api

import (
	"strings"

	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) ensureBuiltinSymptoms(userID uint) error {
	if err := handler.db.
		Model(&models.SymptomType{}).
		Where("user_id = ? AND lower(trim(name)) = ?", userID, "fatique").
		Update("name", "Fatigue").Error; err != nil {
		return err
	}

	existing := make([]models.SymptomType, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&existing).Error; err != nil {
		return err
	}

	existingByName := make(map[string]struct{}, len(existing))
	for _, symptom := range existing {
		key := strings.ToLower(strings.TrimSpace(symptom.Name))
		if key != "" {
			existingByName[key] = struct{}{}
		}
	}

	missing := missingBuiltinSymptomsForUser(userID, existingByName)
	if len(missing) == 0 {
		return nil
	}
	return handler.db.Create(&missing).Error
}

func missingBuiltinSymptomsForUser(userID uint, existingByName map[string]struct{}) []models.SymptomType {
	missing := make([]models.SymptomType, 0)
	for _, symptom := range models.DefaultBuiltinSymptoms() {
		key := strings.ToLower(strings.TrimSpace(symptom.Name))
		if _, ok := existingByName[key]; ok {
			continue
		}
		missing = append(missing, models.SymptomType{
			UserID:    userID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}
	return missing
}
