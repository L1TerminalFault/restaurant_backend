package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StringArray lets us store a []string as JSON in a single column (portable across
// drivers without needing Postgres-specific array types).
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for StringArray")
	}
	return json.Unmarshal(b, a)
}

const (
	MaxImagesFree    = 3
	MaxImagesPremium = 6
)

type Restaurant struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`

	// Multilingual name
	NameEn string `gorm:"not null" json:"name_en"`
	NameAm string `json:"name_am"`

	CustomSubLink string `gorm:"uniqueIndex;not null" json:"custom_sub_link"` // e.g. yourapp.com/r/custom-sub-link

	Logo   string `json:"logo"`
	Banner string `json:"banner"`
	Slogan string `json:"slogan"`

	// Gallery images. Count enforced in handler based on subscription plan
	// (3 for free tier, 6 for premium).
	Images StringArray `gorm:"type:jsonb" json:"images"`

	LongerDescription string `json:"longer_description"`

	Location  string `json:"location"`
	Phone     string `json:"phone"`
	OpenHours string `json:"open_hours"`

	// For chains/branches
	AvailableLocations StringArray `gorm:"type:jsonb" json:"available_locations"`

	FoodSpecifications string `json:"food_specifications"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Owner        *User         `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Categories   []Category    `gorm:"foreignKey:RestaurantID;constraint:OnDelete:CASCADE;" json:"categories,omitempty"`
	Subscription *Subscription `gorm:"foreignKey:RestaurantID;constraint:OnDelete:CASCADE;" json:"subscription,omitempty"`
	QRCodes      []QRCode      `gorm:"foreignKey:RestaurantID;constraint:OnDelete:CASCADE;" json:"qr_codes,omitempty"`
}

func (r *Restaurant) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

// MaxImages returns the image cap based on whether the restaurant has an
// active premium subscription.
func (r *Restaurant) MaxImages(isPremium bool) int {
	if isPremium {
		return MaxImagesPremium
	}
	return MaxImagesFree
}
