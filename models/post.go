package models

import "time"

type Post struct {
	ID        int       `json:"id" gorm:"primary_key"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
