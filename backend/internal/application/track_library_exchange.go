package application

import (
	"context"
	"errors"
	"strings"

	"railkeeper/backend/internal/domain"
)

var (
	ErrTrackLibraryValidation = errors.New("track library validation failed")
	ErrTrackLibraryConflict   = errors.New("track library version conflict")
	ErrTrackLibraryNotFound   = errors.New("track library not found")
)

const maxTrackLibraryVerificationNote = 500

type TrackLibraryRepository interface {
	ListTrackLibraries(context.Context) ([]domain.TrackGeometryLibrary, error)
	TrackLibraryVersionExists(context.Context, domain.TrackLibraryPackageMetadata) (bool, error)
	ImportTrackLibrary(
		context.Context, domain.TrackLibraryPackage, string,
	) (*domain.TrackGeometryLibrary, error)
	ExportTrackLibrary(context.Context, string) (*domain.TrackLibraryPackage, error)
	UpdateTrackLibraryStatus(
		context.Context, string, domain.TrackGeometryStatus, string, string,
	) (*domain.TrackGeometryLibrary, error)
}

type TrackLibraryService struct {
	repository TrackLibraryRepository
}

type TrackLibraryImportPreview struct {
	Package         domain.TrackLibraryPackage `json:"package"`
	DefinitionCount int                        `json:"definitionCount"`
	Warnings        []string                   `json:"warnings"`
	Conflict        bool                       `json:"conflict"`
	CanImport       bool                       `json:"canImport"`
}

type ImportTrackLibraryInput struct {
	Confirmed bool                       `json:"confirmed"`
	Package   domain.TrackLibraryPackage `json:"package"`
}

type UpdateTrackLibraryStatusInput struct {
	Confirmed        bool                       `json:"confirmed"`
	Status           domain.TrackGeometryStatus `json:"status"`
	VerificationNote string                     `json:"verificationNote"`
}

func NewTrackLibraryService(repository TrackLibraryRepository) *TrackLibraryService {
	return &TrackLibraryService{repository: repository}
}

func (service *TrackLibraryService) List(ctx context.Context) ([]domain.TrackGeometryLibrary, error) {
	return service.repository.ListTrackLibraries(ctx)
}

func (service *TrackLibraryService) PreviewImport(
	ctx context.Context,
	doc domain.TrackLibraryPackage,
) (*TrackLibraryImportPreview, error) {
	doc = normalizeTrackLibraryPackage(doc)
	if err := domain.ValidateTrackLibraryPackage(doc); err != nil {
		return nil, ErrTrackLibraryValidation
	}
	conflict, err := service.repository.TrackLibraryVersionExists(ctx, doc.Library)
	if err != nil {
		return nil, err
	}
	warnings := []string{}
	if doc.Library.Status != domain.TrackGeometryDraft || hasNonDraftTrackLibraryDefinition(doc.Definitions) {
		warnings = append(warnings, "verification_status_reset")
	}
	return &TrackLibraryImportPreview{
		Package: doc, DefinitionCount: len(doc.Definitions), Warnings: warnings,
		Conflict: conflict, CanImport: !conflict,
	}, nil
}

func (service *TrackLibraryService) Import(
	ctx context.Context,
	input ImportTrackLibraryInput,
	actor string,
) (*domain.TrackGeometryLibrary, error) {
	if !input.Confirmed || strings.TrimSpace(actor) == "" {
		return nil, ErrTrackLibraryValidation
	}
	preview, err := service.PreviewImport(ctx, input.Package)
	if err != nil {
		return nil, err
	}
	if preview.Conflict {
		return nil, ErrTrackLibraryConflict
	}
	doc := preview.Package
	doc.ExportedAt = ""
	doc.Library.Status = domain.TrackGeometryDraft
	for index := range doc.Definitions {
		doc.Definitions[index].Status = domain.TrackGeometryDraft
	}
	return service.repository.ImportTrackLibrary(ctx, doc, strings.TrimSpace(actor))
}

func (service *TrackLibraryService) Export(
	ctx context.Context,
	id string,
) (*domain.TrackLibraryPackage, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrTrackLibraryValidation
	}
	return service.repository.ExportTrackLibrary(ctx, id)
}

func (service *TrackLibraryService) UpdateStatus(
	ctx context.Context,
	id string,
	input UpdateTrackLibraryStatusInput,
	actor string,
) (*domain.TrackGeometryLibrary, error) {
	id = strings.TrimSpace(id)
	actor = strings.TrimSpace(actor)
	input.VerificationNote = strings.TrimSpace(input.VerificationNote)
	validStatus := input.Status == domain.TrackGeometryVerified || input.Status == domain.TrackGeometryRetired
	validNote := input.Status != domain.TrackGeometryVerified ||
		(input.VerificationNote != "" && len([]rune(input.VerificationNote)) <= maxTrackLibraryVerificationNote)
	if id == "" || actor == "" || !input.Confirmed || !validStatus || !validNote {
		return nil, ErrTrackLibraryValidation
	}
	return service.repository.UpdateTrackLibraryStatus(
		ctx, id, input.Status, input.VerificationNote, actor,
	)
}

func normalizeTrackLibraryPackage(doc domain.TrackLibraryPackage) domain.TrackLibraryPackage {
	doc.Format = strings.TrimSpace(doc.Format)
	doc.ExportedAt = strings.TrimSpace(doc.ExportedAt)
	doc.Library.Manufacturer = strings.TrimSpace(doc.Library.Manufacturer)
	doc.Library.TrackSystem = strings.TrimSpace(doc.Library.TrackSystem)
	doc.Library.Gauge = strings.TrimSpace(doc.Library.Gauge)
	doc.Library.Scale = strings.TrimSpace(doc.Library.Scale)
	doc.Library.Version = strings.TrimSpace(doc.Library.Version)
	doc.Library.SourceURL = strings.TrimSpace(doc.Library.SourceURL)
	for index := range doc.Definitions {
		definition := &doc.Definitions[index]
		definition.ArticleNumber = strings.TrimSpace(definition.ArticleNumber)
		definition.Name = strings.TrimSpace(definition.Name)
		definition.SourceURL = strings.TrimSpace(definition.SourceURL)
		for portIndex := range definition.Geometry.Ports {
			definition.Geometry.Ports[portIndex].ID = strings.TrimSpace(definition.Geometry.Ports[portIndex].ID)
		}
		for routeIndex := range definition.Geometry.Routes {
			definition.Geometry.Routes[routeIndex].ID = strings.TrimSpace(
				definition.Geometry.Routes[routeIndex].ID,
			)
		}
	}
	return doc
}

func hasNonDraftTrackLibraryDefinition(definitions []domain.TrackLibraryPackageDefinition) bool {
	for _, definition := range definitions {
		if definition.Status != domain.TrackGeometryDraft {
			return true
		}
	}
	return false
}
