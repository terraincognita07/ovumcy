package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) buildPasswordResetToken(userID uint, passwordHash string, ttl time.Duration) (string, error) {
	handler.ensureDependencies()
	return handler.authService.BuildPasswordResetToken(handler.secretKey, userID, passwordHash, ttl, time.Now())
}

func (handler *Handler) parsePasswordResetToken(rawToken string) (*passwordResetClaims, error) {
	handler.ensureDependencies()
	claims, err := services.ParsePasswordResetToken(handler.secretKey, rawToken, time.Now())
	if err != nil {
		return nil, err
	}

	return &passwordResetClaims{
		UserID:           claims.UserID,
		Purpose:          claims.Purpose,
		PasswordState:    claims.PasswordState,
		RegisteredClaims: claims.RegisteredClaims,
	}, nil
}
