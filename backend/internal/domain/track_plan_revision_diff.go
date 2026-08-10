package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type TrackPlanObjectChangeType string

const (
	TrackPlanObjectAdded   TrackPlanObjectChangeType = "added"
	TrackPlanObjectRemoved TrackPlanObjectChangeType = "removed"
	TrackPlanObjectChanged TrackPlanObjectChangeType = "changed"
)

type TrackPlanObjectChange struct {
	Type      TrackPlanObjectChangeType `json:"type"`
	LineageID string                    `json:"lineageId"`
	Before    *PlanTrackObject          `json:"before,omitempty"`
	After     *PlanTrackObject          `json:"after,omitempty"`
}

type TrackPlanMaterialDelta struct {
	GeometryID      string `json:"geometryId"`
	LibraryID       string `json:"libraryId"`
	ArticleNumber   string `json:"articleNumber"`
	Name            string `json:"name"`
	BaseQuantity    int    `json:"baseQuantity"`
	CurrentQuantity int    `json:"currentQuantity"`
	Delta           int    `json:"delta"`
}

type TrackPlanRevisionDiff struct {
	ObjectChanges  []TrackPlanObjectChange  `json:"objectChanges"`
	MaterialDeltas []TrackPlanMaterialDelta `json:"materialDeltas"`
}

type TrackPlanIssueChange struct {
	Code       TrackPlanIssueCode     `json:"code"`
	Severity   TrackPlanIssueSeverity `json:"severity"`
	LineageIDs []string               `json:"lineageIds"`
	PortIDs    []string               `json:"portIds,omitempty"`
}

type TrackPlanIssueDiff struct {
	Added    []TrackPlanIssueChange `json:"added"`
	Resolved []TrackPlanIssueChange `json:"resolved"`
}

func CompareTrackPlanRevisions(base, current []PlanTrackObject) TrackPlanRevisionDiff {
	baseByLineage := trackObjectsByLineage(base)
	currentByLineage := trackObjectsByLineage(current)
	lineages := map[string]struct{}{}
	for lineageID := range baseByLineage {
		lineages[lineageID] = struct{}{}
	}
	for lineageID := range currentByLineage {
		lineages[lineageID] = struct{}{}
	}
	orderedLineages := make([]string, 0, len(lineages))
	for lineageID := range lineages {
		orderedLineages = append(orderedLineages, lineageID)
	}
	sort.Strings(orderedLineages)

	diff := TrackPlanRevisionDiff{
		ObjectChanges: []TrackPlanObjectChange{}, MaterialDeltas: trackMaterialDeltas(base, current),
	}
	for _, lineageID := range orderedLineages {
		before, existedBefore := baseByLineage[lineageID]
		after, existsNow := currentByLineage[lineageID]
		switch {
		case !existedBefore:
			afterCopy := after
			diff.ObjectChanges = append(diff.ObjectChanges, TrackPlanObjectChange{
				Type: TrackPlanObjectAdded, LineageID: lineageID, After: &afterCopy,
			})
		case !existsNow:
			beforeCopy := before
			diff.ObjectChanges = append(diff.ObjectChanges, TrackPlanObjectChange{
				Type: TrackPlanObjectRemoved, LineageID: lineageID, Before: &beforeCopy,
			})
		case trackObjectsDiffer(before, after):
			beforeCopy, afterCopy := before, after
			diff.ObjectChanges = append(diff.ObjectChanges, TrackPlanObjectChange{
				Type: TrackPlanObjectChanged, LineageID: lineageID,
				Before: &beforeCopy, After: &afterCopy,
			})
		}
	}
	return diff
}

func DiffTrackPlanIssues(
	baseIssues, currentIssues []TrackPlanIssue,
	baseObjects, currentObjects []PlanTrackObject,
) TrackPlanIssueDiff {
	base := normalizedTrackIssues(baseIssues, baseObjects)
	current := normalizedTrackIssues(currentIssues, currentObjects)
	diff := TrackPlanIssueDiff{Added: []TrackPlanIssueChange{}, Resolved: []TrackPlanIssueChange{}}
	for key, issue := range current {
		if _, exists := base[key]; !exists {
			diff.Added = append(diff.Added, issue)
		}
	}
	for key, issue := range base {
		if _, exists := current[key]; !exists {
			diff.Resolved = append(diff.Resolved, issue)
		}
	}
	sortTrackIssueChanges(diff.Added)
	sortTrackIssueChanges(diff.Resolved)
	return diff
}

func trackObjectsByLineage(objects []PlanTrackObject) map[string]PlanTrackObject {
	result := make(map[string]PlanTrackObject, len(objects))
	for _, object := range objects {
		lineageID := object.LineageID
		if lineageID == "" {
			lineageID = object.ID
		}
		result[lineageID] = object
	}
	return result
}

func trackObjectsDiffer(first, second PlanTrackObject) bool {
	return first.GeometryID != second.GeometryID ||
		math.Abs(first.PositionXMM-second.PositionXMM) > 1e-9 ||
		math.Abs(first.PositionYMM-second.PositionYMM) > 1e-9 ||
		math.Abs(first.RotationDegrees-second.RotationDegrees) > 1e-9 ||
		math.Abs(first.ElevationStartMM-second.ElevationStartMM) > 1e-9 ||
		math.Abs(first.ElevationEndMM-second.ElevationEndMM) > 1e-9 ||
		flexTrackPathsDiffer(first.FlexPath, second.FlexPath) ||
		transitionCurvePathsDiffer(first.TransitionPath, second.TransitionPath)
}

func transitionCurvePathsDiffer(first, second *TransitionCurvePath) bool {
	if first == nil || second == nil {
		return first != second
	}
	return first.SchemaVersion != second.SchemaVersion || first.Direction != second.Direction ||
		math.Abs(first.LengthMM-second.LengthMM) > 1e-9 ||
		math.Abs(first.EndRadiusMM-second.EndRadiusMM) > 1e-9
}

func flexTrackPathsDiffer(first, second *FlexTrackPath) bool {
	if first == nil || second == nil {
		return first != second
	}
	return first.SchemaVersion != second.SchemaVersion ||
		math.Abs(first.EndXMM-second.EndXMM) > 1e-9 ||
		math.Abs(first.EndYMM-second.EndYMM) > 1e-9 ||
		math.Abs(first.EndDirectionDegrees-second.EndDirectionDegrees) > 1e-9 ||
		math.Abs(first.StartHandleMM-second.StartHandleMM) > 1e-9 ||
		math.Abs(first.EndHandleMM-second.EndHandleMM) > 1e-9
}

func trackMaterialDeltas(base, current []PlanTrackObject) []TrackPlanMaterialDelta {
	type quantities struct {
		line    TrackPlanMaterialDelta
		base    int
		current int
	}
	byGeometry := map[string]*quantities{}
	add := func(objects []PlanTrackObject, currentRevision bool) {
		for _, object := range objects {
			entry := byGeometry[object.GeometryID]
			if entry == nil {
				entry = &quantities{line: TrackPlanMaterialDelta{
					GeometryID: object.GeometryID, LibraryID: object.Geometry.LibraryID,
					ArticleNumber: object.Geometry.ArticleNumber, Name: object.Geometry.Name,
				}}
				byGeometry[object.GeometryID] = entry
			}
			if currentRevision {
				entry.current++
				entry.line.LibraryID = object.Geometry.LibraryID
				entry.line.ArticleNumber = object.Geometry.ArticleNumber
				entry.line.Name = object.Geometry.Name
			} else {
				entry.base++
			}
		}
	}
	add(base, false)
	add(current, true)
	geometryIDs := make([]string, 0, len(byGeometry))
	for geometryID := range byGeometry {
		geometryIDs = append(geometryIDs, geometryID)
	}
	sort.Strings(geometryIDs)
	result := make([]TrackPlanMaterialDelta, 0, len(geometryIDs))
	for _, geometryID := range geometryIDs {
		entry := byGeometry[geometryID]
		entry.line.BaseQuantity = entry.base
		entry.line.CurrentQuantity = entry.current
		entry.line.Delta = entry.current - entry.base
		if entry.line.Delta != 0 {
			result = append(result, entry.line)
		}
	}
	return result
}

func normalizedTrackIssues(
	issues []TrackPlanIssue,
	objects []PlanTrackObject,
) map[string]TrackPlanIssueChange {
	lineageByID := map[string]string{}
	for _, object := range objects {
		lineageID := object.LineageID
		if lineageID == "" {
			lineageID = object.ID
		}
		lineageByID[object.ID] = lineageID
	}
	result := map[string]TrackPlanIssueChange{}
	for _, issue := range issues {
		parts := make([]string, 0, len(issue.ObjectIDs))
		for index, objectID := range issue.ObjectIDs {
			lineageID := lineageByID[objectID]
			if lineageID == "" {
				lineageID = objectID
			}
			portID := ""
			if index < len(issue.PortIDs) {
				portID = issue.PortIDs[index]
			}
			parts = append(parts, lineageID+":"+portID)
		}
		sort.Strings(parts)
		change := TrackPlanIssueChange{Code: issue.Code, Severity: issue.Severity,
			LineageIDs: make([]string, 0, len(parts)), PortIDs: make([]string, 0, len(parts))}
		for _, part := range parts {
			lineageID, portID, _ := strings.Cut(part, ":")
			change.LineageIDs = append(change.LineageIDs, lineageID)
			if portID != "" {
				change.PortIDs = append(change.PortIDs, portID)
			}
		}
		key := fmt.Sprintf("%s|%s|%s", issue.Code, issue.Severity, strings.Join(parts, ","))
		result[key] = change
	}
	return result
}

func sortTrackIssueChanges(changes []TrackPlanIssueChange) {
	sort.Slice(changes, func(i, j int) bool {
		first := fmt.Sprintf("%s|%s|%s|%s", changes[i].Code, changes[i].Severity,
			strings.Join(changes[i].LineageIDs, ","), strings.Join(changes[i].PortIDs, ","))
		second := fmt.Sprintf("%s|%s|%s|%s", changes[j].Code, changes[j].Severity,
			strings.Join(changes[j].LineageIDs, ","), strings.Join(changes[j].PortIDs, ","))
		return first < second
	})
}
