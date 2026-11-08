package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"strings"

	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

func applyTransferExhibitionList(
	ctx context.Context,
	db dataTransferApplyDB,
	record transfer.DataTransferPreviewRecord,
	action string,
	issues []transfer.DataTransferIssue,
) error {
	list := transfer.TransferExhibitionList{}
	if err := json.Unmarshal(record.Data, &list); err != nil {
		return dataTransferApplyConflict("exhibition preview data is invalid")
	}
	now := timestamp()
	transferpolicy.NormalizeTransferExhibitionList(&list)
	listID := record.TargetID
	switch action {
	case "replace":
		var locked int
		if listID == "" || db.QueryRowContext(ctx, `SELECT locked FROM exhibition_lists WHERE id=?`, listID).Scan(&locked) != nil || locked != 0 {
			return dataTransferApplyConflict("locked exhibition list cannot be replaced")
		}
		result, err := db.ExecContext(ctx, `
UPDATE exhibition_lists SET
  designation=?, list_date=?, end_date=?, location=?, description=?, organization_notes=?,
  status=?, locked=?, lock_reason=?, locked_at=?, completed_at=?, archived_at=?, revision=revision+1, updated_at=?
WHERE id=? AND locked=0`,
			list.Designation, list.Date, list.EndDate, list.Location, list.Description, list.OrganizationNotes,
			list.Status, list.Locked, list.LockReason, list.LockedAt, list.CompletedAt, list.ArchivedAt, now, listID)
		if err != nil {
			return err
		}
		if err := requireApplyUpdate(result, "replace exhibition list"); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM exhibition_entries WHERE list_id=?`, listID); err != nil {
			return err
		}
	case "merge":
		var locked int
		if listID == "" || db.QueryRowContext(ctx,
			`SELECT locked FROM exhibition_lists WHERE id=?`, listID).Scan(&locked) != nil || locked != 0 {
			return dataTransferApplyConflict("exhibition list cannot be merged")
		}
	case "create", "copy":
		listID = randomID()
		if _, err := db.ExecContext(ctx, `
INSERT INTO exhibition_lists(
  id, designation, list_date, end_date, location, description, organization_notes,
  status, revision, locked, lock_reason, locked_at, completed_at, archived_at, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?)`, listID, list.Designation, list.Date,
			list.EndDate, list.Location, list.Description, list.OrganizationNotes, list.Status, list.Locked,
			list.LockReason, list.LockedAt, list.CompletedAt, list.ArchivedAt, now, now); err != nil {
			return err
		}
	default:
		return dataTransferApplyConflict("unsupported exhibition-list resolution")
	}
	for entryIndex, entry := range list.Entries {
		transferpolicy.NormalizeTransferExhibitionEntry(&entry)
		vehicleID, err := resolveTransferExhibitionVehicle(ctx, db, entry, issues, entryIndex)
		if err != nil {
			return err
		}
		if action == "merge" {
			updated, err := mergeTransferExhibitionEntry(ctx, db, listID, entry, vehicleID, now)
			if err != nil {
				return err
			}
			if updated {
				continue
			}
		}
		if _, err := db.ExecContext(ctx, `
INSERT INTO exhibition_entries(
  id, list_id, vehicle_id, owner, image_url, locomotive_name, gattung, series, manufacturer, epoch,
  railway_company, day_scope, dt_decoder, decoder_number, decoder_type, adapter, interface_name,
  sx_address, analog, availability, function_keys, notes, sort_order, revision, created_at, updated_at
) VALUES(?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?,
		 NULLIF(?, ''), NULLIF(?, ''), ?, 1, ?, ?)`, randomID(), listID, vehicleID, entry.Owner, entry.ImageURL,
			entry.LocomotiveName, entry.Gattung, entry.Series, entry.Manufacturer, entry.Epoch, entry.RailwayCompany,
			entry.DayScope, entry.DTDecoder, entry.DecoderNumber, entry.DecoderType, entry.Adapter,
			entry.InterfaceName, entry.SXAddress, entry.Analog, entry.Availability, entry.FunctionKeys,
			entry.Notes, entry.SortOrder, now, now); err != nil {
			return err
		}
	}
	return nil
}

func mergeTransferExhibitionEntry(
	ctx context.Context,
	db dataTransferApplyDB,
	listID string,
	entry transfer.TransferExhibitionEntry,
	vehicleID string,
	now string,
) (bool, error) {
	if strings.TrimSpace(entry.ID) == "" {
		return false, nil
	}
	var owner, locomotiveName string
	err := db.QueryRowContext(ctx, `
SELECT owner, locomotive_name FROM exhibition_entries WHERE id=? AND list_id=?`, entry.ID, listID).
		Scan(&owner, &locomotiveName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !strings.EqualFold(strings.TrimSpace(owner), strings.TrimSpace(entry.Owner)) ||
		!strings.EqualFold(strings.TrimSpace(locomotiveName), strings.TrimSpace(entry.LocomotiveName)) {
		return false, nil
	}
	result, err := db.ExecContext(ctx, `
UPDATE exhibition_entries SET
  vehicle_id=?, owner=?, image_url=NULLIF(?, ''), locomotive_name=?, gattung=?, series=?,
  manufacturer=?, epoch=?, railway_company=?, day_scope=?, dt_decoder=?, decoder_number=NULLIF(?, ''),
  decoder_type=?, adapter=?, interface_name=?, sx_address=?, analog=?, availability=?,
  function_keys=NULLIF(?, ''), notes=NULLIF(?, ''), sort_order=?, revision=revision+1, updated_at=?
WHERE id=? AND list_id=?`, vehicleID, entry.Owner, entry.ImageURL, entry.LocomotiveName, entry.Gattung,
		entry.Series, entry.Manufacturer, entry.Epoch, entry.RailwayCompany, entry.DayScope, entry.DTDecoder,
		entry.DecoderNumber, entry.DecoderType, entry.Adapter, entry.InterfaceName, entry.SXAddress, entry.Analog,
		entry.Availability, entry.FunctionKeys, entry.Notes, entry.SortOrder, now, entry.ID, listID)
	if err != nil {
		return false, err
	}
	if err := requireApplyUpdate(result, "merge exhibition entry"); err != nil {
		return false, err
	}
	return true, nil
}

func resolveTransferExhibitionVehicle(
	ctx context.Context,
	db dataTransferApplyDB,
	entry transfer.TransferExhibitionEntry,
	issues []transfer.DataTransferIssue,
	entryIndex int,
) (string, error) {
	if entry.VehicleID == "" && entry.VehicleInventoryNumber == "" {
		return "", nil
	}
	var id, inventoryNumber string
	if entry.VehicleID != "" {
		err := db.QueryRowContext(ctx, `SELECT id, inventory_number FROM vehicles WHERE id=?`, entry.VehicleID).Scan(
			&id, &inventoryNumber,
		)
		if err == nil && strings.EqualFold(strings.TrimSpace(inventoryNumber), strings.TrimSpace(entry.VehicleInventoryNumber)) {
			return id, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	var issue *transfer.DataTransferIssue
	field := fmt.Sprintf("entries[%d].vehicleReference", entryIndex)
	for index := range issues {
		if issues[index].Field == field && (issues[index].Code == "exhibition_vehicle_reference" ||
			issues[index].Code == "missing_vehicle_reference") {
			issue = &issues[index]
			break
		}
	}
	if issue == nil {
		return "", dataTransferApplyConflict("exhibition vehicle resolution is missing")
	}
	if issue.SelectedResolution == "skip" {
		return "", nil
	}
	if entry.VehicleInventoryNumber != "" {
		err := db.QueryRowContext(ctx, `SELECT id FROM vehicles WHERE inventory_number=? COLLATE NOCASE`,
			entry.VehicleInventoryNumber).Scan(&id)
		if err == nil {
			if issue.Code == "exhibition_vehicle_reference" && issue.SelectedResolution == "link" {
				return id, nil
			}
			return "", dataTransferApplyConflict("exhibition vehicle resolution does not match preview")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	return "", dataTransferApplyConflict("exhibition vehicle target changed")
}
