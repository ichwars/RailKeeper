package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	ecospkg "railkeeper/backend/internal/ecos"
)

var (
	ErrECoSBackupUnavailable     = errors.New("ECoS backup decryption is unavailable")
	ErrECoSBackupInvalid         = errors.New("invalid ECoS backup")
	ErrECoSBackupStationMismatch = errors.New("ECoS backup station confirmation does not match")
)

type ECoSBackupImportInput struct {
	ConfirmStation bool
	StationHost    string
}

type ECoSBackupImportSummary struct {
	StationSerial  uint32 `json:"stationSerial"`
	Profiles       int    `json:"profiles"`
	CVValues       int    `json:"cvValues"`
	ImportedAt     string `json:"importedAt"`
	LinkedVehicles int    `json:"linkedVehicles"`
	IdentitySource string `json:"identitySource"`
}

type ECoSBackupImportResult struct {
	Session DigitalCenterReadSession `json:"session"`
	Summary ECoSBackupImportSummary  `json:"summary"`
}

// Identity labels are cached station observations, not recovered backup identity.
type ECoSBackupContext struct {
	Item             DigitalCenterWorkItem
	VehicleID        string
	VehicleName      string
	DecoderType      string
	ExternalAddress  string
	ExternalProtocol string
}

type digitalCenterBackupRepository interface {
	LoadECoSBackupContexts(context.Context, string, int, []int) (map[int]ECoSBackupContext, error)
	CreateECoSBackupSession(context.Context, DigitalCenterReadSession, []DigitalCenterWorkItem, DigitalCenterSessionMessage) (DigitalCenterReadSession, error)
}

func (service *DigitalCenterWorkspaceService) ImportECoSBackup(
	ctx context.Context, uploaded io.Reader, keyring, passphrase []byte, input ECoSBackupImportInput, actor string,
) (ECoSBackupImportResult, error) {
	if len(keyring) == 0 || len(passphrase) == 0 {
		return ECoSBackupImportResult{}, ErrECoSBackupUnavailable
	}
	backup, err := ecospkg.DecodeBackup(uploaded, keyring, passphrase)
	if err != nil {
		if errors.Is(err, ecospkg.ErrBackupKey) {
			return ECoSBackupImportResult{}, ErrECoSBackupUnavailable
		}
		return ECoSBackupImportResult{}, ErrECoSBackupInvalid
	}
	return service.importDecodedECoSBackup(ctx, backup, input, actor)
}

func (service *DigitalCenterWorkspaceService) importDecodedECoSBackup(
	ctx context.Context, backup ecospkg.Backup, input ECoSBackupImportInput, actor string,
) (result ECoSBackupImportResult, err error) {
	err = service.runECoSOperation(ctx, func() error {
		repository, ok := service.repository.(digitalCenterBackupRepository)
		if !ok {
			return ErrDigitalCenterWorkspaceUnavailable
		}
		center, err := service.configuredCenter(ctx, "ecos")
		if err != nil {
			return err
		}
		if input.ConfirmStation && (input.StationHost == "" || input.StationHost != center.Host) {
			return ErrECoSBackupStationMismatch
		}
		if len(backup.Locomotives) == 0 || len(backup.Locomotives) > maxDigitalCenterLocomotives {
			return ErrECoSBackupInvalid
		}
		ids := make([]int, 0, len(backup.Locomotives))
		seen := map[int]bool{}
		for _, locomotive := range backup.Locomotives {
			if locomotive.ObjectID < 1 || locomotive.ObjectID > maxDigitalCenterObjectID || seen[locomotive.ObjectID] {
				return ErrECoSBackupInvalid
			}
			seen[locomotive.ObjectID] = true
			ids = append(ids, locomotive.ObjectID)
		}
		contexts, err := repository.LoadECoSBackupContexts(ctx, center.Host, center.Port, ids)
		if err != nil {
			return err
		}
		importedAt := service.now().UTC().Format(time.RFC3339)
		summary := ECoSBackupImportSummary{StationSerial: backup.Serial, Profiles: len(ids),
			ImportedAt: importedAt, IdentitySource: "saved_comparison"}
		items := make([]DigitalCenterWorkItem, 0, len(ids))
		for _, locomotive := range backup.Locomotives {
			values := make([]ECoSCVValue, 0, len(locomotive.Values))
			for _, value := range locomotive.Values {
				values = append(values, ECoSCVValue{Number: value.Number, Value: value.Value,
					IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32})
			}
			values, err = validateStoredECoSCVs(values)
			if err != nil {
				return ErrECoSBackupInvalid
			}
			item := newECoSBackupItem(locomotive.ObjectID, contexts[locomotive.ObjectID], input.ConfirmStation)
			item.Center["cvs"] = values
			item.Center["cvRead"] = ECoSCVReadState{Source: "backup", NextCV: 1025,
				ReadAt: importedAt, SourceHost: center.Host}
			summary.CVValues += len(values)
			if item.VehicleID != "" {
				summary.LinkedVehicles++
			}
			items = append(items, item)
		}
		message := DigitalCenterSessionMessage{Severity: DigitalCenterMessageInfo,
			Code: DigitalCenterMessageReadCompleted,
			Message: fmt.Sprintf("ECoS-Sicherung offline importiert: %d Konfigurationen, %d CV-Werte. "+
				"Namen und Profilkontext stammen aus früheren gespeicherten Abgleichen. Zeitangabe ist die Importzeit. "+
				"Keine Geräteabfrage oder Fahrzeugübernahme. / Offline backup preview; labels use saved comparisons; timestamp is import time.",
				len(items), summary.CVValues),
			NextAction: "CV-Werte und Decoderbibliothek prüfen; vorhandene Angaben bleiben geschützt.",
		}
		session, err := repository.CreateECoSBackupSession(ctx, DigitalCenterReadSession{
			Provider: "ecos", State: DigitalCenterSessionReady, Host: center.Host, Port: center.Port,
			ReadStartedAt: importedAt, ReadCompletedAt: importedAt, CreatedByUserID: strings.TrimSpace(actor),
		}, items, message)
		if err != nil {
			return err
		}
		result = ECoSBackupImportResult{Session: session, Summary: summary}
		return nil
	})
	return result, err
}

func newECoSBackupItem(objectID int, context ECoSBackupContext, confirmed bool) DigitalCenterWorkItem {
	item := DigitalCenterWorkItem{CenterObjectID: strconv.Itoa(objectID), Name: fmt.Sprintf("ECoS-Objekt %d", objectID),
		CompareStatus: DigitalCompareNew, StationStatus: "read", Center: map[string]any{},
		RailKeeper: map[string]any{}, Proposed: map[string]any{}, Conflicts: []map[string]any{}}
	previous := context.Item
	if previous.CenterObjectID != item.CenterObjectID || (previous.StationStatus != "read" && previous.StationStatus != "incomplete") ||
		previous.CompareStatus == DigitalCompareConflict || len(previous.Conflicts) != 0 {
		return item
	}
	name, nameErr := normalizeDigitalCenterName(previous.Name)
	protocol, protocolErr := normalizeDigitalCenterProtocol(previous.Protocol)
	if nameErr != nil || protocolErr != nil || previous.Address < 1 || previous.Address > maxDigitalCenterAddress {
		return item
	}
	item.Name, item.Address, item.Protocol = name, previous.Address, protocol
	item.Center = map[string]any{"objectId": objectID, "name": name, "decoderAddress": item.Address, "protocol": protocol}
	if profile, ok := previous.Center["profile"].(string); ok && len(profile) <= 160 && utf8.ValidString(profile) {
		item.Center["profile"] = strings.Join(strings.Fields(profile), " ")
	}
	if !confirmed || context.VehicleID == "" || previous.VehicleID != context.VehicleID ||
		context.ExternalAddress == "" || context.ExternalProtocol == "" {
		return item
	}
	address, addressErr := strconv.Atoi(context.ExternalAddress)
	mappedProtocol, protocolErr := normalizeDigitalCenterProtocol(context.ExternalProtocol)
	if addressErr != nil || address != item.Address || protocolErr != nil || mappedProtocol != protocol {
		return item
	}
	item.VehicleID, item.CompareStatus = context.VehicleID, previous.CompareStatus
	item.RailKeeper = map[string]any{"vehicleId": context.VehicleID, "name": context.VehicleName,
		"decoderType": context.DecoderType, "decoderAddress": address, "protocol": mappedProtocol}
	return item
}
