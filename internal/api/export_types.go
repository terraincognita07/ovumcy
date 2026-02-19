package api

type exportSymptomFlags struct {
	Cramps           bool
	Headache         bool
	Acne             bool
	Mood             bool
	Bloating         bool
	Fatigue          bool
	BreastTenderness bool
	BackPain         bool
	Nausea           bool
	Spotting         bool
	Irritability     bool
	Insomnia         bool
	FoodCravings     bool
	Diarrhea         bool
	Constipation     bool
}

type exportJSONSymptomFlags struct {
	Cramps           bool `json:"cramps"`
	Headache         bool `json:"headache"`
	Acne             bool `json:"acne"`
	Mood             bool `json:"mood"`
	Bloating         bool `json:"bloating"`
	Fatigue          bool `json:"fatigue"`
	BreastTenderness bool `json:"breast_tenderness"`
	BackPain         bool `json:"back_pain"`
	Nausea           bool `json:"nausea"`
	Spotting         bool `json:"spotting"`
	Irritability     bool `json:"irritability"`
	Insomnia         bool `json:"insomnia"`
	FoodCravings     bool `json:"food_cravings"`
	Diarrhea         bool `json:"diarrhea"`
	Constipation     bool `json:"constipation"`
}

type exportJSONEntry struct {
	Date          string                 `json:"date"`
	Period        bool                   `json:"period"`
	Flow          string                 `json:"flow"`
	Symptoms      exportJSONSymptomFlags `json:"symptoms"`
	OtherSymptoms []string               `json:"other_symptoms"`
	Notes         string                 `json:"notes"`
}
