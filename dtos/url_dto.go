package dtos

import "time"

type CreateShortURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	CustomCode  string `json:"custom_code,omitempty" binding:"omitempty,min=3,max=20,alphanum"`
}

type ShortURLResponse struct {
	ID          uint      `json:"id"`
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
