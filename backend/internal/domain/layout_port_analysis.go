package domain

import (
	"math"
	"sort"
	"strings"
)

const (
	ModulePortSnapDistanceMM                = 25.0
	ModulePortSnapDirectionToleranceDegrees = 10.0
	ModulePortConnectionDistanceMM          = 0.25
	ModulePortConnectionDirectionDegrees    = 0.5
)

type ModulePortPlacement struct {
	UnitID           string             `json:"unitId"`
	UnitName         string             `json:"unitName"`
	PortID           string             `json:"portId"`
	PortName         string             `json:"portName"`
	Kind             LayoutUnitPortKind `json:"kind"`
	InterfaceKey     string             `json:"interfaceKey"`
	XMM              float64            `json:"xMm"`
	YMM              float64            `json:"yMm"`
	DirectionDegrees float64            `json:"directionDegrees"`
	Archived         bool               `json:"archived"`
	UnitPose         TrackPose          `json:"unitPose"`
}

type TransformedModulePort struct {
	ModulePortPlacement
	XMM              float64 `json:"xMm"`
	YMM              float64 `json:"yMm"`
	DirectionDegrees float64 `json:"directionDegrees"`
}

type ModulePortConnection struct {
	UnitAID   string `json:"unitAId"`
	UnitAName string `json:"unitAName"`
	PortAID   string `json:"portAId"`
	PortAName string `json:"portAName"`
	UnitBID   string `json:"unitBId"`
	UnitBName string `json:"unitBName"`
	PortBID   string `json:"portBId"`
	PortBName string `json:"portBName"`
}

type ModulePortIssueCode string

const (
	ModulePortIssueOpen         ModulePortIssueCode = "open_port"
	ModulePortIssueIncompatible ModulePortIssueCode = "incompatible_port"
)

type ModulePortIssue struct {
	Code      ModulePortIssueCode `json:"code"`
	UnitIDs   []string            `json:"unitIds"`
	UnitNames []string            `json:"unitNames"`
	PortIDs   []string            `json:"portIds"`
	PortNames []string            `json:"portNames"`
}

type ModulePortAnalysis struct {
	Connections []ModulePortConnection `json:"connections"`
	Issues      []ModulePortIssue      `json:"issues"`
}

type ModulePortSnapResult struct {
	Snapped      bool      `json:"snapped"`
	Pose         TrackPose `json:"pose"`
	MovingPortID string    `json:"movingPortId,omitempty"`
	TargetUnitID string    `json:"targetUnitId,omitempty"`
	TargetPortID string    `json:"targetPortId,omitempty"`
	DistanceMM   float64   `json:"distanceMm,omitempty"`
}

func TransformModulePort(placement ModulePortPlacement) TransformedModulePort {
	point := TransformTrackPoint(TrackPoint{XMM: placement.XMM, YMM: placement.YMM}, placement.UnitPose)
	return TransformedModulePort{
		ModulePortPlacement: placement,
		XMM:                 point.XMM, YMM: point.YMM,
		DirectionDegrees: NormalizeTrackRotation(placement.DirectionDegrees + placement.UnitPose.RotationDegrees),
	}
}

func AnalyzeModulePorts(placements []ModulePortPlacement) ModulePortAnalysis {
	ports := activeTransformedModulePorts(placements)
	analysis := ModulePortAnalysis{Connections: []ModulePortConnection{}, Issues: []ModulePortIssue{}}
	connected := map[string]bool{}
	for i := 0; i < len(ports); i++ {
		for j := i + 1; j < len(ports); j++ {
			first, second := ports[i], ports[j]
			if first.UnitID == second.UnitID {
				continue
			}
			distance := trackPointDistance(first.XMM, first.YMM, second.XMM, second.YMM)
			direction := trackOpposingAngleDifference(first.DirectionDegrees, second.DirectionDegrees)
			compatible := modulePortsCompatible(first.ModulePortPlacement, second.ModulePortPlacement)
			if compatible && distance <= ModulePortConnectionDistanceMM &&
				direction <= ModulePortConnectionDirectionDegrees &&
				!connected[modulePortKey(first)] && !connected[modulePortKey(second)] {
				analysis.Connections = append(analysis.Connections, ModulePortConnection{
					UnitAID: first.UnitID, UnitAName: first.UnitName, PortAID: first.PortID, PortAName: first.PortName,
					UnitBID: second.UnitID, UnitBName: second.UnitName, PortBID: second.PortID, PortBName: second.PortName,
				})
				connected[modulePortKey(first)] = true
				connected[modulePortKey(second)] = true
			} else if !compatible && distance <= ModulePortSnapDistanceMM &&
				direction <= ModulePortSnapDirectionToleranceDegrees {
				analysis.Issues = append(analysis.Issues, modulePortIssue(ModulePortIssueIncompatible, first, second))
			}
		}
	}
	for _, port := range ports {
		if !connected[modulePortKey(port)] {
			analysis.Issues = append(analysis.Issues, modulePortIssue(ModulePortIssueOpen, port))
		}
	}
	sortModulePortAnalysis(&analysis)
	return analysis
}

func FindModulePortSnap(movingUnitID string, pose TrackPose, placements []ModulePortPlacement) ModulePortSnapResult {
	pose.RotationDegrees = NormalizeTrackRotation(pose.RotationDegrees)
	result := ModulePortSnapResult{Pose: pose}
	ports := activeTransformedModulePorts(placements)
	for _, moving := range placements {
		if moving.Archived || moving.UnitID != movingUnitID {
			continue
		}
		moving.UnitPose = pose
		current := TransformModulePort(moving)
		for _, target := range ports {
			if target.UnitID == movingUnitID || !modulePortsCompatible(moving, target.ModulePortPlacement) {
				continue
			}
			distance := trackPointDistance(current.XMM, current.YMM, target.XMM, target.YMM)
			direction := trackOpposingAngleDifference(current.DirectionDegrees, target.DirectionDegrees)
			if distance > ModulePortSnapDistanceMM || direction > ModulePortSnapDirectionToleranceDegrees {
				continue
			}
			if result.Snapped && !betterModulePortSnap(distance, target.UnitID, target.PortID, result) {
				continue
			}
			desiredRotation := NormalizeTrackRotation(target.DirectionDegrees + 180 - moving.DirectionDegrees)
			rotated := TransformTrackPoint(TrackPoint{XMM: moving.XMM, YMM: moving.YMM}, TrackPose{
				RotationDegrees: desiredRotation,
			})
			result = ModulePortSnapResult{
				Snapped: true,
				Pose: TrackPose{
					PositionXMM: target.XMM - rotated.XMM, PositionYMM: target.YMM - rotated.YMM,
					RotationDegrees: desiredRotation,
				},
				MovingPortID: moving.PortID, TargetUnitID: target.UnitID,
				TargetPortID: target.PortID, DistanceMM: distance,
			}
		}
	}
	return result
}

func activeTransformedModulePorts(placements []ModulePortPlacement) []TransformedModulePort {
	ports := make([]TransformedModulePort, 0, len(placements))
	for _, placement := range placements {
		if !placement.Archived {
			ports = append(ports, TransformModulePort(placement))
		}
	}
	sort.Slice(ports, func(i, j int) bool { return modulePortKey(ports[i]) < modulePortKey(ports[j]) })
	return ports
}

func modulePortsCompatible(first, second ModulePortPlacement) bool {
	return first.Kind == second.Kind && first.InterfaceKey != "" &&
		strings.EqualFold(first.InterfaceKey, second.InterfaceKey)
}

func modulePortIssue(code ModulePortIssueCode, ports ...TransformedModulePort) ModulePortIssue {
	issue := ModulePortIssue{Code: code, UnitIDs: []string{}, UnitNames: []string{},
		PortIDs: []string{}, PortNames: []string{}}
	for _, port := range ports {
		issue.UnitIDs = append(issue.UnitIDs, port.UnitID)
		issue.UnitNames = append(issue.UnitNames, port.UnitName)
		issue.PortIDs = append(issue.PortIDs, port.PortID)
		issue.PortNames = append(issue.PortNames, port.PortName)
	}
	return issue
}

func modulePortKey(port TransformedModulePort) string {
	return port.UnitID + "\x00" + port.PortID
}

func betterModulePortSnap(distance float64, unitID, portID string, current ModulePortSnapResult) bool {
	if distance < current.DistanceMM-1e-9 {
		return true
	}
	if math.Abs(distance-current.DistanceMM) > 1e-9 {
		return false
	}
	return unitID+"\x00"+portID < current.TargetUnitID+"\x00"+current.TargetPortID
}

func sortModulePortAnalysis(analysis *ModulePortAnalysis) {
	sort.Slice(analysis.Connections, func(i, j int) bool {
		left := analysis.Connections[i].UnitAID + "\x00" + analysis.Connections[i].PortAID + "\x00" +
			analysis.Connections[i].UnitBID + "\x00" + analysis.Connections[i].PortBID
		right := analysis.Connections[j].UnitAID + "\x00" + analysis.Connections[j].PortAID + "\x00" +
			analysis.Connections[j].UnitBID + "\x00" + analysis.Connections[j].PortBID
		return left < right
	})
	sort.SliceStable(analysis.Issues, func(i, j int) bool {
		left := string(analysis.Issues[i].Code) + strings.Join(analysis.Issues[i].UnitIDs, "\x00") +
			strings.Join(analysis.Issues[i].PortIDs, "\x00")
		right := string(analysis.Issues[j].Code) + strings.Join(analysis.Issues[j].UnitIDs, "\x00") +
			strings.Join(analysis.Issues[j].PortIDs, "\x00")
		return left < right
	})
}
