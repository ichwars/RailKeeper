package valuationdata_test

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/internal/infrastructure/valuationdata"
	"railkeeper/backend/internal/valuation"
)

func fixture(t *testing.T) (*sql.DB, *valuation.Service, string, string) {
	t.Helper()
	db, err := infrastructure.OpenSQLite(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.SeedRoles(db); err != nil {
		t.Fatal(err)
	}
	user, err := application.NewAuthService(db).CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "valuation-editor", Password: "synthetic-only-password", Roles: []string{"Editor"},
	})
	if err != nil {
		t.Fatal(err)
	}
	vehicle, err := application.NewVehicleService(db).Create(t.Context(), application.CreateVehicleInput{
		Name: "Valuation fixture", Manufacturer: "Synthetic", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
		ListPrice: "99.00", PurchasePrice: "80.00",
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	clock := func() time.Time { return time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC) }
	return db, valuation.New(valuationdata.NewStore(db), clock), user.ID, vehicle.ID
}

func input(id int) valuation.Input {
	return valuation.Input{ID: fmt.Sprintf("%032x", id), Amount: "100.123456", Currency: "EUR", ValuedOn: "2026-10-01", Source: "Manual estimate"}
}

func TestImmutableCorrectionsAndIdenticalReplay(t *testing.T) {
	db, service, actor, vehicle := fixture(t)
	first, err := service.Record(t.Context(), vehicle, input(1), actor)
	if err != nil {
		t.Fatal(err)
	}
	correction := input(2)
	correction.ReplacesID, correction.Amount = first.ID, "125.00"
	second, err := service.Record(t.Context(), vehicle, correction, actor)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []valuation.Record{first, second} {
		replayed, err := service.Record(t.Context(), vehicle, record.Input, actor)
		if err != nil || replayed.RecordedAt != record.RecordedAt {
			t.Fatalf("replay: %+v %v", replayed, err)
		}
	}
	changed := correction
	changed.Amount = "300.00"
	if _, err := service.Record(t.Context(), vehicle, changed, actor); !errors.Is(err, valuation.ErrConflict) {
		t.Fatalf("identity reuse: %v", err)
	}
	history, err := service.History(t.Context(), vehicle, 0)
	if err != nil || history.Total != 2 {
		t.Fatalf("history: %+v %v", history, err)
	}
	if history.Entries[1].Amount != first.Amount || history.Entries[1].ReplacedByID != second.ID {
		t.Fatal("original was modified")
	}
	summary, err := service.Summary(t.Context(), "2026-10-01")
	if err != nil || summary.Totals[0].Amount != second.Amount {
		t.Fatalf("correction summary: %+v %v", summary, err)
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE action='RecordValuation'`).Scan(&audits); err != nil || audits != 2 {
		t.Fatalf("duplicate audit: %d %v", audits, err)
	}
	var list, purchase string
	if err := db.QueryRow(`SELECT list_price,purchase_price FROM vehicles WHERE id=?`, vehicle).Scan(&list, &purchase); err != nil || list != "99.00" || purchase != "80.00" {
		t.Fatal("existing prices changed")
	}
}

func TestParallelCorrectionsCommitExactlyOne(t *testing.T) {
	db, service, actor, vehicle := fixture(t)
	first, err := service.Record(t.Context(), vehicle, input(1), actor)
	if err != nil {
		t.Fatal(err)
	}
	const workers = 12
	results := make(chan error, workers)
	var group sync.WaitGroup
	for i := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			correction := input(i + 2)
			correction.ReplacesID = first.ID
			_, err := service.Record(t.Context(), vehicle, correction, actor)
			results <- err
		}()
	}
	group.Wait()
	close(results)
	var committed int
	for err := range results {
		if err == nil {
			committed++
		} else if !errors.Is(err, valuation.ErrConflict) {
			t.Errorf("unexpected result: %v", err)
		}
	}
	var rows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_valuations`).Scan(&rows); err != nil || committed != 1 || rows != 2 {
		t.Fatalf("forked corrections: committed=%d rows=%d err=%v", committed, rows, err)
	}
}

func TestAuditFailureRollsBackAndWithdrawalRetainsHistory(t *testing.T) {
	db, service, actor, vehicle := fixture(t)
	first, err := service.Record(t.Context(), vehicle, input(1), actor)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER reject_valuation_audit BEFORE INSERT ON audit_logs WHEN NEW.action='RecordValuation' BEGIN SELECT RAISE(ABORT,'synthetic audit failure'); END`); err != nil {
		t.Fatal(err)
	}
	withdrawal := input(2)
	withdrawal.ReplacesID, withdrawal.Withdrawn, withdrawal.Source = first.ID, true, "Entered for the wrong item"
	if _, err := service.Record(t.Context(), vehicle, withdrawal, actor); err == nil {
		t.Fatal("audit failure ignored")
	}
	history, err := service.History(t.Context(), vehicle, 0)
	if err != nil || history.Total != 1 || history.Entries[0].ReplacedByID != "" {
		t.Fatal("partial correction survived failed transaction")
	}
	if _, err := db.Exec(`DROP TRIGGER reject_valuation_audit`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Record(t.Context(), vehicle, withdrawal, actor); err != nil {
		t.Fatal(err)
	}
	summary, err := service.Summary(t.Context(), "2026-11-07")
	if err != nil || summary.ValuedVehicles != 0 || summary.MissingVehicles != 1 {
		t.Fatalf("withdrawal counted: %+v %v", summary, err)
	}
	exported, err := service.Export(t.Context(), vehicle)
	if err != nil || len(exported.Entries) != 2 || exported.Entries[1].Amount != first.Amount {
		t.Fatal("export lost history")
	}
}

func TestHistoryPaginationExportAndVehicleDeletion(t *testing.T) {
	db, service, actor, vehicle := fixture(t)
	for i := range 103 {
		if _, err := service.Record(t.Context(), vehicle, input(i+1), actor); err != nil {
			t.Fatal(err)
		}
	}
	page, err := service.History(t.Context(), vehicle, 0)
	if err != nil || len(page.Entries) != 100 || page.Total != 103 {
		t.Fatalf("first page: %+v %v", page, err)
	}
	last, err := service.History(t.Context(), vehicle, 100)
	if err != nil || len(last.Entries) != 3 || last.Total != 103 {
		t.Fatalf("last page: %+v %v", last, err)
	}
	all, err := service.Export(t.Context(), vehicle)
	if err != nil || len(all.Entries) != 103 {
		t.Fatal("export truncated at page size")
	}
	correction := input(104)
	correction.ReplacesID = input(1).ID
	if _, err := service.Record(t.Context(), vehicle, correction, actor); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM vehicles WHERE id=?`, vehicle); err != nil {
		t.Fatalf("new reference prevents existing vehicle deletion: %v", err)
	}
	var records int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_valuations`).Scan(&records); err != nil || records != 0 {
		t.Fatal("orphaned valuation records")
	}
	if _, err := service.History(t.Context(), vehicle, 0); !errors.Is(err, valuation.ErrNotFound) {
		t.Fatal("deleted vehicle did not return not found")
	}
	if _, err := service.Record(t.Context(), vehicle, input(105), actor); !errors.Is(err, valuation.ErrNotFound) {
		t.Fatal("missing vehicle accepted")
	}
}

func TestCorrectionsCannotMoveBetweenVehicles(t *testing.T) {
	db, service, actor, vehicle := fixture(t)
	first, err := service.Record(t.Context(), vehicle, input(1), actor)
	if err != nil {
		t.Fatal(err)
	}
	correction := input(2)
	correction.ReplacesID = first.ID
	other, err := application.NewVehicleService(db).Create(t.Context(), application.CreateVehicleInput{
		Name: "Other vehicle", Manufacturer: "Synthetic", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
	}, actor)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Record(t.Context(), other.ID, correction, actor); !errors.Is(err, valuation.ErrNotFound) {
		t.Fatal("foreign predecessor accepted")
	}
	correction.ReplacesID = input(50).ID
	if _, err := service.Record(t.Context(), vehicle, correction, actor); !errors.Is(err, valuation.ErrNotFound) {
		t.Fatal("unknown predecessor accepted")
	}
}
