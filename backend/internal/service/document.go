package service

import (
	"strings"

	"github.com/google/uuid"

	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/timeutil"
	"gowiki/internal/pkg/validate"
	"gowiki/internal/repository"
	"gowiki/internal/search"
)

type DocumentService struct {
	docs   *repository.DocumentRepo
	spaces *repository.SpaceRepo
	acts   *repository.ActivityRepo
	wb     *repository.WorkbenchRepo
	idx    *search.Engine
}

func NewDocumentService(
	docs *repository.DocumentRepo,
	spaces *repository.SpaceRepo,
	acts *repository.ActivityRepo,
	wb *repository.WorkbenchRepo,
	idx *search.Engine,
) *DocumentService {
	return &DocumentService{docs: docs, spaces: spaces, acts: acts, wb: wb, idx: idx}
}

func (s *DocumentService) Create(actor uuid.UUID, spaceID uuid.UUID, parentID *uuid.UUID, title, mode string) (*model.Document, error) {
	if _, err := s.spaces.ByID(spaceID); err != nil {
		return nil, err
	}
	title = strings.TrimSpace(title)
	if err := validate.Length("标题", title, 1, 120); err != nil {
		return nil, err
	}
	if mode != model.ModeRich {
		mode = model.ModeMarkdown
	}
	var parentPath string
	if parentID != nil {
		p, err := s.docs.ByID(*parentID)
		if err != nil {
			return nil, err
		}
		if p.SpaceID != spaceID {
			return nil, apperr.New(apperr.CodeValidation, 400, "父节点不在同一空间")
		}
		parentPath = p.Path
	} else {
		parentPath = "/"
	}
	sort, err := s.docs.NextSort(spaceID, parentID)
	if err != nil {
		return nil, err
	}
	id := uuid.New()
	doc := &model.Document{
		ID: id, SpaceID: spaceID, ParentID: parentID,
		Title: title, Path: repository.PathJoin(parentPath, id),
		SortOrder: sort, EditorMode: mode,
		ContentMD: "", ContentJSON: "",
	}
	if err := s.docs.Create(doc); err != nil {
		return nil, err
	}
	_ = s.acts.Add(&model.Activity{
		SpaceID: spaceID, ActorID: actor, Action: "create_doc",
		DocumentID: &id, Summary: "创建文档「" + title + "」", CreatedAt: timeutil.Now(),
	})
	s.index(doc)
	return doc, nil
}

func (s *DocumentService) Get(actor, id uuid.UUID) (*model.Document, error) {
	d, err := s.docs.ByID(id)
	if err != nil {
		return nil, err
	}
	_ = s.wb.TouchRecent(actor, id, timeutil.Now())
	return d, nil
}

func (s *DocumentService) UpdateMeta(actor, id uuid.UUID, title, mode, contentMD, contentJSON *string) (*model.Document, error) {
	d, err := s.docs.ByID(id)
	if err != nil {
		return nil, err
	}
	if title != nil {
		if err := validate.Length("标题", *title, 1, 120); err != nil {
			return nil, err
		}
		d.Title = strings.TrimSpace(*title)
	}
	if mode != nil && (*mode == model.ModeMarkdown || *mode == model.ModeRich) {
		d.EditorMode = *mode
	}
	if contentMD != nil {
		d.ContentMD = *contentMD
	}
	if contentJSON != nil {
		d.ContentJSON = *contentJSON
	}
	d.UpdatedAt = timeutil.Now()
	if err := s.docs.Update(d); err != nil {
		return nil, err
	}
	_ = s.acts.Add(&model.Activity{
		SpaceID: d.SpaceID, ActorID: actor, Action: "update_doc",
		DocumentID: &d.ID, Summary: "更新文档「" + d.Title + "」", CreatedAt: timeutil.Now(),
	})
	s.index(d)
	return d, nil
}

func (s *DocumentService) Tree(spaceID uuid.UUID) ([]model.Document, error) {
	if _, err := s.spaces.ByID(spaceID); err != nil {
		return nil, err
	}
	return s.docs.ListBySpace(spaceID)
}

func (s *DocumentService) Delete(actor, id uuid.UUID) error {
	return s.SoftDeleteTree(actor, id)
}

func (s *DocumentService) Recycle(spaceID uuid.UUID) ([]model.Document, error) {
	return s.docs.Recycle(spaceID)
}

func (s *DocumentService) Restore(actor, id uuid.UUID) (*model.Document, error) {
	if err := s.docs.Restore(id); err != nil {
		return nil, err
	}
	d, err := s.docs.ByID(id)
	if err != nil {
		return nil, err
	}
	s.index(d)
	_ = s.acts.Add(&model.Activity{
		SpaceID: d.SpaceID, ActorID: actor, Action: "restore_doc",
		DocumentID: &id, Summary: "恢复文档「" + d.Title + "」", CreatedAt: timeutil.Now(),
	})
	return d, nil
}

func (s *DocumentService) index(d *model.Document) {
	if s.idx == nil {
		return
	}
	_ = s.idx.Upsert(search.Doc{
		ID: d.ID.String(), SpaceID: d.SpaceID.String(),
		Title: d.Title, Content: d.ContentMD, UpdatedAt: d.UpdatedAt,
	})
}

func (s *DocumentService) ReindexAll() error {
	if s.idx == nil {
		return nil
	}
	list, err := s.docs.ListAll()
	if err != nil {
		return err
	}
	for i := range list {
		s.index(&list[i])
	}
	return nil
}

func (s *DocumentService) PersistContent(id uuid.UUID, md, json, crdt string) error {
	d, err := s.docs.ByID(id)
	if err != nil {
		return err
	}
	d.ContentMD = md
	if json != "" {
		d.ContentJSON = json
	}
	d.CRDTState = crdt
	d.UpdatedAt = timeutil.Now()
	if err := s.docs.Update(d); err != nil {
		return err
	}
	s.index(d)
	return nil
}
