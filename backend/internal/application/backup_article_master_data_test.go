package application_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestLegacyBackupRestorePreservesCurrentArticleMasterData(t *testing.T) {
	for _, version := range []int{1, 2} {
		t.Run(backupVersionName(version), func(t *testing.T) {
			ctx := context.Background()
			db := backupTestDB(t, t.TempDir())
			masterData := application.NewMasterDataService(db)
			active := true
			if _, err := masterData.Update(ctx, "accessory_subtype", "track:straight", application.MasterDataInput{
				Label: "Workshop straight", Active: &active,
			}); err != nil {
				t.Fatal(err)
			}
			if _, err := masterData.Create(ctx, "accessory_subtype", application.MasterDataInput{
				Key: "track:club_profile", Label: "Club profile", Active: &active,
			}); err != nil {
				t.Fatal(err)
			}
			beforeTypes := backupMasterDataKeys(t, masterData, "article_type")
			beforeSubtypes := backupMasterDataKeys(t, masterData, "accessory_subtype")

			doc := &application.BackupDocument{
				Format: "railkeeper-backup", Version: version, Tables: backupDocumentTablesThroughVersion(version),
			}
			doc.Tables["master_data_entries"] = []map[string]any{{
				"id": "vehicle-category-lok", "type": "vehicle_category", "key": "lok", "label": "Lok",
				"active": 1, "sort_order": 10, "source_url": "", "metadata_json": "{}",
				"created_at": "2025-01-01T00:00:00Z", "updated_at": "2025-01-01T00:00:00Z",
			}}

			if _, err := application.NewBackupService(db, t.TempDir()).Import(ctx, doc); err != nil {
				t.Fatal(err)
			}
			after := application.NewMasterDataService(db)
			if got := backupMasterDataKeys(t, after, "article_type"); !reflect.DeepEqual(got, beforeTypes) {
				t.Fatalf("legacy restore changed authoritative article types: before=%#v after=%#v", beforeTypes, got)
			}
			if got := backupMasterDataKeys(t, after, "accessory_subtype"); !reflect.DeepEqual(got, beforeSubtypes) {
				t.Fatalf("legacy restore changed authoritative accessory subtypes: before=%#v after=%#v", beforeSubtypes, got)
			}
			straight, err := after.Get(ctx, "accessory_subtype", "track:straight")
			if err != nil {
				t.Fatal(err)
			}
			if straight.Label != "Workshop straight" {
				t.Fatalf("legacy restore lost current configured subtype label: %#v", straight)
			}
		})
	}
}

func TestVersionThreeBackupRestoresConfiguredArticleMasterDataExactly(t *testing.T) {
	ctx := context.Background()
	db := backupTestDB(t, t.TempDir())
	masterData := application.NewMasterDataService(db)
	active := true
	if _, err := masterData.Update(ctx, "accessory_subtype", "track:straight", application.MasterDataInput{
		Label: "Backed-up straight", Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := masterData.Create(ctx, "accessory_subtype", application.MasterDataInput{
		Key: "track:backed_up_profile", Label: "Backed-up profile", Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, t.TempDir())
	doc, err := service.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := masterData.Create(ctx, "accessory_subtype", application.MasterDataInput{
		Key: "track:target_only", Label: "Target only", Active: &active,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Import(ctx, doc); err != nil {
		t.Fatal(err)
	}
	after := application.NewMasterDataService(db)
	if _, err := after.Get(ctx, "accessory_subtype", "track:target_only"); err != application.ErrMasterDataNotFound {
		t.Fatalf("version 3 restore preserved target-only subtype: %v", err)
	}
	straight, err := after.Get(ctx, "accessory_subtype", "track:straight")
	if err != nil {
		t.Fatal(err)
	}
	if straight.Label != "Backed-up straight" {
		t.Fatalf("version 3 restore did not use the backed-up label: %#v", straight)
	}
	if _, err := after.Get(ctx, "accessory_subtype", "track:backed_up_profile"); err != nil {
		t.Fatalf("version 3 restore lost backed-up configured subtype: %v", err)
	}
}

func TestLegacyBackupRestoreReleasesTransactionWhenCurrentArticleTypesAreInvalid(t *testing.T) {
	ctx := context.Background()
	db := backupTestDB(t, t.TempDir())
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, `DELETE FROM master_data_entries WHERE type='article_type' AND key='track'`); err != nil {
		t.Fatal(err)
	}
	doc := &application.BackupDocument{
		Format: "railkeeper-backup", Version: 1, Tables: backupDocumentTablesThroughVersion(1),
	}

	if _, err := application.NewBackupService(db, t.TempDir()).Import(ctx, doc); !errors.Is(err, application.ErrBackupInvalid) {
		t.Fatalf("expected invalid legacy restore, got %v", err)
	}
	if inUse := db.Stats().InUse; inUse != 0 {
		t.Fatalf("failed legacy restore retained %d database connections", inUse)
	}
}

func backupMasterDataKeys(t *testing.T, service *application.MasterDataService, typeName string) []string {
	t.Helper()
	entries, err := service.List(context.Background(), typeName, false)
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(entries))
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}
	return keys
}

func backupVersionName(version int) string {
	if version == 1 {
		return "version one"
	}
	return "version two"
}
