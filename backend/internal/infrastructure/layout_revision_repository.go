package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func (r *LayoutRepository) ListVariants(ctx context.Context, unitID string) ([]application.PlanVariant, error) {
	rows, err := r.db.QueryContext(ctx, planVariantSelect+` WHERE layout_unit_id=? ORDER BY name COLLATE NOCASE, id`, unitID)
	if err != nil {
		return nil, fmt.Errorf("list plan variants: %w", err)
	}
	variants := []application.PlanVariant{}
	for rows.Next() {
		variant, err := scanPlanVariant(rows)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan plan variant: %w", err)
		}
		variants = append(variants, *variant)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate plan variants: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close plan variants: %w", err)
	}
	for index := range variants {
		revisions, err := r.listPlanRevisions(ctx, variants[index].ID)
		if err != nil {
			return nil, err
		}
		variants[index].Revisions = revisions
	}
	return variants, nil
}

func (r *LayoutRepository) CreateVariant(
	ctx context.Context,
	unitID string,
	input application.CreatePlanVariantInput,
	actor string,
) (*application.PlanVariant, error) {
	now := timestamp()
	variant := &application.PlanVariant{
		ID: randomID(), LayoutUnitID: unitID, Name: input.Name, Description: input.Description,
		Revisions: []application.PlanRevision{}, CreatedAt: now, UpdatedAt: now,
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		exists, err := recordExists(ctx, tx, "layout_units", unitID)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrLayoutNotFound
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO plan_variants(id, layout_unit_id, name, description, archived, created_at, updated_at)
VALUES(?, ?, ?, ?, 0, ?, ?)`, variant.ID, unitID, variant.Name, variant.Description, now, now); err != nil {
			return fmt.Errorf("insert plan variant: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "PlanVariantCreated", "plan_variant", variant.ID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return variant, nil
}

func (r *LayoutRepository) CreateDraft(
	ctx context.Context,
	variantID string,
	input application.CreatePlanRevisionInput,
	actor string,
) (*application.PlanRevision, error) {
	now := timestamp()
	revision := &application.PlanRevision{
		ID: randomID(), VariantID: variantID, Status: domain.PlanRevisionDraft,
		BaseRevisionID: input.BaseRevisionID, Version: 1, CreatedBy: actor, CreatedAt: now, UpdatedAt: now,
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		exists, err := planVariantExists(ctx, tx, variantID)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrLayoutNotFound
		}
		if input.BaseRevisionID != "" {
			var status domain.PlanRevisionStatus
			err := tx.QueryRowContext(ctx, `
SELECT status FROM plan_revisions WHERE id=? AND variant_id=?`, input.BaseRevisionID, variantID).Scan(&status)
			if errors.Is(err, sql.ErrNoRows) || (err == nil && status != domain.PlanRevisionPublished && status != domain.PlanRevisionArchived) {
				return application.ErrLayoutValidation
			}
			if err != nil {
				return fmt.Errorf("validate base plan revision: %w", err)
			}
		}
		if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(MAX(revision_number), 0) + 1 FROM plan_revisions WHERE variant_id=?`, variantID).
			Scan(&revision.RevisionNumber); err != nil {
			return fmt.Errorf("allocate plan revision number: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO plan_revisions(
  id, variant_id, revision_number, status, base_revision_id, version, created_by,
  published_by, published_at, created_at, updated_at
) VALUES(?, ?, ?, ?, NULLIF(?, ''), 1, ?, NULL, NULL, ?, ?)`, revision.ID, variantID,
			revision.RevisionNumber, revision.Status, revision.BaseRevisionID, actor, now, now); err != nil {
			return fmt.Errorf("insert plan draft: %w", err)
		}
		if input.BaseRevisionID != "" {
			rows, err := tx.QueryContext(ctx, `
SELECT geometry_id, position_x_mm, position_y_mm, rotation_degrees,
       elevation_start_mm, elevation_end_mm, lineage_id
FROM plan_track_objects WHERE revision_id=? ORDER BY created_at, id`, input.BaseRevisionID)
			if err != nil {
				return fmt.Errorf("list base revision track objects: %w", err)
			}
			type baseTrackObject struct {
				geometryID       string
				positionXMM      float64
				positionYMM      float64
				rotationDegrees  float64
				elevationStartMM float64
				elevationEndMM   float64
				lineageID        string
			}
			objects := []baseTrackObject{}
			for rows.Next() {
				object := baseTrackObject{}
				if err := rows.Scan(&object.geometryID, &object.positionXMM, &object.positionYMM,
					&object.rotationDegrees, &object.elevationStartMM, &object.elevationEndMM,
					&object.lineageID); err != nil {
					_ = rows.Close()
					return fmt.Errorf("scan base revision track object: %w", err)
				}
				objects = append(objects, object)
			}
			if err := rows.Err(); err != nil {
				_ = rows.Close()
				return fmt.Errorf("iterate base revision track objects: %w", err)
			}
			if err := rows.Close(); err != nil {
				return fmt.Errorf("close base revision track objects: %w", err)
			}
			for _, object := range objects {
				if _, err := tx.ExecContext(ctx, `
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
	elevation_start_mm, elevation_end_mm, lineage_id, version, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`, randomID(), revision.ID, object.geometryID,
					object.positionXMM, object.positionYMM, object.rotationDegrees,
					object.elevationStartMM, object.elevationEndMM, object.lineageID, now, now); err != nil {
					return fmt.Errorf("copy base revision track object: %w", err)
				}
			}
		}
		return writeLayoutAudit(ctx, tx, "PlanDraftCreated", "plan_revision", revision.ID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return revision, nil
}

func (r *LayoutRepository) SubmitRevision(
	ctx context.Context,
	id string,
	expectedVersion int,
	actor string,
) (*application.PlanRevision, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		status, version, err := revisionState(ctx, tx, id)
		if err != nil {
			return err
		}
		if status == domain.PlanRevisionPublished || status == domain.PlanRevisionArchived {
			return application.ErrPlanRevisionImmutable
		}
		if version != expectedVersion {
			return application.ErrPlanRevisionConflict
		}
		if status != domain.PlanRevisionDraft {
			return application.ErrPlanRevisionConflict
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE plan_revisions SET status='review', version=version+1, updated_at=? WHERE id=? AND version=?`,
			now, id, expectedVersion); err != nil {
			return fmt.Errorf("submit plan revision: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "PlanRevisionSubmitted", "plan_revision", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return r.getPlanRevision(ctx, id)
}

func (r *LayoutRepository) PublishRevision(
	ctx context.Context,
	id string,
	expectedVersion int,
	actor string,
) (*application.PlanRevision, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		status, version, err := revisionState(ctx, tx, id)
		if err != nil {
			return err
		}
		if status == domain.PlanRevisionPublished || status == domain.PlanRevisionArchived {
			return application.ErrPlanRevisionImmutable
		}
		if version != expectedVersion {
			return application.ErrPlanRevisionConflict
		}
		var variantID string
		if err := tx.QueryRowContext(ctx, `SELECT variant_id FROM plan_revisions WHERE id=?`, id).Scan(&variantID); err != nil {
			return fmt.Errorf("read plan revision variant: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
UPDATE plan_revisions
SET status='archived', version=version+1, updated_at=?
WHERE variant_id=? AND status='published' AND id<>?`, now, variantID, id); err != nil {
			return fmt.Errorf("archive published plan revision: %w", err)
		}
		result, err := tx.ExecContext(ctx, `
UPDATE plan_revisions
SET status='published', published_by=?, published_at=?, version=version+1, updated_at=?
WHERE id=? AND version=? AND status IN ('draft', 'review')`, actor, now, now, id, expectedVersion)
		if err != nil {
			return fmt.Errorf("publish plan revision: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("read publish result: %w", err)
		}
		if affected != 1 {
			return application.ErrPlanRevisionConflict
		}
		return writeLayoutAudit(ctx, tx, "PlanRevisionPublished", "plan_revision", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return r.getPlanRevision(ctx, id)
}

const planVariantSelect = `SELECT id, layout_unit_id, name, description, archived, created_at, updated_at FROM plan_variants`

const planRevisionSelect = `SELECT id, variant_id, revision_number, status, COALESCE(base_revision_id, ''), version, created_by, COALESCE(published_by, ''), COALESCE(published_at, ''), created_at, updated_at FROM plan_revisions`

func scanPlanVariant(scanner rowScanner) (*application.PlanVariant, error) {
	variant := &application.PlanVariant{}
	var archived int
	err := scanner.Scan(&variant.ID, &variant.LayoutUnitID, &variant.Name, &variant.Description, &archived,
		&variant.CreatedAt, &variant.UpdatedAt)
	variant.Archived = archived != 0
	return variant, err
}

func scanPlanRevision(scanner rowScanner) (*application.PlanRevision, error) {
	revision := &application.PlanRevision{}
	err := scanner.Scan(&revision.ID, &revision.VariantID, &revision.RevisionNumber, &revision.Status,
		&revision.BaseRevisionID, &revision.Version, &revision.CreatedBy, &revision.PublishedBy,
		&revision.PublishedAt, &revision.CreatedAt, &revision.UpdatedAt)
	return revision, err
}

func (r *LayoutRepository) listPlanRevisions(ctx context.Context, variantID string) ([]application.PlanRevision, error) {
	rows, err := r.db.QueryContext(ctx, planRevisionSelect+` WHERE variant_id=? ORDER BY revision_number, id`, variantID)
	if err != nil {
		return nil, fmt.Errorf("list plan revisions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	revisions := []application.PlanRevision{}
	for rows.Next() {
		revision, err := scanPlanRevision(rows)
		if err != nil {
			return nil, fmt.Errorf("scan plan revision: %w", err)
		}
		revisions = append(revisions, *revision)
	}
	return revisions, rows.Err()
}

func (r *LayoutRepository) getPlanRevision(ctx context.Context, id string) (*application.PlanRevision, error) {
	revision, err := scanPlanRevision(r.db.QueryRowContext(ctx, planRevisionSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get plan revision: %w", err)
	}
	return revision, nil
}

func planVariantExists(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM plan_variants WHERE id=?`, id).Scan(&count); err != nil {
		return false, fmt.Errorf("check plan variant: %w", err)
	}
	return count > 0, nil
}

func revisionState(ctx context.Context, tx *sql.Tx, id string) (domain.PlanRevisionStatus, int, error) {
	var status domain.PlanRevisionStatus
	var version int
	if err := tx.QueryRowContext(ctx, `SELECT status, version FROM plan_revisions WHERE id=?`, id).
		Scan(&status, &version); errors.Is(err, sql.ErrNoRows) {
		return "", 0, application.ErrLayoutNotFound
	} else if err != nil {
		return "", 0, fmt.Errorf("read plan revision state: %w", err)
	}
	return status, version, nil
}
