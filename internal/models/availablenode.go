package models

import (
	"time"
	"gorm.io/datatypes"
)

type AvailableNode struct {
	ID         string         `gorm:"primaryKey;type:varchar(100)" json:"id"`
	Name       string         `json:"name"`
	Category   string         `json:"category"`
	Icon       string         `json:"icon"`
	Appearance datatypes.JSON `json:"appearance"`
	Fields     datatypes.JSON `json:"fields"`
	Links      int            `json:"links"`
	BaseNode   bool           `json:"baseNode"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}
