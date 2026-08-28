package application_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDigitalCenterWorkspaceListsServerSettingsWithoutEditingThem(t *testing.T) {
	settings := &digitalSettingsStub{value: application.DigitalCenterSettings{
		Provider: "ecos",
		ECoS: application.DigitalProviderSettings{
			Enabled: true,
			Host:    "192.168.2.151",
			Port:    "15471",
		},
		Z21: application.DigitalProviderSettings{
			Enabled: false,
			Host:    "192.168.2.152",
			Port:    "21105",
		},
	}}
	wantSettings := settings.value
	service := application.NewDigitalCenterWorkspaceService(nil, settings, nil, nil, nil, nil)

	centers, err := service.ListConfiguredCenters(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(centers) != 2 {
		t.Fatalf("centers = %#v, want two configured centers", centers)
	}
	if centers[0].Provider != "ecos" || centers[0].Name != "ESU ECoS" || !centers[0].Active ||
		centers[0].Host != "192.168.2.151" || centers[0].Port != 15471 {
		t.Fatalf("ECoS summary = %#v", centers[0])
	}
	if centers[1].Provider != "z21" || centers[1].Name != "Z21" || centers[1].Active ||
		centers[1].Host != "192.168.2.152" || centers[1].Port != 21105 {
		t.Fatalf("inactive Z21 summary = %#v", centers[1])
	}
	if settings.readCalls != 1 {
		t.Fatalf("DigitalSettings calls = %d, want 1", settings.readCalls)
	}
	if !reflect.DeepEqual(settings.value, wantSettings) {
		t.Fatalf("workspace read mutated settings: got %#v want %#v", settings.value, wantSettings)
	}
}

func TestDigitalCenterWorkspaceSkipsMissingCentersInDeterministicProviderOrder(t *testing.T) {
	settings := &digitalSettingsStub{value: application.DigitalCenterSettings{
		Provider: "cs3",
		ECoS: application.DigitalProviderSettings{
			Enabled: true,
			Port:    "15471",
		},
		Z21: application.DigitalProviderSettings{
			Host: "z21.local",
			Port: "21105",
		},
		Intellibox3: application.DigitalProviderSettings{
			Host: "intellibox.local",
			Port: "21106",
		},
		CS3: application.DigitalProviderSettings{
			Enabled: true,
			Host:    "cs3.local",
			Port:    "8080",
		},
	}}
	service := application.NewDigitalCenterWorkspaceService(nil, settings, nil, nil, nil, nil)

	centers, err := service.ListConfiguredCenters(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	providers := make([]string, 0, len(centers))
	for _, center := range centers {
		providers = append(providers, center.Provider)
	}
	if want := []string{"z21", "intellibox3", "cs3"}; !reflect.DeepEqual(providers, want) {
		t.Fatalf("provider order = %v, want %v", providers, want)
	}
	if !centers[2].Selected {
		t.Fatalf("selected center = %#v, want CS3 selected", centers[2])
	}
}

func TestDigitalCenterWorkspaceReportsTruthfulProviderCapabilities(t *testing.T) {
	settings := &digitalSettingsStub{value: application.DigitalCenterSettings{
		Provider:    "ecos",
		ECoS:        application.DigitalProviderSettings{Host: "ecos.local", Port: "15471"},
		Z21:         application.DigitalProviderSettings{Host: "z21.local", Port: "21105"},
		Intellibox3: application.DigitalProviderSettings{Host: "intellibox.local", Port: "21105"},
		CS3:         application.DigitalProviderSettings{Host: "cs3.local", Port: "80"},
	}}
	service := application.NewDigitalCenterWorkspaceService(nil, settings, nil, nil, nil, nil)

	centers, err := service.ListConfiguredCenters(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	byProvider := map[string]application.DigitalCenterCapabilities{}
	transportsByProvider := map[string][]application.DigitalCenterTransport{}
	for _, center := range centers {
		byProvider[center.Provider] = center.Capabilities
		transportsByProvider[center.Provider] = center.Transports
	}
	assertDigitalCenterCapabilities(t, byProvider["ecos"], application.DigitalCenterCapabilities{
		TestConnection: true, ReadLocomotives: true, LiveMonitor: true, WriteLocomotives: true,
	})
	assertDigitalCenterCapabilities(t, byProvider["z21"], application.DigitalCenterCapabilities{
		TestConnection: true, ReadLocomotives: true, Diagnose: true,
	})
	assertDigitalCenterCapabilities(t, byProvider["intellibox3"], application.DigitalCenterCapabilities{
		TestConnection: true, Diagnose: true,
	})
	assertDigitalCenterCapabilities(t, byProvider["cs3"], application.DigitalCenterCapabilities{
		TestConnection: true, ReadLocomotives: true, Diagnose: true,
	})

	intelliboxTransports := transportsByProvider["intellibox3"]
	if len(intelliboxTransports) != 2 {
		t.Fatalf("Intellibox 3 transports = %#v, want Z21 UDP and planned LocoNet TCP", intelliboxTransports)
	}
	if intelliboxTransports[0].ID != "z21_udp" || intelliboxTransports[0].Status != "available" {
		t.Fatalf("Intellibox 3 Z21 transport = %#v, want available", intelliboxTransports[0])
	}
	assertDigitalCenterCapabilities(t, intelliboxTransports[0].Capabilities, application.DigitalCenterCapabilities{
		TestConnection: true, Diagnose: true,
	})
	if intelliboxTransports[1].ID != "loconet_tcp" || intelliboxTransports[1].Status != "planned" {
		t.Fatalf("Intellibox 3 LocoNet transport = %#v, want planned", intelliboxTransports[1])
	}
	assertDigitalCenterCapabilities(t, intelliboxTransports[1].Capabilities, application.DigitalCenterCapabilities{})
}

func TestDigitalCenterWorkspaceReturnsEmptyListWhenNoCenterIsConfigured(t *testing.T) {
	service := application.NewDigitalCenterWorkspaceService(
		nil,
		&digitalSettingsStub{value: application.DigitalCenterSettings{}},
		nil,
		nil,
		nil,
		nil,
	)

	centers, err := service.ListConfiguredCenters(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if centers == nil || len(centers) != 0 {
		t.Fatalf("centers = %#v, want non-nil empty list", centers)
	}
}

func TestDigitalCenterWorkspacePropagatesSettingsReadFailure(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := application.NewDigitalCenterWorkspaceService(
		nil,
		&digitalSettingsStub{err: wantErr},
		nil,
		nil,
		nil,
		nil,
	)

	_, err := service.ListConfiguredCenters(t.Context())
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped %v", err, wantErr)
	}
}

func assertDigitalCenterCapabilities(
	t *testing.T,
	got application.DigitalCenterCapabilities,
	want application.DigitalCenterCapabilities,
) {
	t.Helper()
	if got != want {
		t.Fatalf("capabilities = %#v, want %#v", got, want)
	}
}

type digitalSettingsStub struct {
	value     application.DigitalCenterSettings
	err       error
	readCalls int
}

func (stub *digitalSettingsStub) DigitalSettings(context.Context) (*application.DigitalCenterSettings, error) {
	stub.readCalls++
	if stub.err != nil {
		return nil, stub.err
	}
	value := stub.value
	return &value, nil
}
