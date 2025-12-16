package services

import (
	"URL-Shortener-Service/dtos"
	"URL-Shortener-Service/models"
	"URL-Shortener-Service/repositories"
	"URL-Shortener-Service/utils"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
)

const (
	maxRetries      = 5
	shortCodeLength = 8
)

type URLService interface {
	CreateShortURL(req *dtos.CreateShortURLRequest, baseURL string) (*dtos.ShortURLResponse, error)
	GetOriginalURL(shortCode string) (string, error)
	GetURLInfo(shortCode string, baseURL string) (*dtos.URLInfoResponse, error)
	ListURLs(req *dtos.ListURLsRequest, baseURL string) (*dtos.ListURLsResponse, error)
}

type urlService struct {
	repo    repositories.URLRepository
	baseURL string
}

func NewURLService(repo repositories.URLRepository, baseURL string) URLService {
	return &urlService{
		repo:    repo,
		baseURL: baseURL,
	}
}

var ErrShortURLNotFound = errors.New("short URL not found")

func (s *urlService) CreateShortURL(req *dtos.CreateShortURLRequest, baseURL string) (*dtos.ShortURLResponse, error) {
	if err := utils.ValidateURL(req.OriginalURL, s.baseURL); err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

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

func (s *urlService) GetOriginalURL(shortCode string) (string, error) {
	url, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("short URL not found")
		}
		return "", fmt.Errorf("failed to find short URL: %w", err)
	}

	go func() {
		if err := s.repo.IncrementClickCount(shortCode); err != nil {
			log.Printf("failed to increment click count: %v", err)
		}
	}()

	return url.OriginalURL, nil
}

func (s *urlService) GetURLInfo(shortCode string, baseURL string) (*dtos.URLInfoResponse, error) {

	url, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShortURLNotFound
		}
		return nil, fmt.Errorf("failed to find short URL: %w", err)
	}

	return &dtos.URLInfoResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		ShortURL:    baseURL + "/" + url.ShortCode,
		ClickCount:  url.ClickCount,
		CreatedAt:   url.CreatedAt,
		UpdatedAt:   url.UpdatedAt,
	}, nil
}

func (s *urlService) buildResponse(url *models.ShortURL, baseURL string) *dtos.ShortURLResponse {
	return &dtos.ShortURLResponse{
		ID:          url.ID,
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		ShortURL:    baseURL + "/" + url.ShortCode,
		ClickCount:  url.ClickCount,
		CreatedAt:   url.CreatedAt,
	}
}

func (s *urlService) ListURLs(req *dtos.ListURLsRequest, baseURL string) (*dtos.ListURLsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	allowedSort := map[string]bool{
		"created_at":  true,
		"updated_at":  true,
		"click_count": true,
	}

	if req.SortBy == "" || !allowedSort[req.SortBy] {
		req.SortBy = "created_at"
	}

	if req.Order != "asc" && req.Order != "desc" {
		req.Order = "desc"
	}

	offset := (req.Page - 1) * req.PageSize

	urls, err := s.repo.FindAll(offset, req.PageSize, req.SortBy, req.Order, req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URLs: %w", err)
	}

	totalCount, err := s.repo.Count(req.Search)
	if err != nil {
		return nil, fmt.Errorf("failed to count URLs: %w", err)
	}

	urlInfos := make([]dtos.URLInfoResponse, len(urls))
	for i, url := range urls {
		urlInfos[i] = dtos.URLInfoResponse{
			ID:          url.ID,
			ShortCode:   url.ShortCode,
			OriginalURL: url.OriginalURL,
			ShortURL:    baseURL + "/" + url.ShortCode,
			ClickCount:  url.ClickCount,
			CreatedAt:   url.CreatedAt,
			UpdatedAt:   url.UpdatedAt,
		}
	}

	totalPages := int(totalCount) / req.PageSize
	if int(totalCount)%req.PageSize > 0 {
		totalPages++
	}

	return &dtos.ListURLsResponse{
		URLs:       urlInfos,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
		SortBy:     req.SortBy,
		Order:      req.Order,
		Search:     req.Search,
	}, nil
}
