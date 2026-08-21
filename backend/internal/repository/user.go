package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
)

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepo) ByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &u, err
}

func (r *UserRepo) ByID(id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperr.ErrNotFound
	}
	return &u, err
}
