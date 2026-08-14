package domain

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestBuildFlexTrackGeometryDerivesDeterministicAdaptiveRoute(t *testing.T) {
	path := FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, EndYMM: 100, EndDirectionDegrees: 380,
		StartHandleMM: 180, EndHandleMM: 170,
	}
	first, err := BuildFlexTrackGeometry(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildFlexTrackGeometry(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("flex sampling is not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if len(first.Geometry.Ports) != 2 || first.Geometry.Ports[0].XMM != 0 ||
		first.Geometry.Ports[1].XMM != 500 || first.Geometry.Ports[1].YMM != 100 ||
		first.Geometry.Ports[1].DirectionDegrees != 20 || len(first.Geometry.Routes) != 1 ||
		first.LengthMM <= math.Hypot(500, 100) || first.MinimumRadiusMM == nil ||
		*first.MinimumRadiusMM <= 0 {
		t.Fatalf("unexpected effective geometry: %#v", first)
	}
	assertFlexRouteTolerance(t, path, first.Geometry.Routes[0].Points)
}

func TestBuildFlexTrackGeometryHandlesStraightAndSCurve(t *testing.T) {
	straight, err := BuildFlexTrackGeometry(FlexTrackPath{
		SchemaVersion: 1, EndXMM: 664, StartHandleMM: 200, EndHandleMM: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(straight.LengthMM-664) > 1e-9 || straight.MinimumRadiusMM != nil {
		t.Fatalf("unexpected straight flex geometry: %#v", straight)
	}
	assertMaximumFlexSegmentLength(t, straight.Geometry.Routes[0].Points)

	sCurve, err := BuildFlexTrackGeometry(FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, EndYMM: 100,
		StartHandleMM: 180, EndHandleMM: 180,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sCurve.MinimumRadiusMM == nil || *sCurve.MinimumRadiusMM <= 0 ||
		sCurve.LengthMM <= math.Hypot(500, 100) {
		t.Fatalf("unexpected S-curve geometry: %#v", sCurve)
	}
	assertFlexRouteTolerance(t, FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, EndYMM: 100,
		StartHandleMM: 180, EndHandleMM: 180,
	}, sCurve.Geometry.Routes[0].Points)
}

func TestBuildFlexTrackGeometryRejectsInvalidPaths(t *testing.T) {
	valid := FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, EndYMM: 100,
		StartHandleMM: 180, EndHandleMM: 170,
	}
	invalid := []FlexTrackPath{
		{SchemaVersion: 2, EndXMM: 500, StartHandleMM: 1, EndHandleMM: 1},
		{SchemaVersion: 1, StartHandleMM: 1, EndHandleMM: 1},
		{SchemaVersion: 1, EndXMM: math.NaN(), StartHandleMM: 1, EndHandleMM: 1},
		{SchemaVersion: 1, EndXMM: 500, EndYMM: math.Inf(1), StartHandleMM: 1, EndHandleMM: 1},
		{SchemaVersion: 1, EndXMM: 500, EndDirectionDegrees: math.Inf(-1), StartHandleMM: 1, EndHandleMM: 1},
		{SchemaVersion: 1, EndXMM: 500, StartHandleMM: 0, EndHandleMM: 1},
		{SchemaVersion: 1, EndXMM: 500, StartHandleMM: 1, EndHandleMM: -1},
	}
	for index, path := range invalid {
		if _, err := BuildFlexTrackGeometry(path); !errors.Is(err, ErrInvalidFlexTrackPath) {
			t.Fatalf("invalid path %d returned %v", index, err)
		}
	}
	if _, err := BuildFlexTrackGeometry(valid); err != nil {
		t.Fatalf("valid path rejected: %v", err)
	}
	if _, err := BuildFlexTrackGeometry(FlexTrackPath{
		SchemaVersion: 1, EndXMM: 30000, StartHandleMM: 10000, EndHandleMM: 10000,
	}); !errors.Is(err, ErrInvalidFlexTrackPath) {
		t.Fatalf("path above segment cap returned %v", err)
	}
}

func TestEffectiveTrackGeometryUsesFlexPathAndRigidFallback(t *testing.T) {
	rigid := testG1Object("rigid", 0, 0, 0)
	effective, err := EffectiveGeometryForObject(rigid)
	if err != nil || effective.LengthMM != 166 ||
		!reflect.DeepEqual(effective.Geometry, rigid.Geometry.Geometry) {
		t.Fatalf("unexpected rigid geometry: %#v, %v", effective, err)
	}

	flex := testFlexObject("flex", FlexTrackPath{
		SchemaVersion: 1, EndXMM: 500, EndYMM: 100,
		StartHandleMM: 180, EndHandleMM: 180,
	})
	effective, err = EffectiveGeometryForObject(flex)
	if err != nil || effective.LengthMM <= math.Hypot(500, 100) || effective.MinimumRadiusMM == nil {
		t.Fatalf("unexpected flex geometry: %#v, %v", effective, err)
	}

	flex.FlexPath = nil
	effective, err = EffectiveGeometryForObject(flex)
	if err != nil || effective.LengthMM != 664 || effective.MinimumRadiusMM != nil ||
		!reflect.DeepEqual(effective.Geometry, flex.Geometry.Geometry) {
		t.Fatalf("unexpected unshaped flex fallback: %#v, %v", effective, err)
	}
}

func assertFlexRouteTolerance(t *testing.T, path FlexTrackPath, points []TrackPoint) {
	t.Helper()
	assertMaximumFlexSegmentLength(t, points)
	for sample := 0; sample <= 1000; sample++ {
		point := testFlexCubicPoint(path, float64(sample)/1000)
		minimumDistance := math.Inf(1)
		for index := 1; index < len(points); index++ {
			distance := distanceToTrackSegment(point, points[index-1], points[index])
			if distance < minimumDistance {
				minimumDistance = distance
			}
		}
		if minimumDistance > 0.050001 {
			t.Fatalf("route exceeds chord tolerance at sample %d: %.6f mm", sample, minimumDistance)
		}
	}
}

func assertMaximumFlexSegmentLength(t *testing.T, points []TrackPoint) {
	t.Helper()
	if len(points) < 2 || len(points)-1 > 4096 {
		t.Fatalf("unexpected route point count: %d", len(points))
	}
	for index := 1; index < len(points); index++ {
		if length := trackPointDistance(points[index-1].XMM, points[index-1].YMM,
			points[index].XMM, points[index].YMM); length > 5.000001 {
			t.Fatalf("segment %d is too long: %.6f mm", index-1, length)
		}
	}
}

func testFlexCubicPoint(path FlexTrackPath, parameter float64) TrackPoint {
	direction := NormalizeTrackRotation(path.EndDirectionDegrees) * math.Pi / 180
	control := [4]TrackPoint{
		{},
		{XMM: path.StartHandleMM},
		{XMM: path.EndXMM - path.EndHandleMM*math.Cos(direction),
			YMM: path.EndYMM - path.EndHandleMM*math.Sin(direction)},
		{XMM: path.EndXMM, YMM: path.EndYMM},
	}
	oneMinus := 1 - parameter
	return TrackPoint{
		XMM: oneMinus*oneMinus*oneMinus*control[0].XMM +
			3*oneMinus*oneMinus*parameter*control[1].XMM +
			3*oneMinus*parameter*parameter*control[2].XMM +
			parameter*parameter*parameter*control[3].XMM,
		YMM: oneMinus*oneMinus*oneMinus*control[0].YMM +
			3*oneMinus*oneMinus*parameter*control[1].YMM +
			3*oneMinus*parameter*parameter*control[2].YMM +
			parameter*parameter*parameter*control[3].YMM,
	}
}

func distanceToTrackSegment(point, start, end TrackPoint) float64 {
	dx, dy := end.XMM-start.XMM, end.YMM-start.YMM
	lengthSquared := dx*dx + dy*dy
	if lengthSquared <= 1e-18 {
		return trackPointDistance(point.XMM, point.YMM, start.XMM, start.YMM)
	}
	fraction := ((point.XMM-start.XMM)*dx + (point.YMM-start.YMM)*dy) / lengthSquared
	fraction = math.Max(0, math.Min(1, fraction))
	return trackPointDistance(point.XMM, point.YMM, start.XMM+fraction*dx, start.YMM+fraction*dy)
}

func testFlexObject(id string, path FlexTrackPath) PlanTrackObject {
	radius := 543.0
	return PlanTrackObject{
		ID: id, GeometryID: "tillig-flex", FlexPath: &path,
		Geometry: TrackGeometryDefinition{
			ID: "tillig-flex", LibraryID: "tillig-v1", ArticleNumber: "83125",
			Name: "Flexgleis Holzschwelle", Kind: TrackGeometryFlex, LengthMM: 664,
			MinimumRadiusMM: &radius, Status: TrackGeometryVerified,
			Geometry: TrackGeometry{
				SchemaVersion: 1,
				Ports: []TrackPort{
					{ID: "a", DirectionDegrees: 180},
					{ID: "b", XMM: 664},
				},
				Routes: []TrackRoute{{ID: "main", Points: []TrackPoint{{}, {XMM: 664}}}},
			},
		},
	}
}
