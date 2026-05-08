package model

import "time"

type Link struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id" db:"user_id"`
	Code        string     `json:"code" db:"code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	ViewCount   int64      `json:"view_count" db:"view_count"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
}
