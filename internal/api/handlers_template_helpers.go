package api

import (
	"html/template"
)

func newTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate":    formatTemplateDate,
		"formatFloat":   formatTemplateFloat,
		"t":             templateTranslate,
		"phaseLabel":    templatePhaseLabel,
		"phaseIcon":     templatePhaseIcon,
		"flowLabel":     templateFlowLabel,
		"symptomLabel":  templateSymptomLabel,
		"roleLabel":     templateRoleLabel,
		"userIdentity":  templateUserIdentity,
		"isActiveRoute": isActiveTemplateRoute,
		"hasSymptom":    hasTemplateSymptom,
		"toJSON":        templateToJSON,
		"dict":          templateDict,
	}
}
