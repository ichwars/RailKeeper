package workflow

import (
	"context"
	"strings"

	"railkeeper/backend/internal/decoder"
)

func (s *CVService) Values(ctx context.Context, vehicleID string) ([]decoder.VehicleCVValue, error) {
	records, err := s.repository.CVValues(ctx, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, err
	}
	histories, err := s.repository.CVHistories(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	return CVValuesWithHistory(records, histories), nil
}

func CVValuesWithHistory(records []decoder.CVRecord, histories []decoder.VehicleCVValueHistory) []decoder.VehicleCVValue {
	out := make([]decoder.VehicleCVValue, len(records))
	for i, item := range records {
		out[i] = decoder.CVRecordValue(item)
	}
	positions := make(map[string]int, len(out))
	for i := range out {
		positions[out[i].ID] = i
		out[i].History = []decoder.VehicleCVValueHistory{}
	}
	for _, h := range histories {
		if i, ok := positions[h.CVValueID]; ok {
			out[i].History = append(out[i].History, h)
		}
	}
	return out
}
