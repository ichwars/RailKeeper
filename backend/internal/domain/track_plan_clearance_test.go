package domain

import (
	"math"
	"reflect"
	"testing"
)

func TestAnalyzeTrackClearanceInterpolatesCrossingElevations(t *testing.T) {
	lower := testG1Object("lower", 0, 0, 0)
	upper := testG1Object("upper", 83, -83, 90)
	lower.ElevationStartMM, lower.ElevationEndMM = 0, 20
	upper.ElevationStartMM, upper.ElevationEndMM = 45, 45
	limit := 40.0

	issues := filterTrackIssues(AnalyzeTrackPlanWithLimits(
		[]PlanTrackObject{upper, lower},
		TrackPlanLimits{MinimumTrackClearanceMM: &limit},
	).Issues, TrackPlanIssueInsufficientClearance)
	if len(issues) != 1 {
		t.Fatalf("expected one clearance issue, got %#v", issues)
	}
	issue := issues[0]
	if !reflect.DeepEqual(issue.ObjectIDs, []string{"lower", "upper"}) ||
		issue.ClearanceMM == nil || math.Abs(*issue.ClearanceMM-35) > 1e-9 ||
		issue.ClearanceLimitMM == nil || *issue.ClearanceLimitMM != 40 ||
		issue.IntersectionXMM == nil || math.Abs(*issue.IntersectionXMM-83) > 1e-9 ||
		issue.IntersectionYMM == nil || math.Abs(*issue.IntersectionYMM) > 1e-9 {
		t.Fatalf("unexpected clearance issue: %#v", issue)
	}
}

func TestAnalyzeTrackClearanceHonorsBoundaryAndSkipsNonCrossings(t *testing.T) {
	base := testG1Object("base", 0, 0, 0)
	boundary := testG1Object("boundary", 83, -83, 90)
	boundary.ElevationStartMM, boundary.ElevationEndMM = 40, 40
	limit := 40.0

	if issues := clearanceIssues([]PlanTrackObject{base, boundary}, &limit); len(issues) != 0 {
		t.Fatalf("boundary produced clearance issue: %#v", issues)
	}
	boundary.ElevationStartMM, boundary.ElevationEndMM = 20, 20
	if issues := clearanceIssues([]PlanTrackObject{base, boundary}, nil); len(issues) != 0 {
		t.Fatalf("unset limit produced clearance issue: %#v", issues)
	}

	endpoint := testG1Object("endpoint", 166, 0, 90)
	endpoint.ElevationStartMM, endpoint.ElevationEndMM = 20, 20
	if issues := clearanceIssues([]PlanTrackObject{base, endpoint}, &limit); len(issues) != 0 {
		t.Fatalf("endpoint connection produced clearance issue: %#v", issues)
	}

	collinear := testG1Object("collinear", 80, 0, 0)
	collinear.ElevationStartMM, collinear.ElevationEndMM = 20, 20
	if issues := clearanceIssues([]PlanTrackObject{base, collinear}, &limit); len(issues) != 0 {
		t.Fatalf("collinear overlap produced clearance issue: %#v", issues)
	}
}

func TestAnalyzeTrackClearanceSupportsReversedRouteAndSkipsAmbiguousGeometry(t *testing.T) {
	lower := testG1Object("lower", 0, 0, 0)
	lower.ElevationStartMM, lower.ElevationEndMM = 0, 20
	lower.Geometry.Geometry.Routes[0].Points = []TrackPoint{{XMM: 166}, {XMM: 0}}
	upper := testG1Object("upper", 41.5, -83, 90)
	upper.ElevationStartMM, upper.ElevationEndMM = 50, 50
	limit := 46.0

	issues := clearanceIssues([]PlanTrackObject{lower, upper}, &limit)
	if len(issues) != 1 || issues[0].ClearanceMM == nil ||
		math.Abs(*issues[0].ClearanceMM-45) > 1e-9 {
		t.Fatalf("reversed route was not interpolated from its ports: %#v", issues)
	}

	for _, mutate := range []func(*PlanTrackObject){
		func(object *PlanTrackObject) {
			object.Geometry.Geometry.Ports = append(object.Geometry.Geometry.Ports,
				TrackPort{ID: "branch", XMM: 83, YMM: 20, DirectionDegrees: 90})
		},
		func(object *PlanTrackObject) {
			object.Geometry.Geometry.Routes = append(object.Geometry.Geometry.Routes,
				TrackRoute{ID: "branch", Points: []TrackPoint{{XMM: 0}, {XMM: 83, YMM: 20}}})
		},
		func(object *PlanTrackObject) {
			object.Geometry.Geometry.Routes[0].Points[0] = TrackPoint{XMM: 10}
		},
	} {
		ambiguous := lower
		mutate(&ambiguous)
		if issues := clearanceIssues([]PlanTrackObject{ambiguous, upper}, &limit); len(issues) != 0 {
			t.Fatalf("ambiguous geometry produced clearance issue: %#v", issues)
		}
	}
}

func TestAnalyzeTrackClearanceKeepsWorstCrossingPerObjectPair(t *testing.T) {
	lower := testG1Object("lower", 0, 0, 0)
	lower.Geometry.Geometry.Routes[0].Points = []TrackPoint{
		{XMM: 0}, {XMM: 50, YMM: 50}, {XMM: 100, YMM: -50}, {XMM: 166},
	}
	lower.ElevationStartMM, lower.ElevationEndMM = 0, 30
	upper := testG1Object("upper", 0, 0, 0)
	upper.Geometry.Geometry.Ports = []TrackPort{
		{ID: "a", XMM: 0, YMM: 20, DirectionDegrees: 180},
		{ID: "b", XMM: 166, YMM: 20, DirectionDegrees: 0},
	}
	upper.Geometry.Geometry.Routes[0].Points = []TrackPoint{{XMM: 0, YMM: 20}, {XMM: 166, YMM: 20}}
	upper.ElevationStartMM, upper.ElevationEndMM = 45, 45
	limit := 40.0

	issues := clearanceIssues([]PlanTrackObject{upper, lower}, &limit)
	if len(issues) != 1 || issues[0].ClearanceMM == nil {
		t.Fatalf("expected one worst-crossing issue, got %#v", issues)
	}
}

func TestAnalyzeTrackClearanceUsesEffectiveFlexRoute(t *testing.T) {
	flex := testFlexObject("flex", FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, EndYMM: 100,
		StartHandleMM: 180, EndHandleMM: 180,
	})
	flex.ElevationStartMM, flex.ElevationEndMM = 0, 20
	crossing := testG1Object("crossing", 250, -50, 90)
	crossing.ElevationStartMM, crossing.ElevationEndMM = 45, 45
	limit := 40.0

	issues := clearanceIssues([]PlanTrackObject{crossing, flex}, &limit)
	if len(issues) != 1 || issues[0].ClearanceMM == nil ||
		issues[0].IntersectionXMM == nil || issues[0].IntersectionYMM == nil {
		t.Fatalf("effective flex crossing not analyzed: %#v", issues)
	}
}

func clearanceIssues(objects []PlanTrackObject, limit *float64) []TrackPlanIssue {
	return filterTrackIssues(AnalyzeTrackPlanWithLimits(objects,
		TrackPlanLimits{MinimumTrackClearanceMM: limit}).Issues,
		TrackPlanIssueInsufficientClearance)
}
