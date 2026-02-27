package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (handler *Handler) buildPasswordResetToken(userID uint, passwordHash string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	passwordState := passwordStateFingerprint(passwordHash)
	if passwordState == "" {
		return "", errors.New("invalid password state")
	}

	now := time.Now()
	claims := passwordResetClaims{
		UserID:        userID,
		Purpose:       "password_reset",
		PasswordState: passwordState,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(handler.secretKey)
}

func (handler *Handler) parsePasswordResetToken(rawToken string) (*passwordResetClaims, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, errors.New("missing token")
	}

	claims := &passwordResetClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return handler.secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Purpose != "password_reset" {
		return nil, errors.New("invalid token purpose")
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}
	if claims.UserID == 0 {
		return nil, errors.New("invalid user id")
	}
	if strings.TrimSpace(claims.PasswordState) == "" {
		return nil, errors.New("invalid password state")
	}
	return claims, nil
}

func passwordStateFingerprint(passwordHash string) string {
	normalizedHash := strings.TrimSpace(passwordHash)
	if normalizedHash == "" {
		return ""
	}

	sum := sha256.Sum256([]byte("ovumcy.reset.password-state.v1:" + normalizedHash))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func isPasswordStateFingerprintMatch(expected string, passwordHash string) bool {
	actual := passwordStateFingerprint(passwordHash)
	if strings.TrimSpace(expected) == "" || strings.TrimSpace(actual) == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
