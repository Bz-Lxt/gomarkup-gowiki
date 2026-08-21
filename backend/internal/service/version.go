package service

import (
	"strings"

	"github.com/google/uuid"

	"gowiki/internal/diff"
	"gowiki/internal/model"
	"gowiki/internal/pkg/apperr"
	"gowiki/internal/pkg/timeutil"
	"gowiki/internal/pkg/validate"
	"gowiki/internal/repository"
	"gowiki/internal/search"
)

type VersionService struct {
	vers *repository.VersionRepo
	docs *repository.DocumentRepo
	acts *repository.ActivityRepo
	idx  *search.Engine
}

func NewVersionService(
	vers *repository.VersionRepo,
	docs *repository.DocumentRepo,
	acts *repository.ActivityRepo,
	idx *search.Engine,
) *VersionService {
	return &VersionService{vers: vers, docs: docs, acts: acts, idx: idx}
}

func (s *VersionService) SaveNamed(actor, docID uuid.UUID, label string) (*model.DocumentVersion, error) {
	if err := validate.Length("版本说明", label, 1, 120); err != nil {
		return nil, err
	}
	d, err := s.docs.ByID(docID)
	if err != nil {
		return nil, err
	}
	v := &model.DocumentVersion{
		DocumentID: docID, Layer: model.LayerL3, Label: strings.TrimSpace(label),
		ContentMD: d.ContentMD, ContentJSON: d.ContentJSON, AuthorID: actor,
		CreatedAt: timeutil.Now(),
	}
	if err := s.vers.Create(v); err != nil {
		return nil, err
	}
	_ = s.acts.Add(&model.Activity{
		SpaceID: d.SpaceID, ActorID: actor, Action: "save_version",
		DocumentID: &docID, Summary: "保存版本「" + v.Label + "」", CreatedAt: timeutil.Now(),
	})
	return v, nil
}

func (s *VersionService) AutoSnapshot(actor, docID uuid.UUID, reason string) (*model.DocumentVersion, error) {
	d, err := s.docs.ByID(docID)
	if err != nil {
		return nil, err
	}
	v := &model.DocumentVersion{
		DocumentID: docID, Layer: model.LayerL2, Label: reason,
		ContentMD: d.ContentMD, ContentJSON: d.ContentJSON, AuthorID: actor,
		CreatedAt: timeutil.Now(),
	}
	if err := s.vers.Create(v); err != nil {
		return nil, err
	}
	n, _ := s.vers.CountLayer(docID, model.LayerL2)
	if n > 50 {
		if old, err := s.vers.OldestL2(docID); err == nil {
			_ = s.vers.Delete(old.ID)
		}
	}
	return v, nil
}

func (s *VersionService) List(docID uuid.UUID) ([]model.DocumentVersion, error) {
	if _, err := s.docs.ByID(docID); err != nil {
		return nil, err
	}
	return s.vers.List(docID)
}

func (s *VersionService) Get(id uuid.UUID) (*model.DocumentVersion, error) {
	return s.vers.ByID(id)
}

func (s *VersionService) Diff(leftID, rightID uuid.UUID) (diff.Result, error) {
	l, err := s.vers.ByID(leftID)
	if err != nil {
		return diff.Result{}, err
	}
	r, err := s.vers.ByID(rightID)
	if err != nil {
		return diff.Result{}, err
	}
	if l.DocumentID != r.DocumentID {
		return diff.Result{}, apperr.New(apperr.CodeValidation, 400, "只能对比同一文档的版本")
	}
	return diff.Compare(l.ContentMD, r.ContentMD), nil
}

func (s *VersionService) DiffAgainstCurrent(versionID uuid.UUID) (diff.Result, error) {
	v, err := s.vers.ByID(versionID)
	if err != nil {
		return diff.Result{}, err
	}
	d, err := s.docs.ByID(v.DocumentID)
	if err != nil {
		return diff.Result{}, err
	}
	return diff.Compare(v.ContentMD, d.ContentMD), nil
}

func (s *VersionService) Rollback(actor, versionID uuid.UUID) (*model.DocumentVersion, error) {
	v, err := s.vers.ByID(versionID)
	if err != nil {
		return nil, err
	}
	d, err := s.docs.ByID(v.DocumentID)
	if err != nil {
		return nil, err
	}
	d.ContentMD = v.ContentMD
	d.ContentJSON = v.ContentJSON
	d.UpdatedAt = timeutil.Now()
	if err := s.docs.Update(d); err != nil {
		return nil, err
	}
	nv := &model.DocumentVersion{
		DocumentID: d.ID, Layer: model.LayerL3,
		Label: "回滚至「" + v.Label + "」",
		ContentMD: v.ContentMD, ContentJSON: v.ContentJSON,
		AuthorID: actor, CreatedAt: timeutil.Now(),
	}
	if err := s.vers.Create(nv); err != nil {
		return nil, err
	}
	if s.idx != nil {
		_ = s.idx.Upsert(search.Doc{
			ID: d.ID.String(), SpaceID: d.SpaceID.String(),
			Title: d.Title, Content: d.ContentMD, UpdatedAt: d.UpdatedAt,
		})
	}
	_ = s.acts.Add(&model.Activity{
		SpaceID: d.SpaceID, ActorID: actor, Action: "rollback",
		DocumentID: &d.ID, Summary: "回滚文档「" + d.Title + "」", CreatedAt: timeutil.Now(),
	})
	return nv, nil
}
