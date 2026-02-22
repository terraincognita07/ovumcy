package api

import "github.com/terraincognita07/lume/internal/models"

func (handler *Handler) registrationEmailExists(normalizedEmail string) (bool, error) {
	var matchedByQuery int64
	if err := handler.db.Model(&models.User{}).
		Where("email = ? OR lower(trim(email)) = ?", normalizedEmail, normalizedEmail).
		Count(&matchedByQuery).Error; err != nil {
		return false, err
	}
	if matchedByQuery > 0 {
		return true, nil
	}

	// Defensive fallback for legacy rows with atypical whitespace/casing forms.
	users := make([]models.User, 0)
	if err := handler.db.Model(&models.User{}).Select("email").Find(&users).Error; err != nil {
		return false, err
	}
	for _, user := range users {
		if normalizeLoginEmail(user.Email) == normalizedEmail {
			return true, nil
		}
	}
	return false, nil
}
