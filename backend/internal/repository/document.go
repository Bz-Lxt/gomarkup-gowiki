package repository

import (
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
)

type DocumentRepo struct{ db *gorm.DB }

func NewDocumentRepo(db *gorm.DB) *DocumentRepo { return &DocumentRepo{db: db} }

func (r *DocumentRepo) DB() *gorm.DB { return r.db }

func (r *DocumentRepo) Create(d *model.Document) error { return r.db.Create(d).Error }

func (r *DocumentRepo) ByID(id uuid.UUID) (*model.Document, error) {
	var d model.Document
	err := r.db.First(&d, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &d, err
}

func (r *DocumentRepo) ByIDUnscoped(id uuid.UUID) (*model.Document, error) {
	var d model.Document
	err := r.db.Unscoped().First(&d, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &d, err
}

func (r *DocumentRepo) ListBySpace(spaceID uuid.UUID) ([]model.Document, error) {
	var list []model.Document
	err := r.db.Where("space_id = ?", spaceID).Order("path asc, sort_order asc").Find(&list).Error
	return list, err
}

func (r *DocumentRepo) ListAll() ([]model.Document, error) {
	var list []model.Document
	err := r.db.Find(&list).Error
	return list, err
}

func (r *DocumentRepo) Update(d *model.Document) error { return r.db.Save(d).Error }

func (r *DocumentRepo) SoftDelete(id uuid.UUID) error {
	return r.db.Delete(&model.Document{}, "id = ?", id).Error
}

func (r *DocumentRepo) Restore(id uuid.UUID) error {
	return r.db.Unscoped().Model(&model.Document{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *DocumentRepo) Recycle(spaceID uuid.UUID) ([]model.Document, error) {
	var list []model.Document
	q := r.db.Unscoped().Where("deleted_at is not null")
	if spaceID != uuid.Nil {
		q = q.Where("space_id = ?", spaceID)
	}
	err := q.Order("deleted_at desc").Find(&list).Error
	return list, err
}

func (r *DocumentRepo) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&model.Document{}, "id = ?", id).Error
}

func (r *DocumentRepo) LockByID(tx *gorm.DB, id uuid.UUID) (*model.Document, error) {
	var d model.Document
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&d, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &d, err
}

func (r *DocumentRepo) Descendants(tx *gorm.DB, pathPrefix string) ([]model.Document, error) {
	var list []model.Document
	err := tx.Where("path like ?", pathPrefix+"%").Find(&list).Error
	return list, err
}

func (r *DocumentRepo) NextSort(spaceID uuid.UUID, parent *uuid.UUID) (int, error) {
	var max int
	q := r.db.Model(&model.Document{}).Select("coalesce(max(sort_order), -1)").Where("space_id = ?", spaceID)
	if parent == nil {
		q = q.Where("parent_id is null")
	} else {
		q = q.Where("parent_id = ?", *parent)
	}
	err := q.Scan(&max).Error
	return max + 1, err
}

func PathJoin(parentPath string, id uuid.UUID) string {
	if parentPath == "" {
		return "/" + id.String() + "/"
	}
	if !strings.HasSuffix(parentPath, "/") {
		parentPath += "/"
	}
	return parentPath + id.String() + "/"
}

func IsAncestorPath(nodePath, candidateParentPath string) bool {
	if nodePath == "" {
		return false
	}
	return strings.HasPrefix(candidateParentPath, nodePath)
}
