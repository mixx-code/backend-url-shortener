package models

import "time"

type User struct {
	ID        int       `json:"id" gorm:"primary_key"`
	Name      string    `json:"name"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-"` // Don't expose password in JSON responses
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
