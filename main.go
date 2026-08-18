package main

import (
	"log"

	"restaurant-backend/config"
	"restaurant-backend/database"
	"restaurant-backend/routes"

	// "restaurant-backend/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Load()
	database.Connect()
	// utils.SeedSuperAdmin()

	if config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	routes.Setup(r)

	log.Printf("server listening on :%s", config.App.Port)
	if err := r.Run(":" + config.App.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
