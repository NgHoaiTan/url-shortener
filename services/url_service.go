package services

import (
	"URL-Shortener-Service/dtos"
	"URL-Shortener-Service/models"
	"URL-Shortener-Service/repositories"
	"URL-Shortener-Service/utils"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	maxRetries      = 5
	shortCodeLength = 8
)

type URLService interface {
	CreateShortURL(req *dtos.CreateShortURLRequest, baseURL string) (*dtos.ShortURLResponse, error)
}

type urlService struct {
	repo repositories.URLRepository
}

func NewURLService(repo repositories.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) CreateShortURL(req *dtos.CreateShortURLRequest, baseURL string) (*dtos.ShortURLResponse, error) {
	existingURL, err := s.repo.FindByOriginalURL(req.OriginalURL)
	if err == nil && existingURL != nil {
		return s.buildResponse(existingURL, baseURL), nil
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		shortCode, err := utils.GenerateShortCode(shortCodeLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}

		newURL := &models.ShortURL{
			OriginalURL: req.OriginalURL,
			ShortCode:   shortCode,
		}

		err = s.repo.Create(newURL)
		if err == nil {
			return s.buildResponse(newURL, baseURL), nil
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			existingURL, fetchErr := s.repo.FindByOriginalURL(req.OriginalURL)
			if fetchErr == nil && existingURL != nil {
				return s.buildResponse(existingURL, baseURL), nil
			}

			if attempt < maxRetries-1 {
				continue
			}
			return nil, errors.New("failed to generate unique short code after max retries")
		}

		return nil, fmt.Errorf("failed to create short URL: %w", err)
	}

	return nil, errors.New("failed to create short URL after max retries")
}

func (s *urlService) buildResponse(url *models.ShortURL, baseURL string) *dtos.ShortURLResponse {
	return &dtos.ShortURLResponse{
		ID:          url.ID,
		OriginalURL: url.OriginalURL,
		ShortURL:    baseURL + "/" + url.ShortCode,
		ClickCount:  url.ClickCount,
		CreatedAt:   url.CreatedAt,
	}
}
