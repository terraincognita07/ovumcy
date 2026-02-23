package api

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) GetSymptoms(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if user.Role != models.RoleOwner {
		return apiError(c, fiber.StatusForbidden, "owner access required")
	}

	symptoms, err := handler.fetchSymptoms(user.ID)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch symptoms")
	}
	return c.JSON(symptoms)
}

func (handler *Handler) CreateSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	payload := symptomPayload{}
	if err := c.BodyParser(&payload); err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid payload")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	payload.Icon = strings.TrimSpace(payload.Icon)
	payload.Color = strings.TrimSpace(payload.Color)

	if payload.Name == "" || len(payload.Name) > 80 {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom name")
	}
	if payload.Icon == "" {
		payload.Icon = "✨"
	}
	if !hexColorRegex.MatchString(payload.Color) {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom color")
	}

	symptom := models.SymptomType{
		UserID:    user.ID,
		Name:      payload.Name,
		Icon:      payload.Icon,
		Color:     payload.Color,
		IsBuiltin: false,
	}

	if err := handler.db.Create(&symptom).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create symptom")
	}
	return c.Status(fiber.StatusCreated).JSON(symptom)
}

func (handler *Handler) DeleteSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom id")
	}

	var symptom models.SymptomType
	if err := handler.db.Where("id = ? AND user_id = ?", id, user.ID).First(&symptom).Error; err != nil {
		return apiError(c, fiber.StatusNotFound, "symptom not found")
	}
	if symptom.IsBuiltin {
		return apiError(c, fiber.StatusBadRequest, "built-in symptom cannot be deleted")
	}

	if err := handler.db.Delete(&symptom).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete symptom")
	}

	if err := handler.removeSymptomFromLogs(user.ID, symptom.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to clean symptom logs")
	}

	return c.JSON(fiber.Map{"ok": true})
}
