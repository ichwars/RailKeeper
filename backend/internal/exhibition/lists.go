package exhibition

import (
	"context"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *Service) List(ctx context.Context) ([]ExhibitionList, error) { return s.repository.Lists(ctx) }
func (s *Service) Get(ctx context.Context, id string) (ExhibitionList, error) {
	list, err := s.repository.List(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	entries, err := s.ListEntries(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	list.Entries = entries
	list.EntryCount = len(entries)
	return list, nil
}
func (s *Service) Create(ctx context.Context, input ExhibitionListInput) (ExhibitionList, error) {
	input, err := normalizeExhibitionListInput(input)
	if err != nil {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	if input.Status != ExhibitionStatusDraft && input.Status != ExhibitionStatusOpen {
		return ExhibitionList{}, ErrExhibitionValidation
	}

	id := credentials.RandomID()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.repository.InsertList(ctx, ListWrite{ID: id, At: now, Input: input}); err != nil {
		return ExhibitionList{}, err
	}
	return s.Get(ctx, id)
}
func (s *Service) Update(ctx context.Context, id string, input ExhibitionListInput) (ExhibitionList, error) {
	input, err := normalizeExhibitionListInput(input)
	if err != nil {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	current, err := s.repository.List(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	if !exhibitionAllowsPlanningChanges(current) {
		return ExhibitionList{}, ErrExhibitionLocked
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != current.Revision {
		return ExhibitionList{}, ErrExhibitionStale
	}
	input.Status = current.Status

	now := time.Now().UTC().Format(time.RFC3339)
	changed, err := s.repository.UpdateList(ctx, ListWrite{ID: id, At: now, Input: input})
	if err != nil {
		return ExhibitionList{}, err
	}
	if !changed {
		return ExhibitionList{}, ErrExhibitionStale
	}
	return s.Get(ctx, id)
}
func (s *Service) SetLocked(ctx context.Context, id string, locked bool) (ExhibitionList, error) {
	status := ExhibitionStatusOpen
	if locked {
		status = ExhibitionStatusLocked
	}
	return s.SetStatus(ctx, id, ExhibitionStatusInput{Status: status, ConfirmConflicts: true, Reason: "Legacy lock action"})
}
func (s *Service) Delete(ctx context.Context, id string) error {
	changed, err := s.repository.DeleteList(ctx, id)
	if err != nil {
		return err
	}
	if !changed {
		return ErrExhibitionNotFound
	}
	return nil
}
