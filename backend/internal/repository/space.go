package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
)

type SpaceRepo struct{ db *gorm.DB }

func NewSpaceRepo(db *gorm.DB) *SpaceRepo { return &SpaceRepo{db: db} }

func (r *SpaceRepo) Create(s *model.Space) error { return r.db.Create(s).Error }

func (r *SpaceRepo) ByID(id uuid.UUID) (*model.Space, error) {
	var s model.Space
	err := r.db.First(&s, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &s, err
}

func (r *SpaceRepo) List() ([]model.Space, error) {
	var list []model.Space
	err := r.db.Order("created_at asc").Find(&list).Error
	return list, err
}

func (r *SpaceRepo) Update(s *model.Space) error { return r.db.Save(s).Error }

func (r *SpaceRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Space{}, "id = ?", id).Error
}
