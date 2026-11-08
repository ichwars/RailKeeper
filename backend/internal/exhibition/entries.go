package exhibition

import (
	"context"
	"fmt"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *Service) ListEntries(ctx context.Context, listID string) ([]ExhibitionEntry, error) {
	if _, err := s.repository.List(ctx, listID); err != nil {
		return nil, err
	}
	return s.repository.Entries(ctx, listID)
}
func (s *Service) CreateEntry(ctx context.Context, listID string, input ExhibitionEntryInput) (ExhibitionEntry, error) {
	list, err := s.repository.List(ctx, listID)
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if !exhibitionAllowsPlanningChanges(list) {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	input = normalizeExhibitionEntryInput(input)
	if input.LocomotiveName == "" {
		return ExhibitionEntry{}, ErrExhibitionValidation
	}
	if input.SortOrder == 0 {
		input.SortOrder = s.repository.NextSort(ctx, listID)
	}

	id := credentials.RandomID()
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return ExhibitionEntry{}, fmt.Errorf("begin exhibition entry create: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := tx.InsertEntry(ctx, EntryWrite{ID: id, ListID: listID, At: now, Input: input}); err != nil {
		return ExhibitionEntry{}, err
	}
	if err := touchUnlockedList(ctx, tx, listID, now); err != nil {
		return ExhibitionEntry{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("commit exhibition entry create: %w", err)
	}
	return s.repository.Entry(ctx, listID, id)
}
func (s *Service) UpdateEntry(ctx context.Context, listID, entryID string, input ExhibitionEntryInput) (ExhibitionEntry, error) {
	list, err := s.repository.List(ctx, listID)
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if !exhibitionAllowsPlanningChanges(list) {
		return ExhibitionEntry{}, ErrExhibitionLocked
	}
	if _, err := s.repository.Entry(ctx, listID, entryID); err != nil {
		return ExhibitionEntry{}, err
	}
	input = normalizeExhibitionEntryInput(input)
	if input.LocomotiveName == "" {
		return ExhibitionEntry{}, ErrExhibitionValidation
	}
	current, err := s.repository.Entry(ctx, listID, entryID)
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != current.Revision {
		return ExhibitionEntry{}, ErrExhibitionStale
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return ExhibitionEntry{}, fmt.Errorf("begin exhibition entry update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	changed, err := tx.UpdateEntry(ctx, EntryWrite{ID: entryID, ListID: listID, At: now, Input: input})
	if err != nil {
		return ExhibitionEntry{}, err
	}
	if !changed {
		return ExhibitionEntry{}, ErrExhibitionStale
	}
	if err := touchUnlockedList(ctx, tx, listID, now); err != nil {
		return ExhibitionEntry{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExhibitionEntry{}, fmt.Errorf("commit exhibition entry update: %w", err)
	}
	return s.repository.Entry(ctx, listID, entryID)
}
func (s *Service) DeleteEntry(ctx context.Context, listID, entryID string) error {
	list, err := s.repository.List(ctx, listID)
	if err != nil {
		return err
	}
	if !exhibitionAllowsPlanningChanges(list) {
		return ErrExhibitionLocked
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin exhibition entry delete: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	changed, err := tx.DeleteEntry(ctx, listID, entryID)
	if err != nil {
		return err
	}
	if !changed {
		return ErrExhibitionNotFound
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if err := touchUnlockedList(ctx, tx, listID, now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit exhibition entry delete: %w", err)
	}
	return nil
}
func touchUnlockedList(ctx context.Context, tx Mutation, listID, now string) error {
	changed, err := tx.TouchUnlockedList(ctx, listID, now)
	if err != nil {
		return err
	}
	if !changed {
		return ErrExhibitionLocked
	}
	return nil
}
