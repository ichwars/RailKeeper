package catalogdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/inventory/catalog"
	"railkeeper/backend/internal/revision"
)

type Store struct{ db *sql.DB }
type mutation struct{ tx *sql.Tx }

func New(db *sql.DB) *Store { return &Store{db: db} }
func (s *Store) Begin(ctx context.Context) (catalog.Mutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &mutation{tx: tx}, nil
}
func (m *mutation) Commit() error   { return m.tx.Commit() }
func (m *mutation) Rollback() error { return m.tx.Rollback() }

type vehicleCatalogTarget struct {
	table, view, links, column string
}

func catalogTarget(kind string) (vehicleCatalogTarget, error) {
	switch kind {
	case "vehicle":
		return vehicleCatalogTarget{"vehicles", "vehicle_catalog_legacy_fields", "vehicle_article_links", "vehicle_id"}, nil
	case "set":
		return vehicleCatalogTarget{"vehicle_sets", "vehicle_set_catalog_legacy_fields", "vehicle_set_article_links", "vehicle_set_id"}, nil
	default:
		return vehicleCatalogTarget{}, catalog.ErrVehicleCatalogValidation
	}
}
func (s *Store) Context(ctx context.Context, kind, id string) (*catalog.VehicleCatalogContext, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return (&mutation{tx: tx}).Context(ctx, kind, strings.TrimSpace(id))
}
func (m *mutation) Context(ctx context.Context, kind, id string) (*catalog.VehicleCatalogContext, error) {
	target, err := catalogTarget(kind)
	if err != nil {
		return nil, err
	}
	result := &catalog.VehicleCatalogContext{TargetType: kind, TargetID: id, Overrides: catalog.VehicleCatalogFields{}, Revisions: []catalog.VehicleCatalogRevision{}}
	var current, overrides, articleID, pin, clock string
	var source sql.NullString
	query := `SELECT v.fields_json, COALESCE(l.authority, 'unlinked'), l.source_authority,
COALESCE(l.overrides_json, '{}'), COALESCE(r.article_id, ''), COALESCE(l.revision_id, ''), COALESCE(l.updated_at, '')
FROM ` + target.view + ` v LEFT JOIN ` + target.links + ` l ON l.` + target.column + `=v.entity_id
LEFT JOIN catalog_article_revisions r ON r.id=l.revision_id WHERE v.entity_id=?`
	if err := m.tx.QueryRowContext(ctx, query, id).Scan(&current, &result.Authority, &source, &overrides, &articleID, &pin, &clock); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, catalog.ErrVehicleCatalogNotFound
		}
		return nil, fmt.Errorf("read catalog assignment: %w", err)
	}
	if source.Valid {
		result.SourceAuthority = &source.String
	}
	if result.CurrentFields, err = catalog.DecodeFields(current, true); err != nil {
		return nil, err
	}
	if result.Overrides, err = catalog.DecodeFields(overrides, false); err != nil {
		return nil, err
	}
	rows, err := m.tx.QueryContext(ctx, `SELECT id,article_id,revision_number,fields_json,note,created_at
FROM catalog_article_revisions WHERE article_id=? ORDER BY revision_number DESC`, articleID)
	if err != nil {
		return nil, fmt.Errorf("read catalog revisions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var revision catalog.VehicleCatalogRevision
		var raw string
		if err := rows.Scan(&revision.ID, &revision.ArticleID, &revision.RevisionNumber, &raw, &revision.Note, &revision.CreatedAt); err != nil {
			return nil, err
		}
		if revision.Fields, err = catalog.DecodeFields(raw, true); err != nil {
			return nil, err
		}
		result.Revisions = append(result.Revisions, revision)
		if revision.ID == pin {
			copy := revision
			result.Pinned = &copy
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result.Revision = revision.Hash(struct {
		Context *catalog.VehicleCatalogContext
		Clock   string
	}{result, clock})
	return result, nil
}
func (m *mutation) RevisionExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := m.tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM catalog_article_revisions WHERE id=?)`, id).Scan(&exists)
	return exists, err
}
func (m *mutation) CreateArticle(ctx context.Context, article, now string) error {
	if _, err := m.tx.ExecContext(ctx, `INSERT INTO catalog_articles(id,origin,source_entity_id,created_at) VALUES(?,'manual',NULL,?)`, article, now); err != nil {
		return fmt.Errorf("create independent article: %w", err)
	}
	return nil
}

func (m *mutation) InsertRevision(ctx context.Context, id, article string, fields catalog.VehicleCatalogFields, note, now string) (*catalog.VehicleCatalogRevision, error) {
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	var number int
	if err := m.tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision_number),0)+1 FROM catalog_article_revisions WHERE article_id=?`, article).Scan(&number); err != nil {
		return nil, err
	}
	if _, err := m.tx.ExecContext(ctx, `INSERT INTO catalog_article_revisions(id,article_id,revision_number,fields_json,note,created_at) VALUES(?,?,?,?,?,?)`, id, article, number, string(raw), note, now); err != nil {
		return nil, fmt.Errorf("create immutable catalog revision: %w", err)
	}
	return &catalog.VehicleCatalogRevision{ID: id, ArticleID: article, RevisionNumber: number, Fields: fields, Note: note, CreatedAt: now}, nil
}
func (m *mutation) Audit(ctx context.Context, before *catalog.VehicleCatalogContext, action, revision, actor, now string) error {
	details, err := json.Marshal(map[string]string{"revisionId": revision})
	if err != nil {
		return err
	}
	if _, err := m.tx.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,action,target_type,target_id,created_at,details_json) VALUES(?,NULLIF(?,''),?,?,?,?,?)`, credentials.RandomID(), actor, action, before.TargetType, before.TargetID, now, string(details)); err != nil {
		return fmt.Errorf("audit catalog change: %w", err)
	}
	return nil
}
