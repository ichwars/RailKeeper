package catalog

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/revision"
)

func (s *Service) CreateCatalogRevision(ctx context.Context, kind, id string, input VehicleCatalogRevisionInput, actor string) (*VehicleCatalogContext, error) {
	if err := validateCatalogRevisionInput(input); err != nil {
		return nil, err
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := tx.Context(ctx, kind, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	exists, err := tx.RevisionExists(ctx, input.ID)
	if err != nil {
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
		article = credentials.RandomID()
		if err := tx.CreateArticle(ctx, article, now); err != nil {
			return nil, err
		}
	}
	revision, err := tx.InsertRevision(ctx, input.ID, article, input.Fields, input.Note, now)
	if err != nil {
		return nil, err
	}
	if err := applyCatalogRevision(ctx, tx, before, revision, input.SelectedFields, now); err != nil {
		return nil, err
	}
	return finishCatalogChange(ctx, tx, before, "VehicleCatalogRevisionCreated", input.ID, actor, now)
}
func (s *Service) AdoptCatalogRevision(ctx context.Context, kind, id string, input VehicleCatalogAdoptionInput, actor string) (*VehicleCatalogContext, error) {
	if !revision.Valid(input.ExpectedRevision) || input.RevisionID == "" || !validCatalogSelection(input.SelectedFields) {
		return nil, ErrVehicleCatalogValidation
	}
	tx, err := s.repository.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := tx.Context(ctx, kind, strings.TrimSpace(id))
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
func preservePendingCatalogProposal(ctx context.Context, tx Mutation, before *VehicleCatalogContext, now string) error {
	if before.Authority != "legacy" || before.Pinned == nil {
		return nil
	}
	fields := maps.Clone(before.Pinned.Fields)
	maps.Copy(fields, before.Overrides)
	if maps.Equal(fields, before.CurrentFields) {
		return nil
	}
	_, err := tx.InsertRevision(ctx, credentials.RandomID(), before.Pinned.ArticleID, fields,
		"Ungeklärten Vorschlag vor bewusster Änderung erhalten, Quellenzeitpunkt unbekannt. / Retained unresolved proposal before explicit change, source time unknown.", now)
	return err
}
func finishCatalogChange(ctx context.Context, tx Mutation, before *VehicleCatalogContext, action, revision, actor, now string) (*VehicleCatalogContext, error) {
	if err := tx.Audit(ctx, before, action, revision, actor, now); err != nil {
		return nil, err
	}
	result, err := tx.Context(ctx, before.TargetType, before.TargetID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit catalog change: %w", err)
	}
	return result, nil
}
func applyCatalogRevision(ctx context.Context, tx Mutation, before *VehicleCatalogContext, chosen *VehicleCatalogRevision, selected []string, now string) error {
	projection, err := PlanProjection(before, chosen, selected)
	if err != nil {
		return err
	}
	return tx.Project(ctx, before, chosen, projection, now)
}
