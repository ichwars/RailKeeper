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
