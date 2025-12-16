package controllers

import (
	"URL-Shortener-Service/dtos"
	"URL-Shortener-Service/services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type URLController struct {
	service services.URLService
	baseURL string
}

func NewURLController(service services.URLService, baseURL string) *URLController {
	return &URLController{
		service: service,
		baseURL: baseURL,
	}
}

func (c *URLController) CreateShortURL(ctx *gin.Context) {
	var req dtos.CreateShortURLRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	result, err := c.service.CreateShortURL(&req, c.baseURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error:   "Failed to create short URL",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    result,
		Message: "Short URL created successfully",
	})
}

func (c *URLController) GetURLInfo(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")

	if shortCode == "" {
		ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
		})
		return
	}

	info, err := c.service.GetURLInfo(shortCode, c.baseURL)
	if err != nil {
		if errors.Is(err, services.ErrShortURLNotFound) {
			ctx.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error:   "Short URL not found",
				Message: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error:   "Failed to get URL info",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    info,
		Message: "URL info retrieved successfully",
	})
}

func (c *URLController) RedirectToOriginalURL(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")

	if shortCode == "" {
		ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
		})
		return
	}

	originalURL, err := c.service.GetOriginalURL(shortCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Error:   "Short URL not found",
			Message: err.Error(),
		})
		return
	}

	ctx.Redirect(http.StatusFound, originalURL)
}

func (c *URLController) ListURLs(ctx *gin.Context) {
	var req dtos.ListURLsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	response, err := c.service.ListURLs(&req, c.baseURL)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error:   "Failed to list URLs",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "URLs retrieved successfully",
	})
}
