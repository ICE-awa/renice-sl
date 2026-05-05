package v1

import "time"

type CreateLinkReq struct {
	UserID      int64      `json:"user_id"`
	OriginalURL string     `json:"original_url" binding:"required,url"`
	Code        string     `json:"code"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type GetLinksReq struct {
	UserID       int64      `json:"user_id"`
	OriginalURL  *string    `json:"original_url"`
	Code         *string    `json:"code"`
	Status       *string    `json:"status"`
	ExpiresBegin *time.Time `json:"expires_begin"`
	ExpiresEnd   *time.Time `json:"expires_end"`
	PageNum      int64      `json:"page_num"`
	PageSize     int64      `json:"page_size"`
}

type LinkItem struct {
	ID          int64      `json:"id"`
	OriginalURL string     `json:"original_url"`
	Code        string     `json:"code"`
	ViewCount   int64      `json:"view_count"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type UpdateLinkReq struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	OriginalURL *string    `json:"original_url" binding:"url"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type DeleteLinkReq struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}
