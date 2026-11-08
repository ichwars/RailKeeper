package application_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/exhibition"
	"railkeeper/backend/internal/infrastructure/exhibitiondata"
)

func exhibitionPortFixture(t *testing.T) (*sql.DB, *application.ExhibitionService, application.ExhibitionList, application.ExhibitionEntry) {
	t.Helper()
	db := testDB(t)
	db.SetMaxOpenConns(1)
	service := application.NewExhibitionService(db)
	list, err := service.Create(t.Context(), application.ExhibitionListInput{Designation: "Preserved event", Date: "2026-11-08", EndDate: "2026-11-10"})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := service.CreateEntry(t.Context(), list.ID, exhibitionPortInput())
	if err != nil {
		t.Fatal(err)
	}
	list, err = service.Get(t.Context(), list.ID)
	if err != nil {
		t.Fatal(err)
	}
	return db, service, list, entry
}
func exhibitionPortInput() application.ExhibitionEntryInput {
	return application.ExhibitionEntryInput{Owner: "Stored owner", LocomotiveName: "Stored locomotive", Analog: true, DayScope: "day1,day3", ImageURL: "/api/images/synthetic", FunctionKeys: "F0", Notes: "Keep original notes"}
}

func TestExhibitionEntryLateFailuresRollbackAllRowsAndAllowSameRevisionRetry(t *testing.T) {
	for _, operation := range []string{"create", "update", "delete"} {
		for _, failure := range []string{"write", "locked", "commit"} {
			t.Run(operation+"/"+failure, func(t *testing.T) {
				db, service, list, entry := exhibitionPortFixture(t)
				if failure == "commit" {
					if _, err := db.Exec(`CREATE TABLE exhibition_commit_fixture(entry_id TEXT REFERENCES exhibition_entries(id) DEFERRABLE INITIALLY DEFERRED)`); err != nil {
						t.Fatal(err)
					}
				}
				before := databaseRowFingerprints(t, db)
				var trigger string
				switch failure {
				case "write":
					trigger = `CREATE TRIGGER fail_exhibition_port BEFORE UPDATE ON exhibition_lists BEGIN SELECT RAISE(FAIL,'synthetic late exhibition failure'); END`
				case "locked":
					event := map[string]string{"create": "INSERT", "update": "UPDATE", "delete": "DELETE"}[operation]
					row := "NEW"
					if operation == "delete" {
						row = "OLD"
					}
					trigger = `CREATE TRIGGER fail_exhibition_port AFTER ` + event + ` ON exhibition_entries BEGIN UPDATE exhibition_lists SET locked=1 WHERE id=` + row + `.list_id; END`
				case "commit":
					trigger = `CREATE TRIGGER fail_exhibition_port AFTER UPDATE ON exhibition_lists BEGIN INSERT INTO exhibition_commit_fixture(entry_id)VALUES('missing entry'); END`
				}
				if _, err := db.Exec(trigger); err != nil {
					t.Fatal(err)
				}
				input := exhibitionPortInput()
				input.Notes = "Replacement"
				input.ExpectedRevision = entry.Revision
				run := func() error {
					switch operation {
					case "create":
						_, err := service.CreateEntry(t.Context(), list.ID, input)
						return err
					case "update":
						_, err := service.UpdateEntry(t.Context(), list.ID, entry.ID, input)
						return err
					default:
						return service.DeleteEntry(t.Context(), list.ID, entry.ID)
					}
				}
				err := run()
				if err == nil || failure == "locked" && !errors.Is(err, application.ErrExhibitionLocked) {
					t.Fatal("late failure not returned", err)
				}
				if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) || db.Stats().InUse != 0 {
					t.Fatal("partial entry/revision/authentication rows or open transaction")
				}
				if _, err := db.Exec(`DROP TRIGGER fail_exhibition_port`); err != nil {
					t.Fatal(err)
				}
				if err := run(); err != nil {
					t.Fatal("same revision retry failed", err)
				}
				saved, err := service.Get(t.Context(), list.ID)
				if err != nil || saved.Revision != list.Revision+1 {
					t.Fatal("retry did not advance list exactly once", err)
				}
				count := 1
				if operation == "create" {
					count = 2
				}
				if operation == "delete" {
					count = 0
				}
				if len(saved.Entries) != count {
					t.Fatal("retry duplicated or lost entries", len(saved.Entries))
				}
				if operation == "update" && (saved.Entries[0].ID != entry.ID || saved.Entries[0].Notes != input.Notes || saved.Entries[0].Revision != entry.Revision+1) {
					t.Fatal("entry identity or revision changed")
				}
			})
		}
	}
}

func TestExhibitionExceptionLateRevisionFailureRollsBackExceptionAndAllowsRetry(t *testing.T) {
	for _, failure := range []string{"write", "stale", "commit"} {
		t.Run(failure, func(t *testing.T) {
			db, service, list, _ := exhibitionPortFixture(t)
			if _, err := service.CreateEntry(t.Context(), list.ID, application.ExhibitionEntryInput{LocomotiveName: "Missing details"}); err != nil {
				t.Fatal(err)
			}
			workspace, err := service.Workspace(t.Context(), list.ID)
			if err != nil || len(workspace.Conflicts) != 1 {
				t.Fatal("missing conflict fixture", err)
			}
			if failure == "commit" {
				if _, err := db.Exec(`CREATE TABLE exhibition_commit_fixture(entry_id TEXT REFERENCES exhibition_entries(id) DEFERRABLE INITIALLY DEFERRED)`); err != nil {
					t.Fatal(err)
				}
			}
			before := databaseRowFingerprints(t, db)
			trigger := `CREATE TRIGGER fail_exhibition_exception BEFORE UPDATE ON exhibition_lists BEGIN SELECT RAISE(FAIL,'synthetic late exception failure'); END`
			if failure == "stale" {
				trigger = `CREATE TRIGGER fail_exhibition_exception AFTER INSERT ON exhibition_conflict_exceptions BEGIN UPDATE exhibition_lists SET revision=revision+1 WHERE id=NEW.list_id; END`
			}
			if failure == "commit" {
				trigger = `CREATE TRIGGER fail_exhibition_exception AFTER UPDATE ON exhibition_lists BEGIN INSERT INTO exhibition_commit_fixture(entry_id)VALUES('missing entry'); END`
			}
			if _, err := db.Exec(trigger); err != nil {
				t.Fatal(err)
			}
			input := application.ExhibitionConflictExceptionInput{Reason: "  Confirmed exception  ", ExpectedRevision: workspace.List.Revision}
			if _, err := service.SetConflictException(t.Context(), list.ID, workspace.Conflicts[0].ID, input); err == nil || failure == "stale" && !errors.Is(err, application.ErrExhibitionStale) {
				t.Fatal("late exception failure not observed", err)
			}
			if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) || db.Stats().InUse != 0 {
				t.Fatal("failed exception changed rows or retained transaction")
			}
			if _, err := db.Exec(`DROP TRIGGER fail_exhibition_exception`); err != nil {
				t.Fatal(err)
			}
			saved, err := service.SetConflictException(t.Context(), list.ID, workspace.Conflicts[0].ID, input)
			if err != nil || saved.List.Revision != workspace.List.Revision+1 || !saved.Conflicts[0].Excepted || saved.Conflicts[0].ExceptionReason != "Confirmed exception" {
				t.Fatal("same revision exception retry failed", err)
			}
			if saved.Summary.ConflictCount != 1 || saved.Readiness.Problems != 1 {
				t.Fatal("excepted conflict disappeared from existing readiness semantics")
			}
		})
	}
}

type exhibitionReadbackFault struct {
	exhibition.Repository
	fail bool
}
type exhibitionCommitReadbackFault struct {
	exhibition.Mutation
	owner *exhibitionReadbackFault
}

func (r *exhibitionReadbackFault) Begin(ctx context.Context) (exhibition.Mutation, error) {
	tx, err := r.Repository.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &exhibitionCommitReadbackFault{Mutation: tx, owner: r}, nil
}
func (m *exhibitionCommitReadbackFault) Commit() error {
	err := m.Mutation.Commit()
	if err == nil {
		m.owner.fail = true
	}
	return err
}
func (r *exhibitionReadbackFault) Entry(ctx context.Context, listID, id string) (exhibition.ExhibitionEntry, error) {
	if r.fail {
		r.fail = false
		return exhibition.ExhibitionEntry{}, errors.New("synthetic lost entry readback")
	}
	return r.Repository.Entry(ctx, listID, id)
}
func TestExhibitionLostReadbackPreservesCommittedEntryAndRejectsStaleRetry(t *testing.T) {
	db, normal, list, entry := exhibitionPortFixture(t)
	service := exhibition.New(&exhibitionReadbackFault{Repository: exhibitiondata.New(db)})
	input := exhibitionPortInput()
	input.ExpectedRevision = entry.Revision
	input.Notes = "Already saved"
	if _, err := service.UpdateEntry(t.Context(), list.ID, entry.ID, input); err == nil || !strings.Contains(err.Error(), "lost entry readback") {
		t.Fatal("readback loss not surfaced", err)
	}
	saved, err := normal.Get(t.Context(), list.ID)
	if err != nil || len(saved.Entries) != 1 || saved.Entries[0].Notes != input.Notes || saved.Entries[0].Revision != entry.Revision+1 || saved.Revision != list.Revision+1 {
		t.Fatal("readback failure undid commit", err)
	}
	before := databaseRowFingerprints(t, db)
	if _, err := service.UpdateEntry(t.Context(), list.ID, entry.ID, input); !errors.Is(err, application.ErrExhibitionStale) {
		t.Fatal("lost response retry overwrote newer data", err)
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
		t.Fatal("stale retry changed rows")
	}
	input.ExpectedRevision = saved.Entries[0].Revision
	if _, err := normal.UpdateEntry(t.Context(), list.ID, entry.ID, input); err != nil {
		t.Fatal("explicit refreshed revision retry failed", err)
	}
}

func TestExhibitionRawLegacyFieldsAndStickyLifecycleTimesRemainCompatible(t *testing.T) {
	db, service, list, entry := exhibitionPortFixture(t)
	if _, err := db.Exec(`UPDATE exhibition_entries SET id='fixed-entry',owner='  Stored Owner  ',vehicle_id='Vehicle-A',image_url=NULL,decoder_number=NULL,function_keys=NULL,notes=NULL,day_scope='',dt_decoder=2,analog=2 WHERE id=?`, entry.ID); err != nil {
		t.Fatal(err)
	}
	workspace, err := service.Workspace(t.Context(), list.ID)
	if err != nil || len(workspace.Entries) != 1 || len(workspace.Conflicts) != 1 {
		t.Fatal("legacy fixture not readable", err)
	}
	stored := workspace.Entries[0]
	if stored.Owner != "  Stored Owner  " || stored.DayScope != "" || stored.DTDecoder || stored.Analog || stored.ImageURL != "" || stored.DecoderNumber != "" || stored.FunctionKeys != "" || stored.Notes != "" {
		t.Fatal("raw fields were normalized while reading")
	}
	conflict := workspace.Conflicts[0]
	// Fixed legacy conflict identity calculated from the original ordered preimage.
	if conflict.ID != "conflict-6983dcb5ea0a6785cf27dde8" || !reflect.DeepEqual(conflict.Fields, []string{"dayScope", "decoderNumber", "interfaceName"}) {
		t.Fatal("legacy conflict identity changed", conflict.ID, conflict.Fields)
	}
	body, err := json.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["status"] != "missing" {
		t.Fatal("workspace status was shadowed by embedded legacy entry status")
	}
	if _, err := db.Exec(`UPDATE exhibition_lists SET locked_at='old locked time',completed_at='old completed time',archived_at='old archived time' WHERE id=?`, list.ID); err != nil {
		t.Fatal(err)
	}
	for _, status := range []application.ExhibitionStatus{application.ExhibitionStatusLocked, application.ExhibitionStatusRunning, application.ExhibitionStatusCompleted, application.ExhibitionStatusArchived, application.ExhibitionStatusOpen} {
		saved, err := service.SetStatus(t.Context(), list.ID, application.ExhibitionStatusInput{Status: status, ConfirmConflicts: true, Reason: "Confirmed conflicts"})
		if err != nil || saved.LockedAt != "old locked time" || saved.CompletedAt != "old completed time" || saved.ArchivedAt != "old archived time" {
			t.Fatal("sticky lifecycle time changed", status, err)
		}
	}
}
