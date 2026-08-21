package service

import (
	"strings"

	"github.com/google/uuid"

	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/repository"
)

type TreeService struct {
	docs *repository.DocumentRepo
}

func NewTreeService(docs *repository.DocumentRepo) *TreeService {
	return &TreeService{docs: docs}
}

type MoveInput struct {
	ParentID  *uuid.UUID `json:"parentId"`
	SortOrder *int       `json:"sortOrder"`
}

func (s *TreeService) Move(id uuid.UUID, in MoveInput) (*model.Document, error) {
	if in.ParentID != nil && *in.ParentID == id {
		return nil, apperr.ErrCycle
	}
	tx := s.docs.DB().Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	node, err := s.docs.LockByID(tx, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var newParentPath string
	if in.ParentID == nil {
		newParentPath = "/"
	} else {
		parent, err := s.docs.LockByID(tx, *in.ParentID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if parent.SpaceID != node.SpaceID {
			tx.Rollback()
			return nil, apperr.New(apperr.CodeValidation, 400, "不能跨空间移动")
		}
		if repository.IsAncestorPath(node.Path, parent.Path) {
			tx.Rollback()
			return nil, apperr.ErrCycle
		}
		newParentPath = parent.Path
	}

	oldPath := node.Path
	newPath := repository.PathJoin(newParentPath, node.ID)
	sortOrder := node.SortOrder
	if in.SortOrder != nil {
		sortOrder = *in.SortOrder
	}

	if err := tx.Model(&model.Document{}).Where("id = ?", node.ID).Updates(map[string]any{
		"parent_id":  in.ParentID,
		"path":       newPath,
		"sort_order": sortOrder,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if oldPath != newPath {
		var descendants []model.Document
		if err := tx.Where("id <> ? AND path LIKE ?", node.ID, oldPath+"%").Find(&descendants).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for i := range descendants {
			np := newPath + strings.TrimPrefix(descendants[i].Path, oldPath)
			if err := tx.Model(&model.Document{}).Where("id = ?", descendants[i].ID).Update("path", np).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.docs.ByID(id)
}
