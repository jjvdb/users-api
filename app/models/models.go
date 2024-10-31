package models

import "time"

type User struct {
	ID          uint      `json:"id"`
	Email       string    `json:"email" gorm:"unique"`
	Password    string    `json:"-"` // This field will be omitted from the JSON output
	Name        string    `json:"name"`
	PhotoUrl    *string   `json:"photo_url"`
	IsActivated bool      `json:"is_activated"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Relationship with RefreshToken
	RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID"`
}

type RefreshToken struct {
	ID        uint
	UserID    uint
	User      User `gorm:"constraint:OnDelete:CASCADE;"` // Reference to User with cascade delete
	Device    string
	Location  *string
	Token     string `gorm:"unique"`
	Remember  bool
	Revoked   bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

type ForgotPassword struct {
	ID        uint
	UserID    uint
	User      User   `gorm:"constraint:OnDelete:CASCADE;"`
	Token     string `gorm:"unique"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

type VerifyEmail struct {
	ID        uint
	UserID    uint
	User      User   `gorm:"constraint:OnDelete:CASCADE;"`
	Token     string `gorm:"unique"`
	ExpiresAt time.Time
	CreatedAt time.Time
}
