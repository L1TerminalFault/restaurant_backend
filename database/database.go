// Package database
package database

import (
	"log"

	"restaurant-backend/config"
	"restaurant-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect opens the DB connection and runs auto-migrations.
func Connect() {
	db, err := gorm.Open(postgres.Open(config.App.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	DB = db
	//
	// _ = DB.Migrator().DropTable(
	// 	&models.Comment{},
	// 	&models.Food{},
	// 	&models.Category{},
	// 	&models.QRCode{},
	// 	&models.Subscription{},
	// 	&models.Restaurant{},
	// 	&models.User{},
	// )

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Restaurant{},
		&models.Category{},
		&models.Food{},
		&models.Comment{},
		&models.Subscription{},
		&models.QRCode{},
	); err != nil {
		log.Fatalf("auto-migration failed: %v", err)
	}

	log.Println("database connected and migrated")
}
