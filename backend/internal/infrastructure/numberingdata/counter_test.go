package numberingdata

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"railkeeper/backend/internal/inventory/numbering"

	_ "modernc.org/sqlite"
)

func counterDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`CREATE TABLE inventory_number_schemes (
id TEXT PRIMARY KEY, category TEXT UNIQUE, prefix TEXT, next_number INTEGER,
padding INTEGER, active INTEGER, created_at TEXT, updated_at TEXT);
INSERT INTO inventory_number_schemes VALUES (
'recorded-id','Fahrzeug','LEGACY',17,4,1,'2020-01-02T03:04:05Z','2020-01-02T03:04:05Z');`); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCounterAdvancementBelongsToCallerCommit(t *testing.T) {
	db := counterDatabase(t)
	for _, commit := range []bool{false, true} {
		tx, err := db.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}
		value, err := numbering.Reserve(t.Context(), NewCounter(tx), "Fahrzeug", "", func(string) error { return nil })
		if err != nil || value != "LEGACY-0017" {
			_ = tx.Rollback()
			t.Fatal("reservation changed the recorded scheme", value, err)
		}
		if commit {
			err = tx.Commit()
		} else {
			err = tx.Rollback()
		}
		if err != nil {
			t.Fatal("counter adapter took ownership of the caller transaction", err)
		}
		var id, prefix, created, updated string
		var next, padding int
		if err := db.QueryRow(`SELECT id,prefix,next_number,padding,created_at,updated_at FROM inventory_number_schemes WHERE category='Fahrzeug'`).Scan(
			&id, &prefix, &next, &padding, &created, &updated); err != nil {
			t.Fatal(err)
		}
		wantNext := 17
		if commit {
			wantNext = 18
		}
		if id != "recorded-id" || prefix != "LEGACY" || next != wantNext || padding != 4 || created != "2020-01-02T03:04:05Z" ||
			!commit && updated != "2020-01-02T03:04:05Z" {
			t.Fatal("caller outcome did not preserve numbering state", id, prefix, next, padding, created, updated)
		}
	}
}

func TestCounterWriteFailureNeverAcknowledgesNumber(t *testing.T) {
	db := counterDatabase(t)
	if _, err := db.Exec(`CREATE TRIGGER reject_counter BEFORE UPDATE ON inventory_number_schemes
BEGIN SELECT RAISE(ABORT, 'synthetic counter rejection'); END`); err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	value, err := numbering.Reserve(t.Context(), NewCounter(tx), "Fahrzeug", "", func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "synthetic counter rejection") || value != "" {
		_ = tx.Rollback()
		t.Fatal("failed counter write acknowledged a reservation", value, err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var next int
	if err := db.QueryRow(`SELECT next_number FROM inventory_number_schemes`).Scan(&next); err != nil || next != 17 {
		t.Fatal("failed advancement altered the recorded counter", next, err)
	}
}

func TestSchemeUpdateRetainsIdentityAndRejectsDuplicates(t *testing.T) {
	db := counterDatabase(t)
	service := numbering.New(New(db))
	input := numbering.SchemeInput{Prefix: "New Prefix", NextNumber: 42, Padding: 5, Active: true}
	value, err := service.Update(t.Context(), "Fahrzeug", input)
	if err != nil || value.ID != "recorded-id" || value.CreatedAt != "2020-01-02T03:04:05Z" ||
		value.Prefix != "NEW-PREFIX" || value.Preview != "NEW-PREFIX-00042" {
		t.Fatal("scheme update changed its identity or counter", value, err)
	}
	if _, err := service.Create(t.Context(), numbering.CreateInput{Category: "Fahrzeug", InventoryNumberSchemeInput: input}); !errors.Is(err, numbering.ErrConflict) {
		t.Fatal("duplicate scheme overwrote an existing category", err)
	}
	if _, err := service.Update(t.Context(), "missing", input); !errors.Is(err, numbering.ErrNotFound) {
		t.Fatal("missing scheme was acknowledged as updated", err)
	}
}
