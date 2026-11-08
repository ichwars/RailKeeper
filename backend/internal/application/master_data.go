package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure/masterdatastore"
	"railkeeper/backend/internal/inventory/masterdata"
)

type MasterDataService struct {
	*masterdata.Service
	db *sql.DB
}

func NewMasterDataService(db *sql.DB) *MasterDataService { return newMasterDataService(db, nil) }
func newMasterDataService(db *sql.DB, loader func(context.Context) (map[string][]MasterDataEntry, error)) *MasterDataService {
	return &MasterDataService{Service: masterdata.New(masterdatastore.New(db), loader), db: db}
}
func slugKey(value string) string { return masterdata.SlugKey(value) }
func ParseAccessoryCustomAttributeDefinition(key string, active bool, metadata map[string]any) (domain.AccessoryAttributeDefinition, error) {
	return masterdata.ParseAccessoryCustomAttributeDefinition(key, active, metadata)
}
func validateProtectedArticleTypes(entries []MasterDataEntry, required bool) error {
	return masterdata.ValidateProtectedArticleTypes(entries, required)
}

type MasterDataEntry = masterdata.MasterDataEntry

type MasterDataInput = masterdata.MasterDataInput

type MasterDataRelation = masterdata.MasterDataRelation

type MasterDataDocument = masterdata.MasterDataDocument

type MasterDataImportResult = masterdata.MasterDataImportResult

var ErrMasterDataValidation = masterdata.ErrMasterDataValidation

var ErrMasterDataNotFound = masterdata.ErrMasterDataNotFound

var ErrMasterDataProtected = masterdata.ErrMasterDataProtected

var standardArticleTypeKeys = masterdata.StandardArticleTypeKeys()

type MasterDataOrigin = masterdata.MasterDataOrigin

const MasterDataOriginBundled = masterdata.MasterDataOriginBundled

const MasterDataOriginCustom = masterdata.MasterDataOriginCustom

type MasterDataCapabilities = masterdata.MasterDataCapabilities

const standardArticleType = masterdata.StandardArticleType

const accessoryCustomField = masterdata.AccessoryCustomField
