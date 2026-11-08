package exhibitiondata

import (
	"context"
	"fmt"

	"railkeeper/backend/internal/exhibition"
)

func (s *Store) Exceptions(ctx context.Context, listID string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT conflict_key, reason FROM exhibition_conflict_exceptions WHERE list_id=?
`, listID)
	if err != nil {
		return nil, fmt.Errorf("list exhibition conflict exceptions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	exceptions := map[string]string{}
	for rows.Next() {
		var key, reason string
		if err := rows.Scan(&key, &reason); err != nil {
			return nil, fmt.Errorf("scan exhibition conflict exception: %w", err)
		}
		exceptions[key] = reason
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exhibition conflict exceptions: %w", err)
	}
	return exceptions, nil
}
func (m *mutation) UpsertException(ctx context.Context, write exhibition.ExceptionWrite) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO exhibition_conflict_exceptions(id, list_id, conflict_key, reason, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?)
ON CONFLICT(list_id, conflict_key) DO UPDATE SET reason=excluded.reason, updated_at=excluded.updated_at
`, write.ID, write.ListID, write.ConflictID, write.Reason, write.At, write.At); err != nil {
		return fmt.Errorf("save exhibition conflict exception: %w", err)
	}
	return nil
}
func (m *mutation) BumpRevision(ctx context.Context, listID string, expected int, now string) (bool, error) {
	result, err := m.tx.ExecContext(ctx, `
UPDATE exhibition_lists SET revision=revision+1, updated_at=? WHERE id=? AND (?=0 OR revision=?)
`, now, listID, expected, expected)
	if err != nil {
		return false, fmt.Errorf("update exhibition revision: %w", err)
	}
	return changed(result), nil
}
