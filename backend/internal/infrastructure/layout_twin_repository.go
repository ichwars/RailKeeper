package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

type layoutTwinUnitPlacement struct {
	unit     application.LayoutUnit
	xMM      float64
	yMM      float64
	rotation float64
}

func (r *LayoutRepository) GetTwin(
	ctx context.Context,
	layoutID string,
	selection application.LayoutTwinSelection,
) (*application.LayoutTwin, error) {
	if _, err := r.GetLayout(ctx, layoutID); err != nil {
		return nil, err
	}
	twin := &application.LayoutTwin{
		LayoutID: layoutID, Units: []application.LayoutTwinUnit{}, Warnings: []application.LayoutTwinWarning{},
	}
	placements, err := r.layoutTwinPlacements(ctx, layoutID, selection, twin)
	if err != nil {
		return nil, err
	}
	bounds := layoutTwinBoundsBuilder{}
	for _, placement := range placements {
		unit, warnings, err := r.buildLayoutTwinUnit(ctx, placement, &bounds)
		if err != nil {
			return nil, err
		}
		twin.Units = append(twin.Units, unit)
		twin.Warnings = append(twin.Warnings, warnings...)
	}
	twin.Bounds, twin.HasGeometry = bounds.result()
	return twin, nil
}

func (r *LayoutRepository) layoutTwinPlacements(
	ctx context.Context,
	layoutID string,
	selection application.LayoutTwinSelection,
	twin *application.LayoutTwin,
) ([]layoutTwinUnitPlacement, error) {
	if selection.ConfigurationID != "" {
		configuration, err := r.getConfiguration(ctx, selection.ConfigurationID)
		if err != nil || configuration.LayoutID != layoutID {
			if err == nil || errors.Is(err, application.ErrLayoutNotFound) {
				return nil, application.ErrLayoutNotFound
			}
			return nil, err
		}
		twin.ConfigurationID = configuration.ID
		twin.ConfigurationName = configuration.Name
		placements := make([]layoutTwinUnitPlacement, 0, len(configuration.Units))
		for _, configured := range configuration.Units {
			unit, err := r.getTwinUnit(ctx, layoutID, configured.UnitID)
			if err != nil {
				return nil, err
			}
			placements = append(placements, layoutTwinUnitPlacement{
				unit: *unit, xMM: configured.PositionXMM, yMM: configured.PositionYMM,
				rotation: configured.RotationDegrees,
			})
		}
		return placements, nil
	}

	unitID := selection.UnitID
	if unitID == "" {
		err := r.db.QueryRowContext(ctx, `
SELECT id FROM layout_units
WHERE layout_id=? AND archived=0 ORDER BY name COLLATE NOCASE, id LIMIT 1`, layoutID).Scan(&unitID)
		if errors.Is(err, sql.ErrNoRows) {
			return []layoutTwinUnitPlacement{}, nil
		}
		if err != nil {
			return nil, fmt.Errorf("select layout twin unit: %w", err)
		}
	}
	unit, err := r.getTwinUnit(ctx, layoutID, unitID)
	if err != nil {
		return nil, err
	}
	twin.UnitID = unit.ID
	return []layoutTwinUnitPlacement{{unit: *unit}}, nil
}

func (r *LayoutRepository) getTwinUnit(
	ctx context.Context,
	layoutID string,
	unitID string,
) (*application.LayoutUnit, error) {
	unit, err := scanLayoutUnit(r.db.QueryRowContext(
		ctx, layoutUnitSelect+` WHERE id=? AND layout_id=?`, unitID, layoutID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get layout twin unit: %w", err)
	}
	return unit, nil
}

func (r *LayoutRepository) buildLayoutTwinUnit(
	ctx context.Context,
	placement layoutTwinUnitPlacement,
	bounds *layoutTwinBoundsBuilder,
) (application.LayoutTwinUnit, []application.LayoutTwinWarning, error) {
	unit := application.LayoutTwinUnit{
		ID: placement.unit.ID, Name: placement.unit.Name, Kind: placement.unit.Kind,
		PositionXMM: placement.xMM, PositionYMM: placement.yMM,
		RotationDegrees: placement.rotation, Version: placement.unit.Version,
		LocalOutline: []application.LayoutTwinPoint{}, Outline: []application.LayoutTwinPoint{},
		Positions: []application.LayoutTwinPosition{},
	}
	warnings := []application.LayoutTwinWarning{}
	outline, err := r.layoutTwinOutline(ctx, placement.unit.ID)
	if err != nil {
		return unit, warnings, err
	}
	if len(outline) == 0 && placement.unit.WidthMM > 0 && placement.unit.HeightMM > 0 {
		outline = []application.LayoutTwinPoint{
			{XMM: 0, YMM: 0}, {XMM: placement.unit.WidthMM, YMM: 0},
			{XMM: placement.unit.WidthMM, YMM: placement.unit.HeightMM},
			{XMM: 0, YMM: placement.unit.HeightMM},
		}
		warnings = append(warnings, application.LayoutTwinWarning{Code: "outline_fallback", UnitID: unit.ID})
	}
	if len(outline) == 0 {
		warnings = append(warnings, application.LayoutTwinWarning{Code: "missing_geometry", UnitID: unit.ID})
	}
	for _, point := range outline {
		unit.LocalOutline = append(unit.LocalOutline, point)
		global := transformLayoutTwinPoint(point.XMM, point.YMM, placement)
		unit.Outline = append(unit.Outline, global)
		bounds.include(global.XMM, global.YMM)
	}
	positions, err := r.layoutTwinPositions(ctx, placement, bounds)
	if err != nil {
		return unit, warnings, err
	}
	unit.Positions = positions
	return unit, warnings, nil
}

func (r *LayoutRepository) layoutTwinOutline(
	ctx context.Context,
	unitID string,
) ([]application.LayoutTwinPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT position_x_mm, position_y_mm FROM layout_unit_outline_points
WHERE layout_unit_id=? ORDER BY point_index`, unitID)
	if err != nil {
		return nil, fmt.Errorf("list layout twin outline: %w", err)
	}
	defer func() { _ = rows.Close() }()
	points := []application.LayoutTwinPoint{}
	for rows.Next() {
		point := application.LayoutTwinPoint{}
		if err := rows.Scan(&point.XMM, &point.YMM); err != nil {
			return nil, fmt.Errorf("scan layout twin outline: %w", err)
		}
		points = append(points, point)
	}
	return points, rows.Err()
}

func (r *LayoutRepository) layoutTwinPositions(
	ctx context.Context,
	placement layoutTwinUnitPlacement,
	bounds *layoutTwinBoundsBuilder,
) ([]application.LayoutTwinPosition, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT position.id, position.label, position.kind, position.position_x_mm, position.position_y_mm,
       position.rotation_degrees, COALESCE(position.product_id, ''), position.description, position.version,
       COALESCE(product.inventory_number, ''), COALESCE(product.manufacturer, ''),
       COALESCE(product.article_number, ''), COALESCE(product.name, '')
FROM layout_technical_positions position
LEFT JOIN accessory_products product ON product.id=position.product_id
WHERE position.layout_unit_id=? AND position.archived=0
ORDER BY position.label COLLATE NOCASE, position.id`, placement.unit.ID)
	if err != nil {
		return nil, fmt.Errorf("list layout twin positions: %w", err)
	}
	positions := []application.LayoutTwinPosition{}
	for rows.Next() {
		position := application.LayoutTwinPosition{LayoutUnitID: placement.unit.ID}
		var localRotation float64
		if err := rows.Scan(&position.ID, &position.Label, &position.Kind, &position.LocalXMM,
			&position.LocalYMM, &localRotation, &position.ProductID, &position.Description, &position.Version,
			&position.InventoryNumber, &position.Manufacturer, &position.ArticleNumber,
			&position.ProductName); err != nil {
			return nil, fmt.Errorf("scan layout twin position: %w", err)
		}
		global := transformLayoutTwinPoint(position.LocalXMM, position.LocalYMM, placement)
		position.GlobalXMM, position.GlobalYMM = global.XMM, global.YMM
		position.LocalRotationDegrees = localRotation
		position.RotationDegrees = normalizeTwinRotation(placement.rotation + localRotation)
		bounds.include(global.XMM, global.YMM)
		positions = append(positions, position)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate layout twin positions: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close layout twin positions: %w", err)
	}
	for index := range positions {
		position := &positions[index]
		position.Reservations, err = r.layoutTwinReservations(ctx, position.ID)
		if err != nil {
			return nil, err
		}
		position.Installations, err = r.layoutTwinInstallations(ctx, position.ID)
		if err != nil {
			return nil, err
		}
		position.Statuses = layoutTwinStatuses(position.Reservations, position.Installations)
	}
	return positions, nil
}

func (r *LayoutRepository) layoutTwinReservations(
	ctx context.Context,
	positionID string,
) ([]application.LayoutTwinAllocation, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT reservation.id, reservation.product_id, product.inventory_number, product.manufacturer,
       product.article_number, product.name, reservation.quantity, reservation.status,
       reservation.placement, reservation.digital_address, reservation.decoder_output,
       reservation.connection, reservation.wiring_notes, reservation.note
FROM accessory_reservation_positions link
JOIN accessory_reservations reservation ON reservation.id=link.reservation_id
JOIN accessory_products product ON product.id=reservation.product_id
WHERE link.position_id=? AND reservation.status='active'
ORDER BY reservation.created_at, reservation.id`, positionID)
	if err != nil {
		return nil, fmt.Errorf("list layout twin reservations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	allocations := []application.LayoutTwinAllocation{}
	for rows.Next() {
		allocation := application.LayoutTwinAllocation{}
		if err := rows.Scan(&allocation.ID, &allocation.ProductID, &allocation.InventoryNumber,
			&allocation.Manufacturer, &allocation.ArticleNumber, &allocation.ProductName,
			&allocation.Quantity, &allocation.ReservationStatus, &allocation.Placement,
			&allocation.DigitalAddress, &allocation.DecoderOutput, &allocation.Connection,
			&allocation.WiringNotes, &allocation.Note); err != nil {
			return nil, fmt.Errorf("scan layout twin reservation: %w", err)
		}
		allocations = append(allocations, allocation)
	}
	return allocations, rows.Err()
}

func (r *LayoutRepository) layoutTwinInstallations(
	ctx context.Context,
	positionID string,
) ([]application.LayoutTwinAllocation, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT installation.id, installation.product_id, product.inventory_number, product.manufacturer,
       product.article_number, product.name, installation.quantity, installation.condition_state,
       installation.placement, installation.digital_address, installation.decoder_output,
       installation.connection, installation.wiring_notes, installation.notes
FROM accessory_installation_positions link
JOIN accessory_installations installation ON installation.id=link.installation_id
JOIN accessory_products product ON product.id=installation.product_id
WHERE link.position_id=? AND installation.removed_at IS NULL
ORDER BY installation.installed_at, installation.id`, positionID)
	if err != nil {
		return nil, fmt.Errorf("list layout twin installations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	allocations := []application.LayoutTwinAllocation{}
	for rows.Next() {
		allocation := application.LayoutTwinAllocation{}
		if err := rows.Scan(&allocation.ID, &allocation.ProductID, &allocation.InventoryNumber,
			&allocation.Manufacturer, &allocation.ArticleNumber, &allocation.ProductName,
			&allocation.Quantity, &allocation.InstallationCondition, &allocation.Placement,
			&allocation.DigitalAddress, &allocation.DecoderOutput, &allocation.Connection,
			&allocation.WiringNotes, &allocation.Note); err != nil {
			return nil, fmt.Errorf("scan layout twin installation: %w", err)
		}
		allocations = append(allocations, allocation)
	}
	return allocations, rows.Err()
}

func layoutTwinStatuses(
	reservations []application.LayoutTwinAllocation,
	installations []application.LayoutTwinAllocation,
) []application.LayoutTwinStatus {
	statuses := map[application.LayoutTwinStatus]bool{}
	if len(reservations) > 0 {
		statuses[application.LayoutTwinReserved] = true
	}
	if len(installations) > 0 {
		statuses[application.LayoutTwinInstalled] = true
	}
	for _, installation := range installations {
		switch installation.InstallationCondition {
		case domain.AccessoryConditionMaintenanceDue:
			statuses[application.LayoutTwinMaintenanceDue] = true
		case domain.AccessoryConditionDefective:
			statuses[application.LayoutTwinDefective] = true
		}
	}
	if len(statuses) == 0 {
		statuses[application.LayoutTwinPlanned] = true
	}
	order := []application.LayoutTwinStatus{
		application.LayoutTwinPlanned, application.LayoutTwinReserved, application.LayoutTwinInstalled,
		application.LayoutTwinMaintenanceDue, application.LayoutTwinDefective,
	}
	result := make([]application.LayoutTwinStatus, 0, len(statuses))
	for _, status := range order {
		if statuses[status] {
			result = append(result, status)
		}
	}
	return result
}

func transformLayoutTwinPoint(
	xMM float64,
	yMM float64,
	placement layoutTwinUnitPlacement,
) application.LayoutTwinPoint {
	radians := placement.rotation * math.Pi / 180
	return application.LayoutTwinPoint{
		XMM: placement.xMM + xMM*math.Cos(radians) - yMM*math.Sin(radians),
		YMM: placement.yMM + xMM*math.Sin(radians) + yMM*math.Cos(radians),
	}
}

func normalizeTwinRotation(rotation float64) float64 {
	rotation = math.Mod(rotation, 360)
	if rotation < 0 {
		rotation += 360
	}
	return rotation
}

type layoutTwinBoundsBuilder struct {
	set        bool
	minX, minY float64
	maxX, maxY float64
}

func (builder *layoutTwinBoundsBuilder) include(xMM float64, yMM float64) {
	if !builder.set {
		builder.set = true
		builder.minX, builder.maxX = xMM, xMM
		builder.minY, builder.maxY = yMM, yMM
		return
	}
	builder.minX, builder.maxX = math.Min(builder.minX, xMM), math.Max(builder.maxX, xMM)
	builder.minY, builder.maxY = math.Min(builder.minY, yMM), math.Max(builder.maxY, yMM)
}

func (builder layoutTwinBoundsBuilder) result() (application.LayoutTwinBounds, bool) {
	if !builder.set {
		return application.LayoutTwinBounds{}, false
	}
	return application.LayoutTwinBounds{
		MinXMM: builder.minX, MinYMM: builder.minY,
		WidthMM: builder.maxX - builder.minX, HeightMM: builder.maxY - builder.minY,
	}, true
}
