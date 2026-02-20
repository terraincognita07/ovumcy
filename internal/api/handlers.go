package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/terraincognita07/lume/internal/i18n"
	"gorm.io/gorm"
)

var pageTemplates = []string{
	"login",
	"register",
	"recovery_code",
	"forgot_password",
	"reset_password",
	"onboarding",
	"dashboard",
	"calendar",
	"stats",
	"settings",
	"privacy",
}

var partialTemplateFiles = []string{"day_editor_partial.html"}

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

func newTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(value time.Time, layout string) string {
			if value.IsZero() {
				return ""
			}
			return value.Format(layout)
		},
		"formatFloat": func(value float64) string {
			return fmt.Sprintf("%.1f", value)
		},
		"t": func(messages map[string]string, key string) string {
			return translateMessage(messages, key)
		},
		"phaseLabel": func(messages map[string]string, phase string) string {
			return translateMessage(messages, phaseTranslationKey(phase))
		},
		"phaseIcon": func(phase string) string {
			switch strings.ToLower(strings.TrimSpace(phase)) {
			case "menstrual":
				return "🌙"
			case "follicular":
				return "🌸"
			case "ovulation":
				return "☀️"
			case "fertile":
				return "🌿"
			case "luteal":
				return "🍂"
			default:
				return "✨"
			}
		},
		"flowLabel": func(messages map[string]string, flow string) string {
			return translateMessage(messages, flowTranslationKey(flow))
		},
		"symptomLabel": func(messages map[string]string, name string) string {
			return localizedSymptomName(messages, name)
		},
		"roleLabel": func(messages map[string]string, role string) string {
			return translateMessage(messages, roleTranslationKey(role))
		},
		"isActiveRoute": func(currentPath string, route string) bool {
			path := strings.TrimSpace(currentPath)
			if path == "" {
				return route == "/"
			}
			if route == "/" {
				return path == "/" || strings.HasPrefix(path, "/?")
			}
			return path == route || strings.HasPrefix(path, route+"?") || strings.HasPrefix(path, route+"/")
		},
		"hasSymptom": func(set map[uint]bool, id uint) bool {
			return set[id]
		},
		"toJSON": func(value any) template.JS {
			serialized, err := json.Marshal(value)
			if err != nil {
				return template.JS("null")
			}
			return template.JS(serialized)
		},
	}
}

func parsePageTemplates(templateDir string, funcMap template.FuncMap, pages []string) (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		templatePath := filepath.Join(templateDir, page+".html")
		parsed, err := template.New("base").Funcs(funcMap).ParseFiles(
			filepath.Join(templateDir, "base.html"),
			templatePath,
		)
		if err != nil {
			return nil, fmt.Errorf("parse page template %s: %w", page, err)
		}
		templates[page] = parsed
	}
	return templates, nil
}

func parsePartialTemplates(templateDir string, funcMap template.FuncMap, partialFiles []string) (map[string]*template.Template, error) {
	partials := make(map[string]*template.Template, len(partialFiles))
	for _, partial := range partialFiles {
		name := strings.TrimSuffix(partial, ".html")
		parsed, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join(templateDir, partial))
		if err != nil {
			return nil, fmt.Errorf("parse partial %s: %w", partial, err)
		}
		partials[name] = parsed
	}
	return partials, nil
}
