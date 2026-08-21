package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ModeMarkdown = "markdown"
	ModeRich     = "rich"
)

type Document struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SpaceID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"spaceId"`
	ParentID    *uuid.UUID     `gorm:"type:uuid;index" json:"parentId"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Path        string         `gorm:"size:2048;index;not null" json:"path"`
	SortOrder   int            `gorm:"not null;default:0" json:"sortOrder"`
	ContentMD   string         `gorm:"type:text" json:"contentMd"`
	ContentJSON string         `gorm:"type:text" json:"contentJson"`
	EditorMode  string         `gorm:"size:16;not null;default:markdown" json:"editorMode"`
	CRDTState   string         `gorm:"type:text" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if d.EditorMode == "" {
		d.EditorMode = ModeMarkdown
	}
	return nil
}

func (d *Document) IsDeleted() bool {
	return d.DeletedAt.Valid
}
