package application

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
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
