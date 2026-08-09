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

type TrackPlannerRepository struct {
	db *sql.DB
}

func NewTrackPlannerRepository(db *sql.DB) *TrackPlannerRepository {
	return &TrackPlannerRepository{db: db}
}

func (repository *TrackPlannerRepository) ListGeometries(
	ctx context.Context,
	gauge string,
) ([]domain.TrackGeometryDefinition, error) {
	rows, err := repository.db.QueryContext(ctx, trackGeometrySelect+`
WHERE library.gauge=? AND library.status='verified' AND geometry.status='verified'
ORDER BY geometry.article_number COLLATE NOCASE, geometry.id`, gauge)
	if err != nil {
		return nil, fmt.Errorf("list track geometries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	geometries := []domain.TrackGeometryDefinition{}
	for rows.Next() {
		geometry, err := scanTrackGeometry(rows)
		if err != nil {
			return nil, fmt.Errorf("scan track geometry: %w", err)
		}
		geometries = append(geometries, *geometry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate track geometries: %w", err)
	}
	return geometries, nil
}

func (repository *TrackPlannerRepository) GetPlan(
	ctx context.Context,
	revisionID string,
) (*application.TrackPlan, error) {
	plan := &application.TrackPlan{RevisionID: revisionID, Objects: []domain.PlanTrackObject{}}
	if err := repository.db.QueryRowContext(ctx, `SELECT status FROM plan_revisions WHERE id=?`, revisionID).
		Scan(&plan.Status); errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrTrackPlanNotFound
	} else if err != nil {
		return nil, fmt.Errorf("read track plan revision: %w", err)
	}
	rows, err := repository.db.QueryContext(ctx, trackObjectSelect+`
WHERE object.revision_id=? ORDER BY object.created_at, object.id`, revisionID)
	if err != nil {
		return nil, fmt.Errorf("list track plan objects: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		object, err := scanTrackObject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan track plan object: %w", err)
		}
		plan.Objects = append(plan.Objects, *object)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate track plan objects: %w", err)
	}
	return plan, nil
}

func (repository *TrackPlannerRepository) GetPlanForObject(
	ctx context.Context,
	objectID string,
) (*application.TrackPlan, error) {
	var revisionID string
	err := repository.db.QueryRowContext(ctx,
		`SELECT revision_id FROM plan_track_objects WHERE id=?`, objectID).Scan(&revisionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrTrackPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read track object revision: %w", err)
	}
	return repository.GetPlan(ctx, revisionID)
}

func (repository *TrackPlannerRepository) TrackMaterialAvailability(
	ctx context.Context,
	bom []domain.TrackBOMLine,
) ([]application.TrackMaterialStatus, error) {
	materials := make([]application.TrackMaterialStatus, 0, len(bom))
	for _, line := range bom {
		material := application.TrackMaterialStatus{
			GeometryID: line.GeometryID, ArticleNumber: line.ArticleNumber, Name: line.Name,
			RequiredQuantity: line.Quantity, ProductIDs: []string{}, InventoryNumbers: []string{},
		}
		err := repository.db.QueryRowContext(ctx, `
SELECT library.manufacturer, geometry.article_number, geometry.name
FROM track_geometry_definitions geometry
JOIN track_geometry_libraries library ON library.id=geometry.library_id
WHERE geometry.id=?`, line.GeometryID).
			Scan(&material.Manufacturer, &material.ArticleNumber, &material.Name)
		if errors.Is(err, sql.ErrNoRows) {
			material.MissingQuantity = material.RequiredQuantity
			materials = append(materials, material)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read track material geometry %s: %w", line.GeometryID, err)
		}

		rows, err := repository.db.QueryContext(ctx, `
SELECT product.id, product.inventory_number,
       COALESCE((SELECT SUM(stock.quantity) FROM accessory_stock stock
                 WHERE stock.product_id=product.id), 0),
       COALESCE((SELECT SUM(reservation.quantity) FROM accessory_reservations reservation
                 WHERE reservation.product_id=product.id AND reservation.status='active'), 0)
FROM accessory_products product
WHERE product.archived=0
  AND lower(trim(product.manufacturer))=lower(trim(?))
  AND lower(trim(product.article_number))=lower(trim(?))
ORDER BY product.inventory_number COLLATE NOCASE, product.id`, material.Manufacturer, material.ArticleNumber)
		if err != nil {
			return nil, fmt.Errorf("list track material products %s: %w", line.GeometryID, err)
		}
		for rows.Next() {
			var productID, inventoryNumber string
			var physical, reserved int
			if err := rows.Scan(&productID, &inventoryNumber, &physical, &reserved); err != nil {
				_ = rows.Close()
				return nil, fmt.Errorf("scan track material product %s: %w", line.GeometryID, err)
			}
			material.ProductIDs = append(material.ProductIDs, productID)
			material.InventoryNumbers = append(material.InventoryNumbers, inventoryNumber)
			material.PhysicalQuantity += physical
			material.ReservedQuantity += reserved
		}
		if err := rows.Close(); err != nil {
			return nil, fmt.Errorf("close track material products %s: %w", line.GeometryID, err)
		}
		material.AvailableQuantity = max(0, material.PhysicalQuantity-material.ReservedQuantity)
		material.MissingQuantity = max(0, material.RequiredQuantity-material.AvailableQuantity)
		materials = append(materials, material)
	}
	return materials, nil
}

func (repository *TrackPlannerRepository) CreateObject(
	ctx context.Context,
	revisionID string,
	input application.CreatePlanTrackObjectInput,
	actor string,
) (*domain.PlanTrackObject, error) {
	now := timestamp()
	object := &domain.PlanTrackObject{
		ID: randomID(), RevisionID: revisionID, GeometryID: input.GeometryID,
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	err := repository.withTx(ctx, func(tx *sql.Tx) error {
		status, err := trackRevisionStatus(ctx, tx, revisionID)
		if err != nil {
			return err
		}
		if status != domain.PlanRevisionDraft {
			return application.ErrTrackPlanImmutable
		}
		placeable, err := trackGeometryPlaceable(ctx, tx, input.GeometryID)
		if err != nil {
			return err
		}
		if !placeable {
			return application.ErrTrackPlanValidation
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO plan_track_objects(
  id, revision_id, geometry_id, position_x_mm, position_y_mm, rotation_degrees,
  version, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, 1, ?, ?)`, object.ID, revisionID, input.GeometryID,
			input.PositionXMM, input.PositionYMM, input.RotationDegrees, now, now); err != nil {
			return fmt.Errorf("insert track plan object: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "PlanTrackObjectCreated", "plan_track_object", object.ID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return repository.getObject(ctx, object.ID)
}

func (repository *TrackPlannerRepository) UpdateObject(
	ctx context.Context,
	id string,
	input application.UpdatePlanTrackObjectInput,
	actor string,
) (*domain.PlanTrackObject, error) {
	now := timestamp()
	err := repository.withTx(ctx, func(tx *sql.Tx) error {
		status, version, err := trackObjectState(ctx, tx, id)
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
UPDATE plan_track_objects
SET position_x_mm=?, position_y_mm=?, rotation_degrees=?, version=version+1, updated_at=?
WHERE id=? AND version=?`, input.PositionXMM, input.PositionYMM, input.RotationDegrees,
			now, id, input.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("update track plan object: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("read track object update result: %w", err)
		}
		if affected != 1 {
			return application.ErrTrackPlanConflict
		}
		return writeLayoutAudit(ctx, tx, "PlanTrackObjectUpdated", "plan_track_object", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return repository.getObject(ctx, id)
}

func (repository *TrackPlannerRepository) DeleteObject(
	ctx context.Context,
	id string,
	expectedVersion int,
	actor string,
) error {
	now := timestamp()
	return repository.withTx(ctx, func(tx *sql.Tx) error {
		status, version, err := trackObjectState(ctx, tx, id)
		if err != nil {
			return err
		}
		if status != domain.PlanRevisionDraft {
			return application.ErrTrackPlanImmutable
		}
		if version != expectedVersion {
			return application.ErrTrackPlanConflict
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM plan_track_objects WHERE id=? AND version=?`,
			id, expectedVersion); err != nil {
			return fmt.Errorf("delete track plan object: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "PlanTrackObjectDeleted", "plan_track_object", id, actor, now)
	})
}

func (repository *TrackPlannerRepository) getObject(
	ctx context.Context,
	id string,
) (*domain.PlanTrackObject, error) {
	object, err := scanTrackObject(repository.db.QueryRowContext(ctx, trackObjectSelect+` WHERE object.id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrTrackPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get track plan object: %w", err)
	}
	return object, nil
}

func (repository *TrackPlannerRepository) withTx(ctx context.Context, work func(*sql.Tx) error) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin track planner transaction: %w", err)
	}
	if err := work(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit track planner transaction: %w", err)
	}
	return nil
}

func trackRevisionStatus(
	ctx context.Context,
	tx *sql.Tx,
	revisionID string,
) (domain.PlanRevisionStatus, error) {
	var status domain.PlanRevisionStatus
	if err := tx.QueryRowContext(ctx, `SELECT status FROM plan_revisions WHERE id=?`, revisionID).
		Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return "", application.ErrTrackPlanNotFound
	} else if err != nil {
		return "", fmt.Errorf("read track plan revision: %w", err)
	}
	return status, nil
}

func trackGeometryPlaceable(ctx context.Context, tx *sql.Tx, geometryID string) (bool, error) {
	var geometryStatus, libraryStatus domain.TrackGeometryStatus
	err := tx.QueryRowContext(ctx, `
SELECT geometry.status, library.status
FROM track_geometry_definitions geometry
JOIN track_geometry_libraries library ON library.id=geometry.library_id
WHERE geometry.id=?`, geometryID).Scan(&geometryStatus, &libraryStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read track geometry status: %w", err)
	}
	return geometryStatus.Placeable() && libraryStatus.Placeable(), nil
}

func trackObjectState(
	ctx context.Context,
	tx *sql.Tx,
	id string,
) (domain.PlanRevisionStatus, int, error) {
	var status domain.PlanRevisionStatus
	var version int
	err := tx.QueryRowContext(ctx, `
SELECT revision.status, object.version
FROM plan_track_objects object
JOIN plan_revisions revision ON revision.id=object.revision_id
WHERE object.id=?`, id).Scan(&status, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, application.ErrTrackPlanNotFound
	}
	if err != nil {
		return "", 0, fmt.Errorf("read track plan object state: %w", err)
	}
	return status, version, nil
}

const trackGeometrySelect = `
SELECT geometry.id, geometry.library_id, geometry.article_number, geometry.name, geometry.kind,
       geometry.length_mm, geometry.geometry_json, geometry.source_url, geometry.status,
       geometry.created_at
FROM track_geometry_definitions geometry
JOIN track_geometry_libraries library ON library.id=geometry.library_id
`

const trackObjectSelect = `
SELECT object.id, object.revision_id, object.geometry_id,
       object.position_x_mm, object.position_y_mm, object.rotation_degrees,
       object.version, object.created_at, object.updated_at,
       geometry.id, geometry.library_id, geometry.article_number, geometry.name, geometry.kind,
       geometry.length_mm, geometry.geometry_json, geometry.source_url, geometry.status,
       geometry.created_at
FROM plan_track_objects object
JOIN track_geometry_definitions geometry ON geometry.id=object.geometry_id
`

type trackScanner interface {
	Scan(...any) error
}

func scanTrackGeometry(scanner trackScanner) (*domain.TrackGeometryDefinition, error) {
	geometry := &domain.TrackGeometryDefinition{}
	var geometryJSON string
	if err := scanner.Scan(&geometry.ID, &geometry.LibraryID, &geometry.ArticleNumber, &geometry.Name,
		&geometry.Kind, &geometry.LengthMM, &geometryJSON, &geometry.SourceURL, &geometry.Status,
		&geometry.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(geometryJSON), &geometry.Geometry); err != nil {
		return nil, fmt.Errorf("decode track geometry %s: %w", geometry.ID, err)
	}
	return geometry, nil
}

func scanTrackObject(scanner trackScanner) (*domain.PlanTrackObject, error) {
	object := &domain.PlanTrackObject{}
	var geometryJSON string
	geometry := &object.Geometry
	if err := scanner.Scan(&object.ID, &object.RevisionID, &object.GeometryID,
		&object.PositionXMM, &object.PositionYMM, &object.RotationDegrees,
		&object.Version, &object.CreatedAt, &object.UpdatedAt,
		&geometry.ID, &geometry.LibraryID, &geometry.ArticleNumber, &geometry.Name,
		&geometry.Kind, &geometry.LengthMM, &geometryJSON, &geometry.SourceURL, &geometry.Status,
		&geometry.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(geometryJSON), &geometry.Geometry); err != nil {
		return nil, fmt.Errorf("decode track geometry %s: %w", geometry.ID, err)
	}
	return object, nil
}
