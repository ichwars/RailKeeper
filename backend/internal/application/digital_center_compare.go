package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	Items        []DigitalCenterWorkItem `json:"items"`
	Page         int                     `json:"page"`
	PageSize     int                     `json:"pageSize"`
	Total        int                     `json:"total"`
	TotalPages   int                     `json:"totalPages"`
	SessionTotal int                     `json:"sessionTotal"`
}

func (service *DigitalCenterWorkspaceService) startReadSessionUnlocked(
	ctx context.Context,
	provider string,
	actor string,
) (DigitalCenterReadSession, error) {
	if service == nil || service.repository == nil || service.vehicles == nil {
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

	vehicles, err := service.vehicles.ListReadOnly(ctx, "")
	if err != nil {
		return service.finishFailedReadSession(ctx, session, fmt.Errorf("list RailKeeper vehicles: %w", err))
	}
	locomotives, skippedConsists, err := service.readDigitalCenterLocomotives(ctx, center, vehicles)
	if err != nil {
		return service.finishFailedReadSession(ctx, session, err)
	}
	if center.Provider == "ecos" {
		current, err := service.configuredCenter(ctx, center.Provider)
		if err != nil || !current.Active || current.Host != center.Host || current.Port != center.Port {
			return service.finishFailedReadSession(ctx, session, ErrDigitalCenterReadNotFresh)
		}
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
	message := fmt.Sprintf("%d Lokomotiven wurden in die Arbeitsliste gelesen.", len(locomotives))
	if skippedConsists > 0 {
		message += fmt.Sprintf(" %d ECoS-Mehrfachtraktionen mit Protokoll MULTI wurden ausgelassen; sie haben keine eigene Decoderadresse."+
			" / %d ECoS consists with protocol MULTI were skipped; they have no decoder address of their own.",
			skippedConsists, skippedConsists)
	}
	if err := service.repository.AddMessage(ctx, DigitalCenterSessionMessage{
		SessionID:  session.ID,
		Severity:   DigitalCenterMessageInfo,
		Code:       DigitalCenterMessageReadCompleted,
		Message:    message,
		NextAction: "Arbeitsliste prüfen und Konflikte klären.",
	}); err != nil {
		return DigitalCenterReadSession{}, fmt.Errorf("record completed digital center read: %w", err)
	}
	return session, nil
}

func (service *DigitalCenterWorkspaceService) readDigitalCenterLocomotives(
	ctx context.Context,
	center DigitalCenterSummary,
	vehicles []Vehicle,
) ([]ECoSRawLocomotive, int, error) {
	var locomotives []ECoSRawLocomotive
	switch center.Provider {
	case "ecos":
		if service.ecos == nil {
			return nil, 0, ErrDigitalCenterWorkspaceUnavailable
		}
		var probe *ECoSRawProbe
		var err error
		if reader, ok := service.ecos.(digitalCenterStoredRosterReader); ok {
			probe, err = reader.ReadStoredLocomotiveRoster(ctx, ECoSConnectionInput{Host: center.Host, Port: center.Port})
		} else {
			probe, err = service.ecos.ProbeLocomotiveRaw(ctx, ECoSConnectionInput{Host: center.Host, Port: center.Port})
		}
		if err != nil {
			return nil, 0, err
		}
		return normalizeDigitalCenterRoster(center.Provider, probe)
	case "cs3":
		if service.digitalCenters == nil {
			return nil, 0, ErrDigitalCenterWorkspaceUnavailable
		}
		var err error
		locomotives, _, err = service.digitalCenters.ReadCS3Locomotives(
			ctx,
			DigitalCenterConnectionInput{Host: center.Host, Port: center.Port},
		)
		if err != nil {
			return nil, 0, err
		}
	case "z21":
		if service.digitalCenters == nil {
			return nil, 0, ErrDigitalCenterWorkspaceUnavailable
		}
		var err error
		locomotives, err = service.digitalCenters.ReadZ21Locomotives(
			ctx,
			DigitalCenterConnectionInput{Host: center.Host, Port: center.Port},
			knownDigitalCenterAddresses(vehicles, center.Provider),
		)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, ErrDigitalCenterCapabilityUnavailable
	}
	return normalizeDigitalCenterRoster(center.Provider, &ECoSRawProbe{Locomotives: locomotives})
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
	sessionID = strings.TrimSpace(sessionID)
	if _, err := service.repository.GetSession(ctx, sessionID); err != nil {
		return DigitalCenterWorkItemPage{}, err
	}
	filter.Query = strings.ToLower(strings.TrimSpace(filter.Query))
	filter.Page = max(1, filter.Page)
	if filter.PageSize < 1 {
		filter.PageSize = 10
	}
	filter.PageSize = min(100, filter.PageSize)
	return service.repository.ListWorkItemPage(ctx, sessionID, filter)
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
