package infrastructure_test

import (
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestTrackLibraryWorkflowKeepsImportsDraftUntilReviewed(t *testing.T) {
	db := openTrackPlannerSchemaDB(t)
	repository := infrastructure.NewTrackPlannerRepository(db)
	service := application.NewTrackLibraryService(repository)
	planner := application.NewTrackPlannerService(repository)
	ctx := t.Context()

	libraries, err := service.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(libraries) != 1 || libraries[0].DefinitionCount != 2 ||
		libraries[0].Status != domain.TrackGeometryVerified {
		t.Fatalf("unexpected seeded libraries: %#v", libraries)
	}

	doc := trackLibraryRepositoryTestPackage()
	created, err := service.Import(ctx, application.ImportTrackLibraryInput{
		Confirmed: true, Package: doc,
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != domain.TrackGeometryDraft || created.DefinitionCount != 1 {
		t.Fatalf("unexpected imported library: %#v", created)
	}
	geometries, err := planner.ListGeometries(ctx, "TT")
	if err != nil {
		t.Fatal(err)
	}
	if len(geometries) != 2 {
		t.Fatalf("draft library is placeable: %#v", geometries)
	}

	exported, err := service.Export(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if exported.Format != domain.TrackLibraryPackageFormat || exported.SchemaVersion != 1 ||
		exported.ExportedAt == "" || exported.Library.Status != domain.TrackGeometryDraft ||
		len(exported.Definitions) != 1 || exported.Definitions[0].Status != domain.TrackGeometryDraft ||
		exported.Definitions[0].Geometry.Ports[1].XMM != 128 {
		t.Fatalf("unexpected library export: %#v", exported)
	}

	verified, err := service.UpdateStatus(ctx, created.ID, application.UpdateTrackLibraryStatusInput{
		Confirmed: true, Status: domain.TrackGeometryVerified,
		VerificationNote: "Herstellerkatalog 2026 geprüft",
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if verified.Status != domain.TrackGeometryVerified || verified.VerifiedAt == "" ||
		verified.VerifiedBy != "admin-1" || verified.VerificationNote == "" {
		t.Fatalf("unexpected verified library: %#v", verified)
	}
	geometries, err = planner.ListGeometries(ctx, "TT")
	if err != nil {
		t.Fatal(err)
	}
	if len(geometries) != 3 || geometries[0].ArticleNumber != "72620" ||
		geometries[0].Manufacturer != "Kühn" {
		t.Fatalf("verified geometry missing: %#v", geometries)
	}

	retired, err := service.UpdateStatus(ctx, created.ID, application.UpdateTrackLibraryStatusInput{
		Confirmed: true, Status: domain.TrackGeometryRetired,
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if retired.Status != domain.TrackGeometryRetired {
		t.Fatalf("unexpected retired library: %#v", retired)
	}
	geometries, err = planner.ListGeometries(ctx, "TT")
	if err != nil {
		t.Fatal(err)
	}
	if len(geometries) != 2 {
		t.Fatalf("retired geometry remains placeable: %#v", geometries)
	}

	preview, err := service.PreviewImport(ctx, doc)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Conflict || preview.CanImport {
		t.Fatalf("expected version conflict: %#v", preview)
	}
	if _, err := service.Import(ctx, application.ImportTrackLibraryInput{
		Confirmed: true, Package: doc,
	}, "admin-1"); !errors.Is(err, application.ErrTrackLibraryConflict) {
		t.Fatalf("expected import conflict, got %v", err)
	}

	doc.Library.Version = "2026.2"
	if _, err := service.Import(ctx, application.ImportTrackLibraryInput{
		Confirmed: true, Package: doc,
	}, "admin-1"); err != nil {
		t.Fatalf("import next version: %v", err)
	}
	var audits int
	if err := db.QueryRow(`
SELECT COUNT(*) FROM audit_logs
WHERE target_type='track_geometry_library' AND action IN(
  'TrackGeometryLibraryImported', 'TrackGeometryLibraryVerified', 'TrackGeometryLibraryRetired'
)`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 4 {
		t.Fatalf("unexpected library audit count: %d", audits)
	}
}

func trackLibraryRepositoryTestPackage() domain.TrackLibraryPackage {
	return domain.TrackLibraryPackage{
		Format: domain.TrackLibraryPackageFormat, SchemaVersion: 1,
		Library: domain.TrackLibraryPackageMetadata{
			Manufacturer: "Kühn", TrackSystem: "TT", Gauge: "TT", Scale: "1:120",
			Version: "2026.1", SourceURL: "https://example.com/catalogue.pdf",
			Status: domain.TrackGeometryVerified,
		},
		Definitions: []domain.TrackLibraryPackageDefinition{{
			ArticleNumber: "72620", Name: "Gerades Gleis", Kind: domain.TrackGeometryStraight,
			LengthMM: 128, SourceURL: "https://example.com/72620", Status: domain.TrackGeometryVerified,
			Geometry: domain.TrackGeometry{SchemaVersion: 1,
				Ports: []domain.TrackPort{
					{ID: "a", DirectionDegrees: 180}, {ID: "b", XMM: 128},
				},
				Routes: []domain.TrackRoute{{ID: "main", Points: []domain.TrackPoint{{}, {XMM: 128}}}},
			},
		}},
	}
}
