// Package models
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;index" json:"restaurant_id"`
	Name         string    `gorm:"not null" json:"name"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Foods []Food `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;" json:"foods,omitempty"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

type Food struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`

	Name  string  `gorm:"not null" json:"name"`
	Price float64 `gorm:"not null" json:"price"`

	RatingAmount float64 `gorm:"default:0" json:"rating_amount"` // sum of all ratings
	RatingCount  int     `gorm:"default:0" json:"rating_count"`  // number of ratings

	Tag         string      `json:"tag"`
	Description string      `json:"description"`
	Ingredients StringArray `gorm:"type:jsonb" json:"ingredients,omitempty"` // optional
	Pic         string      `json:"pic"`

	PrepTimeMinutes int `json:"prep_time_minutes"`
	Calories        int `json:"calories"`

	// Names (or IDs) of foods that pair well with this one.
	BestPairings StringArray `gorm:"type:jsonb" json:"best_pairings,omitempty"`

	IsSpicy bool `gorm:"default:false" json:"is_spicy"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Comments []Comment `gorm:"foreignKey:FoodID;constraint:OnDelete:CASCADE;" json:"comments,omitempty"`
}

func (f *Food) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}

// AverageRating is a convenience accessor, not stored directly.
func (f *Food) AverageRating() float64 {
	if f.RatingCount == 0 {
		return 0
	}
	return f.RatingAmount / float64(f.RatingCount)
}

type Comment struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FoodID uuid.UUID `gorm:"type:uuid;not null;index" json:"food_id"`

	AuthorName string    `json:"author_name"` // diner name, no account required
	Rating     int       `gorm:"check:rating >= 1 AND rating <= 5" json:"rating"`
	Message    string    `json:"message"`
	Time       time.Time `json:"time"`

	CreatedAt time.Time `json:"created_at"`
}

func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Time.IsZero() {
		c.Time = time.Now()
	}
	return
}
