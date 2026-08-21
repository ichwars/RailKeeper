package application_test

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDigitalCenterReadSessionLeavesEveryVehicleRowUnchanged(t *testing.T) {
	db := testDB(t)
	vehicles := application.NewVehicleService(db)
	exhibitions := application.NewExhibitionService(db)
	settings := application.NewSettingsService(db)
	ctx := t.Context()

	vehicle, err := vehicles.Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Piko", Name: "BR 218", Gauge: "H0", Category: "Lokomotive",
		Gattung: "Diesellok", Digital: true, DigitalDecoderNumber: "3", Exhibition: true,
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	expired, err := exhibitions.Create(ctx, application.ExhibitionListInput{
		Designation: "Vergangene Messe", Date: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exhibitions.CreateEntry(ctx, expired.ID, application.ExhibitionEntryInput{
		VehicleID: vehicle.ID, Owner: "Test", LocomotiveName: vehicle.Name, DecoderNumber: "3",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := settings.UpdateDigitalSettings(ctx, application.DigitalCenterSettings{
		Provider: "ecos",
		ECoS:     application.DigitalProviderSettings{Enabled: true, Host: "ecos.local", Port: "15471"},
	}); err != nil {
		t.Fatal(err)
	}

	before := snapshotVehicleRows(t, db)
	workspace := application.NewDigitalCenterWorkspaceService(
		infrastructure.NewDigitalCenterWorkspaceRepository(db), settings,
		&readOnlyECoSStub{probe: application.ECoSRawProbe{Locomotives: []application.ECoSRawLocomotive{{
			ObjectID: 3, Name: "BR 218", Address: 3, Protocol: "DCC128",
		}}}}, nil, vehicles, nil,
	)
	if _, err := workspace.StartReadSession(ctx, "ecos", "admin-1"); err != nil {
		t.Fatal(err)
	}
	after := snapshotVehicleRows(t, db)

	if !reflect.DeepEqual(after, before) {
		t.Fatalf("digital center read mutated vehicle rows:\nbefore=%#v\nafter=%#v", before, after)
	}
}

func snapshotVehicleRows(t *testing.T, db *sql.DB) [][]string {
	t.Helper()
	rows, err := db.Query(`SELECT * FROM vehicles ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	result := [][]string{}
	for rows.Next() {
		values := make([]any, len(columns))
		destinations := make([]any, len(columns))
		for index := range values {
			destinations[index] = &values[index]
		}
		if err := rows.Scan(destinations...); err != nil {
			t.Fatal(err)
		}
		serialized := make([]string, len(values))
		for index, value := range values {
			serialized[index] = fmt.Sprintf("%T:%v", value, value)
		}
		result = append(result, serialized)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

type readOnlyECoSStub struct {
	probe application.ECoSRawProbe
}

func (stub *readOnlyECoSStub) ProbeLocomotiveRaw(
	context.Context,
	application.ECoSConnectionInput,
) (*application.ECoSRawProbe, error) {
	probe := stub.probe
	return &probe, nil
}
