package api

type credentialsInput struct {
	Email           string `json:"email" form:"email"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
	RememberMe      bool   `json:"remember_me" form:"remember_me"`
}

type dayPayload struct {
	IsPeriod   bool   `json:"is_period"`
	Flow       string `json:"flow"`
	SymptomIDs []uint `json:"symptom_ids"`
	Notes      string `json:"notes"`
}

type symptomPayload struct {
	Name  string `json:"name" form:"name"`
	Icon  string `json:"icon" form:"icon"`
	Color string `json:"color" form:"color"`
}

type forgotPasswordInput struct {
	RecoveryCode string `json:"recovery_code" form:"recovery_code"`
}

type resetPasswordInput struct {
	Token           string `json:"token" form:"token"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
}

type changePasswordInput struct {
	CurrentPassword string `json:"current_password" form:"current_password"`
	NewPassword     string `json:"new_password" form:"new_password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
}

type cycleSettingsInput struct {
	CycleLength  int `json:"cycle_length" form:"cycle_length"`
	PeriodLength int `json:"period_length" form:"period_length"`
}

type deleteAccountInput struct {
	Password string `json:"password" form:"password"`
}
