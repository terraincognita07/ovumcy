package api

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func parseCredentials(c *fiber.Ctx) (credentialsInput, error) {
	credentials := credentialsInput{}
	if err := c.BodyParser(&credentials); err != nil {
		return credentialsInput{}, err
	}

	credentials.Email = strings.ToLower(strings.TrimSpace(credentials.Email))
	credentials.Password = strings.TrimSpace(credentials.Password)
	credentials.ConfirmPassword = strings.TrimSpace(credentials.ConfirmPassword)
	credentials.RememberMe = credentials.RememberMe || parseBoolValue(c.FormValue("remember_me"))

	if credentials.Email == "" || credentials.Password == "" {
		return credentialsInput{}, errors.New("missing credentials")
	}
	if _, err := mail.ParseAddress(credentials.Email); err != nil {
		return credentialsInput{}, errors.New("invalid email")
	}

	return credentials, nil
}
