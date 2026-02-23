package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) setAuthCookie(c *fiber.Ctx, user *models.User, rememberMe bool) error {
	tokenTTL := defaultAuthTokenTTL
	if rememberMe {
		tokenTTL = rememberAuthTokenTTL
	}

	token, err := handler.buildToken(user, tokenTTL)
	if err != nil {
		return err
	}

	cookie := &fiber.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: "Lax",
	}
	if rememberMe {
		cookie.Expires = time.Now().Add(tokenTTL)
	}
	c.Cookie(cookie)
	return nil
}

func (handler *Handler) clearAuthCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: "Lax",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}

func (handler *Handler) buildToken(user *models.User, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = defaultAuthTokenTTL
	}
	now := time.Now()

	claims := authClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(handler.secretKey)
}
