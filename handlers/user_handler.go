package handlers

import (
	"net/http"

	"restaurant-backend/database"
	"restaurant-backend/models"
	"restaurant-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var user models.User
	if err := database.DB.Preload("Restaurants").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type updateProfileInput struct {
	FullName     string `json:"full_name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"` // optional or handled on change
	ProfileImage string `json:"profile_image"`
}

func UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input updateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.FullName = input.FullName
	user.Email = input.Email
	user.ProfileImage = input.ProfileImage

	if input.Password != "" {
		hashed, err := utils.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = hashed
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func ListUsers(c *gin.Context) {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

type adminCreateUserInput struct {
	FullName     string      `json:"full_name" binding:"required"`
	Email        string      `json:"email" binding:"required,email"`
	Password     string      `json:"password" binding:"required,min=6"`
	ProfileImage string      `json:"profile_image"`
	Role         models.Role `json:"role" binding:"required"`
}

func CreateUser(c *gin.Context) {
	var input adminCreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if err := database.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		FullName:     input.FullName,
		Email:        input.Email,
		Password:     hashed,
		ProfileImage: input.ProfileImage,
		Role:         input.Role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

type adminUpdateUserInput struct {
	FullName     string      `json:"full_name" binding:"required"`
	Email        string      `json:"email" binding:"required,email"`
	Password     string      `json:"password"`
	ProfileImage string      `json:"profile_image"`
	Role         models.Role `json:"role" binding:"required"`
}

func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input adminUpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.FullName = input.FullName
	user.Email = input.Email
	user.ProfileImage = input.ProfileImage
	user.Role = input.Role

	if input.Password != "" {
		hashed, err := utils.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = hashed
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type updateUserRoleInput struct {
	Role models.Role `json:"role" binding:"required"`
}

func UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input updateUserRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.Role = input.Role
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
