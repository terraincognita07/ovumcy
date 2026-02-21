package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func buildPrivacyMetaDescription(messages map[string]string) string {
	metaDescription := translateMessage(messages, "meta.description.privacy")
	if metaDescription == "meta.description.privacy" {
		metaDescription = "Lume Privacy Policy - Zero data collection, self-hosted period tracker."
	}
	return metaDescription
}

func buildPrivacyPageData(messages map[string]string, backQuery string, user *models.User) fiber.Map {
	backFallback := "/login"
	data := fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.privacy", "Lume | Privacy Policy"),
		"MetaDescription": buildPrivacyMetaDescription(messages),
	}

	if user != nil {
		data["CurrentUser"] = user
		backFallback = "/dashboard"
	}
	data["BackPath"] = sanitizeRedirectPath(backQuery, backFallback)
	return data
}
