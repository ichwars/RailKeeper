package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"railkeeper/backend/internal/application"
)

func (r *AccessoryRepository) ListLocations(ctx context.Context) ([]application.StorageLocation, error) {
	rows, err := r.db.QueryContext(ctx, storageLocationSelect+` ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("list storage locations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	locations := []application.StorageLocation{}
	for rows.Next() {
		location, err := scanStorageLocation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan storage location: %w", err)
		}
		locations = append(locations, *location)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate storage locations: %w", err)
	}
	return locations, nil
}

func (r *AccessoryRepository) CreateLocation(
	ctx context.Context,
	input application.CreateStorageLocationInput,
	actor string,
) (*application.StorageLocation, error) {
	now := timestamp()
	location := &application.StorageLocation{
		ID: randomID(), ParentID: input.ParentID, Name: input.Name, Description: input.Description,
		Archived: input.Archived, CreatedAt: now, UpdatedAt: now,
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if err := requireStorageLocation(ctx, tx, input.ParentID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO storage_locations(id, parent_id, name, description, archived, created_at, updated_at)
VALUES(?, NULLIF(?, ''), ?, ?, ?, ?, ?)`, location.ID, location.ParentID, location.Name,
			location.Description, boolToInt(location.Archived), now, now); err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("insert storage location: %w", err)
		}
		return writeAccessoryAudit(ctx, tx, "StorageLocationCreated", "storage_location", location.ID, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	return location, nil
}

func (r *AccessoryRepository) UpdateLocation(
	ctx context.Context,
	id string,
	input application.UpdateStorageLocationInput,
	actor string,
) (*application.StorageLocation, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		exists, err := accessoryRecordExists(
			ctx,
			tx,
			`SELECT COUNT(*) FROM storage_locations WHERE id=?`,
			id,
		)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrAccessoryNotFound
		}
		if err := requireStorageLocation(ctx, tx, input.ParentID); err != nil {
			return err
		}
		if input.ParentID != "" {
			var cycleCount int
			if err := tx.QueryRowContext(ctx, `
WITH RECURSIVE descendants(id) AS (
  SELECT id FROM storage_locations WHERE parent_id=?
  UNION ALL
  SELECT location.id FROM storage_locations location JOIN descendants ON location.parent_id=descendants.id
)
SELECT COUNT(*) FROM descendants WHERE id=?`, id, input.ParentID).Scan(&cycleCount); err != nil {
				return fmt.Errorf("check storage location cycle: %w", err)
			}
			if cycleCount > 0 {
				return application.ErrAccessoryValidation
			}
		}
		result, err := tx.ExecContext(ctx, `
UPDATE storage_locations SET parent_id=NULLIF(?, ''), name=?, description=?, archived=?, updated_at=? WHERE id=?`,
			input.ParentID, input.Name, input.Description, boolToInt(input.Archived), now, id)
		if err != nil {
			if isSQLiteConstraint(err) {
				return application.ErrAccessoryConflict
			}
			return fmt.Errorf("update storage location: %w", err)
		}
		if err := requireAccessoryUpdated(result); err != nil {
			return err
		}
		return writeAccessoryAudit(ctx, tx, "StorageLocationUpdated", "storage_location", id, actor, now, "{}")
	})
	if err != nil {
		return nil, err
	}
	location, err := scanStorageLocation(r.db.QueryRowContext(ctx, storageLocationSelect+` WHERE id=?`, id))
	if err != nil {
		return nil, fmt.Errorf("get updated storage location: %w", err)
	}
	return location, nil
}

const storageLocationSelect = `SELECT id, COALESCE(parent_id, ''), name, description, archived, created_at, updated_at FROM storage_locations`

func scanStorageLocation(scanner rowScanner) (*application.StorageLocation, error) {
	location := &application.StorageLocation{}
	var archived int
	err := scanner.Scan(&location.ID, &location.ParentID, &location.Name, &location.Description,
		&archived, &location.CreatedAt, &location.UpdatedAt)
	location.Archived = archived != 0
	return location, err
}

func requireStorageLocation(ctx context.Context, tx *sql.Tx, id string) error {
	if id == "" {
		return nil
	}
	exists, err := accessoryRecordExists(
		ctx,
		tx,
		`SELECT COUNT(*) FROM storage_locations WHERE id=?`,
		id,
	)
	if err != nil {
		return err
	}
	if !exists {
		return application.ErrAccessoryNotFound
	}
	return nil
}
