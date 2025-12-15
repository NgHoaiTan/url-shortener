package repositories

import (
	"URL-Shortener-Service/models"

	"gorm.io/gorm"
)

type URLRepository interface {
	Create(url *models.ShortURL) error
	FindByOriginalURL(originalURL string) (*models.ShortURL, error)
	FindByShortCode(shortCode string) (*models.ShortURL, error)
	IncrementClickCount(shortCode string) error
}

type urlRepository struct {
	db *gorm.DB
}

func NewURLRepository(db *gorm.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(url *models.ShortURL) error {
	return r.db.Create(url).Error
}

func (r *urlRepository) FindByOriginalURL(originalURL string) (*models.ShortURL, error) {
	var url models.ShortURL
	err := r.db.Where("original_url = ?", originalURL).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *urlRepository) FindByShortCode(shortCode string) (*models.ShortURL, error) {
	var url models.ShortURL
	err := r.db.Where("short_code = ?", shortCode).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *urlRepository) IncrementClickCount(shortCode string) error {
	return r.db.Model(&models.ShortURL{}).
		Where("short_code = ?", shortCode).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}
