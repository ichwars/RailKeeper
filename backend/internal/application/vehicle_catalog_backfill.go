package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"railkeeper/backend/migrations"
)

type VehicleCatalogAssignment struct {
	TargetType      string   `json:"targetType"`
	TargetID        string   `json:"targetId"`
	Name            string   `json:"name"`
	SetID           string   `json:"setId,omitempty"`
	ArticleID       string   `json:"articleId,omitempty"`
	RevisionID      string   `json:"revisionId,omitempty"`
	Authority       string   `json:"authority"`
	OverrideFields  []string `json:"overrideFields"`
	DifferentFields []string `json:"differentFields"`
}

type VehicleCatalogBackfillReport struct {
	Vehicles                 int                        `json:"vehicles"`
	Sets                     int                        `json:"sets"`
	LinkedVehicles           int                        `json:"linkedVehicles"`
	LinkedSets               int                        `json:"linkedSets"`
	DifferentRepresentations int                        `json:"differentRepresentations"`
	Assignments              []VehicleCatalogAssignment `json:"assignments"`
}

func backfillVehicleCatalogTx(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, migrations.LegacyVehicleCatalogSQL); err != nil {
		return fmt.Errorf("backfill vehicle catalog: %w", err)
	}
	return nil
}

func (s *VehicleService) CatalogBackfillReport(ctx context.Context) (*VehicleCatalogBackfillReport, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return readVehicleCatalogBackfillReport(ctx, tx)
}

func (s *VehicleService) BackfillLegacyCatalog(ctx context.Context, actor string) (*VehicleCatalogBackfillReport, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var before int
	if err := tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM vehicle_article_links)+(SELECT COUNT(*) FROM vehicle_set_article_links)`).Scan(&before); err != nil {
		return nil, err
	}
	if err := backfillVehicleCatalogTx(ctx, tx); err != nil {
		return nil, err
	}
	report, err := readVehicleCatalogBackfillReport(ctx, tx)
	if err != nil {
		return nil, err
	}
	if added := report.LinkedVehicles + report.LinkedSets - before; added > 0 {
		details, err := json.Marshal(map[string]int{"linksAdded": added})
		if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, NULLIF(?, ''), 'VehicleCatalogLegacyBackfilled', 'catalog', 'legacy-backfill', ?, ?)`, randomID(), actor, time.Now().UTC().Format(time.RFC3339), string(details)); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return report, nil
}

func readVehicleCatalogBackfillReport(ctx context.Context, tx *sql.Tx) (*VehicleCatalogBackfillReport, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT 'vehicle', v.entity_id, v.fields_json, COALESCE(m.vehicle_set_id, ''),
  COALESCE(r.article_id, ''), COALESCE(l.revision_id, ''), COALESCE(l.authority, 'unlinked'),
  COALESCE(r.fields_json, '{}'), COALESCE(l.overrides_json, '{}')
FROM vehicle_catalog_legacy_fields v LEFT JOIN vehicle_article_links l ON l.vehicle_id=v.entity_id
LEFT JOIN catalog_article_revisions r ON r.id=l.revision_id
LEFT JOIN vehicle_set_members m ON m.vehicle_id=v.entity_id
UNION ALL
SELECT 'set', s.entity_id, s.fields_json, '', COALESCE(r.article_id, ''),
  COALESCE(l.revision_id, ''), COALESCE(l.authority, 'unlinked'),
  COALESCE(r.fields_json, '{}'), COALESCE(l.overrides_json, '{}')
FROM vehicle_set_catalog_legacy_fields s LEFT JOIN vehicle_set_article_links l ON l.vehicle_set_id=s.entity_id
LEFT JOIN catalog_article_revisions r ON r.id=l.revision_id
ORDER BY 1, 2`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	report := &VehicleCatalogBackfillReport{Assignments: []VehicleCatalogAssignment{}}
	for rows.Next() {
		var assignment VehicleCatalogAssignment
		var legacyJSON, fieldsJSON, overridesJSON string
		if err := rows.Scan(&assignment.TargetType, &assignment.TargetID, &legacyJSON, &assignment.SetID,
			&assignment.ArticleID, &assignment.RevisionID, &assignment.Authority, &fieldsJSON, &overridesJSON); err != nil {
			return nil, err
		}
		legacy, err := decodeVehicleCatalogFields(legacyJSON, true)
		if err != nil {
			return nil, err
		}
		assignment.Name = legacy["name"]
		assignment.OverrideFields, assignment.DifferentFields = []string{}, []string{}
		if assignment.RevisionID != "" {
			fields, err := decodeVehicleCatalogFields(fieldsJSON, true)
			if err != nil {
				return nil, err
			}
			overrides, err := decodeVehicleCatalogFields(overridesJSON, false)
			if err != nil {
				return nil, err
			}
			for field, value := range overrides {
				assignment.OverrideFields = append(assignment.OverrideFields, field)
				fields[field] = value
			}
			for field, value := range legacy {
				if fields[field] != value {
					assignment.DifferentFields = append(assignment.DifferentFields, field)
				}
			}
		}
		sort.Strings(assignment.OverrideFields)
		sort.Strings(assignment.DifferentFields)
		if assignment.TargetType == "vehicle" {
			report.Vehicles++
			if assignment.RevisionID != "" {
				report.LinkedVehicles++
			}
		} else {
			report.Sets++
			if assignment.RevisionID != "" {
				report.LinkedSets++
			}
		}
		if len(assignment.DifferentFields) > 0 {
			report.DifferentRepresentations++
		}
		report.Assignments = append(report.Assignments, assignment)
	}
	return report, rows.Err()
}
