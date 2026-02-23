package api

import "time"

type onboardingStep1Input struct {
	LastPeriodStart string `json:"last_period_start" form:"last_period_start"`
	PeriodStatus    string `json:"period_status" form:"period_status"`
	PeriodEnd       string `json:"period_end" form:"period_end"`
}

type onboardingStep1Values struct {
	Start                time.Time
	InferredPeriodLength int
}

type onboardingStep2Input struct {
	CycleLength    int  `json:"cycle_length" form:"cycle_length"`
	PeriodLength   int  `json:"period_length" form:"period_length"`
	AutoPeriodFill bool `json:"auto_period_fill" form:"auto_period_fill"`
}
