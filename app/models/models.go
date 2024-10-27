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
}
