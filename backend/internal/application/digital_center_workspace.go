package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type digitalCenterSettingsReader interface {
	DigitalSettings(context.Context) (*DigitalCenterSettings, error)
}

type digitalCenterECoSReader interface {
	ProbeLocomotiveRaw(context.Context, ECoSConnectionInput) (*ECoSRawProbe, error)
}

type digitalCenterVehicleReader interface {
	List(context.Context, string) ([]Vehicle, error)
}

type DigitalCenterSummary struct {
	Provider     string                    `json:"provider"`
	Name         string                    `json:"name"`
	Active       bool                      `json:"active"`
	Selected     bool                      `json:"selected"`
	Host         string                    `json:"host"`
	Port         int                       `json:"port"`
	Capabilities DigitalCenterCapabilities `json:"capabilities"`
}

type DigitalCenterWorkspaceService struct {
	repository     DigitalCenterWorkspaceRepository
	settings       digitalCenterSettingsReader
	ecos           digitalCenterECoSReader
	digitalCenters *DigitalCenterService
	vehicles       digitalCenterVehicleReader
	auth           *AuthService
}

func NewDigitalCenterWorkspaceService(
	repository DigitalCenterWorkspaceRepository,
	settings digitalCenterSettingsReader,
	ecos digitalCenterECoSReader,
	digitalCenters *DigitalCenterService,
	vehicles digitalCenterVehicleReader,
	auth *AuthService,
) *DigitalCenterWorkspaceService {
	return &DigitalCenterWorkspaceService{
		repository:     repository,
		settings:       settings,
		ecos:           ecos,
		digitalCenters: digitalCenters,
		vehicles:       vehicles,
		auth:           auth,
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
		})
	}
	return centers, nil
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
		return DigitalCenterCapabilities{TestConnection: true}
	default:
		return DigitalCenterCapabilities{}
	}
}
