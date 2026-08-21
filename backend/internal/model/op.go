package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentOp struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentID uuid.UUID `gorm:"type:uuid;index;not null" json:"documentId"`
	SiteID     uint64    `gorm:"not null" json:"siteId"`
	Clock      uint64    `gorm:"not null" json:"clock"`
	OpJSON     string    `gorm:"type:text;not null" json:"opJson"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (o *DocumentOp) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
