package domain

import "testing"

func TestCompareTrackPlanRevisionsFindsObjectsAndMaterialDeltasByLineage(t *testing.T) {
	base := []PlanTrackObject{
		testRevisionTrack("base-a", "lineage-a", "g1", 0),
		testRevisionTrack("base-b", "lineage-b", "g1", 166),
		testRevisionTrack("base-c", "lineage-c", "g1", 332),
	}
	current := []PlanTrackObject{
		testRevisionTrack("draft-a", "lineage-a", "g1", 0),
		testRevisionTrack("draft-b", "lineage-b", "g1", 170),
		testRevisionTrack("draft-d", "lineage-d", "g2", 336),
	}

	diff := CompareTrackPlanRevisions(base, current)
	if len(diff.ObjectChanges) != 3 {
		t.Fatalf("unexpected object changes: %#v", diff.ObjectChanges)
	}
	for index, want := range []struct {
		lineage string
		change  TrackPlanObjectChangeType
	}{{"lineage-b", TrackPlanObjectChanged}, {"lineage-c", TrackPlanObjectRemoved},
		{"lineage-d", TrackPlanObjectAdded}} {
		if diff.ObjectChanges[index].LineageID != want.lineage || diff.ObjectChanges[index].Type != want.change {
			t.Fatalf("unexpected change %d: %#v", index, diff.ObjectChanges[index])
		}
	}
	if len(diff.MaterialDeltas) != 2 || diff.MaterialDeltas[0].GeometryID != "g1" ||
		diff.MaterialDeltas[0].BaseQuantity != 3 || diff.MaterialDeltas[0].CurrentQuantity != 2 ||
		diff.MaterialDeltas[0].Delta != -1 || diff.MaterialDeltas[1].GeometryID != "g2" ||
		diff.MaterialDeltas[1].Delta != 1 {
		t.Fatalf("unexpected material deltas: %#v", diff.MaterialDeltas)
	}
}

func TestCompareTrackPlanRevisionsWithoutBaseTreatsEveryObjectAsAdded(t *testing.T) {
	diff := CompareTrackPlanRevisions(nil, []PlanTrackObject{
		testRevisionTrack("draft-a", "lineage-a", "g1", 0),
		testRevisionTrack("draft-b", "lineage-b", "g1", 166),
	})
	if len(diff.ObjectChanges) != 2 || diff.ObjectChanges[0].Type != TrackPlanObjectAdded ||
		len(diff.MaterialDeltas) != 1 || diff.MaterialDeltas[0].BaseQuantity != 0 ||
		diff.MaterialDeltas[0].CurrentQuantity != 2 {
		t.Fatalf("unexpected no-base diff: %#v", diff)
	}
}

func TestDiffTrackPlanIssuesUsesLineageAndPortIdentity(t *testing.T) {
	baseObjects := []PlanTrackObject{
		testRevisionTrack("base-a", "lineage-a", "g1", 0),
		testRevisionTrack("base-b", "lineage-b", "g1", 80),
	}
	currentObjects := []PlanTrackObject{
		testRevisionTrack("draft-a", "lineage-a", "g1", 0),
		testRevisionTrack("draft-b", "lineage-b", "g1", 166),
	}
	baseIssues := []TrackPlanIssue{{
		Code: TrackPlanIssueOverlap, Severity: TrackPlanIssueWarning,
		ObjectIDs: []string{"base-a", "base-b"},
	}}
	currentIssues := []TrackPlanIssue{{
		Code: TrackPlanIssueOpenEnd, Severity: TrackPlanIssueWarning,
		ObjectIDs: []string{"draft-a"}, PortIDs: []string{"a"},
	}}

	diff := DiffTrackPlanIssues(baseIssues, currentIssues, baseObjects, currentObjects)
	if len(diff.Added) != 1 || diff.Added[0].Code != TrackPlanIssueOpenEnd ||
		len(diff.Resolved) != 1 || diff.Resolved[0].Code != TrackPlanIssueOverlap ||
		diff.Added[0].LineageIDs[0] != "lineage-a" {
		t.Fatalf("unexpected issue diff: %#v", diff)
	}
}

func testRevisionTrack(id, lineageID, geometryID string, x float64) PlanTrackObject {
	object := testG1Object(id, x, 0, 0)
	object.LineageID = lineageID
	object.GeometryID = geometryID
	object.Geometry.ID = geometryID
	object.Geometry.ArticleNumber = geometryID
	object.Geometry.Name = geometryID
	return object
}
