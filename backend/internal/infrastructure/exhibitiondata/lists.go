package exhibitiondata

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/exhibition"
)

func (s *Store) InsertList(ctx context.Context, write exhibition.ListWrite) error {
	input, id, now := write.Input, write.ID, write.At
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO exhibition_lists(
  id, designation, list_date, end_date, location, description, organization_notes,
  status, revision, locked, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, 1, 0, ?, ?)
`, id, input.Designation, input.Date, input.EndDate, input.Location, input.Description,
		input.OrganizationNotes, input.Status, now, now); err != nil {
		return fmt.Errorf("create exhibition list: %w", err)
	}
	return nil
}
func (s *Store) UpdateList(ctx context.Context, write exhibition.ListWrite) (bool, error) {
	input, id, now := write.Input, write.ID, write.At
	result, err := s.db.ExecContext(ctx, `
UPDATE exhibition_lists
SET designation=?, list_date=?, end_date=?, location=?, description=?, organization_notes=?,
    status=?, revision=revision+1, updated_at=?
WHERE id=? AND (?=0 OR revision=?)
`, input.Designation, input.Date, input.EndDate, input.Location, input.Description,
		input.OrganizationNotes, input.Status, now, id, input.ExpectedRevision, input.ExpectedRevision)
	if err != nil {
		return false, fmt.Errorf("update exhibition list: %w", err)
	}
	return changed(result), nil
}
func (s *Store) DeleteList(ctx context.Context, id string) (bool, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM exhibition_lists WHERE id=?`, id)
	if err != nil {
		return false, fmt.Errorf("delete exhibition list: %w", err)
	}
	return changed(result), nil
}
func (s *Store) SetStatus(ctx context.Context, change exhibition.StatusChange) (bool, error) {
	result, err := s.db.ExecContext(ctx, `
UPDATE exhibition_lists
SET status=?, locked=?, lock_reason=?, locked_at=?, completed_at=?, archived_at=?,
    revision=revision+1, updated_at=?
WHERE id=? AND (?=0 OR revision=?)
`, change.Status, boolToInt(change.Locked), change.Reason, change.LockedAt, change.CompletedAt, change.ArchivedAt, change.At,
		change.ID, change.ExpectedRevision, change.ExpectedRevision)
	if err != nil {
		return false, fmt.Errorf("set exhibition status: %w", err)
	}
	return changed(result), nil
}
