package handlers

import (
	"net/http"
	"time"

	"restaurant-backend/database"
	"restaurant-backend/models"

	"github.com/gin-gonic/gin"
)

// GetSubscription returns the subscription for a restaurant (owner or admin only).
func GetSubscription(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}

	var sub models.Subscription
	if err := database.DB.Where("restaurant_id = ?", restaurantID).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

type upgradeInput struct {
	Plan         models.Plan `json:"plan" binding:"required"` // "free" or "premium"
	DurationDays int         `json:"duration_days"`           // e.g. 30 for monthly, 365 for yearly
}

// UpgradeSubscription is called after payment is confirmed (e.g. by a payment
// webhook or an admin action) to move a restaurant onto a plan.
// NOTE: this endpoint does not process payment itself — wire it up to
// whatever payment provider (Chapa, Telebirr, Stripe, etc.) you use, and call
// this once the payment is verified.
func UpgradeSubscription(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}

	var input upgradeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.DurationDays <= 0 {
		input.DurationDays = 30
	}

	var sub models.Subscription
	if err := database.DB.Where("restaurant_id = ?", restaurantID).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	now := time.Now()
	expires := now.AddDate(0, 0, input.DurationDays)

	sub.Plan = input.Plan
	sub.LastPaymentRenewal = &now
	sub.LastRenewal = &now
	sub.ExpiresAt = &expires

	if err := database.DB.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

// ListSubscriptions is a super_admin-only view across all restaurants —
// useful for the admin dashboard mentioned in your requirements.
func ListSubscriptions(c *gin.Context) {
	var subs []models.Subscription
	if err := database.DB.Find(&subs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch subscriptions"})
		return
	}
	c.JSON(http.StatusOK, subs)
}

func UpdateSubscription(c *gin.Context) {
	restaurantID := c.Param("id")
	var input upgradeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sub models.Subscription
	if err := database.DB.Where("restaurant_id = ?", restaurantID).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found for restaurant"})
		return
	}

	sub.Plan = input.Plan
	if err := database.DB.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, sub)
}
