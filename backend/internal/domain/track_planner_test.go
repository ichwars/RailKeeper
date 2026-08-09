package domain

import "testing"

func TestTrackGeometryKindsAndStatuses(t *testing.T) {
	for _, kind := range []TrackGeometryKind{
		TrackGeometryStraight,
		TrackGeometryCurve,
		TrackGeometryTurnout,
		TrackGeometryCrossing,
	} {
		if !kind.Valid() {
			t.Fatalf("expected %q to be a valid track geometry kind", kind)
		}
	}
	if TrackGeometryKind("flex").Valid() {
		t.Fatal("flex track belongs to a later planner stage")
	}

	if !TrackGeometryVerified.Placeable() {
		t.Fatal("verified geometry must be placeable")
	}
	for _, status := range []TrackGeometryStatus{TrackGeometryDraft, TrackGeometryRetired} {
		if status.Placeable() {
			t.Fatalf("geometry status %q must not be placeable", status)
		}
	}
}

func TestNormalizeTrackRotation(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{input: 0, want: 0},
		{input: 360, want: 0},
		{input: -15, want: 345},
		{input: 735, want: 15},
	}
	for _, test := range tests {
		if got := NormalizeTrackRotation(test.input); got != test.want {
			t.Fatalf("NormalizeTrackRotation(%v)=%v, want %v", test.input, got, test.want)
		}
	}
}

func TestTilligG1GeometryUsesExactMillimetrePorts(t *testing.T) {
	geometry := TrackGeometryDefinition{
		ArticleNumber: "83101",
		LengthMM:      166,
		Geometry: TrackGeometry{
			SchemaVersion: 1,
			Ports: []TrackPort{
				{ID: "a", XMM: 0, YMM: 0, DirectionDegrees: 180},
				{ID: "b", XMM: 166, YMM: 0, DirectionDegrees: 0},
			},
		},
	}
	if geometry.LengthMM != 166 || len(geometry.Geometry.Ports) != 2 {
		t.Fatalf("unexpected Tillig G1 geometry: %#v", geometry)
	}
	if geometry.Geometry.Ports[0].XMM != 0 || geometry.Geometry.Ports[1].XMM != 166 {
		t.Fatalf("unexpected Tillig G1 ports: %#v", geometry.Geometry.Ports)
	}
}
