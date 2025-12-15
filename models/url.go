package models

import (
	"gorm.io/gorm"
)

type ShortURL struct {
	gorm.Model
	OriginalURL string `gorm:"type:text;uniqueIndex;not null" json:"original_url"`
	ShortCode   string `gorm:"size:20;uniqueIndex;not null" json:"short_code"`
	ClickCount  uint64 `gorm:"default:0" json:"click_count"`
}

func (ShortURL) TableName() string {
	return "short_urls"
}
