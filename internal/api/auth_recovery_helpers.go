package api

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func normalizeLoginEmail(raw string) string {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return ""
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ""
	}
	return email
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
