package api

import "github.com/terraincognita07/lume/internal/models"

func (handler *Handler) registrationEmailExists(normalizedEmail string) (bool, error) {
	var matched int64
	if err := handler.db.Model(&models.User{}).
		Where("lower(trim(email)) = ?", normalizedEmail).
		Count(&matched).Error; err != nil {
		return false, err
	}
	return matched > 0, nil
}
