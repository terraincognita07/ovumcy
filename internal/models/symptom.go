package models

type SymptomType struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	Name      string `gorm:"not null"`
	Icon      string `gorm:"not null"`
	Color     string `gorm:"not null"`
	IsBuiltin bool   `gorm:"not null;default:false"`
}

type BuiltinSymptom struct {
	Name  string
	Icon  string
	Color string
}

func DefaultBuiltinSymptoms() []BuiltinSymptom {
	return []BuiltinSymptom{
		{Name: "Cramps", Icon: "🩸", Color: "#FF4444"},
		{Name: "Headache", Icon: "🤕", Color: "#FFA500"},
		{Name: "Mood swings", Icon: "😢", Color: "#9B59B6"},
		{Name: "Bloating", Icon: "🎈", Color: "#3498DB"},
		{Name: "Fatigue", Icon: "😴", Color: "#95A5A6"},
		{Name: "Breast tenderness", Icon: "💔", Color: "#E91E63"},
		{Name: "Acne", Icon: "🔴", Color: "#E74C3C"},
		{Name: "Back pain", Icon: "🦴", Color: "#8E6E53"},
		{Name: "Nausea", Icon: "🤢", Color: "#7CB342"},
		{Name: "Spotting", Icon: "🩹", Color: "#C55A7A"},
		{Name: "Irritability", Icon: "😤", Color: "#FF7043"},
		{Name: "Insomnia", Icon: "🌙", Color: "#5C6BC0"},
		{Name: "Food cravings", Icon: "🍫", Color: "#A1887F"},
		{Name: "Diarrhea", Icon: "🚽", Color: "#26A69A"},
		{Name: "Constipation", Icon: "🪨", Color: "#8D6E63"},
	}
}
