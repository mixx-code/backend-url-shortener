package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type URL struct {
	ID          int       `json:"id" gorm:"primary_key"`
	OriginalURL string    `json:"original_url" gorm:"not null"`
	ShortCode   string    `json:"short_code" gorm:"unique;not null"`
	ClickCount  int       `json:"click_count" gorm:"default:0"`
	UserID      int       `json:"user_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func GenerateShortCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8]
}
