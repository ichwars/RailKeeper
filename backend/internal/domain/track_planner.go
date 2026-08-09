package domain

import "math"

type TrackGeometryKind string

const (
	TrackGeometryStraight TrackGeometryKind = "straight"
	TrackGeometryCurve    TrackGeometryKind = "curve"
	TrackGeometryTurnout  TrackGeometryKind = "turnout"
	TrackGeometryCrossing TrackGeometryKind = "crossing"
)

func (kind TrackGeometryKind) Valid() bool {
	switch kind {
	case TrackGeometryStraight, TrackGeometryCurve, TrackGeometryTurnout, TrackGeometryCrossing:
		return true
	default:
		return false
	}
}

type TrackGeometryStatus string

const (
	TrackGeometryDraft    TrackGeometryStatus = "draft"
	TrackGeometryVerified TrackGeometryStatus = "verified"
	TrackGeometryRetired  TrackGeometryStatus = "retired"
)

func (status TrackGeometryStatus) Valid() bool {
	switch status {
	case TrackGeometryDraft, TrackGeometryVerified, TrackGeometryRetired:
		return true
	default:
		return false
	}
}

func (status TrackGeometryStatus) Placeable() bool {
	return status == TrackGeometryVerified
}

type TrackPoint struct {
	XMM float64 `json:"xMm"`
	YMM float64 `json:"yMm"`
}

type TrackPort struct {
	ID               string  `json:"id"`
	XMM              float64 `json:"xMm"`
	YMM              float64 `json:"yMm"`
	DirectionDegrees float64 `json:"directionDegrees"`
}

type TrackRoute struct {
	ID     string       `json:"id"`
	Points []TrackPoint `json:"points"`
}

type TrackGeometry struct {
	SchemaVersion int          `json:"schemaVersion"`
	Ports         []TrackPort  `json:"ports"`
	Routes        []TrackRoute `json:"routes"`
}

type TrackGeometryLibrary struct {
	ID           string              `json:"id"`
	Manufacturer string              `json:"manufacturer"`
	TrackSystem  string              `json:"trackSystem"`
	Gauge        string              `json:"gauge"`
	Scale        string              `json:"scale"`
	Version      string              `json:"version"`
	SourceURL    string              `json:"sourceUrl"`
	Status       TrackGeometryStatus `json:"status"`
	CreatedAt    string              `json:"createdAt"`
}

type TrackGeometryDefinition struct {
	ID            string              `json:"id"`
	LibraryID     string              `json:"libraryId"`
	ArticleNumber string              `json:"articleNumber"`
	Name          string              `json:"name"`
	Kind          TrackGeometryKind   `json:"kind"`
	LengthMM      float64             `json:"lengthMm"`
	Geometry      TrackGeometry       `json:"geometry"`
	SourceURL     string              `json:"sourceUrl"`
	Status        TrackGeometryStatus `json:"status"`
	CreatedAt     string              `json:"createdAt"`
}

type PlanTrackObject struct {
	ID              string                  `json:"id"`
	LineageID       string                  `json:"lineageId"`
	RevisionID      string                  `json:"revisionId"`
	GeometryID      string                  `json:"geometryId"`
	Geometry        TrackGeometryDefinition `json:"geometry"`
	PositionXMM     float64                 `json:"positionXMm"`
	PositionYMM     float64                 `json:"positionYMm"`
	RotationDegrees float64                 `json:"rotationDegrees"`
	Version         int                     `json:"version"`
	CreatedAt       string                  `json:"createdAt"`
	UpdatedAt       string                  `json:"updatedAt"`
}

func NormalizeTrackRotation(value float64) float64 {
	normalized := math.Mod(value, 360)
	if normalized < 0 {
		normalized += 360
	}
	return normalized
}
