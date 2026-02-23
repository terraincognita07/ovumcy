package api

import "github.com/terraincognita07/ovumcy/internal/models"

func postLoginRedirectPath(user *models.User) string {
	if requiresOnboarding(user) {
		return "/onboarding"
	}
	return "/dashboard"
}

func requiresOnboarding(user *models.User) bool {
	if user == nil {
		return false
	}
	return user.Role == models.RoleOwner && !user.OnboardingCompleted
}
