package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxDigitalCenterLocomotives = 5000
	maxDigitalCenterObjectID    = 2_000_000_000
	maxDigitalCenterAddress     = 10239
	maxDigitalCenterNameRunes   = 240
	maxDigitalCenterProtocol    = 32
)

var (
	ErrDigitalCenterCapabilityUnavailable = errors.New("digital center capability unavailable")
	ErrDigitalCenterNotConfigured         = errors.New("digital center is not configured")
	ErrDigitalCenterInactive              = errors.New("digital center is inactive")
	ErrDigitalCenterWorkspaceUnavailable  = errors.New("digital center workspace is unavailable")
	ErrDigitalCenterDeviceOutput          = errors.New("digital center device output is invalid")
	ErrDigitalCenterFilterValidation      = errors.New("digital center filter is invalid")
)

type DigitalCenterWorkItemFilter struct {
	Query         string
	CompareStatus DigitalCenterCompareStatus
	Page          int
	PageSize      int
}

type DigitalCenterWorkItemPage struct {
	Items      []DigitalCenterWorkItem `json:"items"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	Total      int                     `json:"total"`
	TotalPages int                     `json:"totalPages"`
}

func (service *DigitalCenterWorkspaceService) startReadSessionUnlocked(
	ctx context.Context,
	provider string,
	actor string,
) (DigitalCenterReadSession, error) {
	if service == nil || service.repository == nil || service.ecos == nil || service.vehicles == nil {
		return DigitalCenterReadSession{}, ErrDigitalCenterWorkspaceUnavailable
	}
	center, err := service.configuredCenter(ctx, provider)
	if err != nil {
		return DigitalCenterReadSession{}, err
	}
	if !center.Active {
		return DigitalCenterReadSession{}, ErrDigitalCenterInactive
	}
	if !center.Capabilities.ReadLocomotives {
		return DigitalCenterReadSession{}, ErrDigitalCenterCapabilityUnavailable
	}
	now := digitalCenterTimestamp()
	session, err := service.repository.CreateSession(ctx, DigitalCenterReadSession{
		Provider:        center.Provider,
		State:           DigitalCenterSessionReading,
		Host:            center.Host,
		Port:            center.Port,
		Capabilities:    center.Capabilities,
		ReadStartedAt:   now,
		CreatedByUserID: strings.TrimSpace(actor),
	})
	if err != nil {
		return DigitalCenterReadSession{}, fmt.Errorf("start digital center read session: %w", err)
	}

	probe, err := service.ecos.ProbeLocomotiveRaw(ctx, ECoSConnectionInput{Host: center.Host, Port: center.Port})
	if err != nil {
		return service.finishFailedReadSession(ctx, session, err)
	}
	locomotives, err := normalizeDigitalCenterLocomotives(probe)
	if err != nil {
		return service.finishFailedReadSession(ctx, session, err)
	}
	vehicles, err := service.vehicles.ListReadOnly(ctx, "")
	if err != nil {
		return service.finishFailedReadSession(ctx, session, fmt.Errorf("list RailKeeper vehicles: %w", err))
	}
	items := compareDigitalCenterLocomotives(center.Provider, locomotives, vehicles)
	if err := service.repository.ReplaceWorkItems(ctx, session.ID, items); err != nil {
		return service.finishFailedReadSession(ctx, session, fmt.Errorf("persist comparison worklist: %w", err))
	}
	session.State = DigitalCenterSessionReady
	session.ReadCompletedAt = digitalCenterTimestamp()
	if err := service.repository.UpdateSession(ctx, session); err != nil {
		return DigitalCenterReadSession{}, fmt.Errorf("complete digital center read session: %w", err)
	}
	if err := service.repository.AddMessage(ctx, DigitalCenterSessionMessage{
		SessionID:  session.ID,
		Severity:   DigitalCenterMessageInfo,
		Code:       DigitalCenterMessageReadCompleted,
		Message:    fmt.Sprintf("%d Lokomotiven wurden in die Arbeitsliste gelesen.", len(locomotives)),
		NextAction: "Arbeitsliste prüfen und Konflikte klären.",
	}); err != nil {
		return DigitalCenterReadSession{}, fmt.Errorf("record completed digital center read: %w", err)
	}
	return session, nil
}

func (service *DigitalCenterWorkspaceService) configuredCenter(
	ctx context.Context,
	provider string,
) (DigitalCenterSummary, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	centers, err := service.ListConfiguredCenters(ctx)
	if err != nil {
		return DigitalCenterSummary{}, err
	}
	for _, center := range centers {
		if center.Provider == provider {
			return center, nil
		}
	}
	return DigitalCenterSummary{}, ErrDigitalCenterNotConfigured
}

func (service *DigitalCenterWorkspaceService) finishFailedReadSession(
	ctx context.Context,
	session DigitalCenterReadSession,
	cause error,
) (DigitalCenterReadSession, error) {
	message := DigitalCenterSessionMessage{
		SessionID:  session.ID,
		Severity:   DigitalCenterMessageError,
		Code:       DigitalCenterMessageReadFailed,
		Message:    "Die Digitalzentrale konnte nicht vollständig gelesen werden.",
		NextAction: "Verbindung und Gerätekonfiguration prüfen und den Lesevorgang erneut starten.",
	}
	session.State = DigitalCenterSessionFailed
	if errors.Is(cause, context.Canceled) || errors.Is(cause, context.DeadlineExceeded) {
		session.State = DigitalCenterSessionInterrupted
		message.Severity = DigitalCenterMessageWarning
		message.Code = DigitalCenterMessageConnectionInterrupted
		message.Message = "Der Lesevorgang wurde unterbrochen."
		message.NextAction = "Verbindung prüfen und den Lesevorgang erneut starten."
	}
	session.ReadCompletedAt = digitalCenterTimestamp()
	persistContext := context.WithoutCancel(ctx)
	if err := service.repository.UpdateSession(persistContext, session); err != nil {
		return DigitalCenterReadSession{}, fmt.Errorf("persist failed digital center read session: %w", err)
	}
	if err := service.repository.AddMessage(persistContext, message); err != nil {
		return DigitalCenterReadSession{}, fmt.Errorf("record failed digital center read: %w", err)
	}
	return session, nil
}

func (service *DigitalCenterWorkspaceService) GetReadSession(
	ctx context.Context,
	id string,
) (DigitalCenterReadSession, error) {
	if service == nil || service.repository == nil {
		return DigitalCenterReadSession{}, ErrDigitalCenterWorkspaceUnavailable
	}
	return service.repository.GetSession(ctx, strings.TrimSpace(id))
}

func (service *DigitalCenterWorkspaceService) ListWorkItems(
	ctx context.Context,
	sessionID string,
	filter DigitalCenterWorkItemFilter,
) (DigitalCenterWorkItemPage, error) {
	if service == nil || service.repository == nil {
		return DigitalCenterWorkItemPage{}, ErrDigitalCenterWorkspaceUnavailable
	}
	if !validDigitalCenterCompareStatus(filter.CompareStatus) {
		return DigitalCenterWorkItemPage{}, ErrDigitalCenterFilterValidation
	}
	if _, err := service.repository.GetSession(ctx, strings.TrimSpace(sessionID)); err != nil {
		return DigitalCenterWorkItemPage{}, err
	}
	items, err := service.repository.ListWorkItems(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return DigitalCenterWorkItemPage{}, err
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	filtered := make([]DigitalCenterWorkItem, 0, len(items))
	for _, item := range items {
		if filter.CompareStatus != "" && item.CompareStatus != filter.CompareStatus {
			continue
		}
		if query != "" && !digitalCenterWorkItemContains(item, query) {
			continue
		}
		filtered = append(filtered, item)
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	totalPages := 0
	if len(filtered) > 0 {
		totalPages = int(math.Ceil(float64(len(filtered)) / float64(pageSize)))
	}
	start := len(filtered)
	if page <= totalPages {
		start = (page - 1) * pageSize
	}
	end := min(start+pageSize, len(filtered))
	return DigitalCenterWorkItemPage{
		Items:      append([]DigitalCenterWorkItem(nil), filtered[start:end]...),
		Page:       page,
		PageSize:   pageSize,
		Total:      len(filtered),
		TotalPages: totalPages,
	}, nil
}

func (service *DigitalCenterWorkspaceService) GetWorkItem(
	ctx context.Context,
	sessionID string,
	itemID string,
) (DigitalCenterWorkItem, error) {
	if service == nil || service.repository == nil {
		return DigitalCenterWorkItem{}, ErrDigitalCenterWorkspaceUnavailable
	}
	return service.repository.GetWorkItem(ctx, strings.TrimSpace(sessionID), strings.TrimSpace(itemID))
}

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
		items = append(items, compareDigitalCenterLocomotive(provider, locomotive, vehicles))
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
	return DigitalCenterWorkItem{
		CenterObjectID: strconv.Itoa(locomotive.ObjectID),
		Name:           locomotive.Name,
		Address:        locomotive.Address,
		Protocol:       locomotive.Protocol,
		StationStatus:  stationStatus,
		Center: map[string]any{
			"objectId": locomotive.ObjectID, "name": locomotive.Name,
			"decoderAddress": locomotive.Address, "protocol": locomotive.Protocol,
		},
		RailKeeper: map[string]any{}, Proposed: map[string]any{}, Conflicts: []map[string]any{},
	}
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
	if strings.TrimSpace(vehicle.Name) != "" && strings.TrimSpace(vehicle.Name) != locomotive.Name {
		return DigitalCompareDeviation
	}
	if address := digitalCenterVehicleAddress(vehicle, provider); address > 0 && address != locomotive.Address {
		return DigitalCompareDeviation
	}
	if protocol := digitalCenterVehicleProtocol(vehicle, provider); protocol != "" && protocol != locomotive.Protocol {
		return DigitalCompareDeviation
	}
	return DigitalCompareOK
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

func normalizeDigitalCenterLocomotives(probe *ECoSRawProbe) ([]ECoSRawLocomotive, error) {
	if probe == nil {
		return nil, fmt.Errorf("%w: empty ECoS response", ErrDigitalCenterDeviceOutput)
	}
	if len(probe.Locomotives) > maxDigitalCenterLocomotives {
		return nil, fmt.Errorf("%w: too many locomotives", ErrDigitalCenterDeviceOutput)
	}
	normalized := make([]ECoSRawLocomotive, 0, len(probe.Locomotives))
	seen := make(map[int]struct{}, len(probe.Locomotives))
	for _, locomotive := range probe.Locomotives {
		if locomotive.ObjectID < 1 || locomotive.ObjectID > maxDigitalCenterObjectID {
			return nil, fmt.Errorf("%w: invalid object ID", ErrDigitalCenterDeviceOutput)
		}
		if _, found := seen[locomotive.ObjectID]; found {
			return nil, fmt.Errorf("%w: duplicate object ID", ErrDigitalCenterDeviceOutput)
		}
		seen[locomotive.ObjectID] = struct{}{}
		if locomotive.Address < 1 || locomotive.Address > maxDigitalCenterAddress {
			return nil, fmt.Errorf("%w: invalid decoder address", ErrDigitalCenterDeviceOutput)
		}
		name, err := normalizeDigitalCenterName(locomotive.Name)
		if err != nil {
			return nil, err
		}
		protocol, err := normalizeDigitalCenterProtocol(locomotive.Protocol)
		if err != nil {
			return nil, err
		}
		normalized = append(normalized, ECoSRawLocomotive{
			ObjectID: locomotive.ObjectID, Name: name, Address: locomotive.Address,
			Protocol: protocol, DetailError: strings.TrimSpace(locomotive.DetailError),
		})
	}
	return normalized, nil
}

func normalizeDigitalCenterName(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", fmt.Errorf("%w: locomotive name is not valid UTF-8", ErrDigitalCenterDeviceOutput)
	}
	value = strings.Map(func(character rune) rune {
		if unicode.IsControl(character) || unicode.In(character, unicode.Cf, unicode.Co, unicode.Cs) {
			return ' '
		}
		return character
	}, value)
	value = strings.Join(strings.Fields(value), " ")
	if utf8.RuneCountInString(value) > maxDigitalCenterNameRunes {
		return "", fmt.Errorf("%w: locomotive name is too long", ErrDigitalCenterDeviceOutput)
	}
	return value, nil
}

func normalizeDigitalCenterProtocol(value string) (string, error) {
	compact := strings.Map(func(character rune) rune {
		if unicode.IsSpace(character) || strings.ContainsRune("-_/.+", character) {
			return -1
		}
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return 0
		}
		return unicode.ToUpper(character)
	}, strings.TrimSpace(value))
	if compact == "" && strings.TrimSpace(value) != "" {
		return "", fmt.Errorf("%w: invalid protocol", ErrDigitalCenterDeviceOutput)
	}
	if len(compact) > maxDigitalCenterProtocol {
		return "", fmt.Errorf("%w: protocol is too long", ErrDigitalCenterDeviceOutput)
	}
	for _, character := range compact {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return "", fmt.Errorf("%w: invalid protocol", ErrDigitalCenterDeviceOutput)
		}
	}
	switch {
	case strings.HasPrefix(compact, "DCC"):
		return "DCC", nil
	case strings.HasPrefix(compact, "MOTOROLA"), strings.HasPrefix(compact, "MM"):
		return "MOTOROLA", nil
	case strings.HasPrefix(compact, "MFX"), strings.HasPrefix(compact, "M4"):
		return "MFX", nil
	case strings.HasPrefix(compact, "SELECTRIX"), strings.HasPrefix(compact, "SX"):
		return "SELECTRIX", nil
	default:
		return compact, nil
	}
}

func validDigitalCenterCompareStatus(status DigitalCenterCompareStatus) bool {
	switch status {
	case "", DigitalCompareOK, DigitalCompareDeviation, DigitalCompareMissing, DigitalCompareNew,
		DigitalCompareConflict:
		return true
	default:
		return false
	}
}

func digitalCenterWorkItemContains(item DigitalCenterWorkItem, query string) bool {
	fields := []string{
		item.Name, item.Protocol, item.VehicleID, item.CenterObjectID, strconv.Itoa(item.Address),
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func digitalCenterTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
