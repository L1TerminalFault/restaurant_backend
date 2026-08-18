package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleSuperAdmin      Role = "super_admin" // platform admin
	RoleRestaurantOwner Role = "restaurant_owner"
	RoleStaff           Role = "staff"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FullName     string    `gorm:"not null" json:"full_name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Password     string    `gorm:"not null" json:"-"`
	Role         Role      `gorm:"type:varchar(30);not null;default:restaurant_owner" json:"role"`
	ProfileImage string    `json:"profile_image"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// A restaurant owner can own multiple restaurants (chain support).
	Restaurants []Restaurant `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;" json:"restaurants,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
