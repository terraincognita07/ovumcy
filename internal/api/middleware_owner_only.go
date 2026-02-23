package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

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
