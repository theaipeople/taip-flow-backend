package models

import (
	"encoding/json"
	"time"
)

type Workflow struct {
	ID         string          `gorm:"primaryKey;type:varchar(100)" json:"id"`
	Name       string          `json:"name"`
	Categories json.RawMessage `gorm:"type:json" json:"categories"`
	NodesCount int             `json:"nodes"`
	Status     string          `json:"status"`
	Topology   json.RawMessage `gorm:"type:json" json:"topology"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}
