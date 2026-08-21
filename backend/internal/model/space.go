package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Space struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:120;not null" json:"name"`
	OwnerID   uuid.UUID `gorm:"type:uuid;index;not null" json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *Space) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
