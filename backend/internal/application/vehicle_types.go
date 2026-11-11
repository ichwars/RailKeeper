package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/infrastructure/workshopdata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
	"railkeeper/backend/internal/inventory/workshop"
)

type VehicleService struct {
	db             *sql.DB
	imageLocalizer VehicleImageLocalizer
	workshop       *workshop.Service
	workshopRows   *workshopdata.Store
}
type VehicleImageLocalizer func(ctx context.Context, vehicleID string, images []VehicleImageInput) ([]VehicleImageInput, error)

type Vehicle = vehiclemodel.Vehicle

type VehicleSetSummary = vehiclemodel.VehicleSetSummary

type VehicleSetMainImageMode = vehiclemodel.VehicleSetMainImageMode

type VehicleSetMainImageInput = vehiclemodel.VehicleSetMainImageInput

type VehicleSetMainImage = vehiclemodel.VehicleSetMainImage

type VehicleSetImage = vehiclemodel.VehicleSetImage

type VehicleSetImageInput = vehiclemodel.VehicleSetImageInput

type VehicleSetInput = vehiclemodel.VehicleSetInput

type CreateVehicleSetInput = vehiclemodel.CreateVehicleSetInput

type VehicleSet = vehiclemodel.VehicleSet

type VehicleExternalMap = vehiclemodel.VehicleExternalMap

type VehicleExternalMapInput = vehiclemodel.VehicleExternalMapInput

type VehicleImage = vehiclemodel.VehicleImage

type VehicleImageCreation = vehiclemodel.VehicleImageCreation

type VehicleImageInput = vehiclemodel.VehicleImageInput

type VehicleAttachment = vehiclemodel.VehicleAttachment

type VehicleAttachmentInput = vehiclemodel.VehicleAttachmentInput

type VehicleAttachmentUpdateInput = vehiclemodel.VehicleAttachmentUpdateInput

type VehicleMaintenance = vehiclemodel.VehicleMaintenance

type VehicleMaintenanceInput = vehiclemodel.VehicleMaintenanceInput

type VehicleSparePart = vehiclemodel.VehicleSparePart

type VehicleSparePartInput = vehiclemodel.VehicleSparePartInput

type VehicleFunction = vehiclemodel.VehicleFunction

type VehicleFunctionInput = vehiclemodel.VehicleFunctionInput

type VehicleFunctionChange = vehiclemodel.VehicleFunctionChange

type VehicleCVValue = vehiclemodel.VehicleCVValue

type VehicleCVValueHistory = vehiclemodel.VehicleCVValueHistory

type VehicleCVValueInput = vehiclemodel.VehicleCVValueInput

type VehicleCVFile = vehiclemodel.VehicleCVFile

type VehicleCVFileInput = vehiclemodel.VehicleCVFileInput

type CreateVehicleInput = vehiclemodel.CreateVehicleInput

var (
	ErrVehicleValidation              = vehiclemodel.ErrVehicleValidation
	ErrVehicleOperationalValidation   = vehiclemodel.ErrVehicleOperationalValidation
	ErrVehicleNotFound                = vehiclemodel.ErrVehicleNotFound
	ErrVehicleExternalMappingConflict = vehiclemodel.ErrVehicleExternalMappingConflict
	ErrVehicleSetValidation           = vehiclemodel.ErrVehicleSetValidation
	ErrVehicleSetNotFound             = vehiclemodel.ErrVehicleSetNotFound
	ErrVehicleSetImageValidation      = vehiclemodel.ErrVehicleSetImageValidation
	ErrVehicleSetImageNotFound        = vehiclemodel.ErrVehicleSetImageNotFound
	ErrVehicleImageInUse              = vehiclemodel.ErrVehicleImageInUse
)

const (
	VehicleSetMainImageModeAutomatic = vehiclemodel.VehicleSetMainImageModeAutomatic
	VehicleSetMainImageModeMember    = vehiclemodel.VehicleSetMainImageModeMember
	VehicleSetMainImageModeDedicated = vehiclemodel.VehicleSetMainImageModeDedicated
)
