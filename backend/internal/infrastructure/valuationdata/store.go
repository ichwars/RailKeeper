// Package valuationdata owns SQLite transactions for manual vehicle valuations.
package valuationdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/valuation"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

var _ valuation.Store = (*Store)(nil)

const recordColumns = `v.id,v.amount,v.currency,v.valued_on,v.source,v.notes,
COALESCE(v.replaces_id,''),v.withdrawn,v.vehicle_id,v.recorded_at,
COALESCE((SELECT successor.id FROM vehicle_valuations successor WHERE successor.replaces_id=v.id),'')`

type scanner interface{ Scan(...any) error }

func scanRecord(row scanner) (valuation.Record, error) {
	var r valuation.Record
	err := row.Scan(&r.ID, &r.Amount, &r.Currency, &r.ValuedOn, &r.Source, &r.Notes,
		&r.ReplacesID, &r.Withdrawn, &r.VehicleID, &r.RecordedAt, &r.ReplacedByID)
	return r, err
}

func vehicleExists(ctx context.Context, tx *sql.Tx, id string) error {
	var found string
	err := tx.QueryRowContext(ctx, `SELECT id FROM vehicles WHERE id=?`, id).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return valuation.ErrNotFound
	}
	return err
}

func (s *Store) History(ctx context.Context, vehicleID string, offset int) (valuation.History, error) {
	return s.history(ctx, vehicleID, historyPage{offset: offset, limit: valuation.HistoryPageSize})
}

func (s *Store) Export(ctx context.Context, vehicleID string) (valuation.History, error) {
	return s.history(ctx, vehicleID, historyPage{})
}

type historyPage struct{ offset, limit int }

func (s *Store) history(ctx context.Context, vehicleID string, page historyPage) (valuation.History, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return valuation.History{}, fmt.Errorf("begin valuation history: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := vehicleExists(ctx, tx, vehicleID); err != nil {
		return valuation.History{}, err
	}
	result := valuation.History{Entries: []valuation.Record{}}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_valuations WHERE vehicle_id=?`, vehicleID).Scan(&result.Total); err != nil {
		return valuation.History{}, fmt.Errorf("count valuation history: %w", err)
	}
	query := `SELECT ` + recordColumns + ` FROM vehicle_valuations v WHERE vehicle_id=? ORDER BY valued_on DESC,recorded_at DESC,id DESC`
	args := []any{vehicleID}
	if page.limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, page.limit, page.offset)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return valuation.History{}, fmt.Errorf("list valuation history: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return valuation.History{}, fmt.Errorf("read valuation history: %w", err)
		}
		result.Entries = append(result.Entries, r)
	}
	if err := rows.Err(); err != nil {
		return valuation.History{}, err
	}
	if err := rows.Close(); err != nil {
		return valuation.History{}, err
	}
	if err := tx.Commit(); err != nil {
		return valuation.History{}, fmt.Errorf("commit valuation history: %w", err)
	}
	return result, nil
}

func (s *Store) Save(ctx context.Context, record valuation.Record, actor string) (valuation.Record, error) {
	if err := valuation.ValidateRecord(record); err != nil {
		return valuation.Record{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return valuation.Record{}, fmt.Errorf("begin valuation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	existing, err := scanRecord(tx.QueryRowContext(ctx, `SELECT `+recordColumns+` FROM vehicle_valuations v WHERE id=?`, record.ID))
	if err == nil {
		if !existing.SameContent(record) {
			return valuation.Record{}, valuation.ErrConflict
		}
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return valuation.Record{}, fmt.Errorf("check valuation identity: %w", err)
	}
	if err := vehicleExists(ctx, tx, record.VehicleID); err != nil {
		return valuation.Record{}, err
	}
	if record.ReplacesID != "" {
		predecessor, err := scanRecord(tx.QueryRowContext(ctx, `SELECT `+recordColumns+` FROM vehicle_valuations v WHERE id=? AND vehicle_id=?`, record.ReplacesID, record.VehicleID))
		if errors.Is(err, sql.ErrNoRows) {
			return valuation.Record{}, valuation.ErrNotFound
		}
		if err != nil {
			return valuation.Record{}, fmt.Errorf("read previous valuation: %w", err)
		}
		if predecessor.ReplacedByID != "" {
			return valuation.Record{}, valuation.ErrConflict
		}
	}
	var predecessor any
	if record.ReplacesID != "" {
		predecessor = record.ReplacesID
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO vehicle_valuations(id,vehicle_id,amount,currency,valued_on,source,notes,recorded_at,replaces_id,withdrawn)
VALUES(?,?,?,?,?,?,?,?,?,?)`, record.ID, record.VehicleID, record.Amount, record.Currency, record.ValuedOn,
		record.Source, record.Notes, record.RecordedAt, predecessor, record.Withdrawn); err != nil {
		return valuation.Record{}, fmt.Errorf("insert valuation: %w", err)
	}
	details, err := json.Marshal(map[string]any{"replacesId": record.ReplacesID, "withdrawn": record.Withdrawn})
	if err != nil {
		return valuation.Record{}, fmt.Errorf("encode valuation audit: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,action,target_type,target_id,created_at,details_json)
VALUES(?,?,'RecordValuation','vehicle',?,?,?)`, credentials.RandomID(), actor, record.VehicleID, record.RecordedAt, string(details)); err != nil {
		return valuation.Record{}, fmt.Errorf("audit valuation: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return valuation.Record{}, fmt.Errorf("commit valuation: %w", err)
	}
	return record, nil
}

func (s *Store) Snapshot(ctx context.Context, asOf string) (valuation.Snapshot, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return valuation.Snapshot{}, fmt.Errorf("begin valuation snapshot: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var snapshot valuation.Snapshot
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicles`).Scan(&snapshot.VehicleCount); err != nil {
		return valuation.Snapshot{}, fmt.Errorf("count valuation vehicles: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+recordColumns+` FROM vehicle_valuations v WHERE valued_on<=? AND withdrawn=0
AND NOT EXISTS (SELECT 1 FROM vehicle_valuations successor WHERE successor.replaces_id=v.id)`, asOf)
	if err != nil {
		return valuation.Snapshot{}, fmt.Errorf("list valuation snapshot: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		r, err := scanRecord(rows)
		if err != nil {
			return valuation.Snapshot{}, fmt.Errorf("read valuation snapshot: %w", err)
		}
		snapshot.Entries = append(snapshot.Entries, r)
	}
	if err := rows.Err(); err != nil {
		return valuation.Snapshot{}, err
	}
	if err := rows.Close(); err != nil {
		return valuation.Snapshot{}, err
	}
	if err := tx.Commit(); err != nil {
		return valuation.Snapshot{}, fmt.Errorf("commit valuation snapshot: %w", err)
	}
	return snapshot, nil
}
