package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Activity struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	SpaceID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"spaceId"`
	ActorID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"actorId"`
	Action     string     `gorm:"size:40;not null" json:"action"`
	DocumentID *uuid.UUID `gorm:"type:uuid;index" json:"documentId"`
	Summary    string     `gorm:"size:300;not null" json:"summary"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (a *Activity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type Favorite struct {
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	DocumentID uuid.UUID `gorm:"type:uuid;primaryKey" json:"documentId"`
	CreatedAt  time.Time `json:"createdAt"`
}

type RecentView struct {
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	DocumentID uuid.UUID `gorm:"type:uuid;primaryKey" json:"documentId"`
	ViewedAt   time.Time `gorm:"index;not null" json:"viewedAt"`
}
