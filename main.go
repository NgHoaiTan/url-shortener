package main

import (
	"URL-Shortener-Service/config"
	"URL-Shortener-Service/controllers"
	"URL-Shortener-Service/repositories"
	"URL-Shortener-Service/routes"
	"URL-Shortener-Service/services"
	"log"
	"os"
)

func main() {
	config.LoadEnv()

	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}
	log.Println("Database connected successfully")

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	urlRepo := repositories.NewURLRepository(db)
	urlService := services.NewURLService(urlRepo, baseURL)
	urlController := controllers.NewURLController(urlService, baseURL)

	router := routes.SetupRoutes(urlController)

	log.Println("Server is running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
