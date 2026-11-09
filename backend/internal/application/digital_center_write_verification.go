package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

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
