package domain

import (
	"math"
	"sort"
	"strings"
)

const (
	TrackSnapDistanceMM                 = 8.0
	TrackSnapDirectionToleranceDegrees  = 5.0
	TrackConnectionDistanceMM           = 0.25
	TrackConnectionDirectionDegrees     = 0.5
	TrackElevationConnectionToleranceMM = 0.01
	TrackGradeLimitTolerancePercent     = 1e-9
)

type TrackPose struct {
	PositionXMM     float64 `json:"positionXMm"`
	PositionYMM     float64 `json:"positionYMm"`
	RotationDegrees float64 `json:"rotationDegrees"`
}

type TrackSnapResult struct {
	Snapped        bool      `json:"snapped"`
	Pose           TrackPose `json:"pose"`
	MovingPortID   string    `json:"movingPortId,omitempty"`
	TargetObjectID string    `json:"targetObjectId,omitempty"`
	TargetPortID   string    `json:"targetPortId,omitempty"`
	DistanceMM     float64   `json:"distanceMm,omitempty"`
}

type TrackPlanConnection struct {
	ObjectAID string `json:"objectAId"`
	PortAID   string `json:"portAId"`
	ObjectBID string `json:"objectBId"`
	PortBID   string `json:"portBId"`
}

type TrackPlanIssueCode string

const (
	TrackPlanIssueOpenEnd                TrackPlanIssueCode = "open_end"
	TrackPlanIssueIncompatibleConnection TrackPlanIssueCode = "incompatible_connection"
	TrackPlanIssueOverlap                TrackPlanIssueCode = "overlap"
	TrackPlanIssueBrokenGeometry         TrackPlanIssueCode = "broken_geometry"
	TrackPlanIssueElevationMismatch      TrackPlanIssueCode = "elevation_mismatch"
	TrackPlanIssueGradeLimitExceeded     TrackPlanIssueCode = "grade_limit_exceeded"
	TrackPlanIssueInsufficientClearance  TrackPlanIssueCode = "insufficient_clearance"
)

type TrackPlanIssueSeverity string

const (
	TrackPlanIssueWarning TrackPlanIssueSeverity = "warning"
	TrackPlanIssueError   TrackPlanIssueSeverity = "error"
)

type TrackPlanIssue struct {
	Code                  TrackPlanIssueCode     `json:"code"`
	Severity              TrackPlanIssueSeverity `json:"severity"`
	ObjectIDs             []string               `json:"objectIds"`
	PortIDs               []string               `json:"portIds,omitempty"`
	ElevationDifferenceMM *float64               `json:"elevationDifferenceMm,omitempty"`
	GradePercent          *float64               `json:"gradePercent,omitempty"`
	GradeLimitPercent     *float64               `json:"gradeLimitPercent,omitempty"`
	ClearanceMM           *float64               `json:"clearanceMm,omitempty"`
	ClearanceLimitMM      *float64               `json:"clearanceLimitMm,omitempty"`
	IntersectionXMM       *float64               `json:"intersectionXMm,omitempty"`
	IntersectionYMM       *float64               `json:"intersectionYMm,omitempty"`
}

type TrackBOMLine struct {
	GeometryID    string `json:"geometryId"`
	LibraryID     string `json:"libraryId"`
	ArticleNumber string `json:"articleNumber"`
	Name          string `json:"name"`
	Quantity      int    `json:"quantity"`
}

type TrackGrade struct {
	ObjectID         string  `json:"objectId"`
	ElevationStartMM float64 `json:"elevationStartMm"`
	ElevationEndMM   float64 `json:"elevationEndMm"`
	LengthMM         float64 `json:"lengthMm"`
	GradePercent     float64 `json:"gradePercent"`
}

type TrackPlanAnalysis struct {
	Connections []TrackPlanConnection `json:"connections"`
	Issues      []TrackPlanIssue      `json:"issues"`
	BOM         []TrackBOMLine        `json:"bom"`
	Grades      []TrackGrade          `json:"grades"`
}

type TrackPlanLimits struct {
	MaxGradePercent         *float64
	MinimumTrackClearanceMM *float64
}

type placedTrackPort struct {
	ObjectID       string
	Port           TrackPort
	ElevationMM    float64
	ElevationKnown bool
}

type trackSegment struct {
	A TrackPoint
	B TrackPoint
}

func TransformTrackPoint(point TrackPoint, pose TrackPose) TrackPoint {
	radians := pose.RotationDegrees * math.Pi / 180
	cosine, sine := math.Cos(radians), math.Sin(radians)
	return TrackPoint{
		XMM: pose.PositionXMM + point.XMM*cosine - point.YMM*sine,
		YMM: pose.PositionYMM + point.XMM*sine + point.YMM*cosine,
	}
}

func TransformTrackPort(port TrackPort, pose TrackPose) TrackPort {
	point := TransformTrackPoint(TrackPoint{XMM: port.XMM, YMM: port.YMM}, pose)
	return TrackPort{
		ID: port.ID, XMM: point.XMM, YMM: point.YMM,
		DirectionDegrees: NormalizeTrackRotation(port.DirectionDegrees + pose.RotationDegrees),
	}
}

func FindTrackSnap(moving PlanTrackObject, objects []PlanTrackObject) TrackSnapResult {
	basePose := poseForTrackObject(moving)
	result := TrackSnapResult{Pose: basePose}
	for _, movingPort := range moving.Geometry.Geometry.Ports {
		currentPort := TransformTrackPort(movingPort, basePose)
		for _, targetObject := range objects {
			if targetObject.ID == moving.ID || !trackGeometryUsable(targetObject.Geometry) {
				continue
			}
			for _, targetLocalPort := range targetObject.Geometry.Geometry.Ports {
				targetPort := TransformTrackPort(targetLocalPort, poseForTrackObject(targetObject))
				distance := trackPointDistance(currentPort.XMM, currentPort.YMM, targetPort.XMM, targetPort.YMM)
				if distance > TrackSnapDistanceMM ||
					trackOpposingAngleDifference(currentPort.DirectionDegrees, targetPort.DirectionDegrees) >
						TrackSnapDirectionToleranceDegrees {
					continue
				}
				if result.Snapped && !betterTrackSnap(distance, targetObject.ID, targetPort.ID, result) {
					continue
				}
				desiredRotation := NormalizeTrackRotation(
					targetPort.DirectionDegrees + 180 - movingPort.DirectionDegrees,
				)
				rotatedPort := TransformTrackPort(movingPort, TrackPose{
					PositionXMM: moving.PositionXMM, PositionYMM: moving.PositionYMM,
					RotationDegrees: desiredRotation,
				})
				result = TrackSnapResult{
					Snapped: true,
					Pose: TrackPose{
						PositionXMM:     moving.PositionXMM + targetPort.XMM - rotatedPort.XMM,
						PositionYMM:     moving.PositionYMM + targetPort.YMM - rotatedPort.YMM,
						RotationDegrees: desiredRotation,
					},
					MovingPortID: movingPort.ID, TargetObjectID: targetObject.ID,
					TargetPortID: targetPort.ID, DistanceMM: distance,
				}
			}
		}
	}
	return result
}

func AnalyzeTrackPlan(objects []PlanTrackObject) TrackPlanAnalysis {
	return AnalyzeTrackPlanWithLimits(objects, TrackPlanLimits{})
}

func AnalyzeTrackPlanWithLimits(objects []PlanTrackObject, limits TrackPlanLimits) TrackPlanAnalysis {
	ordered := append([]PlanTrackObject(nil), objects...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	analysis := TrackPlanAnalysis{
		Connections: []TrackPlanConnection{}, Issues: []TrackPlanIssue{}, BOM: []TrackBOMLine{},
		Grades: []TrackGrade{},
	}
	ports := make([]placedTrackPort, 0)
	bom := map[string]*TrackBOMLine{}
	for _, object := range ordered {
		line := bom[object.GeometryID]
		if line == nil {
			line = &TrackBOMLine{GeometryID: object.GeometryID, LibraryID: object.Geometry.LibraryID,
				ArticleNumber: object.Geometry.ArticleNumber, Name: object.Geometry.Name}
			bom[object.GeometryID] = line
		}
		line.Quantity++
		if !trackGeometryUsable(object.Geometry) {
			analysis.Issues = append(analysis.Issues, TrackPlanIssue{
				Code: TrackPlanIssueBrokenGeometry, Severity: TrackPlanIssueError,
				ObjectIDs: []string{object.ID},
			})
			continue
		}
		if object.Geometry.LengthMM > 0 {
			gradePercent := (object.ElevationEndMM - object.ElevationStartMM) /
				object.Geometry.LengthMM * 100
			analysis.Grades = append(analysis.Grades, TrackGrade{
				ObjectID: object.ID, ElevationStartMM: object.ElevationStartMM,
				ElevationEndMM: object.ElevationEndMM, LengthMM: object.Geometry.LengthMM,
				GradePercent: gradePercent,
			})
			if limits.MaxGradePercent != nil &&
				math.Abs(gradePercent)-*limits.MaxGradePercent > TrackGradeLimitTolerancePercent {
				gradeLimitPercent := *limits.MaxGradePercent
				analysis.Issues = append(analysis.Issues, TrackPlanIssue{
					Code: TrackPlanIssueGradeLimitExceeded, Severity: TrackPlanIssueWarning,
					ObjectIDs: []string{object.ID}, GradePercent: &gradePercent,
					GradeLimitPercent: &gradeLimitPercent,
				})
			}
		}
		for _, port := range object.Geometry.Geometry.Ports {
			elevation, known := trackPortElevation(object, port.ID)
			ports = append(ports, placedTrackPort{ObjectID: object.ID,
				Port:        TransformTrackPort(port, poseForTrackObject(object)),
				ElevationMM: elevation, ElevationKnown: known})
		}
	}

	connected := map[string]bool{}
	for i := 0; i < len(ports); i++ {
		for j := i + 1; j < len(ports); j++ {
			first, second := ports[i], ports[j]
			if first.ObjectID == second.ObjectID {
				continue
			}
			distance := trackPointDistance(first.Port.XMM, first.Port.YMM, second.Port.XMM, second.Port.YMM)
			directionDifference := trackOpposingAngleDifference(
				first.Port.DirectionDegrees, second.Port.DirectionDegrees,
			)
			if distance <= TrackConnectionDistanceMM && directionDifference <= TrackConnectionDirectionDegrees &&
				!connected[trackPortKey(first)] && !connected[trackPortKey(second)] {
				analysis.Connections = append(analysis.Connections, TrackPlanConnection{
					ObjectAID: first.ObjectID, PortAID: first.Port.ID,
					ObjectBID: second.ObjectID, PortBID: second.Port.ID,
				})
				if first.ElevationKnown && second.ElevationKnown {
					difference := math.Abs(first.ElevationMM - second.ElevationMM)
					if difference > TrackElevationConnectionToleranceMM {
						analysis.Issues = append(analysis.Issues, TrackPlanIssue{
							Code: TrackPlanIssueElevationMismatch, Severity: TrackPlanIssueWarning,
							ObjectIDs:             []string{first.ObjectID, second.ObjectID},
							PortIDs:               []string{first.Port.ID, second.Port.ID},
							ElevationDifferenceMM: &difference,
						})
					}
				}
				connected[trackPortKey(first)] = true
				connected[trackPortKey(second)] = true
			} else if distance <= TrackSnapDistanceMM && directionDifference > TrackSnapDirectionToleranceDegrees {
				analysis.Issues = append(analysis.Issues, TrackPlanIssue{
					Code: TrackPlanIssueIncompatibleConnection, Severity: TrackPlanIssueWarning,
					ObjectIDs: []string{first.ObjectID, second.ObjectID},
					PortIDs:   []string{first.Port.ID, second.Port.ID},
				})
			}
		}
	}
	for _, port := range ports {
		if !connected[trackPortKey(port)] {
			analysis.Issues = append(analysis.Issues, TrackPlanIssue{
				Code: TrackPlanIssueOpenEnd, Severity: TrackPlanIssueWarning,
				ObjectIDs: []string{port.ObjectID}, PortIDs: []string{port.Port.ID},
			})
		}
	}
	for i := 0; i < len(ordered); i++ {
		if !trackGeometryUsable(ordered[i].Geometry) {
			continue
		}
		for j := i + 1; j < len(ordered); j++ {
			if !trackGeometryUsable(ordered[j].Geometry) {
				continue
			}
			if trackObjectsOverlap(ordered[i], ordered[j]) {
				analysis.Issues = append(analysis.Issues, TrackPlanIssue{
					Code: TrackPlanIssueOverlap, Severity: TrackPlanIssueWarning,
					ObjectIDs: []string{ordered[i].ID, ordered[j].ID},
				})
			}
		}
	}
	if limits.MinimumTrackClearanceMM != nil {
		analysis.Issues = append(analysis.Issues,
			analyzeTrackClearances(ordered, *limits.MinimumTrackClearanceMM)...)
	}
	for _, line := range bom {
		analysis.BOM = append(analysis.BOM, *line)
	}
	sort.Slice(analysis.BOM, func(i, j int) bool {
		if analysis.BOM[i].ArticleNumber == analysis.BOM[j].ArticleNumber {
			return analysis.BOM[i].GeometryID < analysis.BOM[j].GeometryID
		}
		return analysis.BOM[i].ArticleNumber < analysis.BOM[j].ArticleNumber
	})
	sort.SliceStable(analysis.Issues, func(i, j int) bool {
		left := string(analysis.Issues[i].Code) + strings.Join(analysis.Issues[i].ObjectIDs, "\x00") +
			strings.Join(analysis.Issues[i].PortIDs, "\x00")
		right := string(analysis.Issues[j].Code) + strings.Join(analysis.Issues[j].ObjectIDs, "\x00") +
			strings.Join(analysis.Issues[j].PortIDs, "\x00")
		return left < right
	})
	return analysis
}

func poseForTrackObject(object PlanTrackObject) TrackPose {
	return TrackPose{PositionXMM: object.PositionXMM, PositionYMM: object.PositionYMM,
		RotationDegrees: object.RotationDegrees}
}

func trackGeometryUsable(definition TrackGeometryDefinition) bool {
	if definition.ID == "" || definition.Geometry.SchemaVersion < 1 ||
		len(definition.Geometry.Ports) == 0 || len(definition.Geometry.Routes) == 0 {
		return false
	}
	for _, route := range definition.Geometry.Routes {
		if len(route.Points) < 2 {
			return false
		}
	}
	return true
}

func trackPortElevation(object PlanTrackObject, portID string) (float64, bool) {
	if len(object.Geometry.Geometry.Ports) != 2 {
		return 0, false
	}
	for index, port := range object.Geometry.Geometry.Ports {
		if port.ID != portID {
			continue
		}
		if index == 0 {
			return object.ElevationStartMM, true
		}
		return object.ElevationEndMM, true
	}
	return 0, false
}

func betterTrackSnap(distance float64, objectID, portID string, current TrackSnapResult) bool {
	if distance < current.DistanceMM-1e-9 {
		return true
	}
	if math.Abs(distance-current.DistanceMM) > 1e-9 {
		return false
	}
	return objectID+"\x00"+portID < current.TargetObjectID+"\x00"+current.TargetPortID
}

func trackPointDistance(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}

func trackOpposingAngleDifference(first, second float64) float64 {
	return trackAngleDifference(NormalizeTrackRotation(first+180), second)
}

func trackAngleDifference(first, second float64) float64 {
	difference := math.Abs(NormalizeTrackRotation(first) - NormalizeTrackRotation(second))
	if difference > 180 {
		return 360 - difference
	}
	return difference
}

func trackPortKey(port placedTrackPort) string {
	return port.ObjectID + "\x00" + port.Port.ID
}

func trackObjectsOverlap(first, second PlanTrackObject) bool {
	for _, firstRoute := range first.Geometry.Geometry.Routes {
		firstSegments := transformedTrackSegments(firstRoute, first)
		for _, secondRoute := range second.Geometry.Geometry.Routes {
			secondSegments := transformedTrackSegments(secondRoute, second)
			for _, firstSegment := range firstSegments {
				for _, secondSegment := range secondSegments {
					if collinearTrackOverlap(firstSegment, secondSegment) > TrackConnectionDistanceMM {
						return true
					}
				}
			}
		}
	}
	return false
}

func transformedTrackSegments(route TrackRoute, object PlanTrackObject) []trackSegment {
	segments := make([]trackSegment, 0, len(route.Points)-1)
	pose := poseForTrackObject(object)
	for i := 1; i < len(route.Points); i++ {
		segments = append(segments, trackSegment{
			A: TransformTrackPoint(route.Points[i-1], pose),
			B: TransformTrackPoint(route.Points[i], pose),
		})
	}
	return segments
}

func collinearTrackOverlap(first, second trackSegment) float64 {
	dx, dy := first.B.XMM-first.A.XMM, first.B.YMM-first.A.YMM
	length := math.Hypot(dx, dy)
	if length <= 1e-9 {
		return 0
	}
	crossA := math.Abs(dx*(second.A.YMM-first.A.YMM)-dy*(second.A.XMM-first.A.XMM)) / length
	crossB := math.Abs(dx*(second.B.YMM-first.A.YMM)-dy*(second.B.XMM-first.A.XMM)) / length
	if crossA > TrackConnectionDistanceMM || crossB > TrackConnectionDistanceMM {
		return 0
	}
	unitX, unitY := dx/length, dy/length
	project := func(point TrackPoint) float64 {
		return (point.XMM-first.A.XMM)*unitX + (point.YMM-first.A.YMM)*unitY
	}
	secondStart, secondEnd := project(second.A), project(second.B)
	if secondStart > secondEnd {
		secondStart, secondEnd = secondEnd, secondStart
	}
	return math.Max(0, math.Min(length, secondEnd)-math.Max(0, secondStart))
}
