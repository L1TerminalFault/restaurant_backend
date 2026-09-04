package handlers

import (
	"net/http"

	"restaurant-backend/database"
	"restaurant-backend/models"
	"restaurant-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createRestaurantInput struct {
	NameEn             string   `json:"name_en" binding:"required"`
	NameAm             string   `json:"name_am"`
	CustomSubLink      string   `json:"custom_sub_link" binding:"required"`
	Logo               string   `json:"logo"`
	Banner             string   `json:"banner"`
	Slogan             string   `json:"slogan"`
	Images             []string `json:"images"`
	LongerDescription  string   `json:"longer_description"`
	Location           string   `json:"location"`
	Phone              string   `json:"phone"`
	OpenHours          string   `json:"open_hours"`
	AvailableLocations []string `json:"available_locations"`
	FoodSpecifications string   `json:"food_specifications"`
	Email              string   `json:"email"`
	Password           string   `json:"password"`
}

// CreateRestaurant creates a restaurant owned by the authenticated user and
// gives it a default free-tier subscription.
func CreateRestaurant(c *gin.Context) {
	var input createRestaurantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownerID := c.MustGet("user_id").(uuid.UUID)
	role := c.MustGet("role").(models.Role)

	var existing models.Restaurant
	if err := database.DB.Where("custom_sub_link = ?", input.CustomSubLink).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "custom sub-link already taken"})
		return
	}

	if len(input.Images) > models.MaxImagesFree {
		input.Images = input.Images[:models.MaxImagesFree] // trim; upgrade to premium to add more
	}

	tx := database.DB.Begin()

	// If super_admin provides email and password, create an owner user for the restaurant
	if role == models.RoleSuperAdmin && input.Email != "" && input.Password != "" {
		hashed, err := utils.HashPassword(input.Password)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		var ownerUser models.User
		if err := tx.Where("email = ?", input.Email).First(&ownerUser).Error; err != nil {
			// Email not found, create new user
			ownerUser = models.User{
				FullName: input.NameEn + " Owner",
				Email:    input.Email,
				Password: hashed,
				Role:     models.RoleRestaurantOwner,
			}
			if err := tx.Create(&ownerUser).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create restaurant owner"})
				return
			}
		}
		ownerID = ownerUser.ID
	}

	restaurant := models.Restaurant{
		OwnerID:            ownerID,
		NameEn:             input.NameEn,
		NameAm:             input.NameAm,
		CustomSubLink:      input.CustomSubLink,
		Logo:               input.Logo,
		Banner:             input.Banner,
		Slogan:             input.Slogan,
		Images:             input.Images,
		LongerDescription:  input.LongerDescription,
		Location:           input.Location,
		Phone:              input.Phone,
		OpenHours:          input.OpenHours,
		AvailableLocations: input.AvailableLocations,
		FoodSpecifications: input.FoodSpecifications,
	}

	if err := tx.Create(&restaurant).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create restaurant"})
		return
	}

	subscription := models.Subscription{
		RestaurantID: restaurant.ID,
		Plan:         models.PlanFree,
	}
	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create default subscription"})
		return
	}
	tx.Commit()

	restaurant.Subscription = &subscription
	c.JSON(http.StatusCreated, restaurant)
}

// ListRestaurants is public — used for browsing/search.
func ListRestaurants(c *gin.Context) {
	var restaurants []models.Restaurant
	if err := database.DB.Find(&restaurants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch restaurants"})
		return
	}
	c.JSON(http.StatusOK, restaurants)
}

// GetRestaurant is public — returns the full menu tree (categories -> foods -> comments).
func GetRestaurant(c *gin.Context) {
	id := c.Param("id")

	var restaurant models.Restaurant
	query := database.DB.Preload("Categories.Foods.Comments").Preload("Subscription")

	// Allow lookup by ID or by custom sub-link.
	if _, err := uuid.Parse(id); err == nil {
		query = query.Where("id = ?", id)
	} else {
		query = query.Where("custom_sub_link = ?", id)
	}

	if err := query.First(&restaurant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}
	c.JSON(http.StatusOK, restaurant)
}

// UpdateRestaurant lets the owner (or a super_admin) edit restaurant fields.
func UpdateRestaurant(c *gin.Context) {
	id := c.Param("id")

	restaurant, ok := findOwnedRestaurant(c, id)
	if !ok {
		return
	}

	var input createRestaurantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enforce image cap based on active subscription plan.
	maxImages := models.MaxImagesFree
	var sub models.Subscription
	if err := database.DB.Where("restaurant_id = ?", restaurant.ID).First(&sub).Error; err == nil && sub.IsActivePremium() {
		maxImages = models.MaxImagesPremium
	}
	if len(input.Images) > maxImages {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image limit exceeded for current plan", "max_allowed": maxImages})
		return
	}

	restaurant.NameEn = input.NameEn
	restaurant.NameAm = input.NameAm
	restaurant.Logo = input.Logo
	restaurant.Banner = input.Banner
	restaurant.Slogan = input.Slogan
	restaurant.Images = input.Images
	restaurant.LongerDescription = input.LongerDescription
	restaurant.Location = input.Location
	restaurant.Phone = input.Phone
	restaurant.OpenHours = input.OpenHours
	restaurant.AvailableLocations = input.AvailableLocations
	restaurant.FoodSpecifications = input.FoodSpecifications
	// CustomSubLink intentionally not editable here to avoid breaking printed QR codes.

	// Also update the owner's email and password if provided
	if input.Email != "" || input.Password != "" {
		var ownerUser models.User
		if err := database.DB.Where("id = ?", restaurant.OwnerID).First(&ownerUser).Error; err == nil {
			if input.Email != "" {
				ownerUser.Email = input.Email
			}
			if input.Password != "" {
				if hashed, err := utils.HashPassword(input.Password); err == nil {
					ownerUser.Password = hashed
				}
			}
			database.DB.Save(&ownerUser)
		}
	}

	if err := database.DB.Save(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update restaurant"})
		return
	}
	c.JSON(http.StatusOK, restaurant)
}

func DeleteRestaurant(c *gin.Context) {
	id := c.Param("id")
	restaurant, ok := findOwnedRestaurant(c, id)
	if !ok {
		return
	}
	if err := database.DB.Delete(&restaurant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete restaurant"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restaurant deleted"})
}

// findOwnedRestaurant fetches a restaurant by ID and verifies the requester
// is either the owner or a super_admin. Writes an error response and returns
// ok=false if not permitted or not found.
func findOwnedRestaurant(c *gin.Context, id string) (models.Restaurant, bool) {
	var restaurant models.Restaurant
	if err := database.DB.Where("id = ?", id).First(&restaurant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return restaurant, false
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	role := c.MustGet("role").(models.Role)

	if role != models.RoleSuperAdmin && restaurant.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not own this restaurant"})
		return restaurant, false
	}
	return restaurant, true
}
