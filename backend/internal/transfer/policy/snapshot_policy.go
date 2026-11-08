package policy

import (
	"slices"

	"strings"
)

func sortDataTransferSnapshot(snapshot *DataTransferSnapshot) {
	slices.SortStableFunc(snapshot.Vehicles, func(left, right TransferVehicle) int {
		return compareTransferKeys(left.InventoryNumber, right.InventoryNumber, left.ID, right.ID)
	})
	slices.SortStableFunc(snapshot.VehicleSets, func(left, right TransferVehicleSet) int {
		return compareTransferKeys(left.InventoryNumber, right.InventoryNumber, left.ID, right.ID)
	})
	for index := range snapshot.VehicleSets {
		slices.SortStableFunc(snapshot.VehicleSets[index].Members, func(left, right TransferVehicleSetMember) int {
			if left.Position != right.Position {
				return left.Position - right.Position
			}
			return compareTransferKeys(
				left.VehicleInventoryNumber, right.VehicleInventoryNumber,
				left.SourceVehicleID, right.SourceVehicleID,
			)
		})
	}
	slices.SortStableFunc(snapshot.Accessories, func(left, right TransferAccessory) int {
		return compareTransferKeys(left.InventoryNumber, right.InventoryNumber, left.ID, right.ID)
	})
	for index := range snapshot.Accessories {
		slices.SortStableFunc(snapshot.Accessories[index].Stock, func(left, right TransferAccessoryStock) int {
			return compareTransferKeys(left.LocationName, right.LocationName, left.LocationID, right.LocationID)
		})
		slices.SortStableFunc(snapshot.Accessories[index].Assets, func(left, right TransferAccessoryAsset) int {
			return compareTransferKeys(left.InventoryNumber, right.InventoryNumber, left.ID, right.ID)
		})
	}
	slices.SortStableFunc(snapshot.ExhibitionLists, func(left, right TransferExhibitionList) int {
		return compareTransferKeys(left.Date, right.Date, left.Designation, right.Designation, left.ID, right.ID)
	})
	for index := range snapshot.ExhibitionLists {
		slices.SortStableFunc(snapshot.ExhibitionLists[index].Entries, func(left, right TransferExhibitionEntry) int {
			if left.SortOrder != right.SortOrder {
				return left.SortOrder - right.SortOrder
			}
			return compareTransferKeys(left.LocomotiveName, right.LocomotiveName, left.ID, right.ID)
		})
	}
}

func prepareDataTransferSnapshotForExport(snapshot *DataTransferSnapshot) {
	if snapshot.VehicleSets != nil {
		normalizedSets := make([]TransferVehicleSet, 0, len(snapshot.VehicleSets))
		for _, set := range snapshot.VehicleSets {
			if len(set.Members) < 2 {
				continue
			}
			set.Members = slices.Clone(set.Members)
			slices.SortStableFunc(set.Members, func(left, right TransferVehicleSetMember) int {
				if left.Position != right.Position {
					return left.Position - right.Position
				}
				return compareTransferKeys(
					left.VehicleInventoryNumber, right.VehicleInventoryNumber,
					left.SourceVehicleID, right.SourceVehicleID,
				)
			})
			for index := range set.Members {
				set.Members[index].Position = index + 1
			}
			normalizedSets = append(normalizedSets, set)
		}
		snapshot.VehicleSets = normalizedSets
	}
	sortDataTransferSnapshot(snapshot)
}

func compareTransferKeys(values ...string) int {
	for index := 0; index+1 < len(values); index += 2 {
		left := strings.ToLower(values[index])
		right := strings.ToLower(values[index+1])
		if comparison := strings.Compare(left, right); comparison != 0 {
			return comparison
		}
	}
	return 0
}
