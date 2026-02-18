package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/lume/internal/models"
)

const (
	authCookieName     = "lume_auth"
	languageCookieName = "lume_lang"
	contextUserKey     = "current_user"
	contextLanguageKey = "current_language"
	contextMessagesKey = "current_messages"
)

type authClaims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (handler *Handler) LanguageMiddleware(c *fiber.Ctx) error {
	cookieLanguage := c.Cookies(languageCookieName)
	language := handler.i18n.DefaultLanguage()

	if cookieLanguage != "" {
		language = handler.i18n.NormalizeLanguage(cookieLanguage)
	} else {
		language = handler.i18n.DetectFromAcceptLanguage(c.Get("Accept-Language"))
	}

	if cookieLanguage != language {
		handler.setLanguageCookie(c, language)
	}

	c.Locals(contextLanguageKey, language)
	c.Locals(contextMessagesKey, handler.i18n.Messages(language))
	return c.Next()
}

func (handler *Handler) AuthRequired(c *fiber.Ctx) error {
	user, err := handler.authenticateRequest(c)
	if err != nil {
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Redirect("/login", fiber.StatusSeeOther)
	}

	c.Locals(contextUserKey, user)
	if requiresOnboarding(user) && !isOnboardingPath(c.Path()) {
		if strings.HasPrefix(c.Path(), "/api/") {
			if c.Path() == "/api/auth/logout" {
				return c.Next()
			}
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "onboarding required"})
		}
		return c.Redirect("/onboarding", fiber.StatusSeeOther)
	}

	return c.Next()
}

func (handler *Handler) OwnerOnly(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if user.Role != models.RoleOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "owner access required"})
	}
	return c.Next()
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

func currentUser(c *fiber.Ctx) (*models.User, bool) {
	user, ok := c.Locals(contextUserKey).(*models.User)
	return user, ok
}

func (handler *Handler) setLanguageCookie(c *fiber.Ctx, language string) {
	c.Cookie(&fiber.Cookie{
		Name:     languageCookieName,
		Value:    handler.i18n.NormalizeLanguage(language),
		Path:     "/",
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().AddDate(1, 0, 0),
	})
}

func isOnboardingPath(path string) bool {
	cleanPath := strings.TrimSpace(path)
	return cleanPath == "/onboarding" || strings.HasPrefix(cleanPath, "/onboarding/")
}
