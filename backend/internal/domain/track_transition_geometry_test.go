package domain

import (
	"errors"
	"math"
	"testing"
)

func TestBuildTransitionTrackGeometryMirrorsDirectionsAndPreservesLimits(t *testing.T) {
	leftPath := TransitionCurvePath{
		SchemaVersion: 1, LengthMM: 500, EndRadiusMM: 700, Direction: TransitionLeft,
	}
	left, err := BuildTransitionTrackGeometry(leftPath)
	if err != nil {
		t.Fatal(err)
	}
	rightPath := leftPath
	rightPath.Direction = TransitionRight
	right, err := BuildTransitionTrackGeometry(rightPath)
	if err != nil {
		t.Fatal(err)
	}
	if left.LengthMM != 500 || left.MinimumRadiusMM == nil || *left.MinimumRadiusMM != 700 ||
		len(left.Geometry.Ports) != 2 || len(left.Geometry.Routes) != 1 {
		t.Fatalf("unexpected left transition geometry: %#v", left)
	}
	leftPoints := left.Geometry.Routes[0].Points
	rightPoints := right.Geometry.Routes[0].Points
	if len(leftPoints) < 2 || len(leftPoints) != len(rightPoints) {
		t.Fatalf("unexpected transition samples: %d / %d", len(leftPoints), len(rightPoints))
	}
	previousSegmentDirection := 0.0
	previousDirectionChange := -1.0
	for index := range leftPoints {
		if math.Abs(leftPoints[index].XMM-rightPoints[index].XMM) > 1e-9 ||
			math.Abs(leftPoints[index].YMM+rightPoints[index].YMM) > 1e-9 {
			t.Fatalf("transition sample %d is not mirrored: %#v / %#v", index,
				leftPoints[index], rightPoints[index])
		}
		if index > 0 {
			distance := trackPointDistance(leftPoints[index-1].XMM, leftPoints[index-1].YMM,
				leftPoints[index].XMM, leftPoints[index].YMM)
			if distance > 5+1e-9 {
				t.Fatalf("transition segment %d exceeds 5 mm: %.6f", index, distance)
			}
			segmentDirection := math.Atan2(leftPoints[index].YMM-leftPoints[index-1].YMM,
				leftPoints[index].XMM-leftPoints[index-1].XMM)
			if index > 1 {
				directionChange := segmentDirection - previousSegmentDirection
				if directionChange+1e-9 < previousDirectionChange {
					t.Fatalf("transition curvature decreased at segment %d: %.12f < %.12f",
						index, directionChange, previousDirectionChange)
				}
				previousDirectionChange = directionChange
			}
			previousSegmentDirection = segmentDirection
		}
	}
	expectedDirection := 500.0 / (2 * 700) * 180 / math.Pi
	if math.Abs(left.Geometry.Ports[1].DirectionDegrees-expectedDirection) > 1e-9 ||
		math.Abs(right.Geometry.Ports[1].DirectionDegrees-(360-expectedDirection)) > 1e-9 {
		t.Fatalf("unexpected end directions: %.12f / %.12f", left.Geometry.Ports[1].DirectionDegrees,
			right.Geometry.Ports[1].DirectionDegrees)
	}
}

func TestBuildTransitionTrackGeometryRejectsInvalidPaths(t *testing.T) {
	valid := TransitionCurvePath{
		SchemaVersion: 1, LengthMM: 500, EndRadiusMM: 700, Direction: TransitionLeft,
	}
	invalid := []TransitionCurvePath{
		{},
		{SchemaVersion: 2, LengthMM: 500, EndRadiusMM: 700, Direction: TransitionLeft},
		{SchemaVersion: 1, LengthMM: 0, EndRadiusMM: 700, Direction: TransitionLeft},
		{SchemaVersion: 1, LengthMM: 500, EndRadiusMM: -1, Direction: TransitionLeft},
		{SchemaVersion: 1, LengthMM: 500, EndRadiusMM: 700, Direction: "up"},
		{SchemaVersion: 1, LengthMM: math.NaN(), EndRadiusMM: 700, Direction: TransitionLeft},
		{SchemaVersion: 1, LengthMM: 500, EndRadiusMM: math.Inf(1), Direction: TransitionLeft},
		{SchemaVersion: 1, LengthMM: 30000, EndRadiusMM: 0.01, Direction: TransitionLeft},
	}
	for _, path := range invalid {
		if _, err := BuildTransitionTrackGeometry(path); !errors.Is(err, ErrInvalidTransitionCurvePath) {
			t.Fatalf("expected invalid transition error for %#v, got %v", path, err)
		}
	}
	if _, err := BuildTransitionTrackGeometry(valid); err != nil {
		t.Fatalf("valid transition rejected: %v", err)
	}
}

func TestEffectiveTrackGeometryRejectsCompetingFlexPaths(t *testing.T) {
	object := testFlexObject("flex", FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, StartHandleMM: 160, EndHandleMM: 160,
	})
	object.TransitionPath = &TransitionCurvePath{
		SchemaVersion: 1, LengthMM: 500, EndRadiusMM: 700, Direction: TransitionLeft,
	}
	if _, err := EffectiveGeometryForObject(object); !errors.Is(err, ErrInvalidFlexTrackPath) {
		t.Fatalf("expected competing flex paths to be rejected, got %v", err)
	}

	object.FlexPath = nil
	effective, err := EffectiveGeometryForObject(object)
	if err != nil || effective.LengthMM != 500 || effective.MinimumRadiusMM == nil ||
		*effective.MinimumRadiusMM != 700 {
		t.Fatalf("unexpected effective transition geometry: %#v, %v", effective, err)
	}
}
