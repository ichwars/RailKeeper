package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

type LayoutRepository struct {
	db *sql.DB
}

func NewLayoutRepository(db *sql.DB) *LayoutRepository {
	return &LayoutRepository{db: db}
}

func (r *LayoutRepository) ListLayouts(ctx context.Context) ([]application.Layout, error) {
	rows, err := r.db.QueryContext(ctx, layoutSelect+` ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("list layouts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	layouts := []application.Layout{}
	for rows.Next() {
		layout, err := scanLayout(rows)
		if err != nil {
			return nil, fmt.Errorf("scan layout: %w", err)
		}
		layouts = append(layouts, *layout)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layouts: %w", err)
	}
	return layouts, nil
}

func (r *LayoutRepository) GetLayout(ctx context.Context, id string) (*application.Layout, error) {
	layout, err := scanLayout(r.db.QueryRowContext(ctx, layoutSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get layout: %w", err)
	}
	return layout, nil
}

func (r *LayoutRepository) CreateLayout(
	ctx context.Context,
	input application.CreateLayoutInput,
	actor string,
) (*application.Layout, error) {
	now := timestamp()
	layout := &application.Layout{
		ID: randomID(), Name: input.Name, Kind: input.Kind, Gauge: input.Gauge, Scale: input.Scale,
		Description: input.Description, MaxGradePercent: input.MaxGradePercent,
		MinimumTrackClearanceMM: input.MinimumTrackClearanceMM, Version: 1,
		Archived: input.Archived, CreatedAt: now, UpdatedAt: now,
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO layouts(
  id, name, kind, gauge, scale, description, max_grade_percent, minimum_track_clearance_mm,
  version, archived, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?)`, layout.ID, layout.Name, layout.Kind, layout.Gauge, layout.Scale,
			layout.Description, layout.MaxGradePercent, layout.MinimumTrackClearanceMM,
			boolToInt(layout.Archived), now, now); err != nil {
			return fmt.Errorf("insert layout: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "LayoutCreated", "layout", layout.ID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return layout, nil
}

func (r *LayoutRepository) UpdateLayout(
	ctx context.Context,
	id string,
	input application.UpdateLayoutInput,
	actor string,
) (*application.Layout, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
UPDATE layouts
SET name=?, kind=?, gauge=?, scale=?, description=?, max_grade_percent=?,
    minimum_track_clearance_mm=?, archived=?,
    version=version+1, updated_at=?
WHERE id=? AND version=?`, input.Name, input.Kind, input.Gauge, input.Scale, input.Description,
			input.MaxGradePercent, input.MinimumTrackClearanceMM, boolToInt(input.Archived), now, id,
			input.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("update layout: %w", err)
		}
		if err := requireUpdated(ctx, tx, result, "layouts", id, application.ErrLayoutVersionConflict); err != nil {
			return err
		}
		return writeLayoutAudit(ctx, tx, "LayoutUpdated", "layout", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return r.GetLayout(ctx, id)
}

func (r *LayoutRepository) ListUnits(ctx context.Context, layoutID string) ([]application.LayoutUnit, error) {
	rows, err := r.db.QueryContext(ctx, layoutUnitSelect+` WHERE layout_id=? ORDER BY name COLLATE NOCASE, id`, layoutID)
	if err != nil {
		return nil, fmt.Errorf("list layout units: %w", err)
	}
	defer func() { _ = rows.Close() }()

	units := []application.LayoutUnit{}
	for rows.Next() {
		unit, err := scanLayoutUnit(rows)
		if err != nil {
			return nil, fmt.Errorf("scan layout unit: %w", err)
		}
		units = append(units, *unit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layout units: %w", err)
	}
	return units, nil
}

func (r *LayoutRepository) GetUnit(ctx context.Context, id string) (*application.LayoutUnit, error) {
	unit, err := scanLayoutUnit(r.db.QueryRowContext(ctx, layoutUnitSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get layout unit: %w", err)
	}
	return unit, nil
}

func (r *LayoutRepository) GetUnitForPort(ctx context.Context, id string) (*application.LayoutUnit, error) {
	unit, err := scanLayoutUnit(r.db.QueryRowContext(ctx, layoutUnitSelect+`
 WHERE id=(SELECT layout_unit_id FROM layout_unit_ports WHERE id=?)`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get layout unit for port: %w", err)
	}
	return unit, nil
}

func (r *LayoutRepository) ListUnitPorts(
	ctx context.Context,
	unitID string,
) ([]application.LayoutUnitPort, error) {
	rows, err := r.db.QueryContext(ctx, layoutUnitPortSelect+`
 WHERE layout_unit_id=? ORDER BY archived, name COLLATE NOCASE, id`, unitID)
	if err != nil {
		return nil, fmt.Errorf("list layout unit ports: %w", err)
	}
	defer func() { _ = rows.Close() }()
	ports := []application.LayoutUnitPort{}
	for rows.Next() {
		port, err := scanLayoutUnitPort(rows)
		if err != nil {
			return nil, fmt.Errorf("scan layout unit port: %w", err)
		}
		ports = append(ports, *port)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layout unit ports: %w", err)
	}
	return ports, nil
}

func (r *LayoutRepository) CreateUnitPort(
	ctx context.Context,
	unitID string,
	input application.CreateLayoutUnitPortInput,
	actor string,
) (*application.LayoutUnitPort, error) {
	now := timestamp()
	port := &application.LayoutUnitPort{
		ID: randomID(), LayoutUnitID: unitID, Name: input.Name, Kind: input.Kind,
		InterfaceKey: input.InterfaceKey, XMM: input.XMM, YMM: input.YMM,
		DirectionDegrees: input.DirectionDegrees, Notes: input.Notes, Version: 1,
		Archived: input.Archived, CreatedAt: now, UpdatedAt: now,
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
INSERT INTO layout_unit_ports(
  id, layout_unit_id, name, kind, interface_key, x_mm, y_mm, direction_degrees, notes,
  version, archived, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?)`, port.ID, port.LayoutUnitID, port.Name,
			port.Kind, port.InterfaceKey, port.XMM, port.YMM, port.DirectionDegrees, port.Notes,
			boolToInt(port.Archived), now, now); err != nil {
			return fmt.Errorf("insert layout unit port: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "LayoutUnitPortCreated", "layout_unit_port", port.ID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return port, nil
}

func (r *LayoutRepository) UpdateUnitPort(
	ctx context.Context,
	id string,
	input application.UpdateLayoutUnitPortInput,
	actor string,
) (*application.LayoutUnitPort, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
UPDATE layout_unit_ports
SET name=?, kind=?, interface_key=?, x_mm=?, y_mm=?, direction_degrees=?, notes=?, archived=?,
    version=version+1, updated_at=?
WHERE id=? AND version=?`, input.Name, input.Kind, input.InterfaceKey, input.XMM, input.YMM,
			input.DirectionDegrees, input.Notes, boolToInt(input.Archived), now, id, input.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("update layout unit port: %w", err)
		}
		if err := requireUpdated(ctx, tx, result, "layout_unit_ports", id,
			application.ErrLayoutVersionConflict); err != nil {
			return err
		}
		return writeLayoutAudit(ctx, tx, "LayoutUnitPortUpdated", "layout_unit_port", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	port, err := scanLayoutUnitPort(r.db.QueryRowContext(ctx, layoutUnitPortSelect+` WHERE id=?`, id))
	if err != nil {
		return nil, fmt.Errorf("get updated layout unit port: %w", err)
	}
	return port, nil
}

func (r *LayoutRepository) CreateUnit(
	ctx context.Context,
	layoutID string,
	input application.CreateLayoutUnitInput,
	actor string,
) (*application.LayoutUnit, error) {
	now := timestamp()
	unit := &application.LayoutUnit{
		ID: randomID(), LayoutID: layoutID, Name: input.Name, Kind: input.Kind, OwnerLabel: input.OwnerLabel,
		WidthMM: input.WidthMM, HeightMM: input.HeightMM, Version: 1, Archived: input.Archived,
		CreatedAt: now, UpdatedAt: now,
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		exists, err := recordExists(ctx, tx, "layouts", layoutID)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrLayoutNotFound
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO layout_units(
  id, layout_id, name, kind, owner_label, width_mm, height_mm, version, archived, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?)`, unit.ID, unit.LayoutID, unit.Name, unit.Kind, unit.OwnerLabel,
			unit.WidthMM, unit.HeightMM, boolToInt(unit.Archived), now, now); err != nil {
			return fmt.Errorf("insert layout unit: %w", err)
		}
		return writeLayoutAudit(ctx, tx, "LayoutUnitCreated", "layout_unit", unit.ID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func (r *LayoutRepository) UpdateUnit(
	ctx context.Context,
	id string,
	input application.UpdateLayoutUnitInput,
	actor string,
) (*application.LayoutUnit, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
UPDATE layout_units
SET name=?, kind=?, owner_label=?, width_mm=?, height_mm=?, archived=?, version=version+1, updated_at=?
WHERE id=? AND version=?`, input.Name, input.Kind, input.OwnerLabel, input.WidthMM, input.HeightMM,
			boolToInt(input.Archived), now, id, input.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("update layout unit: %w", err)
		}
		if err := requireUpdated(ctx, tx, result, "layout_units", id, application.ErrLayoutVersionConflict); err != nil {
			return err
		}
		return writeLayoutAudit(ctx, tx, "LayoutUnitUpdated", "layout_unit", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	unit, err := scanLayoutUnit(r.db.QueryRowContext(ctx, layoutUnitSelect+` WHERE id=?`, id))
	if err != nil {
		return nil, fmt.Errorf("get updated layout unit: %w", err)
	}
	return unit, nil
}

func (r *LayoutRepository) ListConfigurations(
	ctx context.Context,
	layoutID string,
) ([]application.LayoutConfiguration, error) {
	rows, err := r.db.QueryContext(ctx, layoutConfigurationSelect+` WHERE layout_id=? ORDER BY name COLLATE NOCASE, id`, layoutID)
	if err != nil {
		return nil, fmt.Errorf("list layout configurations: %w", err)
	}
	configurations := []application.LayoutConfiguration{}
	for rows.Next() {
		configuration, err := scanLayoutConfiguration(rows)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan layout configuration: %w", err)
		}
		configurations = append(configurations, *configuration)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate layout configurations: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close layout configurations: %w", err)
	}
	for index := range configurations {
		units, err := listConfigurationUnits(ctx, r.db, configurations[index].ID)
		if err != nil {
			return nil, err
		}
		configurations[index].Units = units
	}
	return configurations, nil
}

func (r *LayoutRepository) LoadConfigurationPortPlacements(
	ctx context.Context,
	configurationID string,
) ([]domain.ModulePortPlacement, error) {
	var exists int
	if err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM layout_configurations WHERE id=?`, configurationID).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find layout configuration for port analysis: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT cu.unit_id, u.name, p.id, p.name, p.kind, p.interface_key, p.x_mm, p.y_mm,
       p.direction_degrees, cu.position_x_mm, cu.position_y_mm, cu.rotation_degrees
FROM layout_configuration_units cu
JOIN layout_units u ON u.id=cu.unit_id
JOIN layout_unit_ports p ON p.layout_unit_id=cu.unit_id AND p.archived=0
WHERE cu.configuration_id=?
ORDER BY u.name COLLATE NOCASE, cu.unit_id, p.name COLLATE NOCASE, p.id`, configurationID)
	if err != nil {
		return nil, fmt.Errorf("list layout configuration port placements: %w", err)
	}
	defer func() { _ = rows.Close() }()
	placements := []domain.ModulePortPlacement{}
	for rows.Next() {
		placement := domain.ModulePortPlacement{}
		if err := rows.Scan(&placement.UnitID, &placement.UnitName, &placement.PortID, &placement.PortName,
			&placement.Kind, &placement.InterfaceKey, &placement.XMM, &placement.YMM,
			&placement.DirectionDegrees, &placement.UnitPose.PositionXMM, &placement.UnitPose.PositionYMM,
			&placement.UnitPose.RotationDegrees); err != nil {
			return nil, fmt.Errorf("scan layout configuration port placement: %w", err)
		}
		placements = append(placements, placement)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layout configuration port placements: %w", err)
	}
	return placements, nil
}

func (r *LayoutRepository) ConfigurationContainsUnit(
	ctx context.Context,
	configurationID string,
	unitID string,
) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
SELECT 1 FROM layout_configuration_units WHERE configuration_id=? AND unit_id=?`,
		configurationID, unitID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check layout configuration unit membership: %w", err)
	}
	return true, nil
}

func (r *LayoutRepository) SaveConfiguration(
	ctx context.Context,
	layoutID string,
	input application.SaveLayoutConfigurationInput,
	actor string,
) (*application.LayoutConfiguration, error) {
	now := timestamp()
	id := input.ID
	if id == "" {
		id = randomID()
	}
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if input.ID != "" && layoutID == "" {
			if err := tx.QueryRowContext(ctx, `SELECT layout_id FROM layout_configurations WHERE id=?`, input.ID).
				Scan(&layoutID); errors.Is(err, sql.ErrNoRows) {
				return application.ErrLayoutNotFound
			} else if err != nil {
				return fmt.Errorf("read layout configuration parent: %w", err)
			}
		}
		exists, err := recordExists(ctx, tx, "layouts", layoutID)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrLayoutNotFound
		}
		for _, unit := range input.Units {
			if err := validateConfigurationUnit(ctx, tx, layoutID, unit); err != nil {
				return err
			}
		}
		if input.ID == "" {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO layout_configurations(id, layout_id, name, description, version, archived, created_at, updated_at)
VALUES(?, ?, ?, ?, 1, ?, ?, ?)`, id, layoutID, input.Name, input.Description,
				boolToInt(input.Archived), now, now); err != nil {
				return fmt.Errorf("insert layout configuration: %w", err)
			}
		} else {
			result, err := tx.ExecContext(ctx, `
UPDATE layout_configurations
SET name=?, description=?, archived=?, version=version+1, updated_at=?
WHERE id=? AND layout_id=? AND version=?`, input.Name, input.Description, boolToInt(input.Archived),
				now, id, layoutID, input.ExpectedVersion)
			if err != nil {
				return fmt.Errorf("update layout configuration: %w", err)
			}
			if err := requireUpdated(ctx, tx, result, "layout_configurations", id, application.ErrLayoutVersionConflict); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM layout_configuration_units WHERE configuration_id=?`, id); err != nil {
				return fmt.Errorf("replace layout configuration units: %w", err)
			}
		}
		for _, unit := range input.Units {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO layout_configuration_units(
  configuration_id, unit_id, plan_revision_id, position_x_mm, position_y_mm, rotation_degrees, sort_order
) VALUES(?, ?, NULLIF(?, ''), ?, ?, ?, ?)`, id, unit.UnitID, unit.PlanRevisionID, unit.PositionXMM,
				unit.PositionYMM, unit.RotationDegrees, unit.SortOrder); err != nil {
				return fmt.Errorf("insert layout configuration unit: %w", err)
			}
		}
		return writeLayoutAudit(ctx, tx, "LayoutConfigurationSaved", "layout_configuration", id, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return r.getConfiguration(ctx, id)
}

const layoutSelect = `SELECT id, name, kind, gauge, scale, description, max_grade_percent,
minimum_track_clearance_mm,
version, archived, created_at, updated_at FROM layouts`

const layoutUnitSelect = `SELECT id, layout_id, name, kind, owner_label, COALESCE(width_mm, 0), COALESCE(height_mm, 0), version, archived, created_at, updated_at FROM layout_units`

const layoutConfigurationSelect = `SELECT id, layout_id, name, description, version, archived, created_at, updated_at FROM layout_configurations`

const layoutUnitPortSelect = `SELECT id, layout_unit_id, name, kind, interface_key, x_mm, y_mm,
direction_degrees, notes, version, archived, created_at, updated_at FROM layout_unit_ports`

type rowScanner interface {
	Scan(...any) error
}

func scanLayout(scanner rowScanner) (*application.Layout, error) {
	layout := &application.Layout{}
	var archived int
	var maxGradePercent sql.NullFloat64
	var minimumTrackClearanceMM sql.NullFloat64
	err := scanner.Scan(&layout.ID, &layout.Name, &layout.Kind, &layout.Gauge, &layout.Scale, &layout.Description,
		&maxGradePercent, &minimumTrackClearanceMM, &layout.Version, &archived, &layout.CreatedAt, &layout.UpdatedAt)
	if maxGradePercent.Valid {
		layout.MaxGradePercent = &maxGradePercent.Float64
	}
	if minimumTrackClearanceMM.Valid {
		layout.MinimumTrackClearanceMM = &minimumTrackClearanceMM.Float64
	}
	layout.Archived = archived != 0
	return layout, err
}

func scanLayoutUnit(scanner rowScanner) (*application.LayoutUnit, error) {
	unit := &application.LayoutUnit{}
	var archived int
	err := scanner.Scan(&unit.ID, &unit.LayoutID, &unit.Name, &unit.Kind, &unit.OwnerLabel, &unit.WidthMM,
		&unit.HeightMM, &unit.Version, &archived, &unit.CreatedAt, &unit.UpdatedAt)
	unit.Archived = archived != 0
	return unit, err
}

func scanLayoutUnitPort(scanner rowScanner) (*application.LayoutUnitPort, error) {
	port := &application.LayoutUnitPort{}
	var archived int
	err := scanner.Scan(&port.ID, &port.LayoutUnitID, &port.Name, &port.Kind, &port.InterfaceKey,
		&port.XMM, &port.YMM, &port.DirectionDegrees, &port.Notes, &port.Version, &archived,
		&port.CreatedAt, &port.UpdatedAt)
	port.Archived = archived != 0
	return port, err
}

func scanLayoutConfiguration(scanner rowScanner) (*application.LayoutConfiguration, error) {
	configuration := &application.LayoutConfiguration{}
	var archived int
	err := scanner.Scan(&configuration.ID, &configuration.LayoutID, &configuration.Name, &configuration.Description,
		&configuration.Version, &archived, &configuration.CreatedAt, &configuration.UpdatedAt)
	configuration.Archived = archived != 0
	return configuration, err
}

func (r *LayoutRepository) getConfiguration(ctx context.Context, id string) (*application.LayoutConfiguration, error) {
	configuration, err := scanLayoutConfiguration(r.db.QueryRowContext(ctx, layoutConfigurationSelect+` WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get layout configuration: %w", err)
	}
	configuration.Units, err = listConfigurationUnits(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return configuration, nil
}

func listConfigurationUnits(ctx context.Context, db *sql.DB, configurationID string) ([]application.ConfigurationUnit, error) {
	rows, err := db.QueryContext(ctx, `
SELECT unit_id, COALESCE(plan_revision_id, ''), position_x_mm, position_y_mm, rotation_degrees, sort_order
FROM layout_configuration_units WHERE configuration_id=? ORDER BY sort_order, unit_id`, configurationID)
	if err != nil {
		return nil, fmt.Errorf("list layout configuration units: %w", err)
	}
	defer func() { _ = rows.Close() }()
	units := []application.ConfigurationUnit{}
	for rows.Next() {
		unit := application.ConfigurationUnit{}
		if err := rows.Scan(&unit.UnitID, &unit.PlanRevisionID, &unit.PositionXMM, &unit.PositionYMM,
			&unit.RotationDegrees, &unit.SortOrder); err != nil {
			return nil, fmt.Errorf("scan layout configuration unit: %w", err)
		}
		units = append(units, unit)
	}
	return units, rows.Err()
}

func validateConfigurationUnit(
	ctx context.Context,
	tx *sql.Tx,
	layoutID string,
	unit application.ConfigurationUnit,
) error {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM layout_units WHERE id=? AND layout_id=?`,
		unit.UnitID, layoutID).Scan(&count); err != nil {
		return fmt.Errorf("validate layout configuration unit: %w", err)
	}
	if count == 0 {
		return application.ErrLayoutValidation
	}
	if unit.PlanRevisionID == "" {
		return nil
	}
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM plan_revisions revision
JOIN plan_variants variant ON variant.id=revision.variant_id
WHERE revision.id=? AND variant.layout_unit_id=? AND revision.status IN ('published', 'archived')`,
		unit.PlanRevisionID, unit.UnitID).Scan(&count); err != nil {
		return fmt.Errorf("validate configuration plan revision: %w", err)
	}
	if count == 0 {
		return application.ErrLayoutValidation
	}
	return nil
}

func (r *LayoutRepository) withTx(ctx context.Context, work func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin layout transaction: %w", err)
	}
	if err := work(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit layout transaction: %w", err)
	}
	return nil
}

func requireUpdated(
	ctx context.Context,
	tx *sql.Tx,
	result sql.Result,
	table string,
	id string,
	conflict error,
) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read layout update result: %w", err)
	}
	if affected > 0 {
		return nil
	}
	exists, err := recordExists(ctx, tx, table, id)
	if err != nil {
		return err
	}
	if !exists {
		return application.ErrLayoutNotFound
	}
	return conflict
}

func recordExists(ctx context.Context, tx *sql.Tx, table, id string) (bool, error) {
	query := map[string]string{
		"layouts":               `SELECT COUNT(*) FROM layouts WHERE id=?`,
		"layout_units":          `SELECT COUNT(*) FROM layout_units WHERE id=?`,
		"layout_unit_ports":     `SELECT COUNT(*) FROM layout_unit_ports WHERE id=?`,
		"layout_configurations": `SELECT COUNT(*) FROM layout_configurations WHERE id=?`,
	}[table]
	if query == "" {
		return false, fmt.Errorf("unsupported layout table %q", table)
	}
	var count int
	if err := tx.QueryRowContext(ctx, query, id).Scan(&count); err != nil {
		return false, fmt.Errorf("check %s record: %w", table, err)
	}
	return count > 0, nil
}

func writeLayoutAudit(
	ctx context.Context,
	tx *sql.Tx,
	action, targetType, targetID, actor, createdAt string,
) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, NULLIF(?, ''), ?, ?, ?, ?, '{}')`, randomID(), actor, action, targetType, targetID, createdAt); err != nil {
		return fmt.Errorf("write layout audit log: %w", err)
	}
	return nil
}

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

var _ application.LayoutRepository = (*LayoutRepository)(nil)
var _ application.LayoutUnitPortRepository = (*LayoutRepository)(nil)
