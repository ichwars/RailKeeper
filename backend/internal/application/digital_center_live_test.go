package application

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDigitalCenterWorkspaceLiveMonitorUsesConfiguredTargetAndRecordsMessages(t *testing.T) {
	repository := &workspaceRepositoryMemory{session: DigitalCenterReadSession{ID: "session-1"}}
	ecos := &workspaceECoSLiveStub{status: ECoSLiveStatus{Provider: "ecos", State: ECoSLiveRunning}}
	service := NewDigitalCenterWorkspaceService(
		repository, &workspaceSettingsReaderStub{value: configuredECoSSettings()}, ecos, nil,
		&workspaceVehicleReaderStub{}, nil,
	)

	status, err := service.StartLiveMonitor(t.Context(), "ecos", "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if ecos.startInput != (ECoSConnectionInput{Host: "ecos.local", Port: 15471}) ||
		status.State != ECoSLiveRunning {
		t.Fatalf("start input=%#v status=%#v", ecos.startInput, status)
	}
	if len(repository.messages) != 1 || repository.messages[0].Code != DigitalCenterMessageLiveStarted {
		t.Fatalf("messages = %#v, want live.started", repository.messages)
	}

	status, err = service.StopLiveMonitor(t.Context(), "ecos", "session-1")
	if err != nil || status.State != ECoSLiveStopped {
		t.Fatalf("stop status=%#v error=%v", status, err)
	}
	if len(repository.messages) != 2 || repository.messages[1].Code != DigitalCenterMessageLiveStopped {
		t.Fatalf("messages = %#v, want live.stopped", repository.messages)
	}

	messages, err := service.ListSessionMessages(t.Context(), "session-1")
	if err != nil || len(messages) != 2 {
		t.Fatalf("listed messages=%#v error=%v", messages, err)
	}
}

func TestDigitalCenterWorkspaceLiveMonitorRecordsMaskedConnectionFailure(t *testing.T) {
	repository := &workspaceRepositoryMemory{session: DigitalCenterReadSession{ID: "session-1"}}
	ecos := &workspaceECoSLiveStub{startErr: errors.New("dial 192.168.2.151 password=secret\x00")}
	service := NewDigitalCenterWorkspaceService(
		repository, &workspaceSettingsReaderStub{value: configuredECoSSettings()}, ecos, nil,
		&workspaceVehicleReaderStub{}, nil,
	)

	_, err := service.StartLiveMonitor(t.Context(), "ecos", "session-1")
	if err == nil {
		t.Fatal("expected connection failure")
	}
	if len(repository.messages) != 1 || repository.messages[0].Code != DigitalCenterMessageConnectionFailed ||
		repository.messages[0].Message == "" || repository.messages[0].NextAction == "" {
		t.Fatalf("messages = %#v", repository.messages)
	}
	text := strings.ToLower(repository.messages[0].Message + " " + repository.messages[0].NextAction)
	for _, forbidden := range []string{"192.168.2.151", "password", "secret", "\x00"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("private connection data persisted in %#v", repository.messages[0])
		}
	}
}

func TestDigitalCenterWorkspaceLiveMonitorRejectsUnsupportedProviderExplicitly(t *testing.T) {
	repository := &workspaceRepositoryMemory{session: DigitalCenterReadSession{ID: "session-1"}}
	settings := configuredECoSSettings()
	settings.Z21 = DigitalProviderSettings{Enabled: true, Host: "z21.local", Port: "21105"}
	service := NewDigitalCenterWorkspaceService(
		repository, &workspaceSettingsReaderStub{value: settings}, &workspaceECoSLiveStub{}, nil,
		&workspaceVehicleReaderStub{}, nil,
	)

	_, err := service.StartLiveMonitor(t.Context(), "z21", "session-1")
	if !errors.Is(err, ErrDigitalCenterCapabilityUnavailable) {
		t.Fatalf("error = %v, want capability unavailable", err)
	}
	if len(repository.messages) != 1 || repository.messages[0].Code != DigitalCenterMessageCapabilityUnavailable {
		t.Fatalf("messages = %#v, want capability.unavailable", repository.messages)
	}
}

type workspaceECoSLiveStub struct {
	workspaceECoSReaderStub
	status     ECoSLiveStatus
	startInput ECoSConnectionInput
	startErr   error
}

func (stub *workspaceECoSLiveStub) StartLive(_ context.Context, input ECoSConnectionInput) (*ECoSLiveStatus, error) {
	stub.startInput = input
	status := stub.status
	return &status, stub.startErr
}

func (stub *workspaceECoSLiveStub) StopLive() ECoSLiveStatus {
	stub.status.State = ECoSLiveStopped
	stub.status.Connected = false
	return stub.status
}

func (stub *workspaceECoSLiveStub) LiveStatus() ECoSLiveStatus {
	return stub.status
}
