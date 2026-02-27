package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/ovumcy/internal/models"
)

type authClaims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (handler *Handler) authenticateRequest(c *fiber.Ctx) (*models.User, error) {
	rawToken := strings.TrimSpace(c.Cookies(authCookieName))
	if rawToken == "" {
		return nil, errors.New("missing auth cookie")
	}
	tokenValue := rawToken
	if strings.HasPrefix(rawToken, secureCookieVersion+".") {
		decodedToken, err := handler.decodeSealedAuthCookieToken(rawToken)
		if err != nil {
			return nil, errors.New("invalid token")
		}
		tokenValue = decodedToken
	}

	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return handler.secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	handler.ensureDependencies()
	user, err := handler.authService.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
