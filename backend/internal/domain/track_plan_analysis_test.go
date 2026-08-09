package domain

import (
	"math"
	"testing"
)

func TestTrackTransformRotatesAndTranslatesMillimetres(t *testing.T) {
	for _, test := range []struct {
		name string
		pose TrackPose
		in   TrackPoint
		want TrackPoint
	}{
		{name: "translation", pose: TrackPose{PositionXMM: 10, PositionYMM: 20},
			in: TrackPoint{XMM: 166}, want: TrackPoint{XMM: 176, YMM: 20}},
		{name: "quarter turn", pose: TrackPose{PositionXMM: 10, PositionYMM: 20, RotationDegrees: 90},
			in: TrackPoint{XMM: 166}, want: TrackPoint{XMM: 10, YMM: 186}},
		{name: "half turn", pose: TrackPose{PositionXMM: 10, PositionYMM: 20, RotationDegrees: 180},
			in: TrackPoint{XMM: 166}, want: TrackPoint{XMM: -156, YMM: 20}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := TransformTrackPoint(test.in, test.pose)
			if math.Abs(got.XMM-test.want.XMM) > 1e-9 || math.Abs(got.YMM-test.want.YMM) > 1e-9 {
				t.Fatalf("unexpected transformed point: got %#v, want %#v", got, test.want)
			}
		})
	}

	port := TransformTrackPort(TrackPort{ID: "b", XMM: 166, DirectionDegrees: 0},
		TrackPose{PositionXMM: 10, PositionYMM: 20, RotationDegrees: 90})
	if port.ID != "b" || math.Abs(port.XMM-10) > 1e-9 || math.Abs(port.YMM-186) > 1e-9 ||
		port.DirectionDegrees != 90 {
		t.Fatalf("unexpected transformed port: %#v", port)
	}
}

func TestFindTrackSnapUsesNearestCompatibleEndpoint(t *testing.T) {
	target := testG1Object("target", 0, 0, 0)
	farther := testG1Object("farther", 345, 0, 0)
	moving := testG1Object("moving", 172, 2, 2)

	snap := FindTrackSnap(moving, []PlanTrackObject{farther, target})
	if !snap.Snapped {
		t.Fatal("expected compatible endpoint to snap")
	}
	if snap.MovingPortID != "a" || snap.TargetObjectID != "target" || snap.TargetPortID != "b" {
		t.Fatalf("unexpected snap ports: %#v", snap)
	}
	if math.Abs(snap.Pose.PositionXMM-166) > 1e-9 || math.Abs(snap.Pose.PositionYMM) > 1e-9 ||
		math.Abs(snap.Pose.RotationDegrees) > 1e-9 {
		t.Fatalf("unexpected snapped pose: %#v", snap.Pose)
	}
}

func TestFindTrackSnapHonorsDistanceAndDirectionBoundaries(t *testing.T) {
	target := testG1Object("target", 0, 0, 0)
	for _, test := range []struct {
		name    string
		moving  PlanTrackObject
		snapped bool
	}{
		{name: "distance boundary", moving: testG1Object("moving", 174, 0, 0), snapped: true},
		{name: "outside distance", moving: testG1Object("moving", 174.01, 0, 0), snapped: false},
		{name: "direction boundary", moving: testG1Object("moving", 172, 0, 5), snapped: true},
		{name: "outside direction", moving: testG1Object("moving", 172, 0, 5.01), snapped: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := FindTrackSnap(test.moving, []PlanTrackObject{target}); got.Snapped != test.snapped {
				t.Fatalf("unexpected snap result: %#v", got)
			}
		})
	}
}

func TestAnalyzeTrackPlanDerivesConnectionsOpenEndsAndBOM(t *testing.T) {
	objects := []PlanTrackObject{
		testG1Object("track-2", 166, 0, 0),
		testG1Object("track-1", 0, 0, 0),
	}
	analysis := AnalyzeTrackPlan(objects)
	if len(analysis.Connections) != 1 {
		t.Fatalf("expected one connection, got %#v", analysis.Connections)
	}
	connection := analysis.Connections[0]
	if connection.ObjectAID != "track-1" || connection.PortAID != "b" ||
		connection.ObjectBID != "track-2" || connection.PortBID != "a" {
		t.Fatalf("unexpected connection: %#v", connection)
	}
	if countTrackIssues(analysis.Issues, TrackPlanIssueOpenEnd) != 2 {
		t.Fatalf("expected two open ends, got %#v", analysis.Issues)
	}
	if len(analysis.BOM) != 1 || analysis.BOM[0].GeometryID != "tillig-g1" || analysis.BOM[0].Quantity != 2 {
		t.Fatalf("unexpected BOM: %#v", analysis.BOM)
	}
}

func TestAnalyzeTrackPlanReportsIncompatibleEndsOverlapAndBrokenGeometry(t *testing.T) {
	objects := []PlanTrackObject{
		testG1Object("base", 0, 0, 0),
		testG1Object("overlap", 80, 0, 0),
		testG1Object("incompatible", 0, 6, 90),
		{ID: "broken", GeometryID: "broken", Geometry: TrackGeometryDefinition{
			ID: "broken", ArticleNumber: "X", Name: "Broken", Status: TrackGeometryVerified,
		}},
	}
	analysis := AnalyzeTrackPlan(objects)
	if countTrackIssues(analysis.Issues, TrackPlanIssueOverlap) == 0 {
		t.Fatalf("expected overlap issue, got %#v", analysis.Issues)
	}
	if countTrackIssues(analysis.Issues, TrackPlanIssueIncompatibleConnection) == 0 {
		t.Fatalf("expected incompatible connection issue, got %#v", analysis.Issues)
	}
	if countTrackIssues(analysis.Issues, TrackPlanIssueBrokenGeometry) != 1 {
		t.Fatalf("expected one broken geometry issue, got %#v", analysis.Issues)
	}
}

func testG1Object(id string, x, y, rotation float64) PlanTrackObject {
	return PlanTrackObject{
		ID: id, GeometryID: "tillig-g1", PositionXMM: x, PositionYMM: y, RotationDegrees: rotation,
		Geometry: TrackGeometryDefinition{
			ID: "tillig-g1", LibraryID: "tillig-v1", ArticleNumber: "83101", Name: "Gleisstück G1",
			Kind: TrackGeometryStraight, LengthMM: 166, Status: TrackGeometryVerified,
			Geometry: TrackGeometry{
				SchemaVersion: 1,
				Ports: []TrackPort{
					{ID: "a", XMM: 0, YMM: 0, DirectionDegrees: 180},
					{ID: "b", XMM: 166, YMM: 0, DirectionDegrees: 0},
				},
				Routes: []TrackRoute{{ID: "main", Points: []TrackPoint{{XMM: 0, YMM: 0}, {XMM: 166, YMM: 0}}}},
			},
		},
	}
}

func countTrackIssues(issues []TrackPlanIssue, code TrackPlanIssueCode) int {
	count := 0
	for _, issue := range issues {
		if issue.Code == code {
			count++
		}
	}
	return count
}
