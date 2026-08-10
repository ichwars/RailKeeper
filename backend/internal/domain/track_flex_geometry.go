package domain

import (
	"errors"
	"math"
)

const (
	flexTrackMaximumSegmentLengthMM = 5.0
	flexTrackMaximumChordErrorMM    = 0.05
	flexTrackMaximumSegments        = 4096
	flexTrackCurvatureEpsilon       = 1e-12
)

var ErrInvalidFlexTrackPath = errors.New("invalid flex track path")

type FlexTrackPath struct {
	SchemaVersion       int     `json:"schemaVersion"`
	EndXMM              float64 `json:"endXMm"`
	EndYMM              float64 `json:"endYMm"`
	EndDirectionDegrees float64 `json:"endDirectionDegrees"`
	StartHandleMM       float64 `json:"startHandleMm"`
	EndHandleMM         float64 `json:"endHandleMm"`
}

type EffectiveTrackGeometry struct {
	Geometry        TrackGeometry `json:"geometry"`
	LengthMM        float64       `json:"lengthMm"`
	MinimumRadiusMM *float64      `json:"minimumRadiusMm,omitempty"`
}

type flexBezierSegment struct {
	Control [4]TrackPoint
	StartT  float64
	EndT    float64
}

func BuildFlexTrackGeometry(path FlexTrackPath) (EffectiveTrackGeometry, error) {
	control, endDirection, err := validatedFlexControlPoints(path)
	if err != nil {
		return EffectiveTrackGeometry{}, err
	}
	points, parameters, err := sampleFlexBezier(control)
	if err != nil {
		return EffectiveTrackGeometry{}, err
	}

	length := 0.0
	for index := 1; index < len(points); index++ {
		length += trackPointDistance(points[index-1].XMM, points[index-1].YMM,
			points[index].XMM, points[index].YMM)
	}
	minimumRadius := flexMinimumRadius(control, parameters)
	return EffectiveTrackGeometry{
		Geometry: TrackGeometry{
			SchemaVersion: 1,
			Ports: []TrackPort{
				{ID: "a", DirectionDegrees: 180},
				{ID: "b", XMM: path.EndXMM, YMM: path.EndYMM, DirectionDegrees: endDirection},
			},
			Routes: []TrackRoute{{ID: "main", Points: points}},
		},
		LengthMM: length, MinimumRadiusMM: minimumRadius,
	}, nil
}

func EffectiveGeometryForObject(object PlanTrackObject) (EffectiveTrackGeometry, error) {
	if object.Geometry.Kind != TrackGeometryFlex || object.FlexPath == nil {
		return EffectiveTrackGeometry{
			Geometry: object.Geometry.Geometry, LengthMM: object.Geometry.LengthMM,
		}, nil
	}
	return BuildFlexTrackGeometry(*object.FlexPath)
}

func validatedFlexControlPoints(path FlexTrackPath) ([4]TrackPoint, float64, error) {
	values := []float64{
		path.EndXMM, path.EndYMM, path.EndDirectionDegrees, path.StartHandleMM, path.EndHandleMM,
	}
	if path.SchemaVersion != 1 || path.StartHandleMM <= 0 || path.EndHandleMM <= 0 ||
		math.Hypot(path.EndXMM, path.EndYMM) <= 1e-9 {
		return [4]TrackPoint{}, 0, ErrInvalidFlexTrackPath
	}
	for _, value := range values {
		if !finiteTrackNumber(value) {
			return [4]TrackPoint{}, 0, ErrInvalidFlexTrackPath
		}
	}
	endDirection := NormalizeTrackRotation(path.EndDirectionDegrees)
	endRadians := endDirection * math.Pi / 180
	return [4]TrackPoint{
		{},
		{XMM: path.StartHandleMM},
		{
			XMM: path.EndXMM - path.EndHandleMM*math.Cos(endRadians),
			YMM: path.EndYMM - path.EndHandleMM*math.Sin(endRadians),
		},
		{XMM: path.EndXMM, YMM: path.EndYMM},
	}, endDirection, nil
}

func sampleFlexBezier(control [4]TrackPoint) ([]TrackPoint, []float64, error) {
	points := []TrackPoint{control[0]}
	parameters := []float64{0}
	stack := []flexBezierSegment{{Control: control, StartT: 0, EndT: 1}}
	for len(stack) > 0 {
		last := len(stack) - 1
		segment := stack[last]
		stack = stack[:last]
		if flexBezierSegmentWithinTolerance(segment.Control) {
			if len(points)-1 >= flexTrackMaximumSegments {
				return nil, nil, ErrInvalidFlexTrackPath
			}
			points = append(points, segment.Control[3])
			parameters = append(parameters, segment.EndT)
			continue
		}
		if len(points)+len(stack) >= flexTrackMaximumSegments {
			return nil, nil, ErrInvalidFlexTrackPath
		}
		left, right := splitFlexBezier(segment.Control)
		midpoint := (segment.StartT + segment.EndT) / 2
		stack = append(stack,
			flexBezierSegment{Control: right, StartT: midpoint, EndT: segment.EndT},
			flexBezierSegment{Control: left, StartT: segment.StartT, EndT: midpoint},
		)
	}
	return points, parameters, nil
}

func flexBezierSegmentWithinTolerance(control [4]TrackPoint) bool {
	if trackPointDistance(control[0].XMM, control[0].YMM,
		control[3].XMM, control[3].YMM) > flexTrackMaximumSegmentLengthMM {
		return false
	}
	return flexPointLineDistance(control[1], control[0], control[3]) <= flexTrackMaximumChordErrorMM &&
		flexPointLineDistance(control[2], control[0], control[3]) <= flexTrackMaximumChordErrorMM
}

func flexPointLineDistance(point, start, end TrackPoint) float64 {
	dx, dy := end.XMM-start.XMM, end.YMM-start.YMM
	length := math.Hypot(dx, dy)
	if length <= 1e-12 {
		return math.Max(
			trackPointDistance(point.XMM, point.YMM, start.XMM, start.YMM),
			trackPointDistance(point.XMM, point.YMM, end.XMM, end.YMM),
		)
	}
	return math.Abs(dx*(start.YMM-point.YMM)-(start.XMM-point.XMM)*dy) / length
}

func splitFlexBezier(control [4]TrackPoint) ([4]TrackPoint, [4]TrackPoint) {
	first := flexMidpoint(control[0], control[1])
	second := flexMidpoint(control[1], control[2])
	third := flexMidpoint(control[2], control[3])
	leftMiddle := flexMidpoint(first, second)
	rightMiddle := flexMidpoint(second, third)
	middle := flexMidpoint(leftMiddle, rightMiddle)
	return [4]TrackPoint{control[0], first, leftMiddle, middle},
		[4]TrackPoint{middle, rightMiddle, third, control[3]}
}

func flexMidpoint(first, second TrackPoint) TrackPoint {
	return TrackPoint{XMM: (first.XMM + second.XMM) / 2, YMM: (first.YMM + second.YMM) / 2}
}

func flexMinimumRadius(control [4]TrackPoint, parameters []float64) *float64 {
	minimum := math.Inf(1)
	for _, parameter := range parameters {
		oneMinus := 1 - parameter
		firstX := 3 * (oneMinus*oneMinus*(control[1].XMM-control[0].XMM) +
			2*oneMinus*parameter*(control[2].XMM-control[1].XMM) +
			parameter*parameter*(control[3].XMM-control[2].XMM))
		firstY := 3 * (oneMinus*oneMinus*(control[1].YMM-control[0].YMM) +
			2*oneMinus*parameter*(control[2].YMM-control[1].YMM) +
			parameter*parameter*(control[3].YMM-control[2].YMM))
		secondX := 6 * (oneMinus*(control[2].XMM-2*control[1].XMM+control[0].XMM) +
			parameter*(control[3].XMM-2*control[2].XMM+control[1].XMM))
		secondY := 6 * (oneMinus*(control[2].YMM-2*control[1].YMM+control[0].YMM) +
			parameter*(control[3].YMM-2*control[2].YMM+control[1].YMM))
		speed := math.Hypot(firstX, firstY)
		if speed <= 1e-12 {
			continue
		}
		curvature := math.Abs(firstX*secondY-firstY*secondX) / (speed * speed * speed)
		if curvature < flexTrackCurvatureEpsilon {
			continue
		}
		radius := 1 / curvature
		if radius < minimum {
			minimum = radius
		}
	}
	if math.IsInf(minimum, 1) {
		return nil
	}
	return &minimum
}
