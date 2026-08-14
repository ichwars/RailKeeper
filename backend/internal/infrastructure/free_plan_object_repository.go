package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

const freePlanObjectSelect = `
SELECT id, lineage_id, revision_id, name, category, position_x_mm, position_y_mm,
       rotation_degrees, shape_json, version, created_at, updated_at
FROM plan_free_objects
`

func (repository *TrackPlannerRepository) CreateFreeObject(
	ctx context.Context,
	revisionID string,
	input application.CreateFreePlanObjectInput,
	actor string,
) (*domain.PlanFreeObject, error) {
	shapeJSON, err := encodeFreePlanObjectShape(input.Shape)
	if err != nil {
		return nil, application.ErrTrackPlanValidation
	}
	now := timestamp()
	object := &domain.PlanFreeObject{
		ID: randomID(), RevisionID: revisionID, Name: input.Name, Category: input.Category,
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees, Shape: input.Shape,
		Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	object.LineageID = object.ID
	err = repository.withTx(ctx, func(tx *sql.Tx) error {
		status, err := trackRevisionStatus(ctx, tx, revisionID)
		if err != nil {
			return err
		}
		if status != domain.PlanRevisionDraft {
			return application.ErrTrackPlanImmutable
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO plan_free_objects(
  id, lineage_id, revision_id, name, category, position_x_mm, position_y_mm,
  rotation_degrees, shape_json, version, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`, object.ID, object.LineageID, revisionID,
			input.Name, input.Category, input.PositionXMM, input.PositionYMM, input.RotationDegrees,
			shapeJSON, now, now); err != nil {
			return fmt.Errorf("insert free plan object: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "PlanFreeObjectCreated", "plan_free_object", object.ID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return repository.getFreeObject(ctx, object.ID)
}

func (repository *TrackPlannerRepository) UpdateFreeObject(
	ctx context.Context,
	id string,
	input application.UpdateFreePlanObjectInput,
	actor string,
) (*domain.PlanFreeObject, error) {
	shapeJSON, err := encodeFreePlanObjectShape(input.Shape)
	if err != nil {
		return nil, application.ErrTrackPlanValidation
	}
	now := timestamp()
	err = repository.withTx(ctx, func(tx *sql.Tx) error {
		status, version, err := freePlanObjectState(ctx, tx, id)
		if err != nil {
			return err
		}
		if status != domain.PlanRevisionDraft {
			return application.ErrTrackPlanImmutable
		}
		if version != input.ExpectedVersion {
			return application.ErrTrackPlanConflict
		}
		result, err := tx.ExecContext(ctx, `
UPDATE plan_free_objects
SET name=?, category=?, position_x_mm=?, position_y_mm=?, rotation_degrees=?, shape_json=?,
    version=version+1, updated_at=?
WHERE id=? AND version=?`, input.Name, input.Category, input.PositionXMM, input.PositionYMM,
			input.RotationDegrees, shapeJSON, now, id, input.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("update free plan object: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("read free plan object update result: %w", err)
		}
		if affected != 1 {
			return application.ErrTrackPlanConflict
		}
		return writeLayoutAudit(ctx, tx, "PlanFreeObjectUpdated", "plan_free_object", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return repository.getFreeObject(ctx, id)
}

func (repository *TrackPlannerRepository) DeleteFreeObject(
	ctx context.Context,
	id string,
	expectedVersion int,
	actor string,
) error {
	now := timestamp()
	return repository.withTx(ctx, func(tx *sql.Tx) error {
		status, version, err := freePlanObjectState(ctx, tx, id)
		if err != nil {
			return err
		}
		if status != domain.PlanRevisionDraft {
			return application.ErrTrackPlanImmutable
		}
		if version != expectedVersion {
			return application.ErrTrackPlanConflict
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM plan_free_objects WHERE id=? AND version=?`,
			id, expectedVersion); err != nil {
			return fmt.Errorf("delete free plan object: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "PlanFreeObjectDeleted", "plan_free_object", id, actor, now)
	})
}

func (repository *TrackPlannerRepository) GetPlanForFreeObject(
	ctx context.Context,
	id string,
) (*application.TrackPlan, error) {
	var revisionID string
	err := repository.db.QueryRowContext(ctx,
		`SELECT revision_id FROM plan_free_objects WHERE id=?`, id).Scan(&revisionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrTrackPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read free plan object revision: %w", err)
	}
	return repository.GetPlan(ctx, revisionID)
}

func (repository *TrackPlannerRepository) getFreeObject(
	ctx context.Context,
	id string,
) (*domain.PlanFreeObject, error) {
	object, err := scanFreePlanObject(repository.db.QueryRowContext(ctx, freePlanObjectSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrTrackPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get free plan object: %w", err)
	}
	return object, nil
}

func listFreePlanObjects(
	ctx context.Context,
	db *sql.DB,
	revisionID string,
) ([]domain.PlanFreeObject, error) {
	rows, err := db.QueryContext(ctx, freePlanObjectSelect+` WHERE revision_id=? ORDER BY created_at, id`, revisionID)
	if err != nil {
		return nil, fmt.Errorf("list free plan objects: %w", err)
	}
	defer func() { _ = rows.Close() }()
	objects := []domain.PlanFreeObject{}
	for rows.Next() {
		object, err := scanFreePlanObject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan free plan object: %w", err)
		}
		objects = append(objects, *object)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate free plan objects: %w", err)
	}
	return objects, nil
}

func scanFreePlanObject(scanner trackScanner) (*domain.PlanFreeObject, error) {
	object := &domain.PlanFreeObject{}
	var shapeJSON string
	if err := scanner.Scan(&object.ID, &object.LineageID, &object.RevisionID, &object.Name,
		&object.Category, &object.PositionXMM, &object.PositionYMM, &object.RotationDegrees,
		&shapeJSON, &object.Version, &object.CreatedAt, &object.UpdatedAt); err != nil {
		return nil, err
	}
	if !object.Category.Valid() {
		return nil, fmt.Errorf("decode free plan object %s: invalid category", object.ID)
	}
	if err := json.Unmarshal([]byte(shapeJSON), &object.Shape); err != nil {
		return nil, fmt.Errorf("decode free plan object %s: %w", object.ID, err)
	}
	if err := domain.ValidateFreePlanObjectShape(object.Shape); err != nil {
		return nil, fmt.Errorf("decode free plan object %s: %w", object.ID, err)
	}
	return object, nil
}

func encodeFreePlanObjectShape(shape domain.FreePlanObjectShape) (string, error) {
	if err := domain.ValidateFreePlanObjectShape(shape); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(shape)
	if err != nil {
		return "", fmt.Errorf("encode free plan object shape: %w", err)
	}
	return string(encoded), nil
}

func freePlanObjectState(
	ctx context.Context,
	tx *sql.Tx,
	id string,
) (domain.PlanRevisionStatus, int, error) {
	var status domain.PlanRevisionStatus
	var version int
	err := tx.QueryRowContext(ctx, `
SELECT revision.status, object.version
FROM plan_free_objects object
JOIN plan_revisions revision ON revision.id=object.revision_id
WHERE object.id=?`, id).Scan(&status, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, application.ErrTrackPlanNotFound
	}
	if err != nil {
		return "", 0, fmt.Errorf("read free plan object state: %w", err)
	}
	return status, version, nil
}
