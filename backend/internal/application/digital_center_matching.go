package application

import (
	"sort"
	"strconv"
	"strings"
)

func compareDigitalCenterLocomotives(
	provider string,
	locomotives []ECoSRawLocomotive,
	vehicles []Vehicle,
) []DigitalCenterWorkItem {
	items := make([]DigitalCenterWorkItem, 0, len(locomotives))
	seenObjects := make(map[string]struct{}, len(locomotives))
	for _, locomotive := range locomotives {
		objectID := strconv.Itoa(locomotive.ObjectID)
		seenObjects[objectID] = struct{}{}
		item := compareDigitalCenterLocomotive(provider, locomotive, vehicles)
		if locomotive.CVSources != nil && len(locomotive.CVSources.Conflicts) > 0 && item.CompareStatus == DigitalCompareOK {
			item.CompareStatus = DigitalCompareDeviation
		}
		items = append(items, item)
	}
	missing := missingDigitalCenterLocomotives(provider, seenObjects, vehicles)
	return append(items, missing...)
}

func compareDigitalCenterLocomotive(
	provider string,
	locomotive ECoSRawLocomotive,
	vehicles []Vehicle,
) DigitalCenterWorkItem {
	provider = strings.ToLower(strings.TrimSpace(provider))
	protocol, _ := normalizeDigitalCenterProtocol(locomotive.Protocol)
	locomotive.Protocol = protocol
	locomotive.Name, _ = normalizeDigitalCenterName(locomotive.Name)
	item := newDigitalCenterWorkItem(locomotive)
	objectID := strconv.Itoa(locomotive.ObjectID)
	for _, vehicle := range vehicles {
		if mapping := exactDigitalCenterMapping(vehicle, provider, objectID); mapping != nil {
			item.VehicleID = vehicle.ID
			item.RailKeeper = digitalCenterVehiclePayload(vehicle, provider)
			item.CompareStatus = compareMappedDigitalCenterStatus(locomotive, vehicle, provider)
			return item
		}
	}

	addressCandidates := make([]Vehicle, 0)
	protocolCandidates := make([]Vehicle, 0)
	for _, vehicle := range vehicles {
		if digitalCenterVehicleAddress(vehicle, provider) != locomotive.Address {
			continue
		}
		addressCandidates = append(addressCandidates, vehicle)
		vehicleProtocol := digitalCenterVehicleProtocol(vehicle, provider)
		if protocol != "" && vehicleProtocol == protocol {
			protocolCandidates = append(protocolCandidates, vehicle)
		}
	}
	if len(protocolCandidates) == 1 {
		return proposedDigitalCenterWorkItem(item, protocolCandidates[0], provider, "address_protocol", DigitalCompareOK)
	}
	if len(protocolCandidates) > 1 {
		item.CompareStatus = DigitalCompareConflict
		item.Conflicts = digitalCenterCandidatePayloads(protocolCandidates, provider)
		return item
	}
	if len(addressCandidates) == 1 {
		return proposedDigitalCenterWorkItem(item, addressCandidates[0], provider, "address", DigitalCompareDeviation)
	}
	if len(addressCandidates) > 1 {
		item.CompareStatus = DigitalCompareConflict
		item.Conflicts = digitalCenterCandidatePayloads(addressCandidates, provider)
		return item
	}
	item.CompareStatus = DigitalCompareNew
	return item
}

func newDigitalCenterWorkItem(locomotive ECoSRawLocomotive) DigitalCenterWorkItem {
	stationStatus := "read"
	if strings.TrimSpace(locomotive.DetailError) != "" {
		stationStatus = "incomplete"
	}
	item := DigitalCenterWorkItem{
		CenterObjectID: strconv.Itoa(locomotive.ObjectID),
		Name:           locomotive.Name,
		Address:        locomotive.Address,
		Protocol:       locomotive.Protocol,
		StationStatus:  stationStatus,
		Center: map[string]any{
			"objectId": locomotive.ObjectID, "name": locomotive.Name,
			"decoderAddress": locomotive.Address, "protocol": locomotive.Protocol,
			"profile": locomotive.Profile, "cvs": locomotive.CVs,
		},
		RailKeeper: map[string]any{}, Proposed: map[string]any{}, Conflicts: []map[string]any{},
	}
	if locomotive.CVRead != nil {
		item.Center["cvRead"] = *locomotive.CVRead
	}
	if locomotive.CVSources != nil {
		item.Center["cvSources"] = *locomotive.CVSources
		if len(locomotive.CVSources.Conflicts) != 0 {
			item.Center["sourceProfile"] = locomotive.Profile
			item.Center["profile"] = ""
		}
	}
	if locomotive.FunctionSnapshot != nil {
		item.Center["functionSnapshot"] = *locomotive.FunctionSnapshot
	}
	return item
}

func proposedDigitalCenterWorkItem(
	item DigitalCenterWorkItem,
	vehicle Vehicle,
	provider string,
	match string,
	status DigitalCenterCompareStatus,
) DigitalCenterWorkItem {
	item.CompareStatus = status
	item.RailKeeper = digitalCenterVehiclePayload(vehicle, provider)
	item.Proposed = map[string]any{"vehicleId": vehicle.ID, "match": match}
	return item
}

func missingDigitalCenterLocomotives(
	provider string,
	seen map[string]struct{},
	vehicles []Vehicle,
) []DigitalCenterWorkItem {
	missing := []DigitalCenterWorkItem{}
	for _, vehicle := range vehicles {
		for _, mapping := range vehicle.ExternalMappings {
			if !strings.EqualFold(strings.TrimSpace(mapping.Provider), provider) {
				continue
			}
			externalID := strings.TrimSpace(mapping.ExternalID)
			if externalID == "" {
				continue
			}
			if _, found := seen[externalID]; found {
				continue
			}
			missing = append(missing, DigitalCenterWorkItem{
				CenterObjectID: externalID,
				VehicleID:      vehicle.ID,
				Name:           strings.TrimSpace(vehicle.Name),
				Address:        digitalCenterVehicleAddress(vehicle, provider),
				Protocol:       digitalCenterVehicleProtocol(vehicle, provider),
				CompareStatus:  DigitalCompareMissing,
				StationStatus:  "missing",
				Center:         map[string]any{},
				RailKeeper:     digitalCenterVehiclePayload(vehicle, provider),
				Proposed:       map[string]any{},
				Conflicts:      []map[string]any{},
			})
		}
	}
	sort.SliceStable(missing, func(left, right int) bool {
		if missing[left].Name == missing[right].Name {
			return missing[left].CenterObjectID < missing[right].CenterObjectID
		}
		return strings.ToLower(missing[left].Name) < strings.ToLower(missing[right].Name)
	})
	return missing
}

func exactDigitalCenterMapping(vehicle Vehicle, provider, objectID string) *VehicleExternalMap {
	for index := range vehicle.ExternalMappings {
		mapping := &vehicle.ExternalMappings[index]
		if strings.EqualFold(strings.TrimSpace(mapping.Provider), provider) &&
			strings.TrimSpace(mapping.ExternalID) == objectID {
			return mapping
		}
	}
	return nil
}

func compareMappedDigitalCenterStatus(
	locomotive ECoSRawLocomotive,
	vehicle Vehicle,
	provider string,
) DigitalCenterCompareStatus {
	if locomotive.Name != "" && strings.TrimSpace(vehicle.Name) != "" && strings.TrimSpace(vehicle.Name) != locomotive.Name {
		return DigitalCompareDeviation
	}
	if address := digitalCenterVehicleAddress(vehicle, provider); address > 0 && address != locomotive.Address {
		return DigitalCompareDeviation
	}
	if protocol := digitalCenterVehicleProtocol(vehicle, provider); locomotive.Protocol != "" && protocol != "" && protocol != locomotive.Protocol {
		return DigitalCompareDeviation
	}
	return DigitalCompareOK
}

func knownDigitalCenterAddresses(vehicles []Vehicle, provider string) []int {
	addresses := make([]int, 0, len(vehicles))
	for _, vehicle := range vehicles {
		if address := digitalCenterVehicleAddress(vehicle, provider); address > 0 {
			addresses = append(addresses, address)
		}
	}
	return addresses
}

func digitalCenterVehicleAddress(vehicle Vehicle, provider string) int {
	if address, err := strconv.Atoi(strings.TrimSpace(vehicle.DigitalDecoderNumber)); err == nil && address > 0 {
		return address
	}
	for _, mapping := range vehicle.ExternalMappings {
		if strings.EqualFold(strings.TrimSpace(mapping.Provider), provider) {
			if address, err := strconv.Atoi(strings.TrimSpace(mapping.ExternalAddress)); err == nil && address > 0 {
				return address
			}
		}
	}
	return 0
}

func digitalCenterVehicleProtocol(vehicle Vehicle, provider string) string {
	for _, mapping := range vehicle.ExternalMappings {
		if strings.EqualFold(strings.TrimSpace(mapping.Provider), provider) {
			protocol, err := normalizeDigitalCenterProtocol(mapping.ExternalProtocol)
			if err == nil && protocol != "" {
				return protocol
			}
		}
	}
	for _, value := range vehicle.CVValues {
		protocol, err := normalizeDigitalCenterProtocol(value.Protocol)
		if err == nil && protocol != "" {
			return protocol
		}
	}
	return ""
}

func digitalCenterVehiclePayload(vehicle Vehicle, provider string) map[string]any {
	return map[string]any{
		"vehicleId": vehicle.ID, "name": strings.TrimSpace(vehicle.Name),
		"decoderType":    vehicle.DecoderType,
		"decoderAddress": digitalCenterVehicleAddress(vehicle, provider),
		"protocol":       digitalCenterVehicleProtocol(vehicle, provider),
	}
}

func digitalCenterCandidatePayloads(vehicles []Vehicle, provider string) []map[string]any {
	candidates := make([]map[string]any, 0, len(vehicles))
	sort.SliceStable(vehicles, func(left, right int) bool { return vehicles[left].ID < vehicles[right].ID })
	for _, vehicle := range vehicles {
		candidates = append(candidates, digitalCenterVehiclePayload(vehicle, provider))
	}
	return candidates
}
