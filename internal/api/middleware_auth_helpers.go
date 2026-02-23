package api

import (
	"errors"
	"fmt"
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
	rawToken := c.Cookies(authCookieName)
	if rawToken == "" {
		return nil, errors.New("missing auth cookie")
	}

	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
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

	var user models.User
	if err := handler.db.First(&user, claims.UserID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
