package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"github.com/terraincognita07/lume/internal/models"
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
		"userIdentity": func(user *models.User) string {
			if user == nil {
				return ""
			}
			if displayName := strings.TrimSpace(user.DisplayName); displayName != "" {
				return displayName
			}
			return strings.TrimSpace(user.Email)
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
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires key-value pairs")
			}
			result := make(map[string]any, len(values)/2)
			for index := 0; index < len(values); index += 2 {
				key, ok := values[index].(string)
				if !ok {
					return nil, fmt.Errorf("dict key at index %d is not a string", index)
				}
				result[key] = values[index+1]
			}
			return result, nil
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
		parsed, err := template.New(name).Funcs(funcMap).ParseFiles(
			filepath.Join(templateDir, "base.html"),
			filepath.Join(templateDir, partial),
		)
		if err != nil {
			return nil, fmt.Errorf("parse partial %s: %w", partial, err)
		}
		partials[name] = parsed
	}
	return partials, nil
}
