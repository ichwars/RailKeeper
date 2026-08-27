package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"railkeeper/backend/internal/application"
)

type transferVehicleSetApplyPolicy struct {
	Action           string
	TargetSetID      string
	MemberRecordKeys map[string]bool
}

func transferVehicleSetSnapshot(ctx context.Context, tx *sql.Tx) ([]application.TransferVehicleSet, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, inventory_number, name, manufacturer, article_number, article_source_url, gauge, epoch,
       railway_company, category, gattung, description, ean, production_period, list_price,
       acquisition_type, acquired_from, purchase_price, purchase_date, storage_location,
       storage_details, condition, condition_details, packaging, created_at, updated_at
FROM vehicle_sets
ORDER BY inventory_number COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("query transfer vehicle sets: %w", err)
	}
	sets := []application.TransferVehicleSet{}
	setIndexes := map[string]int{}
	for rows.Next() {
		set := application.TransferVehicleSet{}
		if err := rows.Scan(
			&set.ID, &set.InventoryNumber, &set.Name, &set.Manufacturer, &set.ArticleNumber,
			&set.ArticleSourceURL, &set.Gauge, &set.Epoch, &set.RailwayCompany, &set.Category,
			&set.Gattung, &set.Description, &set.EAN, &set.ProductionPeriod, &set.ListPrice,
			&set.AcquisitionType, &set.AcquiredFrom, &set.PurchasePrice, &set.PurchaseDate,
			&set.StorageLocation, &set.StorageDetails, &set.Condition, &set.ConditionDetails,
			&set.Packaging, &set.CreatedAt, &set.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan transfer vehicle set: %w", err)
		}
		setIndexes[set.ID] = len(sets)
		sets = append(sets, set)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate transfer vehicle sets: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close transfer vehicle sets: %w", err)
	}

	memberRows, err := tx.QueryContext(ctx, `
SELECT member.vehicle_set_id, member.vehicle_id, vehicle.inventory_number, member.position, member.label
FROM vehicle_set_members member
JOIN vehicles vehicle ON vehicle.id=member.vehicle_id
ORDER BY member.vehicle_set_id, member.position, member.vehicle_id`)
	if err != nil {
		return nil, fmt.Errorf("query transfer vehicle set members: %w", err)
	}
	defer func() { _ = memberRows.Close() }()
	for memberRows.Next() {
		var setID string
		member := application.TransferVehicleSetMember{}
		if err := memberRows.Scan(
			&setID, &member.SourceVehicleID, &member.VehicleInventoryNumber, &member.Position, &member.Label,
		); err != nil {
			return nil, fmt.Errorf("scan transfer vehicle set member: %w", err)
		}
		index, found := setIndexes[setID]
		if !found {
			return nil, fmt.Errorf("transfer vehicle set member references unknown set %q", setID)
		}
		sets[index].Members = append(sets[index].Members, member)
	}
	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer vehicle set members: %w", err)
	}
	return sets, nil
}

func revalidateTransferVehicleSetPreview(
	ctx context.Context,
	tx *sql.Tx,
	previews []application.DataTransferVehicleSetPreview,
	records []application.DataTransferPreviewRecord,
	issues []application.DataTransferIssue,
) error {
	if len(previews) == 0 {
		return nil
	}
	_, memberPolicies, err := transferVehicleSetApplyPolicies(previews, issues)
	if err != nil {
		return err
	}
	vehicleRecords := map[string]application.DataTransferPreviewRecord{}
	for _, record := range records {
		if record.Area != application.TransferVehicles {
			continue
		}
		if _, duplicate := vehicleRecords[record.RecordKey]; duplicate {
			return dataTransferApplyConflict("vehicle set preview contains duplicate member record keys")
		}
		vehicleRecords[record.RecordKey] = record
	}
	sets, err := transferVehicleSetSnapshot(ctx, tx)
	if err != nil {
		return err
	}
	setsByID := make(map[string]application.TransferVehicleSet, len(sets))
	setsByInventory := make(map[string]application.TransferVehicleSet, len(sets))
	for _, set := range sets {
		setsByID[set.ID] = set
		setsByInventory[strings.ToLower(strings.TrimSpace(set.InventoryNumber))] = set
	}
	seenMembers := map[string]bool{}
	for _, preview := range previews {
		if preview.Classification == "error" || len(preview.Diagnostics) != 0 ||
			application.ValidateTransferVehicleSet(preview.Data) != nil {
			return dataTransferApplyConflict("persisted vehicle set preview is invalid")
		}
		if len(preview.MemberRecordKeys) != len(preview.Data.Members) {
			return dataTransferApplyConflict("vehicle set member preview is incomplete")
		}
		policy := memberPolicies[preview.MemberRecordKeys[0]]
		for index, recordKey := range preview.MemberRecordKeys {
			record, found := vehicleRecords[recordKey]
			if !found || seenMembers[recordKey] || !policy.MemberRecordKeys[recordKey] {
				return dataTransferApplyConflict("vehicle set member preview is inconsistent")
			}
			seenMembers[recordKey] = true
			vehicle := application.TransferVehicle{}
			if err := json.Unmarshal(record.Data, &vehicle); err != nil ||
				!transferVehicleSetMemberMatches(preview.Data.Members[index], vehicle) {
				return dataTransferApplyConflict("vehicle set member reference changed")
			}
		}
		currentByInventory, inventoryExists := setsByInventory[strings.ToLower(strings.TrimSpace(preview.Data.InventoryNumber))]
		if preview.TargetID != "" {
			target, found := setsByID[preview.TargetID]
			if !found || preview.TargetFingerprint == "" ||
				application.DataTransferTargetFingerprint(target) != preview.TargetFingerprint {
				return dataTransferApplyConflict("vehicle set preview target changed")
			}
			if !inventoryExists || currentByInventory.ID != target.ID {
				return dataTransferApplyConflict("vehicle set inventory target changed")
			}
		} else if inventoryExists && policy.Action != "copy" && policy.Action != "skip" {
			return dataTransferApplyConflict("vehicle set inventory number now exists")
		}
		if err := validateTransferVehicleSetTargetMembers(policy, preview, vehicleRecords, setsByID); err != nil {
			return err
		}
	}
	return nil
}

func transferVehicleSetApplyPolicies(
	previews []application.DataTransferVehicleSetPreview,
	issues []application.DataTransferIssue,
) (map[string]transferVehicleSetApplyPolicy, map[string]transferVehicleSetApplyPolicy, error) {
	setPolicies := make(map[string]transferVehicleSetApplyPolicy, len(previews))
	memberPolicies := map[string]transferVehicleSetApplyPolicy{}
	for _, preview := range previews {
		if preview.RecordKey == "" {
			return nil, nil, dataTransferApplyConflict("vehicle set record key is missing")
		}
		if _, duplicate := setPolicies[preview.RecordKey]; duplicate {
			return nil, nil, dataTransferApplyConflict("duplicate vehicle set record key")
		}
		action := preview.ProposedAction
		resolved := false
		for _, issue := range issues {
			if issue.Area != application.TransferVehicles || issue.RecordKey != preview.RecordKey {
				continue
			}
			switch issue.Code {
			case "duplicate_vehicle_set_inventory_number", "vehicle_set_member_external_conflict":
				action = issue.SelectedResolution
				resolved = true
			}
		}
		if preview.Classification == "warning" && !resolved {
			return nil, nil, dataTransferApplyConflict("vehicle set conflict has no resolution")
		}
		if !slices.Contains([]string{"create", "replace", "copy", "skip"}, action) {
			return nil, nil, dataTransferApplyConflict("unsupported vehicle set resolution")
		}
		if action == "replace" && preview.TargetID == "" {
			return nil, nil, dataTransferApplyConflict("vehicle set replacement target is missing")
		}
		policy := transferVehicleSetApplyPolicy{
			Action: action, TargetSetID: preview.TargetID, MemberRecordKeys: map[string]bool{},
		}
		for _, recordKey := range preview.MemberRecordKeys {
			if recordKey == "" || memberPolicies[recordKey].Action != "" {
				return nil, nil, dataTransferApplyConflict("vehicle belongs to more than one imported set")
			}
			policy.MemberRecordKeys[recordKey] = true
		}
		if len(policy.MemberRecordKeys) == 0 {
			return nil, nil, dataTransferApplyConflict("vehicle set has no member records")
		}
		setPolicies[preview.RecordKey] = policy
		for recordKey := range policy.MemberRecordKeys {
			memberPolicies[recordKey] = policy
		}
	}
	return setPolicies, memberPolicies, nil
}

func validateTransferVehicleSetTargetMembers(
	policy transferVehicleSetApplyPolicy,
	preview application.DataTransferVehicleSetPreview,
	records map[string]application.DataTransferPreviewRecord,
	setsByID map[string]application.TransferVehicleSet,
) error {
	if policy.Action == "skip" || policy.Action == "copy" {
		return nil
	}
	if policy.Action == "create" {
		for recordKey := range policy.MemberRecordKeys {
			if records[recordKey].TargetID != "" {
				return dataTransferApplyConflict("new vehicle set would reuse an existing vehicle")
			}
		}
		return nil
	}
	target, found := setsByID[preview.TargetID]
	if !found {
		return dataTransferApplyConflict("vehicle set replacement target is missing")
	}
	targetMemberIDs := map[string]bool{}
	for _, member := range target.Members {
		targetMemberIDs[member.SourceVehicleID] = true
	}
	for recordKey := range policy.MemberRecordKeys {
		targetID := records[recordKey].TargetID
		if targetID != "" && !targetMemberIDs[targetID] {
			return dataTransferApplyConflict("vehicle set member target is outside the replacement set")
		}
	}
	return nil
}

func applyTransferVehicleSets(
	ctx context.Context,
	tx *sql.Tx,
	previews []application.DataTransferVehicleSetPreview,
	csvMapping []application.DataTransferCSVColumnMapping,
	issues []application.DataTransferIssue,
	vehicles map[string]transferVehicleApplyResult,
	actor string,
) error {
	policies, _, err := transferVehicleSetApplyPolicies(previews, issues)
	if err != nil {
		return err
	}
	existingSetsByID := map[string]application.TransferVehicleSet{}
	if len(previews) > 0 && len(csvMapping) > 0 {
		existingSets, err := transferVehicleSetSnapshot(ctx, tx)
		if err != nil {
			return err
		}
		for _, set := range existingSets {
			existingSetsByID[set.ID] = set
		}
	}
	for _, preview := range previews {
		policy := policies[preview.RecordKey]
		if policy.Action == "skip" {
			continue
		}
		if application.ValidateTransferVehicleSet(preview.Data) != nil ||
			len(preview.MemberRecordKeys) != len(preview.Data.Members) {
			return dataTransferApplyConflict("vehicle set preview changed before apply")
		}
		memberResults := make([]transferVehicleApplyResult, len(preview.MemberRecordKeys))
		seenVehicleIDs := map[string]bool{}
		for index, recordKey := range preview.MemberRecordKeys {
			result, found := vehicles[recordKey]
			if !found || result.ID == "" || seenVehicleIDs[result.ID] {
				return dataTransferApplyConflict("vehicle set member apply result is incomplete")
			}
			seenVehicleIDs[result.ID] = true
			memberResults[index] = result
		}
		setData := preview.Data
		setID := preview.TargetID
		inventoryNumber := setData.InventoryNumber
		now := timestamp()
		switch policy.Action {
		case "create":
			setID = randomID()
			if err := insertTransferVehicleSet(ctx, tx, setID, inventoryNumber, setData.VehicleSetInput, now); err != nil {
				return err
			}
		case "copy":
			setID = randomID()
			inventoryNumber, err = uniqueTransferInventoryNumber(ctx, tx, "vehicle_sets", inventoryNumber)
			if err != nil {
				return err
			}
			if err := insertTransferVehicleSet(ctx, tx, setID, inventoryNumber, setData.VehicleSetInput, now); err != nil {
				return err
			}
		case "replace":
			if len(csvMapping) > 0 {
				existing, found := existingSetsByID[setID]
				if !found {
					return dataTransferApplyConflict("vehicle set replacement target is missing")
				}
				setData = application.PreserveUnmappedTransferVehicleSetFields(setData, existing, csvMapping)
				if !transferCSVFieldMapped(csvMapping, "vehicleSetMemberLabel") {
					existingLabels := make(map[string]string, len(existing.Members))
					for _, member := range existing.Members {
						existingLabels[member.SourceVehicleID] = member.Label
					}
					for index, result := range memberResults {
						if label, found := existingLabels[result.ID]; found {
							setData.Members[index].Label = label
						}
					}
				}
				inventoryNumber = setData.InventoryNumber
			}
			if err := updateTransferVehicleSet(
				ctx, tx, setID, inventoryNumber, setData.VehicleSetInput, now,
			); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM vehicle_set_members WHERE vehicle_set_id=?`, setID); err != nil {
				return fmt.Errorf("clear transfer vehicle set members: %w", err)
			}
		}
		for index, member := range preview.Data.Members {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
VALUES(?, ?, ?, ?)`, setID, memberResults[index].ID, member.Position, member.Label); err != nil {
				return fmt.Errorf("insert transfer vehicle set member: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, NULLIF(?, ''), 'VehicleSetImported', 'vehicle_set', ?, ?, ?)`,
			randomID(), actor, setID, now, fmt.Sprintf(`{"action":%q}`, policy.Action)); err != nil {
			return fmt.Errorf("write transfer vehicle set audit log: %w", err)
		}
	}
	return nil
}

func transferCSVFieldMapped(mapping []application.DataTransferCSVColumnMapping, targetField string) bool {
	for _, column := range mapping {
		if column.TargetField == targetField {
			return true
		}
	}
	return false
}

func insertTransferVehicleSet(
	ctx context.Context,
	tx *sql.Tx,
	id string,
	inventoryNumber string,
	input application.VehicleSetInput,
	now string,
) error {
	arguments := transferVehicleSetArguments(inventoryNumber, input)
	_, err := tx.ExecContext(ctx, `
INSERT INTO vehicle_sets(
  id, inventory_number, name, manufacturer, article_number, article_source_url, gauge, epoch,
  railway_company, category, gattung, description, ean, production_period, list_price, acquisition_type,
  acquired_from, purchase_price, purchase_date, storage_location, storage_details, condition,
  condition_details, packaging, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		append([]any{id}, append(arguments, now, now)...)...)
	if err != nil {
		return fmt.Errorf("insert transfer vehicle set: %w", err)
	}
	return nil
}

func updateTransferVehicleSet(
	ctx context.Context,
	tx *sql.Tx,
	id string,
	inventoryNumber string,
	input application.VehicleSetInput,
	now string,
) error {
	arguments := transferVehicleSetArguments(inventoryNumber, input)
	result, err := tx.ExecContext(ctx, `
UPDATE vehicle_sets SET
  inventory_number=?, name=?, manufacturer=?, article_number=?, article_source_url=?, gauge=?, epoch=?,
  railway_company=?, category=?, gattung=?, description=?, ean=?, production_period=?, list_price=?,
  acquisition_type=?, acquired_from=?, purchase_price=?, purchase_date=?, storage_location=?, storage_details=?,
  condition=?, condition_details=?, packaging=?, updated_at=?
WHERE id=?`, append(append(arguments, now), id)...)
	if err != nil {
		return fmt.Errorf("update transfer vehicle set: %w", err)
	}
	return requireApplyUpdate(result, "replace vehicle set")
}

func transferVehicleSetArguments(inventoryNumber string, input application.VehicleSetInput) []any {
	return []any{
		inventoryNumber, input.Name, input.Manufacturer, input.ArticleNumber, input.ArticleSourceURL, input.Gauge,
		input.Epoch, input.RailwayCompany, input.Category, input.Gattung, input.Description, input.EAN,
		input.ProductionPeriod, input.ListPrice, input.AcquisitionType, input.AcquiredFrom, input.PurchasePrice,
		input.PurchaseDate, input.StorageLocation, input.StorageDetails, input.Condition, input.ConditionDetails,
		input.Packaging,
	}
}

func transferVehicleSetMemberMatches(
	member application.TransferVehicleSetMember,
	vehicle application.TransferVehicle,
) bool {
	if member.SourceVehicleID != "" && member.SourceVehicleID != vehicle.ID {
		return false
	}
	return member.VehicleInventoryNumber == "" || strings.EqualFold(
		strings.TrimSpace(member.VehicleInventoryNumber), strings.TrimSpace(vehicle.InventoryNumber),
	)
}
