package models

import (
	"encoding/json"
	"time"
)

type AvailableNode struct {
	ID         string          `gorm:"primaryKey;type:varchar(100)" json:"id"`
	Name       string          `json:"name"`
	Category   string          `json:"category"`
	Icon       string          `json:"icon"`
	Appearance json.RawMessage `gorm:"type:json" json:"appearance"`
	Fields     json.RawMessage `gorm:"type:json" json:"fields"`
	Links      int             `json:"links"`
	BaseNode   bool            `json:"baseNode"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}
