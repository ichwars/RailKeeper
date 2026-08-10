package domain

import (
	"errors"
	"math"
	"net/url"
	"strings"
)

const (
	TrackLibraryPackageFormat          = "railkeeper.track-library"
	TrackLibraryPackageSchemaVersion   = 1
	MaxTrackLibraryDefinitions         = 500
	MaxTrackLibraryDefinitionName      = 160
	maxTrackLibraryMetadataLength      = 120
	maxTrackLibraryArticleNumberLength = 80
	maxTrackLibraryVersionLength       = 64
	maxTrackLibraryGaugeScaleLength    = 32
	maxTrackLibraryURLLength           = 2048
	maxTrackLibraryGeometryIDLength    = 64
	maxTrackLibraryPorts               = 64
	maxTrackLibraryRoutes              = 64
	maxTrackLibraryRoutePoints         = 256
	maxTrackLibraryCoordinateMM        = 100000
	maxTrackLibraryLengthMM            = 100000
)

var ErrInvalidTrackLibraryPackage = errors.New("invalid track library package")

type TrackLibraryPackage struct {
	Format        string                          `json:"format"`
	SchemaVersion int                             `json:"schemaVersion"`
	ExportedAt    string                          `json:"exportedAt,omitempty"`
	Library       TrackLibraryPackageMetadata     `json:"library"`
	Definitions   []TrackLibraryPackageDefinition `json:"definitions"`
}

type TrackLibraryPackageMetadata struct {
	Manufacturer string              `json:"manufacturer"`
	TrackSystem  string              `json:"trackSystem"`
	Gauge        string              `json:"gauge"`
	Scale        string              `json:"scale"`
	Version      string              `json:"version"`
	SourceURL    string              `json:"sourceUrl"`
	Status       TrackGeometryStatus `json:"status"`
}

type TrackLibraryPackageDefinition struct {
	ArticleNumber   string              `json:"articleNumber"`
	Name            string              `json:"name"`
	Kind            TrackGeometryKind   `json:"kind"`
	LengthMM        float64             `json:"lengthMm"`
	MinimumRadiusMM *float64            `json:"minimumRadiusMm,omitempty"`
	Geometry        TrackGeometry       `json:"geometry"`
	SourceURL       string              `json:"sourceUrl"`
	Status          TrackGeometryStatus `json:"status"`
}

func ValidateTrackLibraryPackage(doc TrackLibraryPackage) error {
	if doc.Format != TrackLibraryPackageFormat || doc.SchemaVersion != TrackLibraryPackageSchemaVersion ||
		!validTrackLibraryMetadata(doc.Library) || len(doc.Definitions) == 0 ||
		len(doc.Definitions) > MaxTrackLibraryDefinitions {
		return ErrInvalidTrackLibraryPackage
	}
	articles := make(map[string]struct{}, len(doc.Definitions))
	for _, definition := range doc.Definitions {
		articleKey := strings.ToLower(strings.TrimSpace(definition.ArticleNumber))
		if _, duplicate := articles[articleKey]; duplicate || !validTrackLibraryDefinition(definition) {
			return ErrInvalidTrackLibraryPackage
		}
		articles[articleKey] = struct{}{}
	}
	return nil
}

func validTrackLibraryMetadata(metadata TrackLibraryPackageMetadata) bool {
	return validBoundedText(metadata.Manufacturer, maxTrackLibraryMetadataLength) &&
		validBoundedText(metadata.TrackSystem, maxTrackLibraryMetadataLength) &&
		validBoundedText(metadata.Gauge, maxTrackLibraryGaugeScaleLength) &&
		validBoundedText(metadata.Scale, maxTrackLibraryGaugeScaleLength) &&
		validBoundedText(metadata.Version, maxTrackLibraryVersionLength) &&
		validTrackLibraryURL(metadata.SourceURL) && metadata.Status.Valid()
}

func validTrackLibraryDefinition(definition TrackLibraryPackageDefinition) bool {
	if !validBoundedText(definition.ArticleNumber, maxTrackLibraryArticleNumberLength) ||
		!validBoundedText(definition.Name, MaxTrackLibraryDefinitionName) || !definition.Kind.Valid() ||
		!finiteTrackLibraryNumber(definition.LengthMM) || definition.LengthMM <= 0 ||
		definition.LengthMM > maxTrackLibraryLengthMM || !validTrackLibraryURL(definition.SourceURL) ||
		!definition.Status.Valid() {
		return false
	}
	if definition.MinimumRadiusMM != nil &&
		(!finiteTrackLibraryNumber(*definition.MinimumRadiusMM) || *definition.MinimumRadiusMM <= 0 ||
			*definition.MinimumRadiusMM > maxTrackLibraryLengthMM) {
		return false
	}
	return validImportedTrackGeometry(definition.Geometry)
}

func validImportedTrackGeometry(geometry TrackGeometry) bool {
	if geometry.SchemaVersion != 1 || len(geometry.Ports) < 2 || len(geometry.Ports) > maxTrackLibraryPorts ||
		len(geometry.Routes) == 0 || len(geometry.Routes) > maxTrackLibraryRoutes {
		return false
	}
	portIDs := make(map[string]struct{}, len(geometry.Ports))
	for _, port := range geometry.Ports {
		if !validGeometryID(port.ID) || !validTrackLibraryCoordinate(port.XMM) ||
			!validTrackLibraryCoordinate(port.YMM) || !finiteTrackLibraryNumber(port.DirectionDegrees) ||
			port.DirectionDegrees < 0 || port.DirectionDegrees >= 360 {
			return false
		}
		if _, duplicate := portIDs[port.ID]; duplicate {
			return false
		}
		portIDs[port.ID] = struct{}{}
	}
	routeIDs := make(map[string]struct{}, len(geometry.Routes))
	for _, route := range geometry.Routes {
		if !validGeometryID(route.ID) || len(route.Points) < 2 || len(route.Points) > maxTrackLibraryRoutePoints {
			return false
		}
		if _, duplicate := routeIDs[route.ID]; duplicate {
			return false
		}
		routeIDs[route.ID] = struct{}{}
		for _, point := range route.Points {
			if !validTrackLibraryCoordinate(point.XMM) || !validTrackLibraryCoordinate(point.YMM) {
				return false
			}
		}
	}
	return true
}

func validBoundedText(value string, maximum int) bool {
	length := len([]rune(strings.TrimSpace(value)))
	return length > 0 && length <= maximum
}

func validTrackLibraryURL(value string) bool {
	if !validBoundedText(value, maxTrackLibraryURLLength) {
		return false
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func validGeometryID(value string) bool {
	return validBoundedText(value, maxTrackLibraryGeometryIDLength)
}

func validTrackLibraryCoordinate(value float64) bool {
	return finiteTrackLibraryNumber(value) && math.Abs(value) <= maxTrackLibraryCoordinateMM
}

func finiteTrackLibraryNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
