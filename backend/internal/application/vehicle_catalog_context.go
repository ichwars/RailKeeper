package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrVehicleCatalogValidation = errors.New("invalid catalog request")
	ErrVehicleCatalogNotFound   = errors.New("catalog target or revision not found")
	ErrVehicleCatalogConflict   = errors.New("catalog comparison changed")
)

type VehicleCatalogRevision struct {
	ID             string               `json:"id"`
	ArticleID      string               `json:"articleId"`
	RevisionNumber int                  `json:"revisionNumber"`
	Fields         VehicleCatalogFields `json:"fields"`
	Note           string               `json:"note"`
	CreatedAt      string               `json:"createdAt"`
}

type VehicleCatalogContext struct {
	TargetType      string                   `json:"targetType"`
	TargetID        string                   `json:"targetId"`
	CurrentFields   VehicleCatalogFields     `json:"currentFields"`
	Authority       string                   `json:"authority"`
	SourceAuthority *string                  `json:"sourceAuthority,omitempty"`
	Pinned          *VehicleCatalogRevision  `json:"pinned,omitempty"`
	Overrides       VehicleCatalogFields     `json:"overrides"`
	Revisions       []VehicleCatalogRevision `json:"revisions"`
	Revision        string                   `json:"revision"`
}

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
		return vehicleCatalogTarget{}, ErrVehicleCatalogValidation
	}
}

func (s *VehicleService) CatalogContext(ctx context.Context, kind, id string) (*VehicleCatalogContext, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return readVehicleCatalogContext(ctx, tx, kind, strings.TrimSpace(id))
}

func readVehicleCatalogContext(ctx context.Context, tx *sql.Tx, kind, id string) (*VehicleCatalogContext, error) {
	target, err := catalogTarget(kind)
	if err != nil {
		return nil, err
	}
	result := &VehicleCatalogContext{TargetType: kind, TargetID: id, Overrides: VehicleCatalogFields{}, Revisions: []VehicleCatalogRevision{}}
	var current, overrides, articleID, pin, clock string
	var source sql.NullString
	query := `SELECT v.fields_json, COALESCE(l.authority, 'unlinked'), l.source_authority,
COALESCE(l.overrides_json, '{}'), COALESCE(r.article_id, ''), COALESCE(l.revision_id, ''), COALESCE(l.updated_at, '')
FROM ` + target.view + ` v LEFT JOIN ` + target.links + ` l ON l.` + target.column + `=v.entity_id
LEFT JOIN catalog_article_revisions r ON r.id=l.revision_id WHERE v.entity_id=?`
	if err := tx.QueryRowContext(ctx, query, id).Scan(&current, &result.Authority, &source, &overrides, &articleID, &pin, &clock); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleCatalogNotFound
		}
		return nil, fmt.Errorf("read catalog assignment: %w", err)
	}
	if source.Valid {
		result.SourceAuthority = &source.String
	}
	if result.CurrentFields, err = decodeVehicleCatalogFields(current, true); err != nil {
		return nil, err
	}
	if result.Overrides, err = decodeVehicleCatalogFields(overrides, false); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,article_id,revision_number,fields_json,note,created_at
FROM catalog_article_revisions WHERE article_id=? ORDER BY revision_number DESC`, articleID)
	if err != nil {
		return nil, fmt.Errorf("read catalog revisions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var revision VehicleCatalogRevision
		var raw string
		if err := rows.Scan(&revision.ID, &revision.ArticleID, &revision.RevisionNumber, &raw, &revision.Note, &revision.CreatedAt); err != nil {
			return nil, err
		}
		if revision.Fields, err = decodeVehicleCatalogFields(raw, true); err != nil {
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
	result.Revision = recordRevision(struct {
		Context *VehicleCatalogContext
		Clock   string
	}{result, clock})
	return result, nil
}
