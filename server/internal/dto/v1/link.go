package v1

import (
	"net/netip"
	"time"
)

type CreateLinkReq struct {
	UserID      int64      `json:"user_id"`
	OriginalURL string     `json:"original_url" binding:"required,url"`
	Code        string     `json:"code"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type GetLinksReq struct {
	UserID       int64      `json:"user_id"`
	OriginalURL  *string    `form:"original_url" json:"original_url"`
	Code         *string    `form:"code" json:"code"`
	Status       *string    `form:"status" json:"status"`
	ExpiresBegin *time.Time `form:"expires_begin" json:"expires_begin"`
	ExpiresEnd   *time.Time `form:"expires_end" json:"expires_end"`
	PageNum      int64      `form:"page_num" json:"page_num"`
	PageSize     int64      `form:"page_size" json:"page_size"`
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
	OriginalURL *string    `json:"original_url" binding:"omitempty,url"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Status      *string    `json:"status"`
}

type DeleteLinkReq struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
}

type GetStatsResponse struct {
	LinkCount int64 `json:"link_count"`
	ViewCount int64 `json:"view_count"`
}

type ClickLinkReq struct {
	Code      string     `json:"code"`
	IP        netip.Addr `json:"ip"`
	UserAgent string     `json:"user_agent"`
	Referer   string     `json:"referer"`
	ClickedAt time.Time  `json:"clicked_at"`
}
