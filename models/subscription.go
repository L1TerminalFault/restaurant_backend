package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plan string

const (
	PlanFree    Plan = "free"
	PlanPremium Plan = "premium"
)

type Subscription struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"restaurant_id"`

	Plan Plan `gorm:"type:varchar(20);not null;default:free" json:"plan"`

	JoinTime           time.Time  `json:"join_time"`
	LastPaymentRenewal *time.Time `json:"last_payment_renewal"`
	LastRenewal        *time.Time `json:"last_renewal"`
	ExpiresAt          *time.Time `json:"expires_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.JoinTime.IsZero() {
		s.JoinTime = time.Now()
	}
	return
}

// IsActivePremium reports whether the subscription currently grants premium perks.
func (s *Subscription) IsActivePremium() bool {
	if s.Plan != PlanPremium {
		return false
	}
	if s.ExpiresAt == nil {
		return true
	}
	return time.Now().Before(*s.ExpiresAt)
}
