package models

import "time"

// Provider constants — extend for Google, GitHub etc.
const (
	ProviderLocal  = "local"
	ProviderGoogle = "google"
)

type User struct {
	ID           string    `gorm:"primaryKey;type:varchar(100)" json:"id"`
	Name         string    `json:"name"`
	Email        string    `gorm:"uniqueIndex;type:varchar(255)" json:"email"`
	Phone        string    `gorm:"type:varchar(30)" json:"phone"`
	PasswordHash string    `gorm:"type:varchar(255)" json:"-"` // never serialised
	Provider     string    `gorm:"type:varchar(30);default:'local'" json:"provider"`
	ProviderID   string    `gorm:"type:varchar(255)" json:"providerId,omitempty"`
	Avatar       string    `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	IsActive     bool      `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// RefreshToken stores issued refresh tokens for revocation support.
type RefreshToken struct {
	ID        string    `gorm:"primaryKey;type:varchar(100)"`
	UserID    string    `gorm:"type:varchar(100);index"`
	TokenHash string    `gorm:"type:varchar(255);uniqueIndex"` // SHA-256 of the raw token
	ExpiresAt time.Time
	CreatedAt time.Time
}
