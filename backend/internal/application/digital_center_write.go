package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	digitalCenterWriteGrantTTL  = 10 * time.Minute
	digitalCenterReadFreshness  = 10 * time.Minute
	digitalCenterWriteAuditName = "DigitalCenterSynchronization"
)

var (
	ErrDigitalCenterReadNotFresh          = errors.New("digital center read is missing or stale")
	ErrDigitalCenterConflictUnresolved    = errors.New("digital center work item conflict is unresolved")
	ErrDigitalCenterWriteFieldUnsupported = errors.New("digital center write field is unsupported")
	ErrDigitalCenterConfirmationRequired  = errors.New("digital center write confirmation is required")
	ErrDigitalCenterGrantMismatch         = errors.New("digital center write grant does not match request")
	ErrDigitalCenterPreviewStale          = errors.New("digital center write preview is stale")
	ErrDigitalCenterWriteNoChanges        = errors.New("digital center write has no changes")
	ErrDigitalCenterDeviceWrite           = errors.New("digital center device write failed")
	ErrDigitalCenterLivePauseFailed       = errors.New("digital center live monitor could not be paused")
)

type DigitalCenterWriteDirection string
type DigitalCenterWriteResultStatus string
type DigitalCenterWriteOperation string

const (
	DigitalCenterWriteRailKeeperToCenter DigitalCenterWriteDirection    = "railkeeper_to_center"
	DigitalCenterWriteVerified           DigitalCenterWriteResultStatus = "verified"
	DigitalCenterWriteVerificationFailed DigitalCenterWriteResultStatus = "verification_failed"
	DigitalCenterWriteFailed             DigitalCenterWriteResultStatus = "failed"
	DigitalCenterWriteUnknown            DigitalCenterWriteResultStatus = "unknown"
	DigitalCenterWriteUpdate             DigitalCenterWriteOperation    = "update"
	DigitalCenterWriteCreate             DigitalCenterWriteOperation    = "create"
)

type DigitalCenterWritePreviewInput struct {
	Operation DigitalCenterWriteOperation `json:"operation,omitempty"`
	Fields    []string                    `json:"fields"`
}

type DigitalCenterWriteConfirmInput struct {
	Operation DigitalCenterWriteOperation `json:"operation,omitempty"`
	Token     string                      `json:"token"`
	Confirm   bool                        `json:"confirm"`
	Fields    []string                    `json:"fields"`
}

type DigitalCenterWritePreview struct {
	SessionID string                      `json:"sessionId"`
	ItemID    string                      `json:"itemId"`
	Provider  string                      `json:"provider"`
	ObjectID  string                      `json:"objectId"`
	Operation DigitalCenterWriteOperation `json:"operation"`
	Direction DigitalCenterWriteDirection `json:"direction"`
	Fields    []string                    `json:"fields"`
	Changes   []ECoSLocomotiveSyncChange  `json:"changes"`
	Token     string                      `json:"token"`
	ExpiresAt string                      `json:"expiresAt"`
}

type DigitalCenterWriteConfirmation struct {
	SessionID      string                         `json:"sessionId"`
	ItemID         string                         `json:"itemId"`
	Provider       string                         `json:"provider"`
	ObjectID       string                         `json:"objectId"`
	Operation      DigitalCenterWriteOperation    `json:"operation"`
	Direction      DigitalCenterWriteDirection    `json:"direction"`
	Fields         []string                       `json:"fields"`
	Applied        bool                           `json:"applied"`
	Verified       bool                           `json:"verified"`
	Result         DigitalCenterWriteResultStatus `json:"result"`
	Message        string                         `json:"message"`
	VerifiedValues *ECoSLocomotiveSyncSnapshot    `json:"verifiedValues,omitempty"`
	LiveMonitor    DigitalCenterWriteLiveResult   `json:"liveMonitor"`
	WorkItem       *DigitalCenterWorkItem         `json:"workItem,omitempty"`
}

type DigitalCenterWriteLiveResult struct {
	WasRunning bool `json:"wasRunning"`
	Restarted  bool `json:"restarted"`
}

type digitalCenterECoSWriter interface {
	SyncLocomotive(context.Context, ECoSLocomotiveSyncInput) (*ECoSLocomotiveSyncResult, error)
}

type digitalCenterECoSPauser interface {
	PauseLive(context.Context) (ECoSLiveStatus, error)
}

type digitalCenterECoSTargetReader interface {
	ReadLocomotive(context.Context, ECoSConnectionInput, int) (ECoSLocomotive, error)
}

type digitalCenterAuditRecorder interface {
	RecordAudit(context.Context, string, string, string, string, string) error
}

type digitalCenterWriteTarget struct {
	session          DigitalCenterReadSession
	item             DigitalCenterWorkItem
	center           DigitalCenterSummary
	objectID         int
	fields           []string
	desired          ECoSLocomotiveSyncDesired
	operation        DigitalCenterWriteOperation
	previousObjectID string
}

type digitalCenterWriteHashPayload struct {
	SessionID string                      `json:"sessionId"`
	ItemID    string                      `json:"itemId"`
	Provider  string                      `json:"provider"`
	ObjectID  string                      `json:"objectId"`
	Operation DigitalCenterWriteOperation `json:"operation"`
	Direction DigitalCenterWriteDirection `json:"direction"`
	Fields    []string                    `json:"fields"`
	Changes   []ECoSLocomotiveSyncChange  `json:"changes"`
}

func (service *DigitalCenterWorkspaceService) previewWriteUnlocked(
	ctx context.Context,
	sessionID string,
	itemID string,
	input DigitalCenterWritePreviewInput,
	actor string,
) (DigitalCenterWritePreview, error) {
	target, err := service.digitalCenterWriteTarget(ctx, sessionID, itemID, input.Fields, input.Operation)
	if err != nil {
		return DigitalCenterWritePreview{}, err
	}
	changes, err := service.previewDigitalCenterChanges(ctx, target)
	if err != nil {
		return DigitalCenterWritePreview{}, err
	}
	if err := service.checkDigitalCenterAddressConflict(ctx, target, changes); err != nil {
		return DigitalCenterWritePreview{}, err
	}
	previewHash, fields, err := hashDigitalCenterWrite(target, changes)
	if err != nil {
		return DigitalCenterWritePreview{}, err
	}
	token, tokenHash, err := newDigitalCenterWriteToken()
	if err != nil {
		return DigitalCenterWritePreview{}, err
	}
	now := service.digitalCenterWriteNow()
	expiresAt := now.Add(digitalCenterWriteGrantTTL).Format(time.RFC3339)
	if err := service.repository.CreateWriteGrant(ctx, DigitalCenterWriteGrant{
		TokenHash: tokenHash, SessionID: target.session.ID, WorkItemID: target.item.ID,
		PreviewHash: previewHash, ActorUserID: strings.TrimSpace(actor),
		ExpiresAt: expiresAt, CreatedAt: now.Format(time.RFC3339),
	}); err != nil {
		return DigitalCenterWritePreview{}, fmt.Errorf("create digital center write preview: %w", err)
	}
	return DigitalCenterWritePreview{
		SessionID: target.session.ID, ItemID: target.item.ID, Provider: target.center.Provider,
		ObjectID: digitalCenterWriteObjectID(target), Operation: target.operation,
		Direction: DigitalCenterWriteRailKeeperToCenter,
		Fields:    fields, Changes: changes, Token: token, ExpiresAt: expiresAt,
	}, nil
}

func (service *DigitalCenterWorkspaceService) confirmWriteUnlocked(
	ctx context.Context,
	sessionID string,
	itemID string,
	input DigitalCenterWriteConfirmInput,
	actor string,
) (result DigitalCenterWriteConfirmation, returnErr error) {
	if !input.Confirm {
		return DigitalCenterWriteConfirmation{}, ErrDigitalCenterConfirmationRequired
	}
	token := strings.TrimSpace(input.Token)
	if token == "" {
		return DigitalCenterWriteConfirmation{}, ErrDigitalCenterGrantMismatch
	}
	target, err := service.digitalCenterWriteTarget(ctx, sessionID, itemID, input.Fields, input.Operation)
	if err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	tokenDigest := sha256.Sum256([]byte(token))
	grant, err := service.repository.ConsumeWriteGrant(ctx, hex.EncodeToString(tokenDigest[:]), strings.TrimSpace(actor))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DigitalCenterWriteConfirmation{}, ErrDigitalCenterGrantMismatch
		}
		return DigitalCenterWriteConfirmation{}, err
	}
	if grant.SessionID != target.session.ID || grant.WorkItemID != target.item.ID {
		return DigitalCenterWriteConfirmation{}, ErrDigitalCenterGrantMismatch
	}
	liveResult, err := service.pauseDigitalCenterLive(ctx)
	if err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	defer func() {
		service.resumeDigitalCenterLive(context.WithoutCancel(ctx), target, &liveResult)
		result.LiveMonitor = liveResult
	}()

	changes, err := service.previewDigitalCenterChanges(ctx, target)
	if err != nil {
		if errors.Is(err, ErrDigitalCenterWriteNoChanges) {
			return DigitalCenterWriteConfirmation{}, ErrDigitalCenterPreviewStale
		}
		return DigitalCenterWriteConfirmation{}, err
	}
	if err := service.checkDigitalCenterAddressConflict(ctx, target, changes); err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	previewHash, fields, err := hashDigitalCenterWrite(target, changes)
	if err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	if subtle.ConstantTimeCompare([]byte(previewHash), []byte(grant.PreviewHash)) != 1 {
		return DigitalCenterWriteConfirmation{}, ErrDigitalCenterPreviewStale
	}
	target.fields = fields
	target.desired = desiredDigitalCenterFields(target.desired, fields)
	result = service.digitalCenterWriteResult(target, fields)
	applied, createdObjectID, err := service.applyDigitalCenterWrite(ctx, target)
	if err != nil {
		if errors.Is(err, ErrECoSWriteStateUnknown) {
			result.Result = DigitalCenterWriteUnknown
			result.Message = "Der Schreibstatus ist unbekannt. Die Änderung wird nicht automatisch wiederholt."
			_ = service.auditDigitalCenterWrite(ctx, actor, target, fields, result.Result)
			_ = service.addSessionMessage(ctx, target.session.ID, DigitalCenterSessionMessage{
				Severity: DigitalCenterMessageWarning, Code: DigitalCenterMessageWriteUnknown,
				Message:    "Der Schreibstatus der Digitalzentrale ist unbekannt.",
				NextAction: "Werte neu auslesen und vor einem weiteren Schreiben prüfen.",
			})
			return result, nil
		}
		result.Result = DigitalCenterWriteFailed
		result.Message = "Die Digitalzentrale hat den Schreibvorgang abgelehnt."
		_ = service.auditDigitalCenterWrite(ctx, actor, target, fields, result.Result)
		return result, nil
	}
	result.Applied = applied
	if target.operation == DigitalCenterWriteCreate {
		target.objectID = createdObjectID
		target.item.CenterObjectID = strconv.Itoa(createdObjectID)
		result.ObjectID = target.item.CenterObjectID
	}
	verifiedLocomotive, verified, err := service.verifyDigitalCenterWrite(ctx, target)
	if err != nil {
		result.Result = DigitalCenterWriteUnknown
		result.Message = "Die geschriebenen Werte konnten nicht verifiziert werden."
		_ = service.auditDigitalCenterWrite(ctx, actor, target, fields, result.Result)
		_ = service.addSessionMessage(ctx, target.session.ID, DigitalCenterSessionMessage{
			Severity: DigitalCenterMessageWarning, Code: DigitalCenterMessageWriteUnknown,
			Message:    "Der Schreibstatus der Digitalzentrale ist unbekannt.",
			NextAction: "Werte neu auslesen und vor einem weiteren Schreiben prüfen.",
		})
		return result, nil
	}
	verifiedValues := ECoSLocomotiveSyncSnapshot{
		Name: verifiedLocomotive.Name, Address: verifiedLocomotive.Address, Protocol: verifiedLocomotive.Protocol,
	}
	result.VerifiedValues = &verifiedValues
	if !verified {
		result.Result = DigitalCenterWriteVerificationFailed
		result.Message = "Die Digitalzentrale meldet nach dem Schreiben abweichende Werte."
		var mappingErr error
		if target.operation == DigitalCenterWriteCreate {
			mappingErr = service.persistDigitalCenterMapping(ctx, target, verifiedLocomotive, actor, "linked")
		}
		updatedItem, updateErr := service.updateVerifiedDigitalCenterWorkItem(ctx, target, verifiedLocomotive)
		if err := service.auditDigitalCenterWrite(ctx, actor, target, fields, result.Result); err != nil {
			return DigitalCenterWriteConfirmation{}, err
		}
		if errors.Is(mappingErr, ErrDigitalCenterWorkspaceUnavailable) {
			return DigitalCenterWriteConfirmation{}, mappingErr
		}
		if mappingErr != nil {
			return service.digitalCenterLocalPersistenceUnknown(ctx, target, result), nil
		}
		if updateErr != nil {
			return service.digitalCenterLocalPersistenceUnknown(ctx, target, result), nil
		}
		result.WorkItem = &updatedItem
		return result, nil
	}
	result.Verified = true
	result.Result = DigitalCenterWriteVerified
	result.Message = "Die Änderung wurde geschrieben und verifiziert."
	if err := service.auditDigitalCenterWrite(ctx, actor, target, fields, result.Result); err != nil {
		return service.digitalCenterLocalPersistenceUnknown(ctx, target, result), nil
	}
	if err := service.persistDigitalCenterMapping(ctx, target, verifiedLocomotive, actor, "synced"); err != nil {
		if errors.Is(err, ErrDigitalCenterWorkspaceUnavailable) {
			return DigitalCenterWriteConfirmation{}, err
		}
		return service.digitalCenterLocalPersistenceUnknown(ctx, target, result), nil
	}
	updatedItem, err := service.updateVerifiedDigitalCenterWorkItem(ctx, target, verifiedLocomotive)
	if err != nil {
		return service.digitalCenterLocalPersistenceUnknown(ctx, target, result), nil
	}
	result.WorkItem = &updatedItem
	return result, nil
}

func (service *DigitalCenterWorkspaceService) digitalCenterLocalPersistenceUnknown(
	ctx context.Context,
	target digitalCenterWriteTarget,
	result DigitalCenterWriteConfirmation,
) DigitalCenterWriteConfirmation {
	result.Result = DigitalCenterWriteUnknown
	result.Message = "Der bekannte Gerätezustand konnte lokal nicht vollständig gespeichert werden."
	_ = service.addSessionMessage(ctx, target.session.ID, DigitalCenterSessionMessage{
		Severity: DigitalCenterMessageWarning, Code: DigitalCenterMessageWriteUnknown,
		Message:    "Der bekannte Gerätezustand konnte lokal nicht vollständig gespeichert werden.",
		NextAction: "Daten neu auslesen und den lokalen Abgleich prüfen.",
	})
	return result
}

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

func (service *DigitalCenterWorkspaceService) verifyDigitalCenterWrite(
	ctx context.Context,
	target digitalCenterWriteTarget,
) (ECoSLocomotive, bool, error) {
	reader, ok := service.ecos.(digitalCenterECoSTargetReader)
	if !ok {
		return ECoSLocomotive{}, false, ErrDigitalCenterWorkspaceUnavailable
	}
	locomotive, err := reader.ReadLocomotive(ctx, ECoSConnectionInput{
		Host: target.center.Host, Port: target.center.Port,
	}, target.objectID)
	if err != nil {
		return ECoSLocomotive{}, false, fmt.Errorf("verify digital center write: %w", err)
	}
	name, nameErr := normalizeDigitalCenterName(locomotive.Name)
	protocol, protocolErr := normalizeDigitalCenterProtocol(locomotive.Protocol)
	if nameErr != nil || protocolErr != nil || locomotive.ObjectID != target.objectID ||
		locomotive.Address < 1 || locomotive.Address > maxDigitalCenterAddress {
		return ECoSLocomotive{}, false, ErrDigitalCenterDeviceOutput
	}
	locomotive.Name = name
	locomotive.Protocol = protocol
	for _, field := range target.fields {
		switch field {
		case "name":
			if locomotive.Name != target.desired.Name {
				return locomotive, false, nil
			}
		case "address":
			if locomotive.Address != target.desired.Address {
				return locomotive, false, nil
			}
		case "protocol":
			if locomotive.Protocol != target.desired.Protocol {
				return locomotive, false, nil
			}
		}
	}
	return locomotive, true, nil
}

func (service *DigitalCenterWorkspaceService) pauseDigitalCenterLive(
	ctx context.Context,
) (DigitalCenterWriteLiveResult, error) {
	monitor, ok := service.ecos.(digitalCenterECoSLiveMonitor)
	if !ok {
		return DigitalCenterWriteLiveResult{}, nil
	}
	status := monitor.LiveStatus()
	live := DigitalCenterWriteLiveResult{WasRunning: status.State == ECoSLiveRunning && status.Connected}
	pauser, ok := service.ecos.(digitalCenterECoSPauser)
	if !ok {
		if !live.WasRunning {
			return live, nil
		}
		return live, ErrDigitalCenterLivePauseFailed
	}
	if _, err := pauser.PauseLive(ctx); err != nil {
		return live, fmt.Errorf("%w: %v", ErrDigitalCenterLivePauseFailed, err)
	}
	return live, nil
}

func (service *DigitalCenterWorkspaceService) resumeDigitalCenterLive(
	ctx context.Context,
	target digitalCenterWriteTarget,
	result *DigitalCenterWriteLiveResult,
) {
	if result == nil || !result.WasRunning {
		return
	}
	_, err := service.startLiveMonitorUnlocked(ctx, target.center.Provider, target.session.ID)
	result.Restarted = err == nil
	if err == nil {
		return
	}
	_ = service.addSessionMessage(ctx, target.session.ID, DigitalCenterSessionMessage{
		Severity: DigitalCenterMessageWarning, Code: DigitalCenterMessageLiveRestartFailed,
		Message:    "Das Live-Monitoring konnte nach dem Schreibvorgang nicht neu gestartet werden.",
		NextAction: "Live-Monitoring manuell starten und Verbindung prüfen.",
	})
}

func (service *DigitalCenterWorkspaceService) updateVerifiedDigitalCenterWorkItem(
	ctx context.Context,
	target digitalCenterWriteTarget,
	verified ECoSLocomotive,
) (DigitalCenterWorkItem, error) {
	item := target.item
	item.CenterObjectID = strconv.Itoa(verified.ObjectID)
	item.Name = verified.Name
	item.Address = verified.Address
	item.Protocol = verified.Protocol
	item.Center = map[string]any{
		"objectId": verified.ObjectID, "name": verified.Name,
		"decoderAddress": verified.Address, "protocol": verified.Protocol,
	}
	item.CompareStatus = compareDigitalCenterPayloads(item.Center, item.RailKeeper)
	item.StationStatus = "read"
	item.Proposed = map[string]any{}
	item.Conflicts = []map[string]any{}
	updated, err := service.repository.UpdateWorkItem(ctx, item)
	if err != nil {
		return DigitalCenterWorkItem{}, fmt.Errorf("update verified digital center work item: %w", err)
	}
	return updated, nil
}

func compareDigitalCenterPayloads(
	center map[string]any,
	railKeeper map[string]any,
) DigitalCenterCompareStatus {
	centerName, centerNameOK := digitalCenterMapString(center, "name")
	railKeeperName, railKeeperNameOK := digitalCenterMapString(railKeeper, "name")
	centerAddress, centerAddressOK := digitalCenterMapPositiveInt(center, "decoderAddress")
	railKeeperAddress, railKeeperAddressOK := digitalCenterMapPositiveInt(railKeeper, "decoderAddress")
	centerProtocol, centerProtocolOK := digitalCenterMapString(center, "protocol")
	railKeeperProtocol, railKeeperProtocolOK := digitalCenterMapString(railKeeper, "protocol")
	if centerNameOK && railKeeperNameOK && centerName == railKeeperName &&
		centerAddressOK && railKeeperAddressOK && centerAddress == railKeeperAddress &&
		centerProtocolOK && railKeeperProtocolOK && centerProtocol == railKeeperProtocol {
		return DigitalCompareOK
	}
	return DigitalCompareDeviation
}

func (service *DigitalCenterWorkspaceService) digitalCenterWriteResult(
	target digitalCenterWriteTarget,
	fields []string,
) DigitalCenterWriteConfirmation {
	return DigitalCenterWriteConfirmation{
		SessionID: target.session.ID, ItemID: target.item.ID, Provider: target.center.Provider,
		ObjectID: digitalCenterWriteObjectID(target), Operation: target.operation,
		Direction: DigitalCenterWriteRailKeeperToCenter,
		Fields:    append([]string(nil), fields...), Result: DigitalCenterWriteFailed,
	}
}

func (service *DigitalCenterWorkspaceService) auditDigitalCenterWrite(
	ctx context.Context,
	actor string,
	target digitalCenterWriteTarget,
	fields []string,
	result DigitalCenterWriteResultStatus,
) error {
	if service.auth == nil {
		return ErrDigitalCenterWorkspaceUnavailable
	}
	details, err := json.Marshal(struct {
		Station   string                         `json:"station"`
		ObjectID  string                         `json:"objectId"`
		Operation DigitalCenterWriteOperation    `json:"operation"`
		Fields    []string                       `json:"fields"`
		Result    DigitalCenterWriteResultStatus `json:"result"`
	}{Station: target.center.Provider, ObjectID: digitalCenterWriteObjectID(target),
		Operation: target.operation, Fields: fields, Result: result})
	if err != nil {
		return fmt.Errorf("encode digital center write audit: %w", err)
	}
	if err := service.auth.RecordAudit(ctx, strings.TrimSpace(actor), digitalCenterWriteAuditName,
		"digital_center_work_item", target.item.ID, string(details)); err != nil {
		return fmt.Errorf("record digital center write audit: %w", err)
	}
	return nil
}
