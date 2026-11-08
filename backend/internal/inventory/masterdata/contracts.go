package masterdata

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrMasterDataValidation = errors.New("master data validation failed")
	ErrMasterDataNotFound   = errors.New("master data not found")
	ErrMasterDataProtected  = fmt.Errorf("%w: standard article type keys are protected", ErrMasterDataValidation)
)

type MasterDataOrigin string

const (
	MasterDataOriginBundled MasterDataOrigin = "bundled"
	MasterDataOriginCustom  MasterDataOrigin = "custom"
)

type MasterDataCapabilities struct {
	CanDeactivate bool `json:"canDeactivate"`
	CanReactivate bool `json:"canReactivate"`
	CanDelete     bool `json:"canDelete"`
}

const ExportFormat = "railkeeper-master-data"

const ExportVersion = 2

const StandardArticleType = "article_type"

const StandardAccessorySubtype = "accessory_subtype"

const LegacyArticleSubtype = "article_subtype"

const AccessoryCustomField = "accessory_custom_field"

var standardArticleTypeKeys = []string{
	"track", "signal", "decoder", "electrical_control", "building_equipment",
	"landscape_consumable", "lighting", "other",
}

type MasterDataEntry struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Key          string                  `json:"key"`
	Label        string                  `json:"label"`
	Active       bool                    `json:"active"`
	SortOrder    int                     `json:"sortOrder"`
	SourceURL    string                  `json:"sourceUrl,omitempty"`
	Metadata     map[string]any          `json:"metadata"`
	Origin       MasterDataOrigin        `json:"origin"`
	Capabilities *MasterDataCapabilities `json:"capabilities,omitempty"`
	CreatedAt    string                  `json:"createdAt"`
	UpdatedAt    string                  `json:"updatedAt"`
}

type MasterDataInput struct {
	Key       string         `json:"key"`
	Label     string         `json:"label"`
	Active    *bool          `json:"active"`
	SortOrder *int           `json:"sortOrder"`
	SourceURL string         `json:"sourceUrl"`
	Metadata  map[string]any `json:"metadata"`
}

type MasterDataRelation struct {
	ID         string `json:"id"`
	ParentType string `json:"parentType"`
	ParentKey  string `json:"parentKey"`
	ChildType  string `json:"childType"`
	ChildKey   string `json:"childKey"`
	SortOrder  int    `json:"sortOrder"`
}

type MasterDataDocument struct {
	Format    string                       `json:"format"`
	Version   int                          `json:"version"`
	CreatedAt string                       `json:"createdAt"`
	Entries   map[string][]MasterDataEntry `json:"entries"`
	Relations []MasterDataRelation         `json:"relations"`
}

type MasterDataImportResult struct {
	ImportedTypes     int `json:"importedTypes"`
	ImportedEntries   int `json:"importedEntries"`
	ImportedRelations int `json:"importedRelations"`
}
type Reader interface {
	List(context.Context, string, bool) ([]MasterDataEntry, error)
	Snapshot(context.Context) (map[string][]MasterDataEntry, error)
	Get(context.Context, string, string) (*MasterDataEntry, error)
	Relations(context.Context, string, string) ([]MasterDataRelation, error)
	AllRelations(context.Context) ([]MasterDataRelation, error)
}

func StandardArticleTypeKeys() []string { return append([]string(nil), standardArticleTypeKeys...) }
