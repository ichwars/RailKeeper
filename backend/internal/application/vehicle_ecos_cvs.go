package application

import (
	"context"
	"strconv"
	"strings"
	"time"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/decoder/workflow"
)

type ECoSStoredCVSource struct {
	Host, ObjectID, ReadAt, Protocol string
	Values                           []ECoSCVValue
}
type ECoSCVImportResult = workflow.ECoSCVImportResult

func (s *VehicleService) ImportECoSStoredCVs(ctx context.Context, vehicleID string, source ECoSStoredCVSource) (ECoSCVImportResult, error) {
	result := ECoSCVImportResult{}
	values, err := validateStoredECoSCVs(source.Values)
	objectID, idErr := strconv.Atoi(source.ObjectID)
	_, timeErr := time.Parse(time.RFC3339, source.ReadAt)
	if err != nil || len(values) == 0 || idErr != nil || objectID < 1 ||
		objectID > maxDigitalCenterObjectID || timeErr != nil || strings.TrimSpace(source.Host) == "" {
		return result, ErrVehicleValidation
	}
	selection := workflow.ECoSStoredCVSource{Host: source.Host, ObjectID: source.ObjectID, ReadAt: source.ReadAt, Protocol: source.Protocol, Values: make([]decoder.VehicleCVValueInput, len(values))}
	for i, value := range values {
		selection.Values[i] = decoder.VehicleCVValueInput{CVNumber: value.Number, Value: value.Value, IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32}
	}
	v, err := s.cvService().ImportECoSStoredCVs(ctx, vehicleID, selection)
	return v, cvRuleError(err)
}
