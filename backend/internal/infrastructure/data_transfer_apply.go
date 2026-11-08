package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

type dataTransferPersistedPreview struct {
	SourceSHA256       string                                   `json:"sourceSha256"`
	FingerprintVersion int                                      `json:"fingerprintVersion,omitempty"`
	Records            []transfer.DataTransferPreviewRecord     `json:"records"`
	VehicleSets        []transfer.DataTransferVehicleSetPreview `json:"vehicleSets,omitempty"`
	CSVMapping         []transfer.DataTransferCSVColumnMapping  `json:"csvMapping,omitempty"`
	VehicleFields      []transfer.VehicleTransferField          `json:"vehicleFields,omitempty"`
}

type dataTransferApplyDB interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type dataTransferApplyResult struct {
	Applied  int
	Skipped  int
	Vehicles map[string]transferVehicleApplyResult
}

type transferVehicleApplyResult struct {
	ID              string
	InventoryNumber string
	Created         bool `json:"-"`
}

func (repository *DataTransferRepository) ApplyImport(
	ctx context.Context,
	job transfer.DataTransferJob,
	actor string,
) error {
	return repository.ApplyImportWithPolicy(ctx, job, actor, transfer.DataTransferImportPolicy{
		CanManageExhibitionLists: true,
	})
}

func (repository *DataTransferRepository) ApplyImportWithPolicy(
	ctx context.Context,
	job transfer.DataTransferJob,
	actor string,
	policy transfer.DataTransferImportPolicy,
) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin data transfer import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	preview, issues, err := revalidateTransferPreview(ctx, tx, job)
	if err != nil {
		return err
	}
	result, err := applyTransferRecords(
		ctx, tx, job, preview.Records, preview.VehicleSets, preview.CSVMapping, issues, policy,
	)
	if err != nil {
		return err
	}
	if err := applyTransferVehicleSets(
		ctx, tx, preview.VehicleSets, preview.CSVMapping, issues, result.Vehicles, actor,
	); err != nil {
		return err
	}
	now := timestamp()
	completedState := transfer.TransferJobCompleted
	if job.WarningRecords > 0 || job.ErrorRecords > 0 || result.Skipped > 0 {
		completedState = transfer.TransferJobCompletedWithWarnings
	}
	update, err := tx.ExecContext(ctx, `
UPDATE data_transfer_jobs
SET state=?, stage='completed', confirmed_by_user_id=NULLIF(?, ''), confirmed_at=?, completed_at=?,
    result_message=?, revision=revision+1, updated_at=?
WHERE id=? AND direction=? AND state=? AND revision=? AND source_sha256=?`, completedState, actor, now, now,
		fmt.Sprintf("Imported %d record(s); skipped %d.", result.Applied, result.Skipped), now, job.ID,
		transfer.TransferImport, transfer.TransferJobReady, job.Revision, job.SourceSHA256)
	if err != nil {
		return fmt.Errorf("complete data transfer import: %w", err)
	}
	affected, err := update.RowsAffected()
	if err != nil {
		return fmt.Errorf("read data transfer completion result: %w", err)
	}
	if affected != 1 {
		return dataTransferApplyConflict("import job changed during apply")
	}
	if err := writeTransferAudit(ctx, tx, job, actor, now, result); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit data transfer import: %w", err)
	}
	return nil
}

func revalidateTransferPreview(
	ctx context.Context,
	tx *sql.Tx,
	job transfer.DataTransferJob,
) (dataTransferPersistedPreview, []transfer.DataTransferIssue, error) {
	var direction transfer.TransferDirection
	var state transfer.TransferJobState
	var revision, totalRecords int
	var sourceSHA, previewJSON string
	err := tx.QueryRowContext(ctx, `
SELECT direction, state, revision, total_records, source_sha256, preview_json
FROM data_transfer_jobs WHERE id=?`, job.ID).Scan(
		&direction, &state, &revision, &totalRecords, &sourceSHA, &previewJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import job no longer exists")
	}
	if err != nil {
		return dataTransferPersistedPreview{}, nil, fmt.Errorf("load data transfer import for apply: %w", err)
	}
	if direction != transfer.TransferImport || state != transfer.TransferJobReady ||
		revision != job.Revision || sourceSHA == "" || sourceSHA != job.SourceSHA256 {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import preview is stale")
	}
	preview := dataTransferPersistedPreview{}
	decoder := json.NewDecoder(strings.NewReader(previewJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&preview); err != nil {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("persisted import preview is invalid")
	}
	if preview.SourceSHA256 == "" || preview.SourceSHA256 != sourceSHA || len(preview.Records) != totalRecords {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import source or preview changed")
	}
	issues, err := transferApplyIssues(ctx, tx, job.ID)
	if err != nil {
		return dataTransferPersistedPreview{}, nil, err
	}
	for _, issue := range issues {
		if !transferpolicy.ValidTransferApplyResolution(issue.Code, issue.SelectedResolution) {
			return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import has unresolved conflicts")
		}
	}
	if err := revalidateTransferVehicleSetPreview(ctx, tx, preview.VehicleSets, preview.Records, issues); err != nil {
		return dataTransferPersistedPreview{}, nil, err
	}
	targetFingerprints, err := currentTransferTargetFingerprints(
		ctx, tx, preview.Records, preview.FingerprintVersion,
	)
	if err != nil {
		return dataTransferPersistedPreview{}, nil, err
	}
	for _, record := range preview.Records {
		if !slices.Contains(job.Areas, record.Area) {
			return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("preview contains an unselected area")
		}
		if record.TargetID != "" {
			if record.TargetFingerprint == "" ||
				targetFingerprints[record.Area][record.TargetID] != record.TargetFingerprint {
				return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("preview target changed")
			}
		}
	}
	return preview, issues, nil
}

func transferApplyIssues(
	ctx context.Context,
	db dataTransferApplyDB,
	jobID string,
) ([]transfer.DataTransferIssue, error) {
	rows, err := db.QueryContext(ctx, dataTransferIssueSelect+`
WHERE job_id=? ORDER BY created_at, id`, jobID)
	if err != nil {
		return nil, fmt.Errorf("load data transfer issues for apply: %w", err)
	}
	defer func() { _ = rows.Close() }()
	issues := []transfer.DataTransferIssue{}
	for rows.Next() {
		issue, err := scanDataTransferIssue(rows)
		if err != nil {
			return nil, fmt.Errorf("scan data transfer issue for apply: %w", err)
		}
		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate data transfer issues for apply: %w", err)
	}
	return issues, nil
}

func currentTransferTargetFingerprints(
	ctx context.Context,
	tx *sql.Tx,
	records []transfer.DataTransferPreviewRecord,
	fingerprintVersion int,
) (map[transfer.TransferArea]map[string]string, error) {
	wanted := map[transfer.TransferArea]bool{}
	for _, record := range records {
		if record.TargetID != "" {
			wanted[record.Area] = true
		}
	}
	fingerprints := map[transfer.TransferArea]map[string]string{}
	if wanted[transfer.TransferVehicles] {
		vehicles, err := transferVehicleSnapshot(ctx, tx)
		if err != nil {
			return nil, err
		}
		fingerprints[transfer.TransferVehicles] = map[string]string{}
		for _, vehicle := range vehicles {
			fingerprints[transfer.TransferVehicles][vehicle.ID] =
				transfer.DataTransferTargetFingerprintForVersion(vehicle, fingerprintVersion)
		}
	}
	if wanted[transfer.TransferAccessories] {
		accessories, err := transferAccessorySnapshot(ctx, tx)
		if err != nil {
			return nil, err
		}
		fingerprints[transfer.TransferAccessories] = map[string]string{}
		for _, accessory := range accessories {
			fingerprints[transfer.TransferAccessories][accessory.ID] =
				transfer.DataTransferTargetFingerprintForVersion(accessory, fingerprintVersion)
		}
	}
	if wanted[transfer.TransferExhibitionLists] {
		lists, err := transferExhibitionSnapshot(ctx, tx)
		if err != nil {
			return nil, err
		}
		fingerprints[transfer.TransferExhibitionLists] = map[string]string{}
		for _, list := range lists {
			fingerprints[transfer.TransferExhibitionLists][list.ID] =
				transfer.DataTransferTargetFingerprintForVersion(list, fingerprintVersion)
		}
	}
	return fingerprints, nil
}

func applyTransferRecords(
	ctx context.Context,
	db *sql.Tx,
	job transfer.DataTransferJob,
	records []transfer.DataTransferPreviewRecord,
	vehicleSets []transfer.DataTransferVehicleSetPreview,
	csvMapping []transfer.DataTransferCSVColumnMapping,
	issues []transfer.DataTransferIssue,
	policy transfer.DataTransferImportPolicy,
) (dataTransferApplyResult, error) {
	_, memberPolicies, err := transferpolicy.TransferVehicleSetApplyPolicies(vehicleSets, issues)
	if err != nil {
		return dataTransferApplyResult{}, err
	}
	result := dataTransferApplyResult{Vehicles: map[string]transferVehicleApplyResult{}}
	var csvVehicleTargets map[string]transfer.TransferVehicle
	csvVehicleTargetsLoaded := false
	for _, record := range records {
		recordIssues := transferpolicy.TransferRecordIssues(issues, record)
		action, skip := transferpolicy.TransferRecordAction(record, recordIssues)
		if setPolicy, member := memberPolicies[record.RecordKey]; record.Area == transfer.TransferVehicles && member {
			switch setPolicy.Action {
			case "skip":
				skip = true
			case "copy":
				if record.TargetID != "" {
					action = "copy"
				} else {
					action = "create"
				}
			case "replace":
				if record.TargetID != "" {
					action = "replace"
				} else {
					action = "create"
				}
			case "create":
				if record.TargetID != "" {
					return dataTransferApplyResult{}, dataTransferApplyConflict(
						"vehicle set member conflict has no group resolution",
					)
				}
				action = "create"
			}
		}
		if skip {
			result.Skipped++
			continue
		}
		if record.Area == transfer.TransferExhibitionLists && !policy.CanManageExhibitionLists &&
			action != "merge" {
			return dataTransferApplyResult{}, transfer.ErrDataTransferForbidden
		}
		var err error
		switch record.Area {
		case transfer.TransferVehicles:
			var existingCSVVehicle *transfer.TransferVehicle
			if action == "replace" && len(csvMapping) > 0 &&
				job.PackageVersion != transfer.DataTransferPackageLegacyVersion {
				if !csvVehicleTargetsLoaded {
					vehicles, snapshotErr := transferVehicleSnapshot(ctx, db)
					if snapshotErr != nil {
						return dataTransferApplyResult{}, snapshotErr
					}
					csvVehicleTargets = make(map[string]transfer.TransferVehicle, len(vehicles))
					for _, vehicle := range vehicles {
						csvVehicleTargets[vehicle.ID] = vehicle
					}
					csvVehicleTargetsLoaded = true
				}
				if existing, found := csvVehicleTargets[record.TargetID]; found {
					existingCSVVehicle = &existing
				}
			}
			vehicleResult, vehicleErr := applyTransferVehicle(
				ctx, db, record, action, job.PackageVersion, csvMapping, existingCSVVehicle,
			)
			err = vehicleErr
			if err == nil {
				if _, duplicate := result.Vehicles[record.RecordKey]; duplicate {
					return dataTransferApplyResult{}, dataTransferApplyConflict("duplicate vehicle record key")
				}
				result.Vehicles[record.RecordKey] = vehicleResult
			}
		case transfer.TransferAccessories:
			err = applyTransferAccessory(ctx, db, record, action)
		case transfer.TransferExhibitionLists:
			err = applyTransferExhibitionList(ctx, db, record, action, recordIssues)
		default:
			err = dataTransferApplyConflict("preview contains an unsupported area")
		}
		if err != nil {
			return dataTransferApplyResult{}, fmt.Errorf("apply %s record %q: %w", record.Area, record.RecordKey, err)
		}
		result.Applied++
	}
	if result.Applied+result.Skipped != job.TotalRecords {
		return dataTransferApplyResult{}, dataTransferApplyConflict("preview record counts changed")
	}
	return result, nil
}

func uniqueTransferInventoryNumber(
	ctx context.Context,
	db dataTransferApplyDB,
	table string,
	base string,
) (string, error) {
	query := map[string]string{
		"vehicles":           `SELECT EXISTS(SELECT 1 FROM vehicles WHERE inventory_number=? COLLATE NOCASE)`,
		"vehicle_sets":       `SELECT EXISTS(SELECT 1 FROM vehicle_sets WHERE inventory_number=? COLLATE NOCASE)`,
		"accessory_products": `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE inventory_number=? COLLATE NOCASE)`,
		"accessory_assets":   `SELECT EXISTS(SELECT 1 FROM accessory_assets WHERE inventory_number=? COLLATE NOCASE)`,
	}[table]
	if query == "" {
		return "", dataTransferApplyConflict("unsupported copied record")
	}
	base = strings.TrimSpace(base) + "-COPY"
	for suffix := 1; suffix <= 10_000; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		var exists int
		if err := db.QueryRowContext(ctx, query, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if exists == 0 {
			return candidate, nil
		}
	}
	return "", dataTransferApplyConflict("could not allocate copy inventory number")
}

func requireApplyUpdate(result sql.Result, operation string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s rows affected: %w", operation, err)
	}
	if affected != 1 {
		return dataTransferApplyConflict(operation + " target changed")
	}
	return nil
}

func writeTransferAudit(
	ctx context.Context,
	db dataTransferApplyDB,
	job transfer.DataTransferJob,
	actor string,
	createdAt string,
	result dataTransferApplyResult,
) error {
	details, err := json.Marshal(map[string]any{
		"areas": job.Areas, "appliedRecords": result.Applied, "skippedRecords": result.Skipped,
		"sourceSha256": job.SourceSHA256, "totalRecords": job.TotalRecords,
	})
	if err != nil {
		return fmt.Errorf("encode data transfer audit: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, NULLIF(?, ''), 'data_transfer.import_applied', 'data_transfer_job', ?, ?, ?)`,
		randomID(), actor, job.ID, createdAt, string(details)); err != nil {
		return fmt.Errorf("write data transfer audit: %w", err)
	}
	return nil
}

func dataTransferApplyConflict(message string) error {
	return fmt.Errorf("%w: %s", transfer.ErrDataTransferConflict, message)
}
