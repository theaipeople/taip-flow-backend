package models

import "gorm.io/gorm"

type Agent struct {
	gorm.Model
	ID           string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name         string `gorm:"type:varchar(255);not null" json:"name"`
	Category     string `gorm:"type:varchar(100)" json:"category"`
	LLM          string `gorm:"type:varchar(100)" json:"llm"`
	SystemPrompt string `gorm:"type:text" json:"systemPrompt"`
	Temperature  float64 `gorm:"default:0.7" json:"temperature"`
	MaxTokens    int     `gorm:"default:4096" json:"maxTokens"`
	Tools        string  `gorm:"type:text" json:"tools"` // JSON array stored as string
}
