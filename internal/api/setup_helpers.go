package api

import "github.com/terraincognita07/lume/internal/models"

func (handler *Handler) requiresInitialSetup() (bool, error) {
	var usersCount int64
	if err := handler.db.Model(&models.User{}).Count(&usersCount).Error; err != nil {
		return false, err
	}
	return usersCount == 0, nil
}
