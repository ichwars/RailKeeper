package ecosmodel

import (
	"fmt"

	"railkeeper/backend/internal/decoder"
)

func ValidateStoredCVs(values []ECoSCVValue, maximumValues int) ([]ECoSCVValue, error) {
	if len(values) > maximumValues {
		return nil, decoder.ErrCVValidation
	}
	seen := map[string]bool{}
	copyValues := make([]ECoSCVValue, 0, len(values))
	for _, value := range values {
		if !decoder.ValidCVNumber(value.Number) || !decoder.ValidCVValue(value.Value) ||
			(value.IndexCV31 == nil) != (value.IndexCV32 == nil) {
			return nil, decoder.ErrCVValidation
		}
		if value.IndexCV31 != nil && (value.Number < 257 || value.Number > 512 || *value.IndexCV31 < 0 ||
			*value.IndexCV31 > 255 || *value.IndexCV32 < 0 || *value.IndexCV32 > 255) {
			return nil, decoder.ErrCVValidation
		}
		key := fmt.Sprintf("%d:%d", ECoSCVPage(value), value.Number)
		if seen[key] {
			return nil, decoder.ErrCVValidation
		}
		seen[key] = true
		copyValues = append(copyValues, value)
	}
	SortECoSCVValues(copyValues)
	return copyValues, nil
}
