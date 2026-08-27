package application

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type DataTransferVehicleSetPreview struct {
	RecordKey         string                         `json:"recordKey"`
	Classification    string                         `json:"classification"`
	ProposedAction    string                         `json:"proposedAction"`
	TargetID          string                         `json:"targetId,omitempty"`
	TargetUpdatedAt   string                         `json:"targetUpdatedAt,omitempty"`
	TargetFingerprint string                         `json:"targetFingerprint,omitempty"`
	MemberRecordKeys  []string                       `json:"memberRecordKeys"`
	RowNumbers        []int                          `json:"rowNumbers,omitempty"`
	Diagnostics       []TransferVehicleSetDiagnostic `json:"diagnostics,omitempty"`
	Data              TransferVehicleSet             `json:"data"`
}

func ValidateTransferVehicleSet(set TransferVehicleSet) error {
	if len(validateTransferVehicleSetStructure(set)) != 0 {
		return ErrDataTransferValidation
	}
	return nil
}

func classifyDataTransferVehicleSets(
	jobID string,
	incoming DataTransferSnapshot,
	current DataTransferSnapshot,
	vehicleRecords []DataTransferPreviewRecord,
) ([]DataTransferVehicleSetPreview, []DataTransferIssue) {
	previews := make([]DataTransferVehicleSetPreview, 0, len(incoming.VehicleSets))
	issues := []DataTransferIssue{}
	vehiclesByID := map[string]TransferVehicle{}
	vehiclesByInventory := map[string]TransferVehicle{}
	recordsByID := map[string]DataTransferPreviewRecord{}
	recordsByInventory := map[string]DataTransferPreviewRecord{}
	for index, vehicle := range incoming.Vehicles {
		if vehicle.ID != "" {
			vehiclesByID[vehicle.ID] = vehicle
		}
		vehiclesByInventory[transferIdentity(vehicle.InventoryNumber)] = vehicle
		if index < len(vehicleRecords) {
			record := vehicleRecords[index]
			if vehicle.ID != "" {
				recordsByID[vehicle.ID] = record
			}
			recordsByInventory[transferIdentity(vehicle.InventoryNumber)] = record
		}
	}

	currentSetsByInventory := map[string]TransferVehicleSet{}
	for _, set := range current.VehicleSets {
		currentSetsByInventory[transferIdentity(set.InventoryNumber)] = set
	}
	seenMembers := map[string]string{}
	seenSetInventories := map[string]bool{}
	for index, set := range incoming.VehicleSets {
		recordKey := transferVehicleSetRecordKey(set, index)
		preview := DataTransferVehicleSetPreview{
			RecordKey: recordKey, Classification: "ready", ProposedAction: "create", Data: set,
		}
		preview.Diagnostics = append(preview.Diagnostics, validateTransferVehicleSetStructure(set)...)
		setInventoryIdentity := transferIdentity(set.InventoryNumber)
		if setInventoryIdentity != "" && seenSetInventories[setInventoryIdentity] {
			preview.Diagnostics = append(preview.Diagnostics, TransferVehicleSetDiagnostic{
				Field: "inventoryNumber", Code: "duplicate_import_vehicle_set_inventory_number",
			})
		} else if setInventoryIdentity != "" {
			seenSetInventories[setInventoryIdentity] = true
		}
		memberTargetIDs := make([]string, 0, len(set.Members))
		for memberIndex, member := range set.Members {
			if member.SourceRowNumber > 0 {
				preview.RowNumbers = append(preview.RowNumbers, member.SourceRowNumber)
			}
			vehicle, record, found := resolveTransferVehicleSetMember(
				member, vehiclesByID, vehiclesByInventory, recordsByID, recordsByInventory,
			)
			if !found {
				preview.Diagnostics = append(preview.Diagnostics, TransferVehicleSetDiagnostic{
					RowNumber: member.SourceRowNumber,
					Field:     transferVehicleSetMemberField(memberIndex),
					Code:      "missing_vehicle_set_member_reference",
				})
				continue
			}
			identity := transferVehicleSetMemberIdentity(vehicle)
			if previousSet, duplicate := seenMembers[identity]; duplicate && previousSet != recordKey {
				preview.Diagnostics = append(preview.Diagnostics, TransferVehicleSetDiagnostic{
					RowNumber: member.SourceRowNumber,
					Field:     transferVehicleSetMemberField(memberIndex),
					Code:      "vehicle_set_member_in_multiple_sets",
				})
			} else if identity != "" {
				seenMembers[identity] = recordKey
			}
			preview.MemberRecordKeys = append(preview.MemberRecordKeys, record.RecordKey)
			if record.TargetID != "" {
				memberTargetIDs = append(memberTargetIDs, record.TargetID)
			}
			if record.Classification == "error" {
				preview.Diagnostics = append(preview.Diagnostics, TransferVehicleSetDiagnostic{
					RowNumber: member.SourceRowNumber,
					Field:     transferVehicleSetMemberField(memberIndex),
					Code:      "invalid_vehicle_set_member",
				})
			}
		}
		slices.Sort(preview.RowNumbers)
		preview.RowNumbers = slices.Compact(preview.RowNumbers)

		start := len(issues)
		for _, diagnostic := range preview.Diagnostics {
			rowNumber := transferVehicleSetDiagnosticRow(diagnostic.RowNumber)
			issues = append(issues, newTransferIssue(
				jobID, TransferVehicles, recordKey, rowNumber, diagnostic.Field, TransferIssueError,
				diagnostic.Code, transferVehicleSetDiagnosticMessage(diagnostic.Code), "",
			))
		}
		if len(preview.Diagnostics) == 0 {
			target, hasTarget := currentSetsByInventory[transferIdentity(set.InventoryNumber)]
			targetMemberIDs := map[string]bool{}
			if hasTarget {
				preview.TargetID = target.ID
				preview.TargetUpdatedAt = target.UpdatedAt
				preview.TargetFingerprint = DataTransferTargetFingerprint(target)
				for _, member := range target.Members {
					targetMemberIDs[member.SourceVehicleID] = true
				}
			}
			externalConflict := false
			for _, targetID := range memberTargetIDs {
				if !hasTarget || !targetMemberIDs[targetID] {
					externalConflict = true
					break
				}
			}
			if externalConflict {
				preview.ProposedAction = "copy"
				issues = append(issues, newTransferIssue(
					jobID, TransferVehicles, recordKey, nil, "members", TransferIssueWarning,
					"vehicle_set_member_external_conflict",
					"A set member conflicts with a vehicle outside the target set.", "copy",
				))
			} else if hasTarget {
				preview.ProposedAction = "replace"
				issues = append(issues, newTransferIssue(
					jobID, TransferVehicles, recordKey, nil, "inventoryNumber", TransferIssueWarning,
					"duplicate_vehicle_set_inventory_number", "Vehicle set inventory number already exists.",
					"replace_or_copy",
				))
			}
		}
		finalizeDataTransferVehicleSetPreview(&preview, issues[start:])
		previews = append(previews, preview)
	}
	return previews, issues
}

func validateTransferVehicleSetStructure(set TransferVehicleSet) []TransferVehicleSetDiagnostic {
	diagnostics := slices.Clone(set.Diagnostics)
	if strings.TrimSpace(set.InventoryNumber) == "" {
		diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{
			Field: "inventoryNumber", Code: "missing_vehicle_set_inventory_number",
		})
	}
	if !isValidVehicleSetInput(cleanVehicleSetInput(set.VehicleSetInput)) {
		diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{Field: "record", Code: "invalid_vehicle_set"})
	}
	if len(set.Members) < 2 {
		diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{Field: "members", Code: "vehicle_set_too_small"})
	}
	if len(set.Members) > maxVehicleSetMembers {
		diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{Field: "members", Code: "vehicle_set_too_large"})
	}
	positions := make(map[int]bool, len(set.Members))
	members := make(map[string]bool, len(set.Members))
	for index, member := range set.Members {
		field := transferVehicleSetMemberField(index)
		if member.Position <= 0 {
			diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{
				RowNumber: member.SourceRowNumber, Field: field + ".position", Code: "invalid_vehicle_set_position",
			})
		} else if positions[member.Position] {
			diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{
				RowNumber: member.SourceRowNumber, Field: field + ".position", Code: "duplicate_vehicle_set_position",
			})
		}
		positions[member.Position] = true
		memberIdentity := transferIdentity(member.SourceVehicleID)
		if memberIdentity == "" {
			memberIdentity = transferIdentity(member.VehicleInventoryNumber)
		}
		if memberIdentity != "" && members[memberIdentity] {
			diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{
				RowNumber: member.SourceRowNumber, Field: field, Code: "duplicate_vehicle_set_member",
			})
		}
		members[memberIdentity] = true
		if member.SourceRowNumber > 0 {
			if member.DeclaredMemberCount != len(set.Members) {
				diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{
					RowNumber: member.SourceRowNumber, Field: "vehicleSetMemberCount",
					Code: "vehicle_set_member_count_mismatch",
				})
			}
			if member.SourceSetInput != set.VehicleSetInput {
				diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{
					RowNumber: member.SourceRowNumber, Field: "vehicleSet", Code: "conflicting_vehicle_set_metadata",
				})
			}
		}
	}
	for position := 1; position <= len(set.Members); position++ {
		if !positions[position] {
			diagnostics = append(diagnostics, TransferVehicleSetDiagnostic{
				Field: "members.position", Code: "non_contiguous_vehicle_set_positions",
			})
			break
		}
	}
	return compactTransferVehicleSetDiagnostics(diagnostics)
}

func resolveTransferVehicleSetMember(
	member TransferVehicleSetMember,
	vehiclesByID map[string]TransferVehicle,
	vehiclesByInventory map[string]TransferVehicle,
	recordsByID map[string]DataTransferPreviewRecord,
	recordsByInventory map[string]DataTransferPreviewRecord,
) (TransferVehicle, DataTransferPreviewRecord, bool) {
	if member.SourceVehicleID != "" {
		vehicle, found := vehiclesByID[member.SourceVehicleID]
		if !found || member.VehicleInventoryNumber != "" &&
			transferIdentity(vehicle.InventoryNumber) != transferIdentity(member.VehicleInventoryNumber) {
			return TransferVehicle{}, DataTransferPreviewRecord{}, false
		}
		return vehicle, recordsByID[member.SourceVehicleID], true
	}
	identity := transferIdentity(member.VehicleInventoryNumber)
	vehicle, found := vehiclesByInventory[identity]
	if !found || identity == "" {
		return TransferVehicle{}, DataTransferPreviewRecord{}, false
	}
	return vehicle, recordsByInventory[identity], true
}

func transferVehicleSetRecordKey(set TransferVehicleSet, index int) string {
	if strings.TrimSpace(set.ID) != "" {
		return "vehicle-set:id:" + strings.TrimSpace(set.ID)
	}
	if identity := transferIdentity(set.InventoryNumber); identity != "" {
		return "vehicle-set:inventory:" + identity
	}
	for _, member := range set.Members {
		if member.SourceRowNumber > 0 {
			return "vehicle-set:row:" + strconv.Itoa(member.SourceRowNumber)
		}
	}
	return "vehicle-set:index:" + strconv.Itoa(index+1)
}

func transferVehicleSetMemberIdentity(vehicle TransferVehicle) string {
	if vehicle.ID != "" {
		return "id:" + vehicle.ID
	}
	return "inventory:" + transferIdentity(vehicle.InventoryNumber)
}

func transferVehicleSetMemberField(index int) string {
	return fmt.Sprintf("members[%d]", index)
}

func transferVehicleSetDiagnosticRow(row int) *int {
	if row <= 0 {
		return nil
	}
	return &row
}

func compactTransferVehicleSetDiagnostics(
	diagnostics []TransferVehicleSetDiagnostic,
) []TransferVehicleSetDiagnostic {
	seen := map[string]bool{}
	result := make([]TransferVehicleSetDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		key := fmt.Sprintf("%d\x00%s\x00%s", diagnostic.RowNumber, diagnostic.Field, diagnostic.Code)
		if !seen[key] {
			seen[key] = true
			result = append(result, diagnostic)
		}
	}
	return result
}

func transferVehicleSetDiagnosticMessage(code string) string {
	messages := map[string]string{
		"missing_vehicle_set_inventory_number":          "Vehicle set inventory number is required.",
		"duplicate_import_vehicle_set_inventory_number": "Vehicle set inventory number occurs more than once in the import.",
		"invalid_vehicle_set":                           "Vehicle set data violates aggregate validation.",
		"vehicle_set_too_small":                         "A vehicle set requires at least two members.",
		"vehicle_set_too_large":                         "A vehicle set cannot contain more than 100 members.",
		"invalid_vehicle_set_position":                  "Vehicle set member position must be a positive integer.",
		"duplicate_vehicle_set_position":                "Vehicle set member positions must be unique.",
		"non_contiguous_vehicle_set_positions":          "Vehicle set member positions must be contiguous.",
		"invalid_vehicle_set_member_count":              "Vehicle set member count must be an integer.",
		"vehicle_set_member_count_mismatch":             "Vehicle set member count does not match the detected group.",
		"conflicting_vehicle_set_metadata":              "Vehicle set rows contain conflicting metadata.",
		"duplicate_vehicle_set_member":                  "A vehicle occurs more than once in the vehicle set.",
		"vehicle_set_member_in_multiple_sets":           "A vehicle belongs to more than one imported set.",
		"missing_vehicle_set_member_reference":          "Referenced vehicle is missing from the import.",
		"invalid_vehicle_set_member":                    "A vehicle set member contains blocking validation errors.",
	}
	if message := messages[code]; message != "" {
		return message
	}
	return "Vehicle set data is invalid."
}

func finalizeDataTransferVehicleSetPreview(
	preview *DataTransferVehicleSetPreview,
	issues []DataTransferIssue,
) {
	preview.Classification = "ready"
	for _, issue := range issues {
		if issue.Severity == TransferIssueError {
			preview.Classification = "error"
			return
		}
		if issue.Severity == TransferIssueWarning {
			preview.Classification = "warning"
		}
	}
}

func suppressDataTransferVehicleSetMemberConflicts(
	records []DataTransferPreviewRecord,
	issues []DataTransferIssue,
	vehicleSets []DataTransferVehicleSetPreview,
) ([]DataTransferPreviewRecord, []DataTransferIssue) {
	groupedMembers := map[string]bool{}
	for _, set := range vehicleSets {
		if set.Classification != "warning" {
			continue
		}
		for _, recordKey := range set.MemberRecordKeys {
			groupedMembers[recordKey] = true
		}
	}
	if len(groupedMembers) == 0 {
		return records, issues
	}
	filtered := make([]DataTransferIssue, 0, len(issues))
	for _, issue := range issues {
		if issue.Area == TransferVehicles && groupedMembers[issue.RecordKey] &&
			issue.Code == "duplicate_inventory_number" {
			continue
		}
		filtered = append(filtered, issue)
	}
	for index := range records {
		if records[index].Area == TransferVehicles && groupedMembers[records[index].RecordKey] {
			finalizeDataTransferRecord(&records[index], transferIssuesForPreviewRecord(filtered, records[index]))
		}
	}
	return records, filtered
}
