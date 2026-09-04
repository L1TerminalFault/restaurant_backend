package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	RestaurantID uuid.UUID   `gorm:"type:uuid;not null;index" json:"restaurant_id"`
	TableNumber  int         `json:"table_number"`
	TimeLeftMins int         `json:"time_left_mins"`
	IsDelayed    bool        `json:"is_delayed"`
	Status       string      `json:"status"` // New, Preparing, Ready, Delivered
	Items        []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;" json:"items"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID       uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	Name          string    `json:"name"`
	Quantity      int       `json:"quantity"`
	Notes         string    `json:"notes"`
	IsUnavailable bool      `json:"is_unavailable"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	// Default status if empty
	if o.Status == "" {
		o.Status = "New"
	}
	return
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return
}
