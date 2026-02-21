package api

import (
	"errors"
	"time"

	"github.com/terraincognita07/lume/internal/i18n"
	"gorm.io/gorm"
)

func NewHandler(database *gorm.DB, secret string, templateDir string, location *time.Location, i18nManager *i18n.Manager, cookieSecure bool) (*Handler, error) {
	if location == nil {
		location = time.Local
	}
	if i18nManager == nil {
		return nil, errors.New("i18n manager is required")
	}

	funcMap := newTemplateFuncMap()

	templates, err := parsePageTemplates(templateDir, funcMap, pageTemplates)
	if err != nil {
		return nil, err
	}

	partials, err := parsePartialTemplates(templateDir, funcMap, partialTemplateFiles)
	if err != nil {
		return nil, err
	}

	return &Handler{
		db:              database,
		secretKey:       []byte(secret),
		location:        location,
		cookieSecure:    cookieSecure,
		lutealPhaseDays: 14,
		i18n:            i18nManager,
		templates:       templates,
		partials:        partials,
		recoveryLimiter: newAttemptLimiter(),
	}, nil
}
