package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
)

type VersionRepo struct{ db *gorm.DB }

func NewVersionRepo(db *gorm.DB) *VersionRepo { return &VersionRepo{db: db} }

func (r *VersionRepo) Create(v *model.DocumentVersion) error { return r.db.Create(v).Error }

func (r *VersionRepo) ByID(id uuid.UUID) (*model.DocumentVersion, error) {
	var v model.DocumentVersion
	err := r.db.First(&v, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &v, err
}

func (r *VersionRepo) List(docID uuid.UUID) ([]model.DocumentVersion, error) {
	var list []model.DocumentVersion
	err := r.db.Where("document_id = ?", docID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *VersionRepo) CountLayer(docID uuid.UUID, layer string) (int64, error) {
	var n int64
	err := r.db.Model(&model.DocumentVersion{}).Where("document_id = ? and layer = ?", docID, layer).Count(&n).Error
	return n, err
}

func (r *VersionRepo) OldestL2(docID uuid.UUID) (*model.DocumentVersion, error) {
	var v model.DocumentVersion
	err := r.db.Where("document_id = ? and layer = ?", docID, model.LayerL2).Order("created_at asc").First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &v, err
}

func (r *VersionRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.DocumentVersion{}, "id = ?", id).Error
}

type OpRepo struct{ db *gorm.DB }

func NewOpRepo(db *gorm.DB) *OpRepo { return &OpRepo{db: db} }

func (r *OpRepo) Append(op *model.DocumentOp) error { return r.db.Create(op).Error }

func (r *OpRepo) List(docID uuid.UUID) ([]model.DocumentOp, error) {
	var list []model.DocumentOp
	err := r.db.Where("document_id = ?", docID).Order("created_at asc").Find(&list).Error
	return list, err
}

func (r *OpRepo) PurgeBefore(before time.Time) (int64, error) {
	res := r.db.Where("created_at < ?", before).Delete(&model.DocumentOp{})
	return res.RowsAffected, res.Error
}

type ActivityRepo struct{ db *gorm.DB }

func NewActivityRepo(db *gorm.DB) *ActivityRepo { return &ActivityRepo{db: db} }

func (r *ActivityRepo) Add(a *model.Activity) error { return r.db.Create(a).Error }

func (r *ActivityRepo) Recent(limit int) ([]model.Activity, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []model.Activity
	err := r.db.Order("created_at desc").Limit(limit).Find(&list).Error
	return list, err
}

type WorkbenchRepo struct{ db *gorm.DB }

func NewWorkbenchRepo(db *gorm.DB) *WorkbenchRepo { return &WorkbenchRepo{db: db} }

func (r *WorkbenchRepo) TouchRecent(userID, docID uuid.UUID, at time.Time) error {
	row := model.RecentView{UserID: userID, DocumentID: docID, ViewedAt: at}
	return r.db.Save(&row).Error
}

func (r *WorkbenchRepo) Recents(userID uuid.UUID, limit int) ([]model.RecentView, error) {
	var list []model.RecentView
	err := r.db.Where("user_id = ?", userID).Order("viewed_at desc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *WorkbenchRepo) AddFavorite(userID, docID uuid.UUID, at time.Time) error {
	return r.db.Save(&model.Favorite{UserID: userID, DocumentID: docID, CreatedAt: at}).Error
}

func (r *WorkbenchRepo) RemoveFavorite(userID, docID uuid.UUID) error {
	return r.db.Delete(&model.Favorite{}, "user_id = ? and document_id = ?", userID, docID).Error
}

func (r *WorkbenchRepo) Favorites(userID uuid.UUID) ([]model.Favorite, error) {
	var list []model.Favorite
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *WorkbenchRepo) IsFavorite(userID, docID uuid.UUID) bool {
	var n int64
	r.db.Model(&model.Favorite{}).Where("user_id = ? and document_id = ?", userID, docID).Count(&n)
	return n > 0
}
