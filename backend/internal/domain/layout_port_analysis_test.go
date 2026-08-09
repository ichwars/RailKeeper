package domain

import (
	"math"
	"testing"
)

func TestTransformModulePortUsesUnitPose(t *testing.T) {
	placement := modulePortPlacement("unit-a", "port-a", LayoutUnitPortTrack,
		"track:tillig-tt-modellgleis", 100, 20, 0, TrackPose{
			PositionXMM: 10, PositionYMM: 30, RotationDegrees: 90,
		})

	port := TransformModulePort(placement)
	if math.Abs(port.XMM-(-10)) > 1e-9 || math.Abs(port.YMM-130) > 1e-9 ||
		port.DirectionDegrees != 90 {
		t.Fatalf("unexpected transformed module port: %#v", port)
	}
}

func TestAnalyzeModulePortsDerivesConnectionsAndOpenPorts(t *testing.T) {
	placements := []ModulePortPlacement{
		modulePortPlacement("unit-b", "west", LayoutUnitPortTrack,
			"track:tillig-tt-modellgleis", 0, 0, 180, TrackPose{PositionXMM: 100}),
		modulePortPlacement("unit-a", "east", LayoutUnitPortTrack,
			"track:tillig-tt-modellgleis", 100, 0, 0, TrackPose{}),
		modulePortPlacement("unit-a", "power", LayoutUnitPortPower,
			"power:16v-ac", 0, 20, 180, TrackPose{}),
	}

	analysis := AnalyzeModulePorts(placements)
	if len(analysis.Connections) != 1 {
		t.Fatalf("expected one module-port connection, got %#v", analysis.Connections)
	}
	connection := analysis.Connections[0]
	if connection.UnitAID != "unit-a" || connection.PortAID != "east" ||
		connection.UnitBID != "unit-b" || connection.PortBID != "west" {
		t.Fatalf("unexpected stable connection: %#v", connection)
	}
	if countModulePortIssues(analysis.Issues, ModulePortIssueOpen) != 1 {
		t.Fatalf("expected one open port, got %#v", analysis.Issues)
	}
}

func TestAnalyzeModulePortsReportsNearbyIncompatibilityAndIgnoresArchivedPorts(t *testing.T) {
	first := modulePortPlacement("unit-a", "east", LayoutUnitPortTrack,
		"track:tillig-tt-modellgleis", 0, 0, 0, TrackPose{})
	second := modulePortPlacement("unit-b", "west", LayoutUnitPortPower,
		"power:16v-ac", 0, 0, 180, TrackPose{PositionXMM: 10})
	archived := modulePortPlacement("unit-c", "old", LayoutUnitPortTrack,
		"track:tillig-tt-modellgleis", 0, 0, 180, TrackPose{})
	archived.Archived = true

	analysis := AnalyzeModulePorts([]ModulePortPlacement{first, second, archived})
	if countModulePortIssues(analysis.Issues, ModulePortIssueIncompatible) != 1 {
		t.Fatalf("expected one incompatible issue, got %#v", analysis.Issues)
	}
	if countModulePortIssues(analysis.Issues, ModulePortIssueOpen) != 2 {
		t.Fatalf("expected two active open ports, got %#v", analysis.Issues)
	}
	for _, issue := range analysis.Issues {
		for _, portID := range issue.PortIDs {
			if portID == "old" {
				t.Fatalf("archived port included in analysis: %#v", analysis)
			}
		}
	}
}

func TestFindModulePortSnapUsesNearestCompatiblePortAndStableTieBreak(t *testing.T) {
	placements := []ModulePortPlacement{
		modulePortPlacement("moving", "east", LayoutUnitPortTrack,
			"track:tillig-tt-modellgleis", 100, 0, 0, TrackPose{PositionXMM: 5}),
		modulePortPlacement("target-b", "west-b", LayoutUnitPortTrack,
			"track:tillig-tt-modellgleis", 0, 0, 180, TrackPose{PositionXMM: 112}),
		modulePortPlacement("target-a", "west-a", LayoutUnitPortTrack,
			"track:tillig-tt-modellgleis", 0, 0, 180, TrackPose{PositionXMM: 112}),
	}

	snap := FindModulePortSnap("moving", TrackPose{PositionXMM: 5}, placements)
	if !snap.Snapped || snap.MovingPortID != "east" || snap.TargetUnitID != "target-a" ||
		snap.TargetPortID != "west-a" {
		t.Fatalf("unexpected module-port snap: %#v", snap)
	}
	if math.Abs(snap.Pose.PositionXMM-12) > 1e-9 || math.Abs(snap.Pose.PositionYMM) > 1e-9 ||
		math.Abs(snap.Pose.RotationDegrees) > 1e-9 {
		t.Fatalf("unexpected snapped pose: %#v", snap.Pose)
	}
}

func TestFindModulePortSnapHonorsCompatibilityDistanceAndDirectionBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		targetX   float64
		direction float64
		kind      LayoutUnitPortKind
		snapped   bool
	}{
		{name: "distance boundary", targetX: 125, direction: 180, kind: LayoutUnitPortTrack, snapped: true},
		{name: "outside distance", targetX: 125.01, direction: 180, kind: LayoutUnitPortTrack},
		{name: "direction boundary", targetX: 110, direction: 190, kind: LayoutUnitPortTrack, snapped: true},
		{name: "outside direction", targetX: 110, direction: 190.01, kind: LayoutUnitPortTrack},
		{name: "wrong kind", targetX: 110, direction: 180, kind: LayoutUnitPortPower},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			placements := []ModulePortPlacement{
				modulePortPlacement("moving", "east", LayoutUnitPortTrack,
					"track:tillig-tt-modellgleis", 100, 0, 0, TrackPose{}),
				modulePortPlacement("target", "west", test.kind,
					"track:tillig-tt-modellgleis", 0, 0, test.direction, TrackPose{PositionXMM: test.targetX}),
			}
			if got := FindModulePortSnap("moving", TrackPose{}, placements); got.Snapped != test.snapped {
				t.Fatalf("unexpected snap result: %#v", got)
			}
		})
	}
}

func modulePortPlacement(
	unitID, portID string,
	kind LayoutUnitPortKind,
	interfaceKey string,
	x, y, direction float64,
	pose TrackPose,
) ModulePortPlacement {
	return ModulePortPlacement{
		UnitID: unitID, UnitName: unitID, PortID: portID, PortName: portID,
		Kind: kind, InterfaceKey: interfaceKey, XMM: x, YMM: y,
		DirectionDegrees: direction, UnitPose: pose,
	}
}

func countModulePortIssues(issues []ModulePortIssue, code ModulePortIssueCode) int {
	count := 0
	for _, issue := range issues {
		if issue.Code == code {
			count++
		}
	}
	return count
}
