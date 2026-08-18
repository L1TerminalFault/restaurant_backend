package utils

import (
	"log"

	"restaurant-backend/database"
	"restaurant-backend/models"
)

func SeedSuperAdmin() {
	var count int64
	// Check if a super admin already exists
	if err := database.DB.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count).Error; err != nil {
		log.Printf("failed to check for existing super admin: %v", err)
		return
	}

	if count > 0 {
		return // Super admin already exists, skip creation
	}

	// Hash the password
	hashedPassword, err := HashPassword("SuperAdminSecurePassword123!")
	if err != nil {
		log.Fatalf("failed to hash super admin password: %v", err)
	}

	superAdmin := models.User{
		FullName:     "System Super Admin",
		Email:        "admin@platform.com",
		Password:     hashedPassword,
		ProfileImage: "",
		Role:         models.RoleSuperAdmin,
	}

	// Create the super admin in the database
	if err := database.DB.Create(&superAdmin).Error; err != nil {
		log.Fatalf("failed to create super admin: %v", err)
	}

	log.Println("Super admin successfully created: admin@platform.com / SuperAdminSecurePassword123!")
}
