package workflow

import (
	"context"
	"errors"

	"railkeeper/backend/internal/decoder"
)

var (
	ErrCVNotFound              = errors.New("CV vehicle or record not found")
	ErrCVFileConflict          = errors.New("vehicle CV file or its source references changed")
	ErrStoredCVMappingConflict = errors.New("stored CV vehicle mapping is unresolved")
)

type CVReader interface {
	CVValue(context.Context, string, string) (decoder.CVRecord, bool, error)
	CVValues(context.Context, string) ([]decoder.CVRecord, error)
	CVHistory(context.Context, string) ([]decoder.VehicleCVValueHistory, error)
	CVHistories(context.Context, string) ([]decoder.VehicleCVValueHistory, error)
	CVFile(context.Context, string, string) (*decoder.VehicleCVFile, error)
	CVFiles(context.Context, string) ([]decoder.VehicleCVFile, error)
}

type CVMutation interface {
	CVReader
	WorkspaceWriter
	CVBinding(context.Context, string, string) (decoder.CVBinding, error)
	CVSourceVehicle(context.Context, string) (string, bool, error)
	FindCV(context.Context, decoder.CVRecord) (decoder.CVRecord, bool, error)
	InsertCV(context.Context, decoder.CVRecord) error
	UpdateCV(context.Context, decoder.CVRecord) error
	DeleteCV(context.Context, string, string) error
	InsertCVHistory(context.Context, decoder.VehicleCVValueHistory) error
	InsertCVFile(context.Context, decoder.VehicleCVFile) error
	DeleteCVFile(context.Context, string, string) error
	ECoSMappingCount(context.Context, string, string) (int, error)
	CVAddressExists(context.Context, string, int, *int, *int) (bool, error)
	Commit() error
	Rollback() error
}

type CVRepository interface {
	CVReader
	BeginVehicleCVWrite(context.Context, string) (CVMutation, error)
	BeginCVTransaction(context.Context) (CVMutation, error)
}

type CVService struct{ repository CVRepository }

func NewCVService(repository CVRepository) *CVService { return &CVService{repository: repository} }
