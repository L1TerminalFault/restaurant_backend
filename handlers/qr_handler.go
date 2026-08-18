package handlers

import (
	// "fmt"
	"net/http"
	// "os"

	"restaurant-backend/database"
	"restaurant-backend/models"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
	// qrcode "github.com/skip2/go-qrcode"
)

// type createQRInput struct {
// 	TableID string `json:"table_id"` // optional
// }

// CreateQR generates a QR code pointing at the restaurant's public menu page
// (with an optional table_id so orders/scans can be tied to a table).
// func CreateQR(c *gin.Context) {
// 	restaurantID := c.Param("id")
// 	restaurant, ok := findOwnedRestaurant(c, restaurantID)
// 	if !ok {
// 		return
// 	}
//
// 	var input createQRInput
// 	_ = c.ShouldBindJSON(&input) // body is optional — table_id may be omitted
//
// 	rid, _ := uuid.Parse(restaurantID)
// 	qr := models.QRCode{
// 		RestaurantID:   rid,
// 		TableID:        input.TableID,
// 		RestaurantName: restaurant.NameEn,
// 	}
// 	if err := database.DB.Create(&qr).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create qr record"})
// 		return
// 	}
//
// 	// Public URL the QR should resolve to. Adjust FRONTEND_URL / route shape
// 	// to match your actual public menu page.
// 	frontendBase := os.Getenv("FRONTEND_URL")
// 	if frontendBase == "" {
// 		frontendBase = "http://localhost:3000"
// 	}
// 	target := fmt.Sprintf("%s/menu/%s", frontendBase, restaurant.CustomSubLink)
// 	if input.TableID != "" {
// 		target = fmt.Sprintf("%s?table=%s", target, input.TableID)
// 	}
//
// 	if err := os.MkdirAll("./static/qrcodes", 0o755); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare storage"})
// 		return
// 	}
// 	filePath := fmt.Sprintf("./static/qrcodes/%s.png", qr.ID.String())
// 	if err := qrcode.WriteFile(target, qrcode.Medium, 512, filePath); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate qr image"})
// 		return
// 	}
//
// 	qr.ImageURL = fmt.Sprintf("/static/qrcodes/%s.png", qr.ID.String())
// 	database.DB.Save(&qr)
//
// 	c.JSON(http.StatusCreated, qr)
// }

func ListQRCodes(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}
	var qrs []models.QRCode
	if err := database.DB.Where("restaurant_id = ?", restaurantID).Find(&qrs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch qr codes"})
		return
	}
	c.JSON(http.StatusOK, qrs)
}

func AdminListQRCodes(c *gin.Context) {
	var qrcodes []models.QRCode
	if err := database.DB.Find(&qrcodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch qrcodes"})
		return
	}
	c.JSON(http.StatusOK, qrcodes)
}

func DeleteQR(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := findOwnedRestaurant(c, restaurantID); !ok {
		return
	}
	if err := database.DB.Where("id = ? AND restaurant_id = ?", c.Param("qrId"), restaurantID).Delete(&models.QRCode{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete qr code"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "qr code deleted"})
}
