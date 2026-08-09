package application

import (
	"context"
	"math"
	"strings"
)

type LayoutUnitOutline struct {
	LayoutUnitID string            `json:"layoutUnitId"`
	Points       []LayoutTwinPoint `json:"points"`
	Version      int               `json:"version"`
}

type UpdateLayoutUnitOutlineInput struct {
	Points          []LayoutTwinPoint `json:"points"`
	ExpectedVersion int               `json:"expectedVersion"`
}

func (s *LayoutService) UpdateUnitOutline(
	ctx context.Context,
	unitID string,
	input UpdateLayoutUnitOutlineInput,
	actor string,
) (*LayoutUnitOutline, error) {
	unitID = strings.TrimSpace(unitID)
	if unitID == "" || input.ExpectedVersion < 1 || !validLayoutOutline(input.Points) {
		return nil, ErrLayoutValidation
	}
	return s.repository.UpdateUnitOutline(ctx, unitID, input, actor)
}

func validLayoutOutline(points []LayoutTwinPoint) bool {
	if len(points) < 3 {
		return false
	}
	area := 0.0
	for index, point := range points {
		if !finite(point.XMM) || !finite(point.YMM) {
			return false
		}
		next := points[(index+1)%len(points)]
		area += point.XMM*next.YMM - next.XMM*point.YMM
	}
	return math.Abs(area) > 0.000001
}
