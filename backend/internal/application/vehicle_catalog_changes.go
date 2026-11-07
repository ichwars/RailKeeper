package application

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"
	"unicode/utf8"
)

type VehicleCatalogRevisionInput struct {
	ID               string               `json:"id"`
	ExpectedRevision string               `json:"expectedRevision"`
	Fields           VehicleCatalogFields `json:"fields"`
	Note             string               `json:"note"`
	Independent      bool                 `json:"independent"`
	SelectedFields   []string             `json:"selectedFields"`
}

type VehicleCatalogAdoptionInput struct {
	ExpectedRevision string   `json:"expectedRevision"`
	RevisionID       string   `json:"revisionId"`
	SelectedFields   []string `json:"selectedFields"`
}

func validCatalogSelection(fields []string) bool {
	if fields == nil || len(fields) > len(vehicleCatalogFieldNames) {
		return false
	}
	seen := map[string]bool{}
	for _, field := range fields {
		if _, exists := vehicleCatalogColumns[field]; !exists || seen[field] {
			return false
		}
		seen[field] = true
	}
	return true
}

func validCatalogRequiredField(field, value string) bool {
	switch field {
	case "name", "manufacturer", "gauge", "category", "gattung":
		return strings.TrimSpace(value) != ""
	default:
		return true
	}
}

func validateCatalogRevisionInput(input VehicleCatalogRevisionInput) error {
	id, err := hex.DecodeString(input.ID)
	if err != nil || len(id) != 16 || hex.EncodeToString(id) != input.ID ||
		!validCVRevision(input.ExpectedRevision) || !validCatalogSelection(input.SelectedFields) ||
		len(input.Fields) != len(vehicleCatalogFieldNames) || utf8.RuneCountInString(input.Note) > 2000 {
		return ErrVehicleCatalogValidation
	}
	for field, value := range input.Fields {
		if _, exists := vehicleCatalogColumns[field]; !exists || !validCatalogRequiredField(field, value) {
			return ErrVehicleCatalogValidation
		}
	}
	return nil
}

func (s *VehicleService) CreateCatalogRevision(ctx context.Context, kind, id string, input VehicleCatalogRevisionInput, actor string) (*VehicleCatalogContext, error) {
	if err := validateCatalogRevisionInput(input); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := readVehicleCatalogContext(ctx, tx, kind, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM catalog_article_revisions WHERE id=?)`, input.ID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists || before.Revision != input.ExpectedRevision {
		return nil, ErrVehicleCatalogConflict
	}
	if before.Pinned == nil && !input.Independent {
		return nil, ErrVehicleCatalogValidation
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := preservePendingCatalogProposal(ctx, tx, before, now); err != nil {
		return nil, err
	}
	article := ""
	if before.Pinned != nil {
		article = before.Pinned.ArticleID
	}
	if input.Independent {
		article = randomID()
		if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_articles(id,origin,source_entity_id,created_at) VALUES(?,'manual',NULL,?)`, article, now); err != nil {
			return nil, fmt.Errorf("create independent article: %w", err)
		}
	}
	revision, err := insertCatalogRevision(ctx, tx, input.ID, article, input.Fields, input.Note, now)
	if err != nil {
		return nil, err
	}
	if err := applyCatalogRevision(ctx, tx, before, revision, input.SelectedFields, now); err != nil {
		return nil, err
	}
	return finishCatalogChange(ctx, tx, before, "VehicleCatalogRevisionCreated", input.ID, actor, now)
}

func (s *VehicleService) AdoptCatalogRevision(ctx context.Context, kind, id string, input VehicleCatalogAdoptionInput, actor string) (*VehicleCatalogContext, error) {
	if !validCVRevision(input.ExpectedRevision) || input.RevisionID == "" || !validCatalogSelection(input.SelectedFields) {
		return nil, ErrVehicleCatalogValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := readVehicleCatalogContext(ctx, tx, kind, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	if before.Revision != input.ExpectedRevision {
		return nil, ErrVehicleCatalogConflict
	}
	var chosen *VehicleCatalogRevision
	for index := range before.Revisions {
		if before.Revisions[index].ID == input.RevisionID {
			chosen = &before.Revisions[index]
		}
	}
	if chosen == nil {
		return nil, ErrVehicleCatalogNotFound
	}
	unchanged := before.Authority == "catalog" && before.Pinned != nil && before.Pinned.ID == chosen.ID
	for _, field := range input.SelectedFields {
		unchanged = unchanged && before.CurrentFields[field] == chosen.Fields[field]
	}
	if unchanged {
		return before, nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := preservePendingCatalogProposal(ctx, tx, before, now); err != nil {
		return nil, err
	}
	if err := applyCatalogRevision(ctx, tx, before, chosen, input.SelectedFields, now); err != nil {
		return nil, err
	}
	return finishCatalogChange(ctx, tx, before, "VehicleCatalogRevisionAdopted", chosen.ID, actor, now)
}

func insertCatalogRevision(ctx context.Context, tx *sql.Tx, id, article string, fields VehicleCatalogFields, note, now string) (*VehicleCatalogRevision, error) {
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	var number int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(revision_number),0)+1 FROM catalog_article_revisions WHERE article_id=?`, article).Scan(&number); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_article_revisions(id,article_id,revision_number,fields_json,note,created_at) VALUES(?,?,?,?,?,?)`, id, article, number, string(raw), note, now); err != nil {
		return nil, fmt.Errorf("create immutable catalog revision: %w", err)
	}
	return &VehicleCatalogRevision{ID: id, ArticleID: article, RevisionNumber: number, Fields: fields, Note: note, CreatedAt: now}, nil
}

func preservePendingCatalogProposal(ctx context.Context, tx *sql.Tx, before *VehicleCatalogContext, now string) error {
	if before.Authority != "legacy" || before.Pinned == nil {
		return nil
	}
	fields := maps.Clone(before.Pinned.Fields)
	maps.Copy(fields, before.Overrides)
	if maps.Equal(fields, before.CurrentFields) {
		return nil
	}
	_, err := insertCatalogRevision(ctx, tx, randomID(), before.Pinned.ArticleID, fields,
		"Ungeklärten Vorschlag vor bewusster Änderung erhalten, Quellenzeitpunkt unbekannt. / Retained unresolved proposal before explicit change, source time unknown.", now)
	return err
}

func finishCatalogChange(ctx context.Context, tx *sql.Tx, before *VehicleCatalogContext, action, revision, actor, now string) (*VehicleCatalogContext, error) {
	details, err := json.Marshal(map[string]string{"revisionId": revision})
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(id,actor_user_id,action,target_type,target_id,created_at,details_json) VALUES(?,NULLIF(?,''),?,?,?,?,?)`, randomID(), actor, action, before.TargetType, before.TargetID, now, string(details)); err != nil {
		return nil, fmt.Errorf("audit catalog change: %w", err)
	}
	result, err := readVehicleCatalogContext(ctx, tx, before.TargetType, before.TargetID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit catalog change: %w", err)
	}
	return result, nil
}
