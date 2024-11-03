package models

import "time"

type User struct {
	ID            uint            `json:"id"`
	Email         string          `json:"email" gorm:"unique"`
	Password      string          `json:"-"` // This field will be omitted from the JSON output
	Name          string          `json:"name"`
	PhotoUrl      *string         `json:"photo_url"`
	IsActivated   bool            `json:"is_activated"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	RefreshTokens []RefreshToken  `json:"-" gorm:"foreignKey:UserID"`
	Preferences   UserPreferences `gorm:"foreignKey:UserID"`
}

type UserPreferences struct {
	ID                uint    `json:"id"`
	UserID            uint    `json:"user_id" gorm:"unique"`
	DarkMode          bool    `json:"dark_mode"`
	Theme             *string `json:"theme"`
	Translation       *string `json:"translation"`
	LastReadBook      *string `json:"last_read_book"`
	LastReadChapter   uint    `json:"last_read_chapter"`
	FontSize          uint    `json:"font_size"`
	FontFamily        uint    `json:"font_family"`
	ReferenceAtBottom bool    `json:"reference_at_bottom"`
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

type ReadHistory struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex:unique_read_history"`
	User      User      `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Book      string    `json:"book" gorm:"uniqueIndex:unique_read_history"`
	Chapter   uint      `json:"chapter" gorm:"uniqueIndex:unique_read_history"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
