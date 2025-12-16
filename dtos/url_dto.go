package dtos

import "time"

type CreateShortURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	CustomCode  string `json:"custom_code,omitempty" binding:"omitempty,min=3,max=20,alphanum"`
}

type ShortURLResponse struct {
	ID          uint      `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	ClickCount  uint64    `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type URLInfoResponse struct {
	ID          uint      `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ShortURL    string    `json:"short_url"`
	ClickCount  uint64    `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ListURLsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy   string `form:"sort_by" binding:"omitempty,oneof=created_at click_count original_url"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`
	Search   string `form:"search" binding:"omitempty"`
}

type ListURLsResponse struct {
	URLs       []URLInfoResponse `json:"urls"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
	SortBy     string            `json:"sort_by"`
	Order      string            `json:"order"`
	Search     string            `json:"search,omitempty"`
}
