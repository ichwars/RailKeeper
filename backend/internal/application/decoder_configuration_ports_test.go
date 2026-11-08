package application_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationFinalReadbackFailureRollsBackInstanceAuditAndVehicle(t *testing.T) {
	db, _, original := decoderConfigurationFixture(t)
	db.SetMaxOpenConns(1)
	service := application.NewVehicleService(db)
	vehicleID := original.Tables["vehicles"][0]["id"].(string)
	context, err := service.DecoderConfigurationContext(t.Context(), vehicleID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_decoder_context_readback AFTER INSERT ON audit_logs
WHEN NEW.action='DecoderInstanceCreated'
BEGIN DELETE FROM vehicles WHERE id=json_extract(NEW.details_json,'$.vehicleId'); END`); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	input := application.DecoderInstanceInput{ID: strings.Repeat("7", 32), ExpectedRevision: context.Revision, Label: "must roll back"}
	result, err := service.CreateDecoderInstance(t.Context(), vehicleID, input, "")
	if result != nil || !errors.Is(err, application.ErrDecoderNotFound) {
		t.Fatal("failed final readback acknowledged", result, err)
	}
	if after := databaseRowFingerprints(t, db); !reflect.DeepEqual(before, after) {
		t.Fatal("readback failure left instance, audit, vehicle or cascaded evidence changed")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_decoder_context_readback`); err != nil {
		t.Fatal(err)
	}
	retry, err := service.CreateDecoderInstance(t.Context(), vehicleID, input, "")
	if err != nil || retry == nil || retry.Revision == context.Revision {
		t.Fatal("retry after complete rollback failed", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM decoder_instances WHERE id=?`, input.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("retry did not retain exactly one instance", count, err)
	}
}
