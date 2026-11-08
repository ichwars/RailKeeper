package ecosmodel

import (
	"strconv"
	"strings"
	"time"

	"railkeeper/backend/internal/decoder"
)

type ECoSStoredCVSource struct {
	Host, ObjectID, ReadAt, Protocol string
	Values                           []ECoSCVValue
}
type StoredCVLimits struct{ MaximumValues, MaximumObjectID int }
type StoredCVAdoption struct {
	Host, ObjectID, ReadAt, Protocol string
	Values                           []decoder.VehicleCVValueInput
}

func PlanStoredCVAdoption(source ECoSStoredCVSource, limits StoredCVLimits) (StoredCVAdoption, error) {
	values, err := ValidateStoredCVs(source.Values, limits.MaximumValues)
	objectID, idErr := strconv.Atoi(source.ObjectID)
	_, timeErr := time.Parse(time.RFC3339, source.ReadAt)
	if err != nil || len(values) == 0 || idErr != nil || objectID < 1 ||
		objectID > limits.MaximumObjectID || timeErr != nil || strings.TrimSpace(source.Host) == "" {
		return StoredCVAdoption{}, decoder.ErrCVValidation
	}
	selection := StoredCVAdoption{Host: source.Host, ObjectID: source.ObjectID, ReadAt: source.ReadAt, Protocol: source.Protocol, Values: make([]decoder.VehicleCVValueInput, len(values))}
	for i, value := range values {
		selection.Values[i] = decoder.VehicleCVValueInput{CVNumber: value.Number, Value: value.Value, IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32}
	}
	return selection, nil
}
