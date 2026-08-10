package application

import (
	"context"
	"errors"
	"testing"

	"railkeeper/backend/internal/domain"
)

type trackLibraryRepositoryStub struct {
	conflict     bool
	imported     *domain.TrackLibraryPackage
	updatedID    string
	updatedState domain.TrackGeometryStatus
	updatedNote  string
}

func (stub *trackLibraryRepositoryStub) ListTrackLibraries(context.Context) ([]domain.TrackGeometryLibrary, error) {
	return []domain.TrackGeometryLibrary{}, nil
}

func (stub *trackLibraryRepositoryStub) TrackLibraryVersionExists(
	context.Context,
	domain.TrackLibraryPackageMetadata,
) (bool, error) {
	return stub.conflict, nil
}

func (stub *trackLibraryRepositoryStub) ImportTrackLibrary(
	_ context.Context,
	doc domain.TrackLibraryPackage,
	_ string,
) (*domain.TrackGeometryLibrary, error) {
	stub.imported = &doc
	return &domain.TrackGeometryLibrary{ID: "library-1", Status: domain.TrackGeometryDraft}, nil
}

func (stub *trackLibraryRepositoryStub) ExportTrackLibrary(
	context.Context,
	string,
) (*domain.TrackLibraryPackage, error) {
	doc := validApplicationTrackLibraryPackage()
	return &doc, nil
}

func (stub *trackLibraryRepositoryStub) UpdateTrackLibraryStatus(
	_ context.Context,
	id string,
	status domain.TrackGeometryStatus,
	note string,
	_ string,
) (*domain.TrackGeometryLibrary, error) {
	stub.updatedID, stub.updatedState, stub.updatedNote = id, status, note
	return &domain.TrackGeometryLibrary{ID: id, Status: status, VerificationNote: note}, nil
}

func TestTrackLibraryPreviewNormalizesWithoutMutation(t *testing.T) {
	repository := &trackLibraryRepositoryStub{}
	service := NewTrackLibraryService(repository)
	doc := validApplicationTrackLibraryPackage()
	doc.Library.Manufacturer = "  Kühn  "
	doc.Definitions[0].Name = "  Gerades Gleis  "

	preview, err := service.PreviewImport(t.Context(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.CanImport || preview.Conflict || preview.DefinitionCount != 1 ||
		preview.Package.Library.Manufacturer != "Kühn" ||
		preview.Package.Definitions[0].Name != "Gerades Gleis" || len(preview.Warnings) != 1 {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	if repository.imported != nil {
		t.Fatal("preview mutated repository")
	}
}

func TestTrackLibraryPreviewReportsVersionConflict(t *testing.T) {
	repository := &trackLibraryRepositoryStub{conflict: true}
	preview, err := NewTrackLibraryService(repository).PreviewImport(
		t.Context(), validApplicationTrackLibraryPackage(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if preview.CanImport || !preview.Conflict {
		t.Fatalf("expected conflict preview: %#v", preview)
	}
}

func TestTrackLibraryImportRequiresConfirmationAndResetsTrust(t *testing.T) {
	repository := &trackLibraryRepositoryStub{}
	service := NewTrackLibraryService(repository)
	doc := validApplicationTrackLibraryPackage()
	if _, err := service.Import(t.Context(), ImportTrackLibraryInput{Package: doc}, "admin"); !errors.Is(err, ErrTrackLibraryValidation) {
		t.Fatalf("expected confirmation validation, got %v", err)
	}
	created, err := service.Import(t.Context(), ImportTrackLibraryInput{
		Confirmed: true, Package: doc,
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != domain.TrackGeometryDraft || repository.imported == nil ||
		repository.imported.Library.Status != domain.TrackGeometryDraft ||
		repository.imported.Definitions[0].Status != domain.TrackGeometryDraft ||
		repository.imported.ExportedAt != "" {
		t.Fatalf("import retained external trust: created=%#v package=%#v", created, repository.imported)
	}
}

func TestTrackLibraryStatusRequiresReviewNote(t *testing.T) {
	repository := &trackLibraryRepositoryStub{}
	service := NewTrackLibraryService(repository)
	if _, err := service.UpdateStatus(t.Context(), "library-1", UpdateTrackLibraryStatusInput{
		Confirmed: true, Status: domain.TrackGeometryVerified,
	}, "admin"); !errors.Is(err, ErrTrackLibraryValidation) {
		t.Fatalf("expected note validation, got %v", err)
	}
	updated, err := service.UpdateStatus(t.Context(), "library-1", UpdateTrackLibraryStatusInput{
		Confirmed: true, Status: domain.TrackGeometryVerified, VerificationNote: "  Herstellerkatalog geprüft  ",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != domain.TrackGeometryVerified || repository.updatedID != "library-1" ||
		repository.updatedNote != "Herstellerkatalog geprüft" {
		t.Fatalf("unexpected status update: %#v", updated)
	}
}

func validApplicationTrackLibraryPackage() domain.TrackLibraryPackage {
	return domain.TrackLibraryPackage{
		Format: domain.TrackLibraryPackageFormat, SchemaVersion: 1, ExportedAt: "2026-08-10T08:00:00Z",
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
