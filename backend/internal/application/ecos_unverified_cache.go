package application

import "fmt"

// Cached bytes have no proven per-register validity markers. They can expose a
// contradiction, but cannot supply new, confirmed decoder readings on their own.
func inspectUnverifiedCVCache(values, export, candidates []ECoSCVValue, status string, info *ECoSCVSourceInfo) ([]ECoSCVValue, error) {
	info.StoredCacheStatus = status
	if len(candidates) == 0 {
		return values, nil
	}
	filtered := make([]ECoSCVValue, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Number < 1 || candidate.Number > 255 || candidate.IndexCV31 != nil || candidate.IndexCV32 != nil {
			return nil, ErrDigitalCenterDeviceOutput
		}
		if candidate.Value != 0 {
			filtered = append(filtered, candidate)
		}
	}
	checked, err := validateStoredECoSCVs(filtered)
	if err != nil {
		return nil, err
	}
	if len(checked) > 0 {
		found := false
		for _, source := range info.Sources {
			found = found || source == "backup"
		}
		if !found {
			info.Sources = append(info.Sources, "backup")
		}
	}
	baseline := map[int]ECoSCVValue{}
	for _, value := range export {
		if value.IndexCV31 == nil && value.IndexCV32 == nil {
			baseline[value.Number] = value
		}
	}
	importable := map[string]ECoSCVValue{}
	for _, value := range values {
		importable[storedCVAddress(value)] = value
	}
	for _, candidate := range checked {
		value, found := baseline[candidate.Number]
		if !found {
			info.UnverifiedValues = append(info.UnverifiedValues, candidate)
			continue
		}
		if value.Value == candidate.Value {
			continue
		}
		delete(importable, storedCVAddress(candidate))
		index := -1
		for position, conflict := range info.Conflicts {
			if conflict.Number == candidate.Number && conflict.IndexCV31 == nil && conflict.IndexCV32 == nil {
				index = position
				break
			}
		}
		if index >= 0 {
			raw := candidate.Value
			info.Conflicts[index].RawCacheValue = &raw
		} else {
			info.Conflicts = append(info.Conflicts, ECoSCVSourceConflict{Number: candidate.Number,
				DatabaseValue: value.Value, BackupValue: candidate.Value, BackupSource: "stored_cache"})
		}
	}
	result := make([]ECoSCVValue, 0, len(importable))
	for _, value := range importable {
		result = append(result, value)
	}
	return validateStoredECoSCVs(result)
}

func storedCVAddress(value ECoSCVValue) string {
	return fmt.Sprintf("%d:%d", value.Number, eCoSCVPage(value))
}

func storedCVReference(database, configuration []ECoSCVValue) []ECoSCVValue {
	values := make(map[string]ECoSCVValue)
	for _, value := range database {
		values[storedCVAddress(value)] = value
	}
	for _, value := range configuration {
		if _, exists := values[storedCVAddress(value)]; !exists {
			values[storedCVAddress(value)] = value
		}
	}
	result := make([]ECoSCVValue, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func countStoredBackupAdditions(values, database []ECoSCVValue) int {
	known := make(map[string]bool)
	for _, value := range database {
		known[storedCVAddress(value)] = true
	}
	added := 0
	for _, value := range values {
		if !known[storedCVAddress(value)] {
			added++
		}
	}
	return added
}
