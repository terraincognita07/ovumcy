package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (handler *Handler) buildPasswordResetToken(userID uint, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}

	now := time.Now()
	claims := passwordResetClaims{
		UserID:  userID,
		Purpose: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(handler.secretKey)
}

func (handler *Handler) parsePasswordResetToken(rawToken string) (uint, error) {
	if strings.TrimSpace(rawToken) == "" {
		return 0, errors.New("missing token")
	}

	claims := &passwordResetClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return handler.secretKey, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	if claims.Purpose != "password_reset" {
		return 0, errors.New("invalid token purpose")
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return 0, errors.New("token expired")
	}
	if claims.UserID == 0 {
		return 0, errors.New("invalid user id")
	}
	return claims.UserID, nil
}
