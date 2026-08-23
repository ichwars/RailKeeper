package application

import (
	"context"
	"errors"
	"math"
	"testing"
)

func TestCompareDigitalCenterLocomotiveUsesExistingMapping(t *testing.T) {
	item := compareDigitalCenterLocomotive("ecos", ECoSRawLocomotive{
		ObjectID: 3, Name: "BR 218", Address: 3, Protocol: "DCC128",
	}, []Vehicle{mappedDigitalCenterVehicle("vehicle-1", "ecos", "3", "3", "DCC")})

	if item.VehicleID != "vehicle-1" || item.CompareStatus != DigitalCompareOK {
		t.Fatalf("comparison = %#v, want mapped OK", item)
	}
}

func TestCompareDigitalCenterLocomotiveUsesRailKeeperAddressAsMappedWriteTarget(t *testing.T) {
	vehicle := mappedDigitalCenterVehicle("vehicle-1", "ecos", "3", "3", "DCC")
	vehicle.DigitalDecoderNumber = "18"
	item := compareDigitalCenterLocomotive("ecos", ECoSRawLocomotive{
		ObjectID: 3, Name: "BR 218", Address: 3, Protocol: "DCC128",
	}, []Vehicle{vehicle})

	if item.CompareStatus != DigitalCompareDeviation || item.RailKeeper["decoderAddress"] != 18 {
		t.Fatalf("comparison = %#v, want mapped deviation with RailKeeper address 18", item)
	}
}

func TestCompareDigitalCenterLocomotiveProposesUniqueAddressAndNormalizedProtocol(t *testing.T) {
	vehicle := mappedDigitalCenterVehicle("vehicle-1", "ecos", "old-object", "3", "dcc-28")
	item := compareDigitalCenterLocomotive("ecos", ECoSRawLocomotive{
		ObjectID: 4, Name: "BR 218", Address: 3, Protocol: " DCC128 ",
	}, []Vehicle{vehicle})

	if item.VehicleID != "" || item.CompareStatus != DigitalCompareOK || item.Proposed["vehicleId"] != "vehicle-1" {
		t.Fatalf("comparison = %#v, want unique address/protocol proposal", item)
	}
}

func TestCompareDigitalCenterLocomotiveMarksUniqueAddressOnlyAsDeviation(t *testing.T) {
	vehicle := mappedDigitalCenterVehicle("vehicle-1", "ecos", "old-object", "18", "Motorola")
	item := compareDigitalCenterLocomotive("ecos", ECoSRawLocomotive{
		ObjectID: 5, Name: "ICE 3", Address: 18, Protocol: "DCC128",
	}, []Vehicle{vehicle})

	if item.CompareStatus != DigitalCompareDeviation || item.Proposed["vehicleId"] != "vehicle-1" {
		t.Fatalf("comparison = %#v, want address-only deviation", item)
	}
}

func TestCompareDigitalCenterLocomotiveMarksMultipleAddressCandidatesAsConflict(t *testing.T) {
	vehicles := []Vehicle{
		mappedDigitalCenterVehicle("vehicle-1", "ecos", "old-1", "24", "DCC"),
		mappedDigitalCenterVehicle("vehicle-2", "ecos", "old-2", "24", "DCC"),
	}
	item := compareDigitalCenterLocomotive("ecos", ECoSRawLocomotive{
		ObjectID: 6, Name: "V 200", Address: 24, Protocol: "DCC128",
	}, vehicles)

	if item.CompareStatus != DigitalCompareConflict || len(item.Conflicts) != 2 {
		t.Fatalf("comparison = %#v, want two-candidate conflict", item)
	}
}

func TestCompareDigitalCenterLocomotiveNeverMatchesByNameAlone(t *testing.T) {
	item := compareDigitalCenterLocomotive("ecos", ECoSRawLocomotive{
		ObjectID: 7, Name: "BR 218", Address: 7, Protocol: "DCC",
	}, []Vehicle{{ID: "vehicle-1", Name: "BR 218", DigitalDecoderNumber: "99"}})

	if item.CompareStatus != DigitalCompareNew || item.VehicleID != "" || len(item.Proposed) != 0 {
		t.Fatalf("comparison = %#v, want new locomotive", item)
	}
}

func TestCompareDigitalCenterLocomotivesAddsMappedVehiclesMissingFromStation(t *testing.T) {
	vehicles := []Vehicle{
		mappedDigitalCenterVehicle("present", "ecos", "3", "3", "DCC"),
		mappedDigitalCenterVehicle("missing", "ecos", "42", "42", "DCC"),
	}
	vehicles[1].Name = "Fehlende Lok"
	items := compareDigitalCenterLocomotives("ecos", []ECoSRawLocomotive{{
		ObjectID: 3, Name: "BR 218", Address: 3, Protocol: "DCC",
	}}, vehicles)

	if len(items) != 2 || items[1].VehicleID != "missing" ||
		items[1].CenterObjectID != "42" || items[1].CompareStatus != DigitalCompareMissing {
		t.Fatalf("items = %#v, want missing mapped vehicle", items)
	}
}

func TestDigitalCenterWorkspaceReadUsesConfiguredTargetAndPersistsReadyWorklist(t *testing.T) {
	repository := &workspaceRepositoryMemory{}
	settings := &workspaceSettingsReaderStub{value: DigitalCenterSettings{
		Provider: "ecos",
		ECoS: DigitalProviderSettings{
			Enabled: true, Host: "trusted-center.local", Port: "15471",
		},
	}}
	ecos := &workspaceECoSReaderStub{probe: &ECoSRawProbe{Locomotives: []ECoSRawLocomotive{{
		ObjectID: 3, Name: "  BR\t218  ", Address: 3, Protocol: "dcc-128",
	}}}}
	vehicles := &workspaceVehicleReaderStub{vehicles: []Vehicle{
		mappedDigitalCenterVehicle("vehicle-1", "ecos", "3", "3", "DCC"),
	}}
	service := NewDigitalCenterWorkspaceService(repository, settings, ecos, nil, vehicles, nil)

	session, err := service.StartReadSession(t.Context(), "ecos", "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if ecos.input != (ECoSConnectionInput{Host: "trusted-center.local", Port: 15471}) {
		t.Fatalf("ECoS target = %#v, want configured server target", ecos.input)
	}
	if session.State != DigitalCenterSessionReady || session.CreatedByUserID != "admin-1" ||
		session.Host != "trusted-center.local" || session.ReadStartedAt == "" || session.ReadCompletedAt == "" {
		t.Fatalf("session = %#v, want completed configured read", session)
	}
	if len(repository.items) != 1 || repository.items[0].Name != "BR 218" ||
		repository.items[0].Protocol != "DCC" {
		t.Fatalf("work items = %#v, want normalized locomotive", repository.items)
	}
	if vehicles.listCalls != 1 || ecos.probeCalls != 1 {
		t.Fatalf("read calls: vehicles=%d ECoS=%d", vehicles.listCalls, ecos.probeCalls)
	}
}

func TestDigitalCenterWorkspaceReadPersistsFailedAndInterruptedStates(t *testing.T) {
	tests := []struct {
		name       string
		probeError error
		wantState  DigitalCenterSessionState
		wantCode   DigitalCenterMessageCode
	}{
		{name: "device failure", probeError: errors.New("dial failed"), wantState: DigitalCenterSessionFailed, wantCode: DigitalCenterMessageReadFailed},
		{name: "cancelled", probeError: context.Canceled, wantState: DigitalCenterSessionInterrupted, wantCode: DigitalCenterMessageConnectionInterrupted},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &workspaceRepositoryMemory{}
			service := NewDigitalCenterWorkspaceService(
				repository,
				&workspaceSettingsReaderStub{value: configuredECoSSettings()},
				&workspaceECoSReaderStub{err: test.probeError},
				nil,
				&workspaceVehicleReaderStub{},
				nil,
			)

			session, err := service.StartReadSession(t.Context(), "ecos", "admin-1")
			if err != nil {
				t.Fatal(err)
			}
			if session.State != test.wantState || len(repository.messages) != 1 ||
				repository.messages[0].Code != test.wantCode || repository.messages[0].NextAction == "" {
				t.Fatalf("session=%#v messages=%#v", session, repository.messages)
			}
		})
	}
}

func TestDigitalCenterWorkspaceRejectsUnboundedDeviceOutputBeforePersistence(t *testing.T) {
	repository := &workspaceRepositoryMemory{}
	service := NewDigitalCenterWorkspaceService(
		repository,
		&workspaceSettingsReaderStub{value: configuredECoSSettings()},
		&workspaceECoSReaderStub{probe: &ECoSRawProbe{Locomotives: []ECoSRawLocomotive{{
			ObjectID: 3, Name: "BR 218", Address: 70000, Protocol: "DCC",
		}}}},
		nil,
		&workspaceVehicleReaderStub{},
		nil,
	)

	session, err := service.StartReadSession(t.Context(), "ecos", "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if session.State != DigitalCenterSessionFailed || len(repository.items) != 0 {
		t.Fatalf("session=%#v items=%#v, want rejected device output", session, repository.items)
	}
}

func TestDigitalCenterWorkspaceFiltersAndBoundsPagedWorkItems(t *testing.T) {
	repository := &workspaceRepositoryMemory{}
	for index := 0; index < 125; index++ {
		status := DigitalCompareOK
		if index%2 == 0 {
			status = DigitalCompareDeviation
		}
		repository.items = append(repository.items, DigitalCenterWorkItem{
			ID: "item-" + string(rune(index+1000)), Name: "BR 218", Address: index + 1,
			Protocol: "DCC", CompareStatus: status,
		})
	}
	service := NewDigitalCenterWorkspaceService(repository, nil, nil, nil, nil, nil)

	page, err := service.ListWorkItems(t.Context(), "session-1", DigitalCenterWorkItemFilter{
		Query: "218", CompareStatus: DigitalCompareDeviation, Page: 1, PageSize: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.PageSize != 100 || len(page.Items) != 63 || page.Total != 63 || page.TotalPages != 1 {
		t.Fatalf("page = %#v", page)
	}
}

func TestDigitalCenterWorkspaceHandlesExcessivePageWithoutIntegerOverflow(t *testing.T) {
	repository := &workspaceRepositoryMemory{
		session: DigitalCenterReadSession{ID: "session-1"},
		items:   []DigitalCenterWorkItem{{ID: "item-1", Name: "BR 218"}},
	}
	service := NewDigitalCenterWorkspaceService(repository, nil, nil, nil, nil, nil)

	page, err := service.ListWorkItems(t.Context(), "session-1", DigitalCenterWorkItemFilter{
		Page: math.MaxInt, PageSize: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 || page.Total != 1 {
		t.Fatalf("page = %#v, want empty out-of-range page", page)
	}
}

func mappedDigitalCenterVehicle(id, provider, externalID, address, protocol string) Vehicle {
	return Vehicle{
		ID: id, Name: "BR 218", Digital: true, DigitalDecoderNumber: address,
		ExternalMappings: []VehicleExternalMap{{
			VehicleID: id, Provider: provider, ExternalID: externalID,
			ExternalAddress: address, ExternalProtocol: protocol,
		}},
	}
}

func configuredECoSSettings() DigitalCenterSettings {
	return DigitalCenterSettings{
		Provider: "ecos",
		ECoS:     DigitalProviderSettings{Enabled: true, Host: "ecos.local", Port: "15471"},
	}
}

type workspaceSettingsReaderStub struct {
	value DigitalCenterSettings
	err   error
}

func (stub *workspaceSettingsReaderStub) DigitalSettings(context.Context) (*DigitalCenterSettings, error) {
	value := stub.value
	return &value, stub.err
}

type workspaceECoSReaderStub struct {
	probe      *ECoSRawProbe
	err        error
	input      ECoSConnectionInput
	probeCalls int
}

func (stub *workspaceECoSReaderStub) ProbeLocomotiveRaw(
	_ context.Context,
	input ECoSConnectionInput,
) (*ECoSRawProbe, error) {
	stub.input = input
	stub.probeCalls++
	return stub.probe, stub.err
}

type workspaceVehicleReaderStub struct {
	vehicles  []Vehicle
	err       error
	listCalls int
}

func (stub *workspaceVehicleReaderStub) ListReadOnly(context.Context, string) ([]Vehicle, error) {
	stub.listCalls++
	return append([]Vehicle(nil), stub.vehicles...), stub.err
}

type workspaceRepositoryMemory struct {
	session  DigitalCenterReadSession
	items    []DigitalCenterWorkItem
	messages []DigitalCenterSessionMessage
}

func (repository *workspaceRepositoryMemory) CreateSession(
	_ context.Context,
	session DigitalCenterReadSession,
) (DigitalCenterReadSession, error) {
	session.ID = "session-1"
	session.CreatedAt = "2026-08-21T00:00:00Z"
	session.UpdatedAt = session.CreatedAt
	repository.session = session
	return session, nil
}

func (repository *workspaceRepositoryMemory) UpdateSession(_ context.Context, session DigitalCenterReadSession) error {
	repository.session = session
	return nil
}

func (repository *workspaceRepositoryMemory) GetSession(context.Context, string) (DigitalCenterReadSession, error) {
	return repository.session, nil
}

func (repository *workspaceRepositoryMemory) ReplaceWorkItems(
	_ context.Context,
	sessionID string,
	items []DigitalCenterWorkItem,
) error {
	repository.items = append([]DigitalCenterWorkItem(nil), items...)
	for index := range repository.items {
		repository.items[index].SessionID = sessionID
		repository.items[index].ID = "item-" + string(rune(index+1+'0'))
	}
	return nil
}

func (repository *workspaceRepositoryMemory) ListWorkItems(context.Context, string) ([]DigitalCenterWorkItem, error) {
	return append([]DigitalCenterWorkItem(nil), repository.items...), nil
}

func (repository *workspaceRepositoryMemory) GetWorkItem(
	_ context.Context,
	_ string,
	id string,
) (DigitalCenterWorkItem, error) {
	for _, item := range repository.items {
		if item.ID == id {
			return item, nil
		}
	}
	return DigitalCenterWorkItem{}, errors.New("not found")
}

func (repository *workspaceRepositoryMemory) UpdateWorkItem(
	_ context.Context,
	item DigitalCenterWorkItem,
) (DigitalCenterWorkItem, error) {
	for index := range repository.items {
		if repository.items[index].ID == item.ID && repository.items[index].SessionID == item.SessionID {
			repository.items[index] = item
			return item, nil
		}
	}
	return DigitalCenterWorkItem{}, errors.New("not found")
}

func (repository *workspaceRepositoryMemory) AddMessage(
	_ context.Context,
	message DigitalCenterSessionMessage,
) error {
	repository.messages = append(repository.messages, message)
	return nil
}

func (repository *workspaceRepositoryMemory) ListMessages(
	context.Context,
	string,
) ([]DigitalCenterSessionMessage, error) {
	return append([]DigitalCenterSessionMessage(nil), repository.messages...), nil
}

func (repository *workspaceRepositoryMemory) CreateWriteGrant(context.Context, DigitalCenterWriteGrant) error {
	return nil
}

func (repository *workspaceRepositoryMemory) ConsumeWriteGrant(
	context.Context,
	string,
	string,
) (DigitalCenterWriteGrant, error) {
	return DigitalCenterWriteGrant{}, nil
}

func TestNormalizeDigitalCenterProtocolAliases(t *testing.T) {
	tests := map[string]string{"DCC128": "DCC", "dcc-28": "DCC", "MM27": "MOTOROLA", "mfx": "MFX"}
	for input, want := range tests {
		got, err := normalizeDigitalCenterProtocol(input)
		if err != nil || got != want {
			t.Fatalf("normalize protocol %q = %q, %v; want %q", input, got, err, want)
		}
	}
}

func TestNormalizeDigitalCenterNameStripsInvisibleAndPrivateUnicode(t *testing.T) {
	got, err := normalizeDigitalCenterName("BR\u202e\u200b\ue000 218")
	if err != nil {
		t.Fatal(err)
	}
	if got != "BR 218" {
		t.Fatalf("normalized name = %q, want %q", got, "BR 218")
	}
}

func TestNormalizeDigitalCenterNameRejectsInvalidSurrogateEncoding(t *testing.T) {
	invalidUTF8 := string([]byte{0xed, 0xa0, 0x80})
	if _, err := normalizeDigitalCenterName(invalidUTF8); !errors.Is(err, ErrDigitalCenterDeviceOutput) {
		t.Fatalf("error = %v, want invalid device output", err)
	}
}
