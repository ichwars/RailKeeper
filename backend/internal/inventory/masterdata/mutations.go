package masterdata

import (
	"context"

	"railkeeper/backend/internal/domain"
)

// Mutation is prepared by the use case before the persistence owner begins a write.
// MetadataJSON retains the original JSON validation/encoding error boundary.
type Mutation struct {
	Entry        MasterDataEntry
	MetadataJSON []byte
	Definition   *domain.AccessoryAttributeDefinition
}

type ImportRequest struct {
	Document                      *MasterDataDocument
	EntriesByType                 map[string][]MasterDataEntry
	ArticleTypesWereImported      bool
	AccessorySubtypesWereImported bool
}

type Repository interface {
	Reader
	Create(context.Context, Mutation) error
	Update(context.Context, Mutation) error
	UpdateCustomField(context.Context, Mutation) error
	SetActive(context.Context, string, string, bool) error
	SetActiveMany(context.Context, string, []string, bool) error
	Delete(context.Context, string, string) error
	ImportReconciled(context.Context, ImportRequest) (*MasterDataImportResult, error)
	Usage(context.Context, MasterDataEntry) (bool, error)
}

// Usage and AttributeReferences are bound to the persistence owner's transaction.
type Usage func(context.Context, MasterDataEntry) (bool, error)

type ReferencedAttribute struct {
	ProductID string
	Value     domain.AccessoryAttributeValue
}

type AttributeReferences func(context.Context, string) ([]ReferencedAttribute, error)
