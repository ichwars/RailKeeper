package exhibition

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

func buildExhibitionConflicts(list ExhibitionList, entries []ExhibitionEntry) []ExhibitionConflict {
	conflicts := make([]ExhibitionConflict, 0)
	dayScopes := exhibitionDayScopes(list)
	vehicleEntries := map[string][]string{}
	for _, entry := range entries {
		missing := exhibitionMissingFields(entry)
		if len(missing) > 0 {
			conflicts = append(conflicts, newExhibitionConflict(
				ExhibitionConflictMissing, []string{entry.ID}, missing, "", "", nil,
			))
		}
		if entry.VehicleID != "" {
			vehicleEntries[entry.VehicleID] = append(vehicleEntries[entry.VehicleID], entry.ID)
		}
	}
	for vehicleID, entryIDs := range vehicleEntries {
		if len(entryIDs) > 1 {
			sort.Strings(entryIDs)
			conflicts = append(conflicts, newExhibitionConflict(
				ExhibitionConflictDuplicateVehicle, entryIDs, []string{"vehicleId"}, "", vehicleID, nil,
			))
		}
	}
	for leftIndex, left := range entries {
		leftAddress := exhibitionAddress(left)
		leftInterface := strings.TrimSpace(left.InterfaceName)
		if leftAddress == "" || leftInterface == "" || left.Analog || left.Availability == "unavailable" {
			continue
		}
		for rightIndex := leftIndex + 1; rightIndex < len(entries); rightIndex++ {
			right := entries[rightIndex]
			if right.Analog || right.Availability == "unavailable" ||
				!strings.EqualFold(leftAddress, exhibitionAddress(right)) ||
				!strings.EqualFold(leftInterface, strings.TrimSpace(right.InterfaceName)) {
				continue
			}
			overlap := overlappingExhibitionDays(left.DayScope, right.DayScope, dayScopes)
			if len(overlap) == 0 {
				continue
			}
			entryIDs := []string{left.ID, right.ID}
			sort.Strings(entryIDs)
			conflicts = append(conflicts, newExhibitionConflict(
				ExhibitionConflictAddress, entryIDs, []string{"decoderNumber"}, leftInterface, leftAddress, overlap,
			))
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].ID < conflicts[j].ID })
	return conflicts
}
func newExhibitionConflict(kind string, entryIDs, fields []string, interfaceName, address string, days []string) ExhibitionConflict {
	key := strings.Join([]string{
		kind, strings.Join(entryIDs, ","), strings.Join(fields, ","),
		strings.ToLower(interfaceName), strings.ToLower(address), strings.Join(days, ","),
	}, "|")
	sum := sha256.Sum256([]byte(key))
	return ExhibitionConflict{
		ID: fmt.Sprintf("conflict-%x", sum[:12]), Kind: kind, EntryIDs: entryIDs,
		Fields: fields, InterfaceName: interfaceName, Address: address, DayScopes: days,
	}
}
func exhibitionMissingFields(entry ExhibitionEntry) []string {
	missing := make([]string, 0, 4)
	if strings.TrimSpace(entry.Owner) == "" {
		missing = append(missing, "owner")
	}
	if strings.TrimSpace(entry.DayScope) == "" {
		missing = append(missing, "dayScope")
	}
	if !entry.Analog {
		if exhibitionAddress(entry) == "" {
			missing = append(missing, "decoderNumber")
		}
		if strings.TrimSpace(entry.InterfaceName) == "" {
			missing = append(missing, "interfaceName")
		}
	}
	return missing
}
func exhibitionAddressReady(entry ExhibitionEntry) bool {
	return entry.Analog || (exhibitionAddress(entry) != "" && strings.TrimSpace(entry.InterfaceName) != "")
}
func exhibitionAddress(entry ExhibitionEntry) string {
	if value := strings.TrimSpace(entry.DecoderNumber); value != "" {
		return value
	}
	return strings.TrimSpace(entry.SXAddress)
}
