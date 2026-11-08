package exhibitiondata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/exhibition"
)

func (s *Store) Lists(ctx context.Context) ([]exhibition.ExhibitionList, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT l.id, l.designation, l.list_date, l.end_date, l.location, l.description,
       l.organization_notes, l.status, l.revision, l.locked, l.lock_reason,
       l.locked_at, l.completed_at, l.archived_at, COUNT(e.id), l.created_at, l.updated_at
FROM exhibition_lists l
LEFT JOIN exhibition_entries e ON e.list_id = l.id
GROUP BY l.id
ORDER BY l.list_date DESC, l.designation COLLATE NOCASE
`)
	if err != nil {
		return nil, fmt.Errorf("list exhibition lists: %w", err)
	}
	defer func() { _ = rows.Close() }()

	lists := []exhibition.ExhibitionList{}
	for rows.Next() {
		var list exhibition.ExhibitionList
		var locked int
		if err := rows.Scan(
			&list.ID, &list.Designation, &list.Date, &list.EndDate, &list.Location, &list.Description,
			&list.OrganizationNotes, &list.Status, &list.Revision, &locked, &list.LockReason,
			&list.LockedAt, &list.CompletedAt, &list.ArchivedAt, &list.EntryCount,
			&list.CreatedAt, &list.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exhibition list: %w", err)
		}
		list.Locked = locked == 1
		lists = append(lists, list)
	}
	return lists, rows.Err()
}
func (s *Store) List(ctx context.Context, id string) (exhibition.ExhibitionList, error) {
	var list exhibition.ExhibitionList
	var locked int
	if err := s.db.QueryRowContext(ctx, `
SELECT id, designation, list_date, end_date, location, description, organization_notes,
       status, revision, locked, lock_reason, locked_at, completed_at, archived_at,
       created_at, updated_at
FROM exhibition_lists WHERE id=?
`, id).Scan(
		&list.ID, &list.Designation, &list.Date, &list.EndDate, &list.Location,
		&list.Description, &list.OrganizationNotes, &list.Status, &list.Revision, &locked,
		&list.LockReason, &list.LockedAt, &list.CompletedAt, &list.ArchivedAt,
		&list.CreatedAt, &list.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exhibition.ExhibitionList{}, exhibition.ErrExhibitionNotFound
		}
		return exhibition.ExhibitionList{}, fmt.Errorf("get exhibition list: %w", err)
	}
	list.Locked = locked == 1
	return list, nil
}
func (s *Store) Entries(ctx context.Context, listID string) ([]exhibition.ExhibitionEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, list_id, owner, image_url, locomotive_name,
       COALESCE(vehicle_id, ''),
       COALESCE(gattung, ''), COALESCE(series, ''), COALESCE(manufacturer, ''), COALESCE(epoch, ''), COALESCE(railway_company, ''),
       COALESCE(day_scope, 'all'),
       dt_decoder, decoder_number, COALESCE(decoder_type, ''), COALESCE(adapter, ''),
       COALESCE(interface_name, ''), COALESCE(sx_address, ''), analog,
       function_keys, notes, sort_order, availability, revision, created_at, updated_at
FROM exhibition_entries
WHERE list_id=?
ORDER BY sort_order, locomotive_name COLLATE NOCASE, owner COLLATE NOCASE
`, listID)
	if err != nil {
		return nil, fmt.Errorf("list exhibition entries: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []exhibition.ExhibitionEntry{}
	for rows.Next() {
		entry, err := scanExhibitionEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
func (s *Store) Entry(ctx context.Context, listID, entryID string) (exhibition.ExhibitionEntry, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, list_id, owner, image_url, locomotive_name,
       COALESCE(vehicle_id, ''),
       COALESCE(gattung, ''), COALESCE(series, ''), COALESCE(manufacturer, ''), COALESCE(epoch, ''), COALESCE(railway_company, ''),
       COALESCE(day_scope, 'all'),
       dt_decoder, decoder_number, COALESCE(decoder_type, ''), COALESCE(adapter, ''),
       COALESCE(interface_name, ''), COALESCE(sx_address, ''), analog,
       function_keys, notes, sort_order, availability, revision, created_at, updated_at
FROM exhibition_entries
WHERE id=? AND list_id=?
`, entryID, listID)
	entry, err := scanExhibitionEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return exhibition.ExhibitionEntry{}, exhibition.ErrExhibitionNotFound
		}
		return exhibition.ExhibitionEntry{}, err
	}
	return entry, nil
}
func (s *Store) NextSort(ctx context.Context, listID string) int {
	var next sql.NullInt64
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order), 0) + 10 FROM exhibition_entries WHERE list_id=?`, listID).Scan(&next)
	if next.Valid && next.Int64 > 0 {
		return int(next.Int64)
	}
	return 10
}
