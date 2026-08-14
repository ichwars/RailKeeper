package domain

import (
	"errors"
	"math"
)

type TransitionDirection string

const (
	TransitionLeft  TransitionDirection = "left"
	TransitionRight TransitionDirection = "right"
)

var ErrInvalidTransitionCurvePath = errors.New("invalid transition curve path")

type TransitionCurvePath struct {
	SchemaVersion int                 `json:"schemaVersion"`
	LengthMM      float64             `json:"lengthMm"`
	EndRadiusMM   float64             `json:"endRadiusMm"`
	Direction     TransitionDirection `json:"direction"`
}

func (direction TransitionDirection) Valid() bool {
	return direction == TransitionLeft || direction == TransitionRight
}

func BuildTransitionTrackGeometry(path TransitionCurvePath) (EffectiveTrackGeometry, error) {
	if path.SchemaVersion != 1 || !finiteTrackNumber(path.LengthMM) || path.LengthMM <= 0 ||
		!finiteTrackNumber(path.EndRadiusMM) || path.EndRadiusMM <= 0 || !path.Direction.Valid() {
		return EffectiveTrackGeometry{}, ErrInvalidTransitionCurvePath
	}
	maximumStep := math.Min(flexTrackMaximumSegmentLengthMM,
		math.Sqrt(8*path.EndRadiusMM*flexTrackMaximumChordErrorMM))
	segments := int(math.Ceil(path.LengthMM / maximumStep))
	if segments < 1 || segments > flexTrackMaximumSegments {
		return EffectiveTrackGeometry{}, ErrInvalidTransitionCurvePath
	}

	sign := 1.0
	if path.Direction == TransitionRight {
		sign = -1
	}
	points := make([]TrackPoint, segments+1)
	segmentLength := path.LengthMM / float64(segments)
	for index := 1; index <= segments; index++ {
		start := float64(index-1) * segmentLength
		end := float64(index) * segmentLength
		middle := (start + end) / 2
		startHeading := transitionHeadingRadians(start, path, sign)
		middleHeading := transitionHeadingRadians(middle, path, sign)
		endHeading := transitionHeadingRadians(end, path, sign)
		deltaX := segmentLength / 6 *
			(math.Cos(startHeading) + 4*math.Cos(middleHeading) + math.Cos(endHeading))
		deltaY := segmentLength / 6 *
			(math.Sin(startHeading) + 4*math.Sin(middleHeading) + math.Sin(endHeading))
		points[index] = TrackPoint{
			XMM: points[index-1].XMM + deltaX,
			YMM: points[index-1].YMM + deltaY,
		}
	}

	end := points[len(points)-1]
	endDirection := NormalizeTrackRotation(sign * path.LengthMM / (2 * path.EndRadiusMM) * 180 / math.Pi)
	minimumRadius := path.EndRadiusMM
	return EffectiveTrackGeometry{
		Geometry: TrackGeometry{
			SchemaVersion: 1,
			Ports: []TrackPort{
				{ID: "a", DirectionDegrees: 180},
				{ID: "b", XMM: end.XMM, YMM: end.YMM, DirectionDegrees: endDirection},
			},
			Routes: []TrackRoute{{ID: "main", Points: points}},
		},
		LengthMM: path.LengthMM, MinimumRadiusMM: &minimumRadius,
	}, nil
}

func transitionHeadingRadians(distance float64, path TransitionCurvePath, sign float64) float64 {
	return sign * distance * distance / (2 * path.EndRadiusMM * path.LengthMM)
}
