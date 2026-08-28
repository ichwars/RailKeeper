package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrDigitalCenterLiveStartFailed = errors.New("digital center live start failed")

type digitalCenterSettingsReader interface {
	DigitalSettings(context.Context) (*DigitalCenterSettings, error)
}

type digitalCenterECoSReader interface {
	ProbeLocomotiveRaw(context.Context, ECoSConnectionInput) (*ECoSRawProbe, error)
}

type digitalCenterECoSLiveMonitor interface {
	StartLive(context.Context, ECoSConnectionInput) (*ECoSLiveStatus, error)
	StopLive() ECoSLiveStatus
	LiveStatus() ECoSLiveStatus
}

type digitalCenterECoSLiveInterruptionMonitor interface {
	StartLiveWithInterruption(context.Context, ECoSConnectionInput, func()) (*ECoSLiveStatus, error)
}

type digitalCenterVehicleReader interface {
	ListReadOnly(context.Context, string) ([]Vehicle, error)
}

type DigitalCenterSummary struct {
	Provider     string                    `json:"provider"`
	Name         string                    `json:"name"`
	Active       bool                      `json:"active"`
	Selected     bool                      `json:"selected"`
	Host         string                    `json:"host"`
	Port         int                       `json:"port"`
	Capabilities DigitalCenterCapabilities `json:"capabilities"`
	Transports   []DigitalCenterTransport  `json:"transports"`
}

type DigitalCenterTransport struct {
	ID           string                    `json:"id"`
	Status       string                    `json:"status"`
	Capabilities DigitalCenterCapabilities `json:"capabilities"`
}

type DigitalCenterWorkspaceService struct {
	operationMu    sync.Mutex
	repository     DigitalCenterWorkspaceRepository
	settings       digitalCenterSettingsReader
	ecos           digitalCenterECoSReader
	digitalCenters *DigitalCenterService
	vehicles       digitalCenterVehicleReader
	auth           digitalCenterAuditRecorder
	now            func() time.Time
}

func NewDigitalCenterWorkspaceService(
	repository DigitalCenterWorkspaceRepository,
	settings digitalCenterSettingsReader,
	ecos digitalCenterECoSReader,
	digitalCenters *DigitalCenterService,
	vehicles digitalCenterVehicleReader,
	auth digitalCenterAuditRecorder,
) *DigitalCenterWorkspaceService {
	return &DigitalCenterWorkspaceService{
		repository:     repository,
		settings:       settings,
		ecos:           ecos,
		digitalCenters: digitalCenters,
		vehicles:       vehicles,
		auth:           auth,
		now:            time.Now,
	}
}

func (service *DigitalCenterWorkspaceService) ListConfiguredCenters(
	ctx context.Context,
) ([]DigitalCenterSummary, error) {
	if service == nil || service.settings == nil {
		return nil, errors.New("digital center workspace settings are unavailable")
	}
	settings, err := service.settings.DigitalSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("list configured digital centers: %w", err)
	}
	if settings == nil {
		return []DigitalCenterSummary{}, nil
	}

	configured := []struct {
		provider string
		name     string
		settings DigitalProviderSettings
	}{
		{provider: "ecos", name: "ESU ECoS", settings: settings.ECoS},
		{provider: "z21", name: "Z21", settings: settings.Z21},
		{provider: "intellibox3", name: "Intellibox 3", settings: settings.Intellibox3},
		{provider: "cs3", name: "CS3", settings: settings.CS3},
	}
	centers := make([]DigitalCenterSummary, 0, len(configured))
	for _, center := range configured {
		host := strings.TrimSpace(center.settings.Host)
		if host == "" {
			continue
		}
		port, err := parseDigitalCenterSettingsPort(center.provider, center.settings.Port)
		if err != nil {
			return nil, err
		}
		centers = append(centers, DigitalCenterSummary{
			Provider:     center.provider,
			Name:         center.name,
			Active:       center.settings.Enabled,
			Selected:     settings.Provider == center.provider,
			Host:         host,
			Port:         port,
			Capabilities: capabilitiesForProvider(center.provider),
			Transports:   transportsForProvider(center.provider),
		})
	}
	return centers, nil
}

func transportsForProvider(provider string) []DigitalCenterTransport {
	available := func(id string, capabilities DigitalCenterCapabilities) DigitalCenterTransport {
		return DigitalCenterTransport{ID: id, Status: "available", Capabilities: capabilities}
	}
	switch provider {
	case "ecos":
		return []DigitalCenterTransport{available("ecos_tcp", capabilitiesForProvider(provider))}
	case "z21":
		return []DigitalCenterTransport{available("z21_udp", capabilitiesForProvider(provider))}
	case "intellibox3":
		return []DigitalCenterTransport{
			available("z21_udp", capabilitiesForProvider(provider)),
			{ID: "loconet_tcp", Status: "planned", Capabilities: DigitalCenterCapabilities{}},
		}
	case "cs3":
		return []DigitalCenterTransport{available("cs3_http", capabilitiesForProvider(provider))}
	default:
		return []DigitalCenterTransport{}
	}
}

func parseDigitalCenterSettingsPort(provider, value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("list configured digital centers: invalid %s port", provider)
	}
	return port, nil
}

func capabilitiesForProvider(provider string) DigitalCenterCapabilities {
	switch provider {
	case "ecos":
		return DigitalCenterCapabilities{
			TestConnection:   true,
			ReadLocomotives:  true,
			LiveMonitor:      true,
			WriteLocomotives: true,
		}
	case "z21", "intellibox3":
		return DigitalCenterCapabilities{TestConnection: true, Diagnose: true}
	case "cs3":
		return DigitalCenterCapabilities{
			TestConnection:  true,
			ReadLocomotives: true,
			Diagnose:        true,
		}
	default:
		return DigitalCenterCapabilities{}
	}
}

func (service *DigitalCenterWorkspaceService) startLiveMonitorUnlocked(
	ctx context.Context,
	provider string,
	sessionID string,
) (*ECoSLiveStatus, error) {
	if service == nil || service.ecos == nil {
		return nil, ErrDigitalCenterWorkspaceUnavailable
	}
	if err := service.requireSession(ctx, sessionID); err != nil {
		return nil, err
	}
	center, err := service.configuredCenter(ctx, provider)
	if err != nil {
		return nil, err
	}
	if !center.Active {
		return nil, ErrDigitalCenterInactive
	}
	if !center.Capabilities.LiveMonitor {
		if err := service.addSessionMessage(ctx, sessionID, DigitalCenterSessionMessage{
			Severity:   DigitalCenterMessageWarning,
			Code:       DigitalCenterMessageCapabilityUnavailable,
			Message:    "Diese Digitalzentrale unterstützt kein Live-Monitoring.",
			NextAction: "Eine Digitalzentrale mit Live-Monitoring auswählen.",
		}); err != nil {
			return nil, err
		}
		return nil, ErrDigitalCenterCapabilityUnavailable
	}
	monitor, ok := service.ecos.(digitalCenterECoSLiveMonitor)
	if !ok {
		return nil, ErrDigitalCenterWorkspaceUnavailable
	}
	start := monitor.StartLive
	if interruptionMonitor, supported := service.ecos.(digitalCenterECoSLiveInterruptionMonitor); supported {
		persistContext := context.WithoutCancel(ctx)
		var interruptionOnce sync.Once
		start = func(ctx context.Context, input ECoSConnectionInput) (*ECoSLiveStatus, error) {
			return interruptionMonitor.StartLiveWithInterruption(ctx, input, func() {
				interruptionOnce.Do(func() {
					_ = service.addSessionMessage(persistContext, sessionID, DigitalCenterSessionMessage{
						Severity:   DigitalCenterMessageWarning,
						Code:       DigitalCenterMessageLiveInterrupted,
						Message:    "Das passive Live-Monitoring wurde unerwartet unterbrochen.",
						NextAction: "Verbindung prüfen und das Monitoring erneut starten.",
					})
				})
			})
		}
	}
	status, err := start(ctx, ECoSConnectionInput{Host: center.Host, Port: center.Port})
	if err != nil {
		if messageErr := service.addSessionMessage(ctx, sessionID, DigitalCenterSessionMessage{
			Severity:   DigitalCenterMessageError,
			Code:       DigitalCenterMessageConnectionFailed,
			Message:    "Die Live-Verbindung zur Digitalzentrale konnte nicht gestartet werden.",
			NextAction: "Verbindung und Gerätekonfiguration prüfen und erneut starten.",
		}); messageErr != nil {
			return nil, messageErr
		}
		return nil, fmt.Errorf("%w: %v", ErrDigitalCenterLiveStartFailed, err)
	}
	if err := service.addSessionMessage(ctx, sessionID, DigitalCenterSessionMessage{
		Severity:   DigitalCenterMessageInfo,
		Code:       DigitalCenterMessageLiveStarted,
		Message:    "Das passive Live-Monitoring wurde gestartet.",
		NextAction: "Live-Status und Ereignisse beobachten.",
	}); err != nil {
		monitor.StopLive()
		return nil, err
	}
	return status, nil
}

func (service *DigitalCenterWorkspaceService) stopLiveMonitorUnlocked(
	ctx context.Context,
	provider string,
	sessionID string,
) (*ECoSLiveStatus, error) {
	if service == nil || service.ecos == nil {
		return nil, ErrDigitalCenterWorkspaceUnavailable
	}
	if err := service.requireSession(ctx, sessionID); err != nil {
		return nil, err
	}
	monitor, err := service.liveMonitorForProvider(provider)
	if err != nil {
		return nil, err
	}
	status := monitor.StopLive()
	if err := service.addSessionMessage(ctx, sessionID, DigitalCenterSessionMessage{
		Severity:   DigitalCenterMessageInfo,
		Code:       DigitalCenterMessageLiveStopped,
		Message:    "Das passive Live-Monitoring wurde beendet.",
		NextAction: "Das Monitoring kann bei Bedarf erneut gestartet werden.",
	}); err != nil {
		return nil, err
	}
	return &status, nil
}

func (service *DigitalCenterWorkspaceService) LiveMonitorStatus(
	ctx context.Context,
	provider string,
) (ECoSLiveStatus, error) {
	if service == nil || service.ecos == nil {
		return ECoSLiveStatus{}, ErrDigitalCenterWorkspaceUnavailable
	}
	monitor, err := service.liveMonitorForProvider(provider)
	if err != nil {
		return ECoSLiveStatus{}, err
	}
	return monitor.LiveStatus(), nil
}

func (service *DigitalCenterWorkspaceService) liveMonitorForProvider(
	provider string,
) (digitalCenterECoSLiveMonitor, error) {
	if strings.ToLower(strings.TrimSpace(provider)) != "ecos" {
		return nil, ErrDigitalCenterCapabilityUnavailable
	}
	monitor, ok := service.ecos.(digitalCenterECoSLiveMonitor)
	if !ok {
		return nil, ErrDigitalCenterWorkspaceUnavailable
	}
	return monitor, nil
}

func (service *DigitalCenterWorkspaceService) ListSessionMessages(
	ctx context.Context,
	sessionID string,
) ([]DigitalCenterSessionMessage, error) {
	if service == nil || service.repository == nil {
		return nil, ErrDigitalCenterWorkspaceUnavailable
	}
	sessionID = strings.TrimSpace(sessionID)
	if _, err := service.repository.GetSession(ctx, sessionID); err != nil {
		return nil, err
	}
	messages, err := service.repository.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list digital center session messages: %w", err)
	}
	return messages, nil
}

func (service *DigitalCenterWorkspaceService) requireSession(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	if service.repository == nil {
		return ErrDigitalCenterWorkspaceUnavailable
	}
	_, err := service.repository.GetSession(ctx, sessionID)
	return err
}

func (service *DigitalCenterWorkspaceService) addSessionMessage(
	ctx context.Context,
	sessionID string,
	message DigitalCenterSessionMessage,
) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	message.SessionID = sessionID
	if err := service.repository.AddMessage(ctx, message); err != nil {
		return fmt.Errorf("record digital center live message: %w", err)
	}
	return nil
}
