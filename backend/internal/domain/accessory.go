package domain

import (
	"errors"
	"strings"
)

var ErrAllocationTarget = errors.New("exactly one allocation target is required")

type AccessoryTrackingMode string

const (
	AccessoryTrackingModeQuantity   AccessoryTrackingMode = "quantity"
	AccessoryTrackingModeIndividual AccessoryTrackingMode = "individual"
)

func (mode AccessoryTrackingMode) Valid() bool {
	return mode == AccessoryTrackingModeQuantity || mode == AccessoryTrackingModeIndividual
}

type AccessoryInventoryStrategy string

const (
	AccessoryInventoryQuantity                AccessoryInventoryStrategy = "quantity"
	AccessoryInventoryIndividual              AccessoryInventoryStrategy = "individual"
	AccessoryInventoryQuantityLaterIndividual AccessoryInventoryStrategy = "quantity_later_individual"
)

func (strategy AccessoryInventoryStrategy) Valid() bool {
	return strategy == AccessoryInventoryQuantity || strategy == AccessoryInventoryIndividual ||
		strategy == AccessoryInventoryQuantityLaterIndividual
}

func (strategy AccessoryInventoryStrategy) TrackingMode() AccessoryTrackingMode {
	if strategy == AccessoryInventoryIndividual {
		return AccessoryTrackingModeIndividual
	}
	return AccessoryTrackingModeQuantity
}

func InventoryStrategyFromTrackingMode(mode AccessoryTrackingMode) AccessoryInventoryStrategy {
	if mode == AccessoryTrackingModeIndividual {
		return AccessoryInventoryIndividual
	}
	return AccessoryInventoryQuantity
}

type AccessoryArticleType string

const (
	AccessoryArticleTrack               AccessoryArticleType = "track"
	AccessoryArticleSignal              AccessoryArticleType = "signal"
	AccessoryArticleDecoder             AccessoryArticleType = "decoder"
	AccessoryArticleElectricalControl   AccessoryArticleType = "electrical_control"
	AccessoryArticleBuildingEquipment   AccessoryArticleType = "building_equipment"
	AccessoryArticleLandscapeConsumable AccessoryArticleType = "landscape_consumable"
	AccessoryArticleLighting            AccessoryArticleType = "lighting"
	AccessoryArticleOther               AccessoryArticleType = "other"
)

func (articleType AccessoryArticleType) Valid() bool {
	switch articleType {
	case AccessoryArticleTrack, AccessoryArticleSignal, AccessoryArticleDecoder,
		AccessoryArticleElectricalControl, AccessoryArticleBuildingEquipment,
		AccessoryArticleLandscapeConsumable, AccessoryArticleLighting, AccessoryArticleOther:
		return true
	default:
		return false
	}
}

type AccessoryCondition string

const (
	AccessoryConditionReady          AccessoryCondition = "ready"
	AccessoryConditionMaintenanceDue AccessoryCondition = "maintenance_due"
	AccessoryConditionDefective      AccessoryCondition = "defective"
	AccessoryConditionUnknown        AccessoryCondition = "unknown"
)

func (condition AccessoryCondition) Valid() bool {
	switch condition {
	case AccessoryConditionReady, AccessoryConditionMaintenanceDue, AccessoryConditionDefective,
		AccessoryConditionUnknown:
		return true
	default:
		return false
	}
}

type AccessoryLifecycle string

const (
	AccessoryLifecycleStored      AccessoryLifecycle = "stored"
	AccessoryLifecycleReserved    AccessoryLifecycle = "reserved"
	AccessoryLifecycleInstalled   AccessoryLifecycle = "installed"
	AccessoryLifecycleMaintenance AccessoryLifecycle = "maintenance"
	AccessoryLifecycleRetired     AccessoryLifecycle = "retired"
)

func (lifecycle AccessoryLifecycle) Valid() bool {
	switch lifecycle {
	case AccessoryLifecycleStored, AccessoryLifecycleReserved, AccessoryLifecycleInstalled,
		AccessoryLifecycleMaintenance, AccessoryLifecycleRetired:
		return true
	default:
		return false
	}
}

type AccessoryReservationStatus string

const (
	AccessoryReservationActive    AccessoryReservationStatus = "active"
	AccessoryReservationFulfilled AccessoryReservationStatus = "fulfilled"
	AccessoryReservationCancelled AccessoryReservationStatus = "cancelled"
)

func (status AccessoryReservationStatus) Valid() bool {
	switch status {
	case AccessoryReservationActive, AccessoryReservationFulfilled, AccessoryReservationCancelled:
		return true
	default:
		return false
	}
}

type AccessoryRemovalDisposition string

const (
	AccessoryRemovalStored      AccessoryRemovalDisposition = "stored"
	AccessoryRemovalMaintenance AccessoryRemovalDisposition = "maintenance"
	AccessoryRemovalDefective   AccessoryRemovalDisposition = "defective"
	AccessoryRemovalRetired     AccessoryRemovalDisposition = "retired"
)

func (disposition AccessoryRemovalDisposition) Valid() bool {
	switch disposition {
	case AccessoryRemovalStored, AccessoryRemovalMaintenance, AccessoryRemovalDefective,
		AccessoryRemovalRetired:
		return true
	default:
		return false
	}
}

type AllocationTarget struct {
	VehicleID    string
	LayoutID     string
	LayoutUnitID string
}

func (target AllocationTarget) Validate() error {
	count := 0
	for _, value := range []string{target.VehicleID, target.LayoutID, target.LayoutUnitID} {
		if strings.TrimSpace(value) != "" {
			count++
		}
	}
	if count != 1 {
		return ErrAllocationTarget
	}
	return nil
}
