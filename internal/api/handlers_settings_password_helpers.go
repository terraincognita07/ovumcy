package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func parseChangePasswordInput(c *fiber.Ctx) (changePasswordInput, string) {
	input := changePasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return changePasswordInput{}, "invalid settings input"
	}

	input.CurrentPassword = strings.TrimSpace(input.CurrentPassword)
	input.NewPassword = strings.TrimSpace(input.NewPassword)
	input.ConfirmPassword = strings.TrimSpace(input.ConfirmPassword)
	return input, ""
}
