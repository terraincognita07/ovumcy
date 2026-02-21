package api

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
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
	users := make([]models.User, 0)
	if err := handler.db.Where("recovery_code_hash <> ''").Find(&users).Error; err != nil {
		return nil, err
	}

	for index := range users {
		hash := strings.TrimSpace(users[index].RecoveryCodeHash)
		if hash == "" {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil {
			return &users[index], nil
		}
	}
	return nil, errors.New("recovery code not found")
}
