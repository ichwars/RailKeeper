package application

import (
	"context"
	"errors"
	"strconv"
	"testing"
)

func TestDigitalCenterWritePreviewBlocksAddressOwnedByAnotherObject(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	targetObjectID, err := strconv.Atoi(fixture.item.CenterObjectID)
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.locomotives = []ECoSLocomotive{
		{ObjectID: targetObjectID, Name: "Target", Address: 3, Protocol: "DCC"},
		{ObjectID: 2002, Name: "Other", Address: 18, Protocol: "DCC"},
	}

	_, err = fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"address"}}, "admin-1")
	var conflict *DigitalCenterAddressConflictError
	if !errors.As(err, &conflict) || conflict.ObjectID != 2002 || conflict.Name != "Other" {
		t.Fatalf("error=%#v", err)
	}
}

func TestDigitalCenterWritePreviewAllowsNameOnlyWithExistingAddressCollision(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	targetObjectID, err := strconv.Atoi(fixture.item.CenterObjectID)
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.locomotives = []ECoSLocomotive{
		{ObjectID: targetObjectID, Name: "Target", Address: 3, Protocol: "DCC"},
		{ObjectID: 2002, Name: "Other", Address: 3, Protocol: "DCC"},
	}

	_, err = fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDigitalCenterWritePreviewFailsClosedWhenTargetIsMissingFromMasterList(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.locomotives = []ECoSLocomotive{
		{ObjectID: 2002, Name: "Other", Address: 99, Protocol: "DCC"},
	}

	_, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"address"}}, "admin-1")
	if !errors.Is(err, ErrDigitalCenterDeviceOutput) {
		t.Fatalf("error=%v", err)
	}
}

func TestDigitalCenterCreatePreviewRejectsReappearedPreviousObject(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	fixture.item.CenterObjectID = "42"
	fixture.item.Address = 4405
	fixture.item.CompareStatus = DigitalCompareMissing
	fixture.item.StationStatus = "missing"
	fixture.item.Center = map[string]any{}
	fixture.item.RailKeeper = map[string]any{
		"vehicleId": "vehicle-1", "name": "BR 18", "decoderAddress": 4405, "protocol": "DCC",
	}
	fixture.repository.item = fixture.item
	fixture.ecos.locomotives = []ECoSLocomotive{
		{ObjectID: 42, Name: "Reappeared", Address: 99, Protocol: "DCC"},
	}

	_, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Operation: DigitalCenterWriteCreate,
			Fields: []string{"name", "address", "protocol"}}, "admin-1")
	if !errors.Is(err, ErrDigitalCenterPreviewStale) {
		t.Fatalf("error=%v", err)
	}
}

func (stub *digitalCenterWriteECoSStub) ListLocomotives(
	context.Context,
	ECoSConnectionInput,
) ([]ECoSLocomotive, error) {
	stub.events = append(stub.events, "list")
	if len(stub.locomotives) == 0 {
		return []ECoSLocomotive{{ObjectID: 3, Name: stub.currentName, Address: 3, Protocol: "DCC"}}, nil
	}
	return append([]ECoSLocomotive(nil), stub.locomotives...), nil
}
