package policy

import (
	"fmt"

	"strings"
)

func ClassifyDataTransferImport(
	jobID string,
	incoming DataTransferSnapshot,
	current DataTransferSnapshot,
	rowNumbers map[TransferArea][]int,
	rules AggregateRules,
) ([]DataTransferPreviewRecord, []DataTransferIssue) {
	records := make([]DataTransferPreviewRecord, 0,
		len(incoming.Vehicles)+len(incoming.Accessories)+len(incoming.ExhibitionLists))
	issues := []DataTransferIssue{}
	currentVehiclesByInventory := map[string]TransferVehicle{}
	currentVehiclesByID := map[string]TransferVehicle{}
	for _, vehicle := range current.Vehicles {
		currentVehiclesByInventory[transferIdentity(vehicle.InventoryNumber)] = vehicle
		if vehicle.ID != "" {
			currentVehiclesByID[vehicle.ID] = vehicle
		}
	}
	seenVehicleInventory := map[string]bool{}
	for index, vehicle := range incoming.Vehicles {
		rowNumber := transferRowNumber(rowNumbers[TransferVehicles], index)
		recordKey := strings.TrimSpace(vehicle.InventoryNumber)
		if recordKey == "" {
			recordKey = fmt.Sprintf("row-%d", transferRowValue(rowNumber, index+1))
		}
		record := newDataTransferPreviewRecord(TransferVehicles, recordKey, rowNumber, vehicle)
		start := len(issues)
		issues = appendRequiredTransferIssues(issues, jobID, TransferVehicles, recordKey, rowNumber,
			vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.Name)
		for _, required := range []struct {
			field, value, code, message string
		}{
			{"gauge", vehicle.Gauge, "missing_gauge", "Gauge is required."},
			{"category", vehicle.Category, "missing_category", "Category is required."},
			{"gattung", vehicle.Gattung, "missing_gattung", "Gattung is required."},
		} {
			if strings.TrimSpace(required.value) == "" {
				issues = append(issues, NewTransferIssue(jobID, TransferVehicles, recordKey, rowNumber,
					required.field, TransferIssueError, required.code, required.message, "skip"))
			}
		}
		if err := rules.validateVehicle(vehicle); err != nil &&
			strings.TrimSpace(vehicle.Manufacturer) != "" && strings.TrimSpace(vehicle.Name) != "" &&
			strings.TrimSpace(vehicle.Gauge) != "" && strings.TrimSpace(vehicle.Category) != "" &&
			strings.TrimSpace(vehicle.Gattung) != "" {
			issues = append(issues, NewTransferIssue(jobID, TransferVehicles, recordKey, rowNumber, "record",
				TransferIssueError, "invalid_vehicle", "Vehicle data violates aggregate validation.", "skip"))
		}
		identity := transferIdentity(vehicle.InventoryNumber)
		if identity != "" && seenVehicleInventory[identity] {
			issues = append(issues, NewTransferIssue(jobID, TransferVehicles, recordKey, rowNumber, "inventoryNumber",
				TransferIssueError, "duplicate_input_inventory_number", "Inventory number occurs more than once in the upload.", "skip"))
		} else if target, found := currentVehiclesByInventory[identity]; identity != "" && found {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			record.ProposedAction = "replace"
			issues = append(issues, NewTransferIssue(jobID, TransferVehicles, recordKey, rowNumber, "inventoryNumber",
				TransferIssueWarning, "duplicate_inventory_number", "Inventory number already exists.", "replace_or_copy"))
		}
		if identity != "" {
			seenVehicleInventory[identity] = true
		}
		FinalizeDataTransferRecord(&record, issues[start:])
		records = append(records, record)
	}

	currentAccessoriesByInventory := map[string]TransferAccessory{}
	currentAccessoriesByArticle := map[string]TransferAccessory{}
	for _, accessory := range current.Accessories {
		currentAccessoriesByInventory[transferIdentity(accessory.InventoryNumber)] = accessory
		if key := transferArticleIdentity(accessory.Manufacturer, accessory.ArticleNumber); key != "" {
			currentAccessoriesByArticle[key] = accessory
		}
	}
	seenAccessoryInventory := map[string]bool{}
	for index, accessory := range incoming.Accessories {
		rowNumber := transferRowNumber(rowNumbers[TransferAccessories], index)
		recordKey := strings.TrimSpace(accessory.InventoryNumber)
		if recordKey == "" {
			recordKey = fmt.Sprintf("row-%d", transferRowValue(rowNumber, index+1))
		}
		record := newDataTransferPreviewRecord(TransferAccessories, recordKey, rowNumber, accessory)
		start := len(issues)
		issues = appendRequiredTransferIssues(issues, jobID, TransferAccessories, recordKey, rowNumber,
			accessory.InventoryNumber, accessory.Manufacturer, accessory.Name)
		if err := rules.validateAccessory(accessory); err != nil {
			issues = append(issues, NewTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "record",
				TransferIssueError, "invalid_accessory", "Accessory data violates aggregate validation.", "skip"))
		}
		identity := transferIdentity(accessory.InventoryNumber)
		if identity != "" && seenAccessoryInventory[identity] {
			issues = append(issues, NewTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "inventoryNumber",
				TransferIssueError, "duplicate_input_inventory_number", "Inventory number occurs more than once in the upload.", "skip"))
		} else if target, found := currentAccessoriesByInventory[identity]; identity != "" && found {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			record.ProposedAction = "replace"
			issues = append(issues, NewTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "inventoryNumber",
				TransferIssueWarning, "duplicate_inventory_number", "Inventory number already exists.", "replace_or_copy"))
		} else if target, found := currentAccessoriesByArticle[transferArticleIdentity(accessory.Manufacturer, accessory.ArticleNumber)]; found {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			record.ProposedAction = "use_existing"
			issues = append(issues, NewTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "articleNumber",
				TransferIssueWarning, "matching_manufacturer_article_number",
				"Manufacturer and article number match an existing accessory.", "use_existing"))
		}
		if identity != "" {
			seenAccessoryInventory[identity] = true
		}
		FinalizeDataTransferRecord(&record, issues[start:])
		records = append(records, record)
	}

	currentListsByID := map[string]TransferExhibitionList{}
	currentListsByIdentity := map[string]TransferExhibitionList{}
	for _, list := range current.ExhibitionLists {
		if list.ID != "" {
			currentListsByID[list.ID] = list
		}
		currentListsByIdentity[transferExhibitionIdentity(list.Designation, list.Date)] = list
	}
	for index, list := range incoming.ExhibitionLists {
		rowNumber := transferRowNumber(rowNumbers[TransferExhibitionLists], index)
		recordKey := strings.TrimSpace(list.ID)
		if recordKey == "" {
			recordKey = strings.TrimSpace(list.Designation) + "|" + strings.TrimSpace(list.Date)
		}
		record := newDataTransferPreviewRecord(TransferExhibitionLists, recordKey, rowNumber, list)
		start := len(issues)
		if strings.TrimSpace(list.Designation) == "" {
			issues = append(issues, NewTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber,
				"designation", TransferIssueError, "missing_name", "Designation is required.", "skip"))
		}
		target, matched := currentListsByID[list.ID]
		if matched && transferExhibitionIdentity(target.Designation, target.Date) !=
			transferExhibitionIdentity(list.Designation, list.Date) {
			matched = false
		}
		if !matched {
			target, matched = currentListsByIdentity[transferExhibitionIdentity(list.Designation, list.Date)]
		}
		if matched {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			if target.Locked {
				record.ProposedAction = "copy"
				issues = append(issues, NewTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber, "id",
					TransferIssueError, "locked_exhibition_list", "Locked exhibition lists cannot be replaced.", "copy"))
			} else {
				record.ProposedAction = "replace"
				issues = append(issues, NewTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber, "id",
					TransferIssueWarning, "duplicate_exhibition_list",
					"Exhibition list already exists; choose replace or copy.", "replace_or_copy"))
			}
		}
		for entryIndex, entry := range list.Entries {
			if entry.VehicleID == "" && entry.VehicleInventoryNumber == "" {
				continue
			}
			if local, found := currentVehiclesByID[entry.VehicleID]; found &&
				strings.TrimSpace(entry.VehicleInventoryNumber) != "" && strings.EqualFold(
				strings.TrimSpace(entry.VehicleInventoryNumber), strings.TrimSpace(local.InventoryNumber)) {
				continue
			}
			if _, found := currentVehiclesByInventory[transferIdentity(entry.VehicleInventoryNumber)]; found {
				issues = append(issues, NewTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber,
					transferExhibitionEntryIssueField(entryIndex), TransferIssueWarning, "exhibition_vehicle_reference",
					"Vehicle reference can be linked by inventory number after confirmation.", "link"))
			} else {
				issues = append(issues, NewTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber,
					transferExhibitionEntryIssueField(entryIndex), TransferIssueError, "missing_vehicle_reference",
					"Referenced vehicle does not exist in this installation.", "skip"))
			}
		}
		FinalizeDataTransferRecord(&record, issues[start:])
		records = append(records, record)
	}
	return records, issues
}
