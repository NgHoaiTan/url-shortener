package main

import (
	"URL-Shortener-Service/config"

	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}
	_ = db
	log.Println("Database connected successfully")

	router := gin.Default()
	log.Println("Server is running on port 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
