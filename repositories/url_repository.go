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
	FindAll(offset, limit int, sortBy, order, search string) ([]*models.ShortURL, error)
	Count(search string) (int64, error)
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

func (r *urlRepository) FindAll(offset, limit int, sortBy, order, search string) ([]*models.ShortURL, error) {
	var urls []*models.ShortURL
	query := r.db.Model(&models.ShortURL{})

	if search != "" {
		query = query.Where("original_url ILIKE ?", "%"+search+"%")
	}

	orderClause := sortBy + " " + order

	err := query.Order(orderClause).Offset(offset).Limit(limit).Find(&urls).Error
	return urls, err
}

func (r *urlRepository) Count(search string) (int64, error) {
	var count int64
	query := r.db.Model(&models.ShortURL{})

	if search != "" {
		query = query.Where("original_url ILIKE ?", "%"+search+"%")
	}

	err := query.Count(&count).Error
	return count, err
}
