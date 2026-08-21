package service

import (
	"github.com/google/uuid"

	"gowiki/internal/pkg/timeutil"
	"gowiki/internal/repository"
)

type WorkbenchService struct {
	wb   *repository.WorkbenchRepo
	docs *repository.DocumentRepo
	acts *repository.ActivityRepo
}

func NewWorkbenchService(wb *repository.WorkbenchRepo, docs *repository.DocumentRepo, acts *repository.ActivityRepo) *WorkbenchService {
	return &WorkbenchService{wb: wb, docs: docs, acts: acts}
}

type Workbench struct {
	Recents    []WorkbenchItem `json:"recents"`
	Favorites  []WorkbenchItem `json:"favorites"`
	Activities []ActivityItem  `json:"activities"`
}

type WorkbenchItem struct {
	DocumentID string `json:"documentId"`
	Title      string `json:"title"`
	SpaceID    string `json:"spaceId"`
	At         string `json:"at"`
}

type ActivityItem struct {
	ID         string `json:"id"`
	Action     string `json:"action"`
	Summary    string `json:"summary"`
	DocumentID string `json:"documentId,omitempty"`
	ActorID    string `json:"actorId"`
	At         string `json:"at"`
}

func (s *WorkbenchService) Home(userID uuid.UUID) (*Workbench, error) {
	recents, err := s.wb.Recents(userID, 12)
	if err != nil {
		return nil, err
	}
	favs, err := s.wb.Favorites(userID)
	if err != nil {
		return nil, err
	}
	acts, err := s.acts.Recent(20)
	if err != nil {
		return nil, err
	}
	out := &Workbench{
		Recents:    make([]WorkbenchItem, 0, len(recents)),
		Favorites:  make([]WorkbenchItem, 0, len(favs)),
		Activities: make([]ActivityItem, 0, len(acts)),
	}
	for _, r := range recents {
		if item, ok := s.item(r.DocumentID, r.ViewedAt.Format("2006-01-02 15:04:05")); ok {
			out.Recents = append(out.Recents, item)
		}
	}
	for _, f := range favs {
		if item, ok := s.item(f.DocumentID, timeutil.Format(f.CreatedAt)); ok {
			out.Favorites = append(out.Favorites, item)
		}
	}
	for _, a := range acts {
		it := ActivityItem{
			ID: a.ID.String(), Action: a.Action, Summary: a.Summary,
			ActorID: a.ActorID.String(), At: timeutil.Format(a.CreatedAt),
		}
		if a.DocumentID != nil {
			it.DocumentID = a.DocumentID.String()
		}
		out.Activities = append(out.Activities, it)
	}
	return out, nil
}

func (s *WorkbenchService) item(id uuid.UUID, at string) (WorkbenchItem, bool) {
	d, err := s.docs.ByID(id)
	if err != nil {
		return WorkbenchItem{}, false
	}
	return WorkbenchItem{
		DocumentID: d.ID.String(), Title: d.Title,
		SpaceID: d.SpaceID.String(), At: at,
	}, true
}

func (s *WorkbenchService) ToggleFavorite(userID, docID uuid.UUID) (bool, error) {
	if _, err := s.docs.ByID(docID); err != nil {
		return false, err
	}
	if s.wb.IsFavorite(userID, docID) {
		return false, s.wb.RemoveFavorite(userID, docID)
	}
	return true, s.wb.AddFavorite(userID, docID, timeutil.Now())
}

func (s *WorkbenchService) IsFavorite(userID, docID uuid.UUID) bool {
	return s.wb.IsFavorite(userID, docID)
}
