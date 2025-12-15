package controllers

import (
	"URL-Shortener-Service/dtos"
	"URL-Shortener-Service/services"
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
