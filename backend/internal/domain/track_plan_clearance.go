package domain

import (
	"math"
	"sort"
)

const trackClearanceComparisonToleranceMM = 1e-9

type clearanceRoute struct {
	ObjectID         string
	Points           []TrackPoint
	CumulativeMM     []float64
	TotalLengthMM    float64
	ElevationStartMM float64
	ElevationEndMM   float64
}

type clearanceCandidate struct {
	ObjectIDs    []string
	ClearanceMM  float64
	Intersection TrackPoint
}

func analyzeTrackClearances(objects []PlanTrackObject, limit float64) []TrackPlanIssue {
	if !finiteTrackNumber(limit) || limit <= 0 {
		return nil
	}
	routes := make([]clearanceRoute, 0, len(objects))
	for _, object := range objects {
		route, ok := clearanceRouteForObject(object)
		if ok {
			routes = append(routes, route)
		}
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].ObjectID < routes[j].ObjectID })

	issues := make([]TrackPlanIssue, 0)
	for firstIndex := 0; firstIndex < len(routes); firstIndex++ {
		for secondIndex := firstIndex + 1; secondIndex < len(routes); secondIndex++ {
			candidate, ok := minimumClearanceCandidate(routes[firstIndex], routes[secondIndex])
			if !ok || candidate.ClearanceMM+trackClearanceComparisonToleranceMM >= limit {
				continue
			}
			clearance, clearanceLimit := candidate.ClearanceMM, limit
			intersectionX, intersectionY := candidate.Intersection.XMM, candidate.Intersection.YMM
			issues = append(issues, TrackPlanIssue{
				Code: TrackPlanIssueInsufficientClearance, Severity: TrackPlanIssueWarning,
				ObjectIDs: candidate.ObjectIDs, ClearanceMM: &clearance, ClearanceLimitMM: &clearanceLimit,
				IntersectionXMM: &intersectionX, IntersectionYMM: &intersectionY,
			})
		}
	}
	return issues
}

func clearanceRouteForObject(object PlanTrackObject) (clearanceRoute, bool) {
	effective, usable := effectiveTrackObjectGeometry(object)
	geometry := effective.Geometry
	if !usable || len(geometry.Ports) != 2 || len(geometry.Routes) != 1 ||
		len(geometry.Routes[0].Points) < 2 {
		return clearanceRoute{}, false
	}
	points := geometry.Routes[0].Points
	firstPoint, lastPoint := points[0], points[len(points)-1]
	forward := trackPointDistance(firstPoint.XMM, firstPoint.YMM, geometry.Ports[0].XMM,
		geometry.Ports[0].YMM) <= TrackConnectionDistanceMM &&
		trackPointDistance(lastPoint.XMM, lastPoint.YMM, geometry.Ports[1].XMM,
			geometry.Ports[1].YMM) <= TrackConnectionDistanceMM
	reversed := trackPointDistance(firstPoint.XMM, firstPoint.YMM, geometry.Ports[1].XMM,
		geometry.Ports[1].YMM) <= TrackConnectionDistanceMM &&
		trackPointDistance(lastPoint.XMM, lastPoint.YMM, geometry.Ports[0].XMM,
			geometry.Ports[0].YMM) <= TrackConnectionDistanceMM
	if !forward && !reversed {
		return clearanceRoute{}, false
	}

	transformed := make([]TrackPoint, len(points))
	cumulative := make([]float64, len(points))
	pose := poseForTrackObject(object)
	for index, point := range points {
		transformed[index] = TransformTrackPoint(point, pose)
		if index > 0 {
			cumulative[index] = cumulative[index-1] + trackPointDistance(
				transformed[index-1].XMM, transformed[index-1].YMM,
				transformed[index].XMM, transformed[index].YMM,
			)
		}
	}
	if cumulative[len(cumulative)-1] <= trackClearanceComparisonToleranceMM {
		return clearanceRoute{}, false
	}
	elevationStart, elevationEnd := object.ElevationStartMM, object.ElevationEndMM
	if reversed {
		elevationStart, elevationEnd = elevationEnd, elevationStart
	}
	return clearanceRoute{
		ObjectID: object.ID, Points: transformed, CumulativeMM: cumulative,
		TotalLengthMM: cumulative[len(cumulative)-1], ElevationStartMM: elevationStart,
		ElevationEndMM: elevationEnd,
	}, true
}

func minimumClearanceCandidate(first, second clearanceRoute) (clearanceCandidate, bool) {
	best := clearanceCandidate{ObjectIDs: []string{first.ObjectID, second.ObjectID}}
	found := false
	for firstIndex := 1; firstIndex < len(first.Points); firstIndex++ {
		firstSegment := trackSegment{A: first.Points[firstIndex-1], B: first.Points[firstIndex]}
		firstLength := trackPointDistance(firstSegment.A.XMM, firstSegment.A.YMM,
			firstSegment.B.XMM, firstSegment.B.YMM)
		for secondIndex := 1; secondIndex < len(second.Points); secondIndex++ {
			secondSegment := trackSegment{A: second.Points[secondIndex-1], B: second.Points[secondIndex]}
			secondLength := trackPointDistance(secondSegment.A.XMM, secondSegment.A.YMM,
				secondSegment.B.XMM, secondSegment.B.YMM)
			intersection, firstPart, secondPart, ok := properTrackSegmentIntersection(firstSegment, secondSegment)
			if !ok {
				continue
			}
			firstDistance := first.CumulativeMM[firstIndex-1] + firstPart*firstLength
			secondDistance := second.CumulativeMM[secondIndex-1] + secondPart*secondLength
			if firstDistance <= TrackConnectionDistanceMM ||
				first.TotalLengthMM-firstDistance <= TrackConnectionDistanceMM ||
				secondDistance <= TrackConnectionDistanceMM ||
				second.TotalLengthMM-secondDistance <= TrackConnectionDistanceMM {
				continue
			}
			firstFraction := firstDistance / first.TotalLengthMM
			secondFraction := secondDistance / second.TotalLengthMM
			firstElevation := interpolateTrackElevation(first, firstFraction)
			secondElevation := interpolateTrackElevation(second, secondFraction)
			candidate := clearanceCandidate{
				ObjectIDs: best.ObjectIDs, ClearanceMM: math.Abs(firstElevation - secondElevation),
				Intersection: intersection,
			}
			if !found || betterClearanceCandidate(candidate, best) {
				best, found = candidate, true
			}
		}
	}
	return best, found
}

func properTrackSegmentIntersection(first, second trackSegment) (TrackPoint, float64, float64, bool) {
	firstX, firstY := first.B.XMM-first.A.XMM, first.B.YMM-first.A.YMM
	secondX, secondY := second.B.XMM-second.A.XMM, second.B.YMM-second.A.YMM
	denominator := crossTrackVectors(firstX, firstY, secondX, secondY)
	if math.Abs(denominator) <= trackClearanceComparisonToleranceMM {
		return TrackPoint{}, 0, 0, false
	}
	offsetX, offsetY := second.A.XMM-first.A.XMM, second.A.YMM-first.A.YMM
	firstPart := crossTrackVectors(offsetX, offsetY, secondX, secondY) / denominator
	secondPart := crossTrackVectors(offsetX, offsetY, firstX, firstY) / denominator
	if firstPart < -trackClearanceComparisonToleranceMM ||
		firstPart > 1+trackClearanceComparisonToleranceMM ||
		secondPart < -trackClearanceComparisonToleranceMM ||
		secondPart > 1+trackClearanceComparisonToleranceMM {
		return TrackPoint{}, 0, 0, false
	}
	return TrackPoint{XMM: first.A.XMM + firstPart*firstX, YMM: first.A.YMM + firstPart*firstY},
		firstPart, secondPart, true
}

func interpolateTrackElevation(route clearanceRoute, fraction float64) float64 {
	return route.ElevationStartMM + fraction*(route.ElevationEndMM-route.ElevationStartMM)
}

func betterClearanceCandidate(candidate, current clearanceCandidate) bool {
	if candidate.ClearanceMM < current.ClearanceMM-trackClearanceComparisonToleranceMM {
		return true
	}
	if math.Abs(candidate.ClearanceMM-current.ClearanceMM) > trackClearanceComparisonToleranceMM {
		return false
	}
	if candidate.Intersection.XMM < current.Intersection.XMM-trackClearanceComparisonToleranceMM {
		return true
	}
	if math.Abs(candidate.Intersection.XMM-current.Intersection.XMM) > trackClearanceComparisonToleranceMM {
		return false
	}
	return candidate.Intersection.YMM < current.Intersection.YMM
}

func crossTrackVectors(firstX, firstY, secondX, secondY float64) float64 {
	return firstX*secondY - firstY*secondX
}

func finiteTrackNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
