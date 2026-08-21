package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	LayerL2 = "L2"
	LayerL3 = "L3"
)

type DocumentVersion struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentID  uuid.UUID `gorm:"type:uuid;index;not null" json:"documentId"`
	Layer       string    `gorm:"size:8;index;not null" json:"layer"`
	Label       string    `gorm:"size:200" json:"label"`
	ContentMD   string    `gorm:"type:text" json:"contentMd"`
	ContentJSON string    `gorm:"type:text" json:"contentJson"`
	AuthorID    uuid.UUID `gorm:"type:uuid;index;not null" json:"authorId"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (v *DocumentVersion) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
