package exhibition

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *Service) SetConflictException(
	ctx context.Context,
	listID string,
	conflictID string,
	input ExhibitionConflictExceptionInput,
) (ExhibitionWorkspace, error) {
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return ExhibitionWorkspace{}, ErrExhibitionValidation
	}
	workspace, err := s.Workspace(ctx, listID)
	if err != nil {
		return ExhibitionWorkspace{}, err
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != workspace.List.Revision {
		return ExhibitionWorkspace{}, ErrExhibitionStale
	}
	found := false
	for _, conflict := range workspace.Conflicts {
		if conflict.ID == conflictID {
			found = true
			break
		}
	}
	if !found {
		return ExhibitionWorkspace{}, ErrExhibitionNotFound
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return ExhibitionWorkspace{}, fmt.Errorf("begin exhibition exception: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := tx.UpsertException(ctx, ExceptionWrite{ID: credentials.RandomID(), ListID: listID, ConflictID: conflictID, Reason: reason, At: now}); err != nil {
		return ExhibitionWorkspace{}, err
	}
	changed, err := tx.BumpRevision(ctx, listID, input.ExpectedRevision, now)
	if err != nil {
		return ExhibitionWorkspace{}, err
	}
	if !changed {
		return ExhibitionWorkspace{}, ErrExhibitionStale
	}
	if err := tx.Commit(); err != nil {
		return ExhibitionWorkspace{}, fmt.Errorf("commit exhibition exception: %w", err)
	}
	return s.Workspace(ctx, listID)
}
func (s *Service) SetStatus(ctx context.Context, id string, input ExhibitionStatusInput) (ExhibitionList, error) {
	if !validExhibitionStatus(input.Status) {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	workspace, err := s.Workspace(ctx, id)
	if err != nil {
		return ExhibitionList{}, err
	}
	if input.ExpectedRevision > 0 && input.ExpectedRevision != workspace.List.Revision {
		return ExhibitionList{}, ErrExhibitionStale
	}
	if input.Status != workspace.List.Status && !validExhibitionTransition(workspace.List.Status, input.Status) {
		return ExhibitionList{}, ErrExhibitionValidation
	}
	reason := strings.TrimSpace(input.Reason)
	if input.Status == ExhibitionStatusLocked && len(workspace.Conflicts) > 0 {
		if !input.ConfirmConflicts || reason == "" {
			return ExhibitionList{}, ErrExhibitionConflicts
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	locked := input.Status == ExhibitionStatusLocked || input.Status == ExhibitionStatusRunning
	lockedAt, completedAt, archivedAt := workspace.List.LockedAt, workspace.List.CompletedAt, workspace.List.ArchivedAt
	if input.Status == ExhibitionStatusLocked && lockedAt == "" {
		lockedAt = now
	}
	if input.Status == ExhibitionStatusCompleted && completedAt == "" {
		completedAt = now
	}
	if input.Status == ExhibitionStatusArchived && archivedAt == "" {
		archivedAt = now
	}
	changed, err := s.repository.SetStatus(ctx, StatusChange{ID: id, Status: input.Status, Locked: locked, Reason: reason,
		LockedAt: lockedAt, CompletedAt: completedAt, ArchivedAt: archivedAt, At: now, ExpectedRevision: input.ExpectedRevision})
	if err != nil {
		return ExhibitionList{}, err
	}
	if !changed {
		return ExhibitionList{}, ErrExhibitionStale
	}
	return s.Get(ctx, id)
}
func validExhibitionTransition(from, to ExhibitionStatus) bool {
	switch from {
	case ExhibitionStatusDraft:
		return to == ExhibitionStatusOpen || to == ExhibitionStatusArchived
	case ExhibitionStatusOpen:
		return to == ExhibitionStatusDraft || to == ExhibitionStatusLocked || to == ExhibitionStatusArchived
	case ExhibitionStatusLocked:
		return to == ExhibitionStatusOpen || to == ExhibitionStatusRunning || to == ExhibitionStatusArchived
	case ExhibitionStatusRunning:
		return to == ExhibitionStatusLocked || to == ExhibitionStatusCompleted
	case ExhibitionStatusCompleted:
		return to == ExhibitionStatusOpen || to == ExhibitionStatusArchived
	case ExhibitionStatusArchived:
		return to == ExhibitionStatusOpen
	default:
		return false
	}
}
