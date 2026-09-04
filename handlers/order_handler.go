package handlers

import (
	"net/http"

	"restaurant-backend/database"
	"restaurant-backend/models"

	"github.com/gin-gonic/gin"
)

type orderInput struct {
	TableNumber  int              `json:"table_number"`
	TimeLeftMins int              `json:"time_left_mins"`
	IsDelayed    bool             `json:"is_delayed"`
	Status       string           `json:"status"` // Typically "New" when created from client
	Items        []orderItemInput `json:"items" binding:"required"`
}

type orderItemInput struct {
	Name          string `json:"name" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required"`
	Notes         string `json:"notes"`
	IsUnavailable bool   `json:"is_unavailable"` // default false when calling from client
}

// CreateOrder is called by customers using the menu directly (no auth needed)
func CreateOrder(c *gin.Context) {
	restaurantID := c.Param("id")

	// Validate restaurant exists securely via find function or raw lookup
	var restaurant models.Restaurant
	if err := database.DB.Where("id = ? OR custom_sub_link = ?", restaurantID, restaurantID).First(&restaurant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}

	var input orderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare Order models
	order := models.Order{
		RestaurantID: restaurant.ID,
		TableNumber:  input.TableNumber,
		TimeLeftMins: input.TimeLeftMins,
		IsDelayed:    input.IsDelayed,
		Status:       "New",
	}

	for _, itemInput := range input.Items {
		order.Items = append(order.Items, models.OrderItem{
			Name:          itemInput.Name,
			Quantity:      itemInput.Quantity,
			Notes:         itemInput.Notes,
			IsUnavailable: itemInput.IsUnavailable,
		})
	}

	tx := database.DB.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusCreated, order)
}

// ListOrders returns the active orders for the given restaurant (requires Auth)
func ListOrders(c *gin.Context) {
	restaurantID := c.Param("id")

	// Ensure caller legitimately owns this restaurant (using the existing findOwnedRestaurant)
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}

	var orders []models.Order
	// Preload Items and return all orders mapping to the actual restaurant ID
	if err := database.DB.Preload("Items").Where("restaurant_id = ?", restaurantID).Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

type updateOrderStatusInput struct {
	Status string `json:"status" binding:"required"`
}

// UpdateOrderStatus modifies order state (New, Preparing, Ready, Delivered)
func UpdateOrderStatus(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}

	orderID := c.Param("orderId")
	var order models.Order
	if err := database.DB.Where("id = ? AND restaurant_id = ?", orderID, restaurantID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	var input updateOrderStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order.Status = input.Status
	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, order)
}
