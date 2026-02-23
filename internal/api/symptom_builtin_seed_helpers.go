package api

import "github.com/terraincognita07/ovumcy/internal/models"

func (handler *Handler) seedBuiltinSymptoms(userID uint) error {
	var count int64
	if err := handler.db.Model(&models.SymptomType{}).
		Where("user_id = ? AND is_builtin = ?", userID, true).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	records := builtinSymptomRecordsForUser(userID)
	return handler.db.Create(&records).Error
}

func builtinSymptomRecordsForUser(userID uint) []models.SymptomType {
	builtin := models.DefaultBuiltinSymptoms()
	records := make([]models.SymptomType, 0, len(builtin))
	for _, symptom := range builtin {
		records = append(records, models.SymptomType{
			UserID:    userID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}
	return records
}
