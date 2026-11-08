package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func (s *VehicleService) Delete(ctx context.Context, id, actorUserID string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrVehicleNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete vehicle: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var vehicleSetID sql.NullString
	if err = tx.QueryRowContext(ctx, `
SELECT vehicle_set_id FROM vehicle_set_members WHERE vehicle_id=?
`, id).Scan(&vehicleSetID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read vehicle set membership: %w", err)
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM vehicles WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete vehicle: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if affected == 0 {
		_ = tx.Rollback()
		return ErrVehicleNotFound
	}
	if vehicleSetID.Valid {
		if _, err = tx.ExecContext(ctx, `
DELETE FROM vehicle_sets
WHERE id=? AND NOT EXISTS (SELECT 1 FROM vehicle_set_members WHERE vehicle_set_id=?)
`, vehicleSetID.String, vehicleSetID.String); err != nil {
			return fmt.Errorf("delete empty vehicle set: %w", err)
		}
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleDeleted', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, id, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("write vehicle audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit delete vehicle: %w", err)
	}

	return nil
}

func readVehicleCore(ctx context.Context, query queryRower, id string) (*Vehicle, error) {
	return vehicledata.ReadCore(ctx, query, id)
}

func (s *VehicleService) get(ctx context.Context, id string) (*Vehicle, error) {
	vehicle, err := readVehicleCore(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	images, err := s.loadVehicleImages(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Images = images
	vehicle.ImagesRevision = vehicleImagesRevision(images)
	if vehicle.VehicleSet != nil {
		_, _, _, mainImage, err := s.resolveSetMainImage(ctx, vehicle.VehicleSet.ID)
		if err != nil {
			return nil, err
		}
		vehicle.VehicleSet.MainImage = mainImage
	}
	attachments, err := s.loadVehicleAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Attachments = attachments
	maintenance, err := s.loadVehicleMaintenance(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Maintenance = maintenance
	spareParts, err := s.loadVehicleSpareParts(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.SpareParts = spareParts
	functions, err := s.loadVehicleFunctions(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Functions = functions
	cvValues, err := s.loadVehicleCVValues(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.CVValues = cvValues
	cvFiles, err := s.loadVehicleCVFiles(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.CVFiles = cvFiles

	return vehicle, nil
}
