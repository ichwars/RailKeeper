package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	var maxGradePercent sql.NullFloat64
	var minimumTrackClearanceMM sql.NullFloat64
	if err := repository.db.QueryRowContext(ctx, `
SELECT revision.status, layout.max_grade_percent, layout.minimum_track_clearance_mm
FROM plan_revisions revision
JOIN plan_variants variant ON variant.id=revision.variant_id
JOIN layout_units unit ON unit.id=variant.layout_unit_id
JOIN layouts layout ON layout.id=unit.layout_id
WHERE revision.id=?`, revisionID).Scan(&plan.Status, &maxGradePercent,
		&minimumTrackClearanceMM); errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrTrackPlanNotFound
	} else if err != nil {
		return nil, fmt.Errorf("read track plan revision: %w", err)
	}
	if maxGradePercent.Valid {
		plan.Limits.MaxGradePercent = &maxGradePercent.Float64
	}
	if minimumTrackClearanceMM.Valid {
		plan.Limits.MinimumTrackClearanceMM = &minimumTrackClearanceMM.Float64
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

func (repository *TrackPlannerRepository) GetBaseRevisionID(
	ctx context.Context,
	revisionID string,
) (string, error) {
	var baseRevisionID string
	err := repository.db.QueryRowContext(ctx, `
SELECT COALESCE(base_revision_id, '') FROM plan_revisions WHERE id=?`, revisionID).
		Scan(&baseRevisionID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", application.ErrTrackPlanNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read track plan base revision: %w", err)
	}
	return baseRevisionID, nil
}

func (repository *TrackPlannerRepository) ListAffectedConfigurations(
	ctx context.Context,
	revisionID string,
) ([]application.TrackPlanAffectedConfiguration, error) {
	rows, err := repository.db.QueryContext(ctx, `
SELECT DISTINCT configuration.id, configuration.name
FROM layout_configurations configuration
JOIN layout_configuration_units unit ON unit.configuration_id=configuration.id
WHERE unit.plan_revision_id=? AND configuration.archived=0
ORDER BY configuration.name COLLATE NOCASE, configuration.id`, revisionID)
	if err != nil {
		return nil, fmt.Errorf("list track plan affected configurations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	configurations := []application.TrackPlanAffectedConfiguration{}
	for rows.Next() {
		configuration := application.TrackPlanAffectedConfiguration{}
		if err := rows.Scan(&configuration.ID, &configuration.Name); err != nil {
			return nil, fmt.Errorf("scan track plan affected configuration: %w", err)
		}
		configurations = append(configurations, configuration)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate track plan affected configurations: %w", err)
	}
	return configurations, nil
}

func (repository *TrackPlannerRepository) ReserveMaterials(
	ctx context.Context,
	revisionID string,
	input application.ReserveTrackPlanMaterialsInput,
	actor string,
) (*application.TrackPlanReservationBatch, error) {
	now := timestamp()
	type pendingReservation struct {
		trackObjectID string
		reservationID string
	}
	pending := make([]pendingReservation, 0, len(input.Items))
	err := repository.withTx(ctx, func(tx *sql.Tx) error {
		var status domain.PlanRevisionStatus
		var layoutUnitID string
		err := tx.QueryRowContext(ctx, `
SELECT revision.status, variant.layout_unit_id
FROM plan_revisions revision
JOIN plan_variants variant ON variant.id=revision.variant_id
WHERE revision.id=?`, revisionID).Scan(&status, &layoutUnitID)
		if errors.Is(err, sql.ErrNoRows) {
			return application.ErrTrackPlanNotFound
		}
		if err != nil {
			return fmt.Errorf("read track reservation revision: %w", err)
		}
		if status != domain.PlanRevisionDraft && status != domain.PlanRevisionReview {
			return application.ErrTrackPlanImmutable
		}

		for _, item := range input.Items {
			var objectVersion int
			var manufacturer, articleNumber string
			err := tx.QueryRowContext(ctx, `
SELECT object.version, library.manufacturer, geometry.article_number
FROM plan_track_objects object
JOIN track_geometry_definitions geometry ON geometry.id=object.geometry_id
JOIN track_geometry_libraries library ON library.id=geometry.library_id
WHERE object.id=? AND object.revision_id=?`, item.TrackObjectID, revisionID).
				Scan(&objectVersion, &manufacturer, &articleNumber)
			if errors.Is(err, sql.ErrNoRows) {
				return application.ErrTrackPlanNotFound
			}
			if err != nil {
				return fmt.Errorf("read track reservation object: %w", err)
			}
			if objectVersion != item.ExpectedObjectVersion {
				return application.ErrTrackPlanConflict
			}
			var productManufacturer, productArticleNumber string
			err = tx.QueryRowContext(ctx, `
SELECT manufacturer, article_number FROM accessory_products WHERE id=? AND archived=0`, item.ProductID).
				Scan(&productManufacturer, &productArticleNumber)
			if errors.Is(err, sql.ErrNoRows) {
				return application.ErrAccessoryNotFound
			}
			if err != nil {
				return fmt.Errorf("read track reservation product: %w", err)
			}
			if !strings.EqualFold(strings.TrimSpace(manufacturer), strings.TrimSpace(productManufacturer)) ||
				!strings.EqualFold(strings.TrimSpace(articleNumber), strings.TrimSpace(productArticleNumber)) {
				return application.ErrTrackPlanValidation
			}
			var activeLinkCount int
			if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM plan_track_object_reservations WHERE track_object_id=? AND active=1`,
				item.TrackObjectID).Scan(&activeLinkCount); err != nil {
				return fmt.Errorf("check active track reservation: %w", err)
			}
			if activeLinkCount > 0 {
				return application.ErrTrackPlanConflict
			}
			strategy, err := accessoryInventoryStrategy(ctx, tx, item.ProductID)
			if err != nil {
				return err
			}
			if err := requireActiveStorageLocation(ctx, tx, item.LocationID); err != nil {
				return err
			}
			reservationInput := application.CreateAccessoryReservationInput{
				ProductID: item.ProductID, AssetID: item.AssetID, LocationID: item.LocationID,
				Quantity: 1, AllocationTargetInput: application.AllocationTargetInput{LayoutUnitID: layoutUnitID},
				Note: "Gleisplanobjekt " + item.TrackObjectID,
			}
			usesAsset, err := accessoryAllocationUsesAsset(strategy, item.AssetID)
			if err != nil {
				return err
			}
			if usesAsset {
				if err := requireReservableAsset(ctx, tx, reservationInput); err != nil {
					return err
				}
			} else {
				available, err := availableAccessoryQuantity(ctx, tx, item.ProductID, item.LocationID)
				if err != nil {
					return err
				}
				if available < 1 {
					return application.ErrAccessoryInsufficientStock
				}
			}

			reservationID := randomID()
			if _, err := tx.ExecContext(ctx, `
INSERT INTO accessory_reservations(
  id, product_id, asset_id, location_id, quantity, layout_unit_id, status, note,
  created_by, created_at, updated_at, placement, digital_address, decoder_output,
  connection, wiring_notes
) VALUES(?, ?, NULLIF(?, ''), ?, 1, ?, ?, ?, ?, ?, ?, '', '', '', '', '')`,
				reservationID, item.ProductID, item.AssetID, item.LocationID, layoutUnitID,
				domain.AccessoryReservationActive, reservationInput.Note, actor, now, now); err != nil {
				if isSQLiteConstraint(err) {
					return application.ErrTrackPlanConflict
				}
				return fmt.Errorf("insert track material reservation: %w", err)
			}
			if item.AssetID != "" {
				result, err := tx.ExecContext(ctx, `
UPDATE accessory_assets SET lifecycle_state=?, updated_at=?
WHERE id=? AND lifecycle_state=?`, domain.AccessoryLifecycleReserved, now, item.AssetID,
					domain.AccessoryLifecycleStored)
				if err != nil {
					return fmt.Errorf("reserve track material asset: %w", err)
				}
				if err := requireAccessoryConflictFreeUpdate(result); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO plan_track_object_reservations(
  reservation_id, track_object_id, active, created_at, updated_at
) VALUES(?, ?, 1, ?, ?)`, reservationID, item.TrackObjectID, now, now); err != nil {
				if isSQLiteConstraint(err) {
					return application.ErrTrackPlanConflict
				}
				return fmt.Errorf("link track material reservation: %w", err)
			}
			if err := writeAccessoryAudit(ctx, tx, "TrackPlanMaterialReserved", "plan_track_object",
				item.TrackObjectID, actor, now, `{"reservationId":"`+reservationID+`"}`); err != nil {
				return err
			}
			pending = append(pending, pendingReservation{
				trackObjectID: item.TrackObjectID, reservationID: reservationID,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	batch := &application.TrackPlanReservationBatch{
		RevisionID: revisionID, Reservations: make([]application.TrackPlanObjectReservation, 0, len(pending)),
	}
	for _, item := range pending {
		reservation, err := getReservationWith(ctx, repository.db, item.reservationID)
		if err != nil {
			return nil, err
		}
		batch.Reservations = append(batch.Reservations, application.TrackPlanObjectReservation{
			TrackObjectID: item.trackObjectID, Reservation: *reservation,
		})
	}
	return batch, nil
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

func (repository *TrackPlannerRepository) ListMaterialReservations(
	ctx context.Context,
	revisionID string,
) ([]application.TrackPlanObjectReservation, error) {
	rows, err := repository.db.QueryContext(ctx, `
SELECT link.track_object_id, link.reservation_id
FROM plan_track_object_reservations link
JOIN plan_track_objects object ON object.id=link.track_object_id
JOIN accessory_reservations reservation ON reservation.id=link.reservation_id
WHERE object.revision_id=? AND link.active=1 AND reservation.status='active'
ORDER BY object.created_at, object.id`, revisionID)
	if err != nil {
		return nil, fmt.Errorf("list track material reservation links: %w", err)
	}
	type reservationLink struct {
		trackObjectID string
		reservationID string
	}
	links := []reservationLink{}
	for rows.Next() {
		link := reservationLink{}
		if err := rows.Scan(&link.trackObjectID, &link.reservationID); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan track material reservation link: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate track material reservation links: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close track material reservation links: %w", err)
	}
	reservations := make([]application.TrackPlanObjectReservation, 0, len(links))
	for _, link := range links {
		reservation, err := getReservationWith(ctx, repository.db, link.reservationID)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, application.TrackPlanObjectReservation{
			TrackObjectID: link.trackObjectID, Reservation: *reservation,
		})
	}
	return reservations, nil
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
		RotationDegrees: input.RotationDegrees, ElevationStartMM: input.ElevationStartMM,
		ElevationEndMM: input.ElevationEndMM, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	object.LineageID = object.ID
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
	elevation_start_mm, elevation_end_mm, lineage_id, version, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`, object.ID, revisionID, input.GeometryID,
			input.PositionXMM, input.PositionYMM, input.RotationDegrees,
			input.ElevationStartMM, input.ElevationEndMM, object.LineageID, now, now); err != nil {
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
SET position_x_mm=?, position_y_mm=?, rotation_degrees=?, elevation_start_mm=?, elevation_end_mm=?,
    version=version+1, updated_at=?
WHERE id=? AND version=?`, input.PositionXMM, input.PositionYMM, input.RotationDegrees,
			input.ElevationStartMM, input.ElevationEndMM, now, id, input.ExpectedVersion)
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
		var activeReservations int
		if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM plan_track_object_reservations WHERE track_object_id=?`, id).
			Scan(&activeReservations); err != nil {
			return fmt.Errorf("check track object reservations: %w", err)
		}
		if activeReservations > 0 {
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
SELECT object.id, object.lineage_id, object.revision_id, object.geometry_id,
       object.position_x_mm, object.position_y_mm, object.rotation_degrees,
	   object.elevation_start_mm, object.elevation_end_mm,
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
	if err := scanner.Scan(&object.ID, &object.LineageID, &object.RevisionID, &object.GeometryID,
		&object.PositionXMM, &object.PositionYMM, &object.RotationDegrees,
		&object.ElevationStartMM, &object.ElevationEndMM,
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
