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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	urlRepo := repositories.NewURLRepository(db)
	urlService := services.NewURLService(urlRepo, baseURL)
	urlController := controllers.NewURLController(urlService, baseURL)

	router := routes.SetupRoutes(urlController)

	log.Printf("Server is running on port %s", port)
	log.Printf("Access the service at: %s", baseURL)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
