package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (service *DigitalCenterWorkspaceService) digitalCenterWriteTarget(
	ctx context.Context,
	sessionID string,
	itemID string,
	requestedFields []string,
	requestedOperation DigitalCenterWriteOperation,
) (digitalCenterWriteTarget, error) {
	if service == nil || service.repository == nil || service.ecos == nil || service.vehicles == nil {
		return digitalCenterWriteTarget{}, ErrDigitalCenterWorkspaceUnavailable
	}
	sessionID = strings.TrimSpace(sessionID)
	itemID = strings.TrimSpace(itemID)
	session, err := service.repository.GetSession(ctx, sessionID)
	if err != nil {
		return digitalCenterWriteTarget{}, err
	}
	if session.State != DigitalCenterSessionReady || strings.TrimSpace(session.ReadCompletedAt) == "" {
		return digitalCenterWriteTarget{}, ErrDigitalCenterReadNotFresh
	}
	completedAt, err := time.Parse(time.RFC3339, session.ReadCompletedAt)
	now := service.digitalCenterWriteNow()
	if err != nil || now.Sub(completedAt) > digitalCenterReadFreshness || completedAt.After(now) {
		return digitalCenterWriteTarget{}, ErrDigitalCenterReadNotFresh
	}
	center, err := service.configuredCenter(ctx, session.Provider)
	if err != nil {
		return digitalCenterWriteTarget{}, err
	}
	if !center.Active || !session.Capabilities.WriteLocomotives || !center.Capabilities.WriteLocomotives {
		return digitalCenterWriteTarget{}, ErrDigitalCenterCapabilityUnavailable
	}
	if center.Provider != session.Provider || center.Host != session.Host || center.Port != session.Port {
		return digitalCenterWriteTarget{}, ErrDigitalCenterGrantMismatch
	}
	item, err := service.repository.GetWorkItem(ctx, sessionID, itemID)
	if err != nil {
		return digitalCenterWriteTarget{}, err
	}
	if item.SessionID != session.ID || item.ID != itemID {
		return digitalCenterWriteTarget{}, ErrDigitalCenterGrantMismatch
	}
	operation, err := normalizeDigitalCenterWriteOperation(requestedOperation)
	if err != nil {
		return digitalCenterWriteTarget{}, err
	}
	if len(item.Conflicts) > 0 || strings.TrimSpace(item.VehicleID) == "" {
		return digitalCenterWriteTarget{}, ErrDigitalCenterConflictUnresolved
	}
	objectID := 0
	if operation == DigitalCenterWriteCreate {
		if item.CompareStatus != DigitalCompareMissing || item.StationStatus != "missing" {
			return digitalCenterWriteTarget{}, ErrDigitalCenterConflictUnresolved
		}
	} else {
		if item.CompareStatus == DigitalCompareConflict || item.CompareStatus == DigitalCompareMissing ||
			item.CompareStatus == DigitalCompareNew || item.StationStatus != "read" {
			return digitalCenterWriteTarget{}, ErrDigitalCenterConflictUnresolved
		}
		objectID, err = strconv.Atoi(strings.TrimSpace(item.CenterObjectID))
		if err != nil || objectID < 1 || objectID > maxDigitalCenterObjectID {
			return digitalCenterWriteTarget{}, ErrDigitalCenterDeviceOutput
		}
	}
	fields, err := normalizeDigitalCenterWriteFields(requestedFields)
	if err != nil {
		return digitalCenterWriteTarget{}, err
	}
	if operation == DigitalCenterWriteCreate && !equalDigitalCenterWriteFields(fields,
		[]string{"address", "name", "protocol"}) {
		return digitalCenterWriteTarget{}, ErrDigitalCenterWriteFieldUnsupported
	}
	desired, err := digitalCenterDesiredFromWorkItem(item, fields)
	if err != nil {
		return digitalCenterWriteTarget{}, err
	}
	return digitalCenterWriteTarget{
		session: session, item: item, center: center, objectID: objectID, fields: fields, desired: desired,
		operation: operation, previousObjectID: strings.TrimSpace(item.CenterObjectID),
	}, nil
}

func (service *DigitalCenterWorkspaceService) digitalCenterWriteNow() time.Time {
	if service != nil && service.now != nil {
		return service.now().UTC()
	}
	return time.Now().UTC()
}

func normalizeDigitalCenterWriteFields(values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{"address", "name", "protocol"}, nil
	}
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		field := strings.ToLower(strings.TrimSpace(value))
		switch field {
		case "address", "name", "protocol":
			unique[field] = struct{}{}
		default:
			return nil, ErrDigitalCenterWriteFieldUnsupported
		}
	}
	fields := make([]string, 0, len(unique))
	for field := range unique {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields, nil
}

func digitalCenterDesiredFromWorkItem(
	item DigitalCenterWorkItem,
	fields []string,
) (ECoSLocomotiveSyncDesired, error) {
	desired := ECoSLocomotiveSyncDesired{}
	for _, field := range fields {
		switch field {
		case "name":
			value, ok := digitalCenterMapString(item.RailKeeper, "name")
			if !ok {
				return ECoSLocomotiveSyncDesired{}, ErrDigitalCenterWriteFieldUnsupported
			}
			normalized, err := normalizeDigitalCenterName(value)
			if err != nil || normalized == "" {
				return ECoSLocomotiveSyncDesired{}, ErrDigitalCenterWriteFieldUnsupported
			}
			desired.Name = normalized
		case "address":
			value, ok := digitalCenterMapPositiveInt(item.RailKeeper, "decoderAddress")
			if !ok || value > maxDigitalCenterAddress {
				return ECoSLocomotiveSyncDesired{}, ErrDigitalCenterWriteFieldUnsupported
			}
			desired.Address = value
		case "protocol":
			value, ok := digitalCenterMapString(item.RailKeeper, "protocol")
			if !ok {
				return ECoSLocomotiveSyncDesired{}, ErrDigitalCenterWriteFieldUnsupported
			}
			normalized, err := normalizeDigitalCenterProtocol(value)
			if err != nil || normalized == "" {
				return ECoSLocomotiveSyncDesired{}, ErrDigitalCenterWriteFieldUnsupported
			}
			desired.Protocol = normalized
		}
	}
	return desired, nil
}

func digitalCenterMapString(values map[string]any, key string) (string, bool) {
	value, ok := values[key].(string)
	return strings.TrimSpace(value), ok
}

func digitalCenterMapPositiveInt(values map[string]any, key string) (int, bool) {
	value, found := values[key]
	if !found {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, typed > 0
	case int64:
		return int(typed), typed > 0
	case float64:
		parsed := int(typed)
		return parsed, typed == float64(parsed) && parsed > 0
	case json.Number:
		parsed, err := strconv.Atoi(typed.String())
		return parsed, err == nil && parsed > 0
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		return parsed, err == nil && parsed > 0
	default:
		return 0, false
	}
}

func (service *DigitalCenterWorkspaceService) previewDigitalCenterChanges(
	ctx context.Context,
	target digitalCenterWriteTarget,
) ([]ECoSLocomotiveSyncChange, error) {
	if target.operation == DigitalCenterWriteCreate {
		if _, err := buildECoSLocomotiveCreateCommand(target.desired); err != nil {
			return nil, ErrDigitalCenterWriteFieldUnsupported
		}
		return []ECoSLocomotiveSyncChange{
			{Field: "address", Current: "", Desired: strconv.Itoa(target.desired.Address)},
			{Field: "name", Current: "", Desired: target.desired.Name},
			{Field: "protocol", Current: "", Desired: target.desired.Protocol},
		}, nil
	}
	writer, ok := service.ecos.(digitalCenterECoSWriter)
	if !ok {
		return nil, ErrDigitalCenterWorkspaceUnavailable
	}
	result, err := writer.SyncLocomotive(ctx, ECoSLocomotiveSyncInput{
		Host: target.center.Host, Port: target.center.Port, ObjectID: target.objectID,
		Desired: target.desired, DryRun: true,
	})
	if err != nil {
		return nil, fmt.Errorf("preview digital center write: %w", err)
	}
	if result == nil {
		return nil, ErrDigitalCenterDeviceOutput
	}
	changes := make([]ECoSLocomotiveSyncChange, 0, len(result.Changes))
	allowed := make(map[string]struct{}, len(target.fields))
	seen := make(map[string]struct{}, len(result.Changes))
	for _, field := range target.fields {
		allowed[field] = struct{}{}
	}
	for _, change := range result.Changes {
		field := strings.ToLower(strings.TrimSpace(change.Field))
		if _, ok := allowed[field]; !ok {
			return nil, ErrDigitalCenterDeviceOutput
		}
		if _, duplicate := seen[field]; duplicate {
			return nil, ErrDigitalCenterDeviceOutput
		}
		seen[field] = struct{}{}
		current, err := normalizeDigitalCenterWriteValue(field, change.Current, field == "name")
		if err != nil {
			return nil, err
		}
		desired, err := normalizeDigitalCenterWriteValue(field, change.Desired, false)
		if err != nil || desired != expectedDigitalCenterWriteValue(field, target.desired) {
			return nil, ErrDigitalCenterDeviceOutput
		}
		if current == desired {
			continue
		}
		changes = append(changes, ECoSLocomotiveSyncChange{Field: field, Current: current, Desired: desired})
	}
	sort.Slice(changes, func(left, right int) bool { return changes[left].Field < changes[right].Field })
	if len(changes) == 0 {
		return nil, ErrDigitalCenterWriteNoChanges
	}
	return changes, nil
}

func normalizeDigitalCenterWriteValue(field string, value string, allowEmpty bool) (string, error) {
	switch field {
	case "name":
		normalized, err := normalizeDigitalCenterName(value)
		if err != nil || (!allowEmpty && normalized == "") {
			return "", ErrDigitalCenterDeviceOutput
		}
		return normalized, nil
	case "protocol":
		normalized, err := normalizeDigitalCenterProtocol(value)
		if err != nil || normalized == "" {
			return "", ErrDigitalCenterDeviceOutput
		}
		return normalized, nil
	case "address":
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || parsed < 1 || parsed > maxDigitalCenterAddress {
			return "", ErrDigitalCenterDeviceOutput
		}
		return strconv.Itoa(parsed), nil
	default:
		return "", ErrDigitalCenterDeviceOutput
	}
}

func expectedDigitalCenterWriteValue(field string, desired ECoSLocomotiveSyncDesired) string {
	switch field {
	case "name":
		return desired.Name
	case "address":
		return strconv.Itoa(desired.Address)
	case "protocol":
		return desired.Protocol
	default:
		return ""
	}
}

func hashDigitalCenterWrite(
	target digitalCenterWriteTarget,
	changes []ECoSLocomotiveSyncChange,
) (string, []string, error) {
	fields := make([]string, 0, len(changes))
	for _, change := range changes {
		fields = append(fields, change.Field)
	}
	payload, err := json.Marshal(digitalCenterWriteHashPayload{
		SessionID: target.session.ID, ItemID: target.item.ID, Provider: target.center.Provider,
		ObjectID: digitalCenterWriteObjectID(target), Operation: target.operation,
		Direction: DigitalCenterWriteRailKeeperToCenter,
		Fields:    fields, Changes: changes,
	})
	if err != nil {
		return "", nil, fmt.Errorf("hash digital center write preview: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), fields, nil
}

func newDigitalCenterWriteToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate digital center write token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	digest := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(digest[:]), nil
}

func desiredDigitalCenterFields(
	desired ECoSLocomotiveSyncDesired,
	fields []string,
) ECoSLocomotiveSyncDesired {
	selected := ECoSLocomotiveSyncDesired{}
	for _, field := range fields {
		switch field {
		case "name":
			selected.Name = desired.Name
		case "address":
			selected.Address = desired.Address
		case "protocol":
			selected.Protocol = desired.Protocol
		}
	}
	return selected
}
