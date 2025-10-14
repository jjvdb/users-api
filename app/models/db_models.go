package models

import (
	"strings"
	"time"
)

type User struct {
	ID            uint           `json:"id"`
	Email         string         `json:"email" gorm:"unique;not null"`
	Username      string         `json:"username" gorm:"unique;not null"`
	Password      string         `json:"-" gorm:"not null"` // This field will be omitted from the JSON output
	Name          string         `json:"name"`
	PhotoUrl      string         `json:"photo_url"`
	IsActivated   bool           `json:"is_activated"`
	Bio           string         `json:"bio"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	RefreshTokens []RefreshToken `json:"-" gorm:"foreignKey:UserID"`
	Preference    UserPreference `json:"preference" gorm:"foreignKey:UserID"`
}

func (user *User) Trim() {
	user.Email = strings.TrimSpace(user.Email)
	user.Username = strings.TrimSpace(user.Username)
	user.Name = strings.TrimSpace(user.Name)
	user.Bio = strings.TrimSpace(user.Bio)
}

type UserPreference struct {
	ID                      uint    `json:"id"`
	UserID                  uint    `json:"user_id" gorm:"unique"`
	DarkMode                bool    `json:"dark_mode"`
	Theme                   *string `json:"theme"`
	PreferredTranslation    *string `json:"preferred_translation"`
	FontSize                int     `json:"font_size"`
	FontFamily              uint    `json:"font_family"`
	MarginSize              int     `json:"margin_size"`
	ReferenceAtBottom       bool    `json:"reference_at_bottom"`
	CopyIncludesUrl         bool    `json:"copy_includes_url"`
	MarkAsReadAutomatically bool    `json:"mark_as_read_automatically"`
	UseAbbreviationsForNav  bool    `json:"use_abrbeviations_for_nav"`
}

type RefreshToken struct {
	ID        uint
	UserID    uint
	User      User `gorm:"constraint:OnDelete:CASCADE;"` // Reference to User with cascade delete
	Device    *string
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
	Book      uint      `json:"book" gorm:"uniqueIndex:unique_read_history"`
	Chapter   uint      `json:"chapter" gorm:"uniqueIndex:unique_read_history"`
	CreatedAt time.Time `json:"created_at"`
}

type Bookmark struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id" gorm:"uniqueIndex:unique_bookmark"`
	User          User      `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Book          string    `json:"book" gorm:"uniqueIndex:unique_bookmark"`
	ChapterNumber uint      `json:"chapter_number" gorm:"uniqueIndex:unique_bookmark"`
	VerseNumber   uint      `json:"verse_number" gorm:"uniqueIndex:unique_bookmark"`
	CreatedAt     time.Time `json:"created_at"`
}

type ParallelTranslations struct {
	ID           uint
	UserID       uint      `json:"user_id" gorm:"uniqueIndex:uniquePT"`
	User         User      `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Translation1 string    `json:"translation_1" gorm:"uniqueIndex:uniquePT"`
	Translation2 string    `json:"translation_2" gorm:"uniqueIndex:uniquePT"`
	CreatedAt    time.Time `json:"created_at"`
}

type Note struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	User          User      `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Book          string    `json:"book"`
	ChapterNumber uint      `json:"chapter_number"`
	VerseNumber   uint      `json:"verse_number"`
	Note          string    `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
