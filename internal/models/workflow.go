package models

import (
	"time"
	"gorm.io/datatypes"
)

type Workflow struct {
	ID         string         `gorm:"primaryKey;type:varchar(100)" json:"id"`
	Name       string         `json:"name"`
	Categories datatypes.JSON `json:"categories"`
	NodesCount int            `json:"nodes"`
	Status     string         `json:"status"`
	Topology   datatypes.JSON `json:"topology"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}
