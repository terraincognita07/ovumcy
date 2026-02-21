package api

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

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
