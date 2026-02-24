package api

import (
	"html/template"
)

func newTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate":          formatTemplateDate,
		"formatLocalizedDate": formatTemplateLocalizedDate,
		"formatFloat":         formatTemplateFloat,
		"t":                   templateTranslate,
		"phaseLabel":          templatePhaseLabel,
		"phaseIcon":           templatePhaseIcon,
		"flowLabel":           templateFlowLabel,
		"symptomLabel":        templateSymptomLabel,
		"roleLabel":           templateRoleLabel,
		"userIdentity":        templateUserIdentity,
		"hasDisplayName":      templateHasDisplayName,
		"isActiveRoute":       isActiveTemplateRoute,
		"hasSymptom":          hasTemplateSymptom,
		"toJSON":              templateToJSON,
		"dict":                templateDict,
	}
}
