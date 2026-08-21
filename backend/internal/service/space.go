package service

import (
	"strings"

	"github.com/google/uuid"

	"gowiki/internal/model"
	"gowiki/internal/pkg/validate"
	"gowiki/internal/repository"
)

type SpaceService struct {
	spaces *repository.SpaceRepo
}

func NewSpaceService(spaces *repository.SpaceRepo) *SpaceService {
	return &SpaceService{spaces: spaces}
}

func (s *SpaceService) List() ([]model.Space, error) { return s.spaces.List() }

func (s *SpaceService) Get(id uuid.UUID) (*model.Space, error) { return s.spaces.ByID(id) }

func (s *SpaceService) Create(owner uuid.UUID, name string) (*model.Space, error) {
	name = strings.TrimSpace(name)
	if err := validate.Length("空间名称", name, 1, 80); err != nil {
		return nil, err
	}
	sp := &model.Space{Name: name, OwnerID: owner}
	if err := s.spaces.Create(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

func (s *SpaceService) Rename(id uuid.UUID, name string) (*model.Space, error) {
	name = strings.TrimSpace(name)
	if err := validate.Length("空间名称", name, 1, 80); err != nil {
		return nil, err
	}
	sp, err := s.spaces.ByID(id)
	if err != nil {
		return nil, err
	}
	sp.Name = name
	if err := s.spaces.Update(sp); err != nil {
		return nil, err
	}
	return sp, nil
}
