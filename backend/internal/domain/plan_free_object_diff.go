package domain

import (
	"math"
	"sort"
)

type PlanFreeObjectChange struct {
	Type      TrackPlanObjectChangeType `json:"type"`
	LineageID string                    `json:"lineageId"`
	Before    *PlanFreeObject           `json:"before,omitempty"`
	After     *PlanFreeObject           `json:"after,omitempty"`
}

func CompareFreePlanObjectRevisions(base, current []PlanFreeObject) []PlanFreeObjectChange {
	baseByLineage := freePlanObjectsByLineage(base)
	currentByLineage := freePlanObjectsByLineage(current)
	lineages := make(map[string]struct{}, len(baseByLineage)+len(currentByLineage))
	for lineageID := range baseByLineage {
		lineages[lineageID] = struct{}{}
	}
	for lineageID := range currentByLineage {
		lineages[lineageID] = struct{}{}
	}
	ordered := make([]string, 0, len(lineages))
	for lineageID := range lineages {
		ordered = append(ordered, lineageID)
	}
	sort.Strings(ordered)

	changes := []PlanFreeObjectChange{}
	for _, lineageID := range ordered {
		before, existedBefore := baseByLineage[lineageID]
		after, existsNow := currentByLineage[lineageID]
		switch {
		case !existedBefore:
			afterCopy := after
			changes = append(changes, PlanFreeObjectChange{
				Type: TrackPlanObjectAdded, LineageID: lineageID, After: &afterCopy,
			})
		case !existsNow:
			beforeCopy := before
			changes = append(changes, PlanFreeObjectChange{
				Type: TrackPlanObjectRemoved, LineageID: lineageID, Before: &beforeCopy,
			})
		case freePlanObjectsDiffer(before, after):
			beforeCopy, afterCopy := before, after
			changes = append(changes, PlanFreeObjectChange{
				Type: TrackPlanObjectChanged, LineageID: lineageID, Before: &beforeCopy, After: &afterCopy,
			})
		}
	}
	return changes
}

func freePlanObjectsByLineage(objects []PlanFreeObject) map[string]PlanFreeObject {
	result := make(map[string]PlanFreeObject, len(objects))
	for _, object := range objects {
		lineageID := object.LineageID
		if lineageID == "" {
			lineageID = object.ID
		}
		result[lineageID] = object
	}
	return result
}

func freePlanObjectsDiffer(first, second PlanFreeObject) bool {
	return first.Name != second.Name || first.Category != second.Category ||
		math.Abs(first.PositionXMM-second.PositionXMM) > 1e-9 ||
		math.Abs(first.PositionYMM-second.PositionYMM) > 1e-9 ||
		math.Abs(first.RotationDegrees-second.RotationDegrees) > 1e-9 ||
		freePlanShapesDiffer(first.Shape, second.Shape)
}

func freePlanShapesDiffer(first, second FreePlanObjectShape) bool {
	return first.SchemaVersion != second.SchemaVersion || first.Kind != second.Kind ||
		first.Text != second.Text || optionalFreeFloatsDiffer(first.WidthMM, second.WidthMM) ||
		optionalFreeFloatsDiffer(first.HeightMM, second.HeightMM) ||
		optionalFreeFloatsDiffer(first.EndXMM, second.EndXMM) ||
		optionalFreeFloatsDiffer(first.EndYMM, second.EndYMM) ||
		optionalFreeFloatsDiffer(first.FontSizeMM, second.FontSizeMM)
}

func optionalFreeFloatsDiffer(first, second *float64) bool {
	if first == nil || second == nil {
		return first != second
	}
	return math.Abs(*first-*second) > 1e-9
}
