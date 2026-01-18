package models

import (
	"time"
)

type Click struct {
	ID        int       `json:"id" gorm:"primary_key"`
	URLID     int       `json:"url_id" gorm:"not null"`
	ClickedAt time.Time `json:"clicked_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
