// Package handlers
package handlers

import (
	"net/http"

	"restaurant-backend/database"
	"restaurant-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------- Categories ----------

type categoryInput struct {
	Name string `json:"name" binding:"required"`
}

func CreateCategory(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}

	var input categoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rid, _ := uuid.Parse(restaurantID)
	category := models.Category{RestaurantID: rid, Name: input.Name}
	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}
	c.JSON(http.StatusCreated, category)
}

func ListCategories(c *gin.Context) {
	restaurantID := c.Param("id")

	// If called by an admin globally without a restaurant scope, return all categories
	if restaurantID == "" {
		var categories []models.Category
		if err := database.DB.Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
			return
		}
		c.JSON(http.StatusOK, categories)
		return
	}

	// Otherwise, scope to the specific restaurant
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}

	var categories []models.Category
	if err := database.DB.Where("restaurant_id = ?", restaurantID).Preload("Foods").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func UpdateCategory(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}

	var category models.Category
	if err := database.DB.Where("id = ? AND restaurant_id = ?", c.Param("categoryId"), restaurantID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	var input categoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = input.Name
	database.DB.Save(&category)
	c.JSON(http.StatusOK, category)
}

func DeleteCategory(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}
	if err := database.DB.Where("id = ? AND restaurant_id = ?", c.Param("categoryId"), restaurantID).Delete(&models.Category{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

// ---------- Foods ----------

func ListFoods(c *gin.Context) {
	var foods []models.Food
	if err := database.DB.Find(&foods).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch foods"})
		return
	}
	c.JSON(http.StatusOK, foods)
}

func GetFood(c *gin.Context) {
	foodID := c.Param("foodId")

	var food models.Food

	if err := database.DB.Where("id = ?", foodID).First(&food).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "food not found"})
		return
	}

	c.JSON(http.StatusOK, food)
}

// verifyCategoryOwnership checks the category belongs to a restaurant the caller owns.
func verifyCategoryOwnership(c *gin.Context, restaurantID, categoryID string) (models.Category, bool) {
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return models.Category{}, false
	}
	var category models.Category
	if err := database.DB.Where("id = ? AND restaurant_id = ?", categoryID, restaurantID).First(&category).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return category, false
	}
	return category, true
}

type foodInput struct {
	Name            string   `json:"name" binding:"required"`
	Price           float64  `json:"price" binding:"required"`
	Rating          uint     `json:"rating"`
	Tag             string   `json:"tag"`
	Description     string   `json:"description"`
	Ingredients     []string `json:"ingredients"`
	Pic             string   `json:"pic"`
	PrepTimeMinutes int      `json:"prep_time_minutes"`
	Calories        int      `json:"calories"`
	BestPairings    []string `json:"best_pairings"`
	IsSpicy         bool     `json:"spicy_or_not"`
}

func CreateFood(c *gin.Context) {
	category, ok := verifyCategoryOwnership(c, c.Param("id"), c.Param("categoryId"))
	if !ok {
		return
	}

	var input foodInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	food := models.Food{
		CategoryID:      category.ID,
		Name:            input.Name,
		Price:           input.Price,
		Tag:             input.Tag,
		Description:     input.Description,
		Ingredients:     input.Ingredients,
		Pic:             input.Pic,
		PrepTimeMinutes: input.PrepTimeMinutes,
		Calories:        input.Calories,
		BestPairings:    input.BestPairings,
		IsSpicy:         input.IsSpicy,
	}
	if err := database.DB.Create(&food).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create food"})
		return
	}
	c.JSON(http.StatusCreated, food)
}

func UpdateFood(c *gin.Context) {
	if _, ok := verifyCategoryOwnership(c, c.Param("id"), c.Param("categoryId")); !ok {
		return
	}

	var food models.Food
	if err := database.DB.Where("id = ? AND category_id = ?", c.Param("foodId"), c.Param("categoryId")).First(&food).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "food not found"})
		return
	}

	var input foodInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	food.Name = input.Name
	food.Price = input.Price
	food.Tag = input.Tag
	food.Description = input.Description
	food.Ingredients = input.Ingredients
	food.Pic = input.Pic
	food.PrepTimeMinutes = input.PrepTimeMinutes
	food.Calories = input.Calories
	food.BestPairings = input.BestPairings
	food.IsSpicy = input.IsSpicy

	database.DB.Save(&food)
	c.JSON(http.StatusOK, food)
}

func DeleteFood(c *gin.Context) {
	if _, ok := verifyCategoryOwnership(c, c.Param("id"), c.Param("categoryId")); !ok {
		return
	}
	if err := database.DB.Where("id = ? AND category_id = ?", c.Param("foodId"), c.Param("categoryId")).Delete(&models.Food{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete food"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "food deleted"})
}

// ---------- Comments (public — diners rate/comment without an account) ----------

func ListComments(c *gin.Context) {
	var comments []models.Comment
	if err := database.DB.Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch comments"})
		return
	}
	c.JSON(http.StatusOK, comments)
}

type commentInput struct {
	AuthorName string `json:"author_name" binding:"required"`
	Rating     int    `json:"rating" binding:"required,min=1,max=5"`
	Message    string `json:"message"`
}

func AddComment(c *gin.Context) {
	foodID := c.Param("foodId")

	var food models.Food
	if err := database.DB.Where("id = ?", foodID).First(&food).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "food not found"})
		return
	}

	var input commentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fid, _ := uuid.Parse(foodID)
	comment := models.Comment{
		FoodID:     fid,
		AuthorName: input.AuthorName,
		Rating:     input.Rating,
		Message:    input.Message,
	}

	tx := database.DB.Begin()
	if err := tx.Create(&comment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add comment"})
		return
	}

	food.RatingAmount += float64(input.Rating)
	food.RatingCount += 1
	if err := tx.Save(&food).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update food rating"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusCreated, comment)
}

func DeleteComment(c *gin.Context) {
	commentID := c.Param("commentId")
	var comment models.Comment
	if err := database.DB.Where("id = ?", commentID).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})
}
