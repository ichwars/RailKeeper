package catalog

import (
	"context"
	"errors"
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
type RawAssignment struct {
	Assignment                            VehicleCatalogAssignment
	LegacyJSON, FieldsJSON, OverridesJSON string
}
type Repository interface {
	Context(context.Context, string, string) (*VehicleCatalogContext, error)
	Report(context.Context) ([]RawAssignment, error)
	Begin(context.Context) (Mutation, error)
}
type Mutation interface {
	Context(context.Context, string, string) (*VehicleCatalogContext, error)
	RevisionExists(context.Context, string) (bool, error)
	CreateArticle(context.Context, string, string) error
	InsertRevision(context.Context, string, string, VehicleCatalogFields, string, string) (*VehicleCatalogRevision, error)
	Project(context.Context, *VehicleCatalogContext, *VehicleCatalogRevision, Projection, string) error
	Audit(context.Context, *VehicleCatalogContext, string, string, string, string) error
	LinkCount(context.Context) (int, error)
	Backfill(context.Context) error
	Report(context.Context) ([]RawAssignment, error)
	AuditBackfill(context.Context, int, string, string) error
	Commit() error
	Rollback() error
}
type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) CatalogContext(ctx context.Context, kind, id string) (*VehicleCatalogContext, error) {
	return s.repository.Context(ctx, kind, strings.TrimSpace(id))
}
func (s *Service) CatalogBackfillReport(ctx context.Context) (*VehicleCatalogBackfillReport, error) {
	rows, err := s.repository.Report(ctx)
	if err != nil {
		return nil, err
	}
	return BuildReport(rows)
}
