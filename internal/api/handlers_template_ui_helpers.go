package api

import (
	"fmt"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func templateUserIdentity(user *models.User) string {
	if user == nil {
		return ""
	}
	if displayName := strings.TrimSpace(user.DisplayName); displayName != "" {
		return displayName
	}
	email := strings.TrimSpace(user.Email)
	if email == "" {
		return "account"
	}
	atIndex := strings.Index(email, "@")
	if atIndex <= 0 {
		return email
	}
	localPart := strings.TrimSpace(email[:atIndex])
	if localPart == "" {
		return "account"
	}
	return localPart
}

func isActiveTemplateRoute(currentPath string, route string) bool {
	path := strings.TrimSpace(currentPath)
	if path == "" {
		return route == "/"
	}
	if route == "/" {
		return path == "/" || strings.HasPrefix(path, "/?")
	}
	return path == route || strings.HasPrefix(path, route+"?") || strings.HasPrefix(path, route+"/")
}

func hasTemplateSymptom(set map[uint]bool, id uint) bool {
	return set[id]
}

func templateDict(values ...any) (map[string]any, error) {
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
}
