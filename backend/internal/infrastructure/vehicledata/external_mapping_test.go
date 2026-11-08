package vehicledata_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func externalMappingFixture(t *testing.T) (*sql.DB, vehiclemodel.ExternalMappingRepository) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE vehicle_external_mappings(
 id TEXT PRIMARY KEY,vehicle_id TEXT,provider TEXT,external_id TEXT,external_name TEXT,external_address TEXT,
 external_protocol TEXT,sync_status TEXT,last_seen_at TEXT,created_at TEXT,updated_at TEXT,UNIQUE(provider,external_id))`); err != nil {
		t.Fatal(err)
	}
	return db, vehicledata.NewExternalMappingStore(db)
}
func mappingInput(id string) vehiclemodel.VehicleExternalMapInput {
	return vehiclemodel.VehicleExternalMapInput{Provider: "ecos", ExternalID: id, ExternalName: "recorded " + id, SyncStatus: "linked"}
}
func externalMappingRows(t *testing.T, db *sql.DB) [][]any {
	t.Helper()
	rows, err := db.Query(`SELECT * FROM vehicle_external_mappings ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	out := [][]any{}
	for rows.Next() {
		values := make([]any, 11)
		targets := make([]any, 11)
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			t.Fatal(err)
		}
		out = append(out, values)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}
func TestExternalMappingConsolidationPreservesAllRowsOnFailure(t *testing.T) {
	for _, failure := range []string{"second write fails", "target belongs to another vehicle"} {
		t.Run(failure, func(t *testing.T) {
			db, repository := externalMappingFixture(t)
			for _, id := range []string{"previous", "target"} {
				owner := "owner"
				if id == "target" && failure == "target belongs to another vehicle" {
					owner = "other"
				}
				if _, err := repository.WriteMapping(t.Context(), vehiclemodel.NewExternalMappingWrite(owner, mappingInput(id), "recorded "+id)); err != nil {
					t.Fatal(err)
				}
			}
			before := externalMappingRows(t, db)
			if failure == "second write fails" {
				if _, err := db.Exec(`CREATE TRIGGER fail_target_update BEFORE UPDATE ON vehicle_external_mappings
WHEN NEW.external_id='target' BEGIN SELECT RAISE(ABORT,'synthetic consolidation failure'); END`); err != nil {
					t.Fatal(err)
				}
			}
			service := vehiclemodel.NewExternalMappingService(repository, nil)
			result, err := service.RebindExternalMapping(t.Context(), "owner", "previous", mappingInput("target"), "actor")
			if err == nil || result != nil || (failure == "target belongs to another vehicle" && !errors.Is(err, vehiclemodel.ErrVehicleExternalMappingConflict)) {
				t.Fatal("failed consolidation acknowledged", result, err)
			}
			if after := externalMappingRows(t, db); !reflect.DeepEqual(before, after) {
				t.Fatal("failure changed original bytes/NULLs/owner/IDs/times", before, after)
			}
		})
	}
}
func TestExternalMappingConsolidationKeepsExistingTargetIdentity(t *testing.T) {
	_, repository := externalMappingFixture(t)
	for _, id := range []string{"previous", "target"} {
		if _, err := repository.WriteMapping(t.Context(), vehiclemodel.NewExternalMappingWrite("owner", mappingInput(id), "recorded "+id)); err != nil {
			t.Fatal(err)
		}
	}
	before, err := repository.Mapping(t.Context(), "ecos", "target")
	if err != nil {
		t.Fatal(err)
	}
	service := vehiclemodel.NewExternalMappingService(repository, nil)
	input := mappingInput(" target ")
	input.Provider = " ECoS "
	input.ExternalName = " renamed "
	after, err := service.RebindExternalMapping(t.Context(), " owner ", " previous ", input, "ignored actor")
	if err != nil {
		t.Fatal(err)
	}
	if after.ID != before.ID || after.CreatedAt != before.CreatedAt || after.VehicleID != before.VehicleID ||
		after.ExternalID != "target" || after.ExternalName != "renamed" || after.ExternalAddress != "" || after.ExternalProtocol != "" {
		t.Fatal("consolidation changed retained identity", before, after)
	}
	if _, found, err := repository.FindMapping(t.Context(), "ecos", "previous"); err != nil || found {
		t.Fatal("previous mapping remains", found, err)
	}
}
func TestExternalMappingCallerTransactionCannotStealAnotherOwner(t *testing.T) {
	db, repository := externalMappingFixture(t)
	if _, err := repository.WriteMapping(t.Context(), vehiclemodel.NewExternalMappingWrite("first-owner", mappingInput("target"), "recorded")); err != nil {
		t.Fatal(err)
	}
	before := externalMappingRows(t, db)
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	mapping, err := vehicledata.WriteExternalMapping(t.Context(), tx, vehiclemodel.NewExternalMappingWrite("other-owner", mappingInput("target"), "must not replace"))
	if mapping != nil || !errors.Is(err, vehiclemodel.ErrVehicleExternalMappingConflict) {
		t.Fatal("owner predicate lost", mapping, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal("writer closed caller transaction", err)
	}
	if after := externalMappingRows(t, db); !reflect.DeepEqual(before, after) {
		t.Fatal(before, after)
	}
}
func TestExternalMappingRenameAndUpsertRetainIdentityAndParentRules(t *testing.T) {
	_, repository := externalMappingFixture(t)
	calls := 0
	service := vehiclemodel.NewExternalMappingService(repository, func(_ context.Context, id string) error {
		calls++
		if id != "owner" {
			t.Fatal(id)
		}
		return nil
	})
	before, err := service.UpsertExternalMapping(t.Context(), "owner", mappingInput("previous"), "actor")
	if err != nil {
		t.Fatal(err)
	}
	after, err := service.RebindExternalMapping(t.Context(), "owner", "previous", mappingInput("new"), "actor")
	if err != nil || after.ID != before.ID || after.CreatedAt != before.CreatedAt || after.ExternalID != "new" || calls != 1 {
		t.Fatal(before, after, err, calls)
	}
}
