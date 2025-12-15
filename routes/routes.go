package routes

import (
	"URL-Shortener-Service/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(urlController *controllers.URLController) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/shorten", urlController.CreateShortURL)
	}

	return router
}
