package application

import (
	"railkeeper/backend/internal/infrastructure/transferfiles"
	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
	"railkeeper/backend/internal/transfer/workflow"
)

type CreateDataTransferProfileInput = workflow.CreateDataTransferProfileInput
type UpdateDataTransferProfileInput = workflow.UpdateDataTransferProfileInput
type DataTransferSummary = workflow.DataTransferSummary
type DataTransferJobDetails = workflow.DataTransferJobDetails
type DataTransferService = workflow.Service

func NewDataTransferService(repository DataTransferRepository, dataDir string,
	options ...DataTransferServiceOption) *DataTransferService {
	return workflow.New(repository, transferfiles.New(dataDir),
		transferpolicy.AggregateRules{Vehicle: ValidateTransferVehicle, Accessory: ValidateTransferAccessory}, options...)
}

var (
	ErrDataTransferValidation = transfer.ErrDataTransferValidation
	ErrDataTransferNotFound   = transfer.ErrDataTransferNotFound
)

func DataTransferProfileNameKey(name string) string { return transfer.DataTransferProfileNameKey(name) }
