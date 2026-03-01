package api

import (
	"errors"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func normalizeLoginEmail(raw string) string {
	return services.NormalizeAuthEmail(raw)
}

func (handler *Handler) findUserByRecoveryCode(code string) (*models.User, error) {
	handler.ensureDependencies()
	user, err := handler.authService.FindUserByRecoveryCode(code)
	if err != nil {
		if errors.Is(err, services.ErrRecoveryCodeNotFound) {
			return nil, errors.New("recovery code not found")
		}
		return nil, err
	}
	return user, nil
}
