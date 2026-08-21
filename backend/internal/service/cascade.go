package service

import (
	"github.com/google/uuid"

	"gowiki/internal/model"
	"gowiki/internal/pkg/timeutil"
	"gowiki/internal/repository"
)

// SoftDeleteTree removes a node and every descendant in one pass so the
// sidebar never shows orphans hanging off a recycled parent.
func (s *DocumentService) SoftDeleteTree(actor, id uuid.UUID) error {
	d, err := s.docs.ByID(id)
	if err != nil {
		return err
	}
	list, err := s.docs.Descendants(s.docs.DB(), d.Path)
	if err != nil {
		return err
	}
	for i := range list {
		_ = s.docs.SoftDelete(list[i].ID)
		if s.idx != nil {
			_ = s.idx.Delete(list[i].ID)
		}
	}
	_ = s.acts.Add(&model.Activity{
		SpaceID: d.SpaceID, ActorID: actor, Action: "delete_tree",
		DocumentID: &id, Summary: "删除文档树「" + d.Title + "」", CreatedAt: timeutil.Now(),
	})
	return nil
}

func CollectIDs(docs []model.Document) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(docs))
	for _, d := range docs {
		out = append(out, d.ID)
	}
	return out
}

func SortTree(docs []model.Document) []model.Document {
	byParent := map[string][]model.Document{}
	for _, d := range docs {
		key := ""
		if d.ParentID != nil {
			key = d.ParentID.String()
		}
		byParent[key] = append(byParent[key], d)
	}
	var walk func(string) []model.Document
	walk = func(parent string) []model.Document {
		kids := byParent[parent]
		var out []model.Document
		for _, k := range kids {
			out = append(out, k)
			out = append(out, walk(k.ID.String())...)
		}
		return out
	}
	return walk("")
}

func WouldCycle(nodePath, parentPath string) bool {
	return repository.IsAncestorPath(nodePath, parentPath)
}
