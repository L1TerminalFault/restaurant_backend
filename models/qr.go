package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QRCode represents one generated QR code for a restaurant, optionally scoped
// to a specific table. Scanning it should resolve to the restaurant's public
// menu page (and pre-select the table if TableID is set).
type QRCode struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;index" json:"restaurant_id"`

	TableID        string `json:"table_id,omitempty"` // optional — omit for a general/menu-only QR
	RestaurantName string `json:"restaurant_name"`    // denormalized snapshot for quick display/printing

	// ImageURL points at the generated QR PNG/SVG once rendered.
	ImageURL string `json:"image_url"`

	CreatedAt time.Time `json:"created_at"`
}

func (q *QRCode) BeforeCreate(tx *gorm.DB) (err error) {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return
}
