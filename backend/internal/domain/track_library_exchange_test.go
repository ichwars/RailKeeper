package domain

import (
	"math"
	"strings"
	"testing"
)

func TestValidateTrackLibraryPackageAcceptsVersionedRailKeeperDocument(t *testing.T) {
	doc := validTrackLibraryPackage()
	if err := ValidateTrackLibraryPackage(doc); err != nil {
		t.Fatalf("validate package: %v", err)
	}
}

func TestValidateTrackLibraryPackageRejectsInvalidDocuments(t *testing.T) {
	tests := map[string]func(*TrackLibraryPackage){
		"format":       func(doc *TrackLibraryPackage) { doc.Format = "scarm" },
		"schema":       func(doc *TrackLibraryPackage) { doc.SchemaVersion = 2 },
		"manufacturer": func(doc *TrackLibraryPackage) { doc.Library.Manufacturer = " " },
		"source scheme": func(doc *TrackLibraryPackage) {
			doc.Library.SourceURL = "file:///tmp/library.json"
		},
		"library status": func(doc *TrackLibraryPackage) { doc.Library.Status = "trusted" },
		"empty":          func(doc *TrackLibraryPackage) { doc.Definitions = nil },
		"too many": func(doc *TrackLibraryPackage) {
			doc.Definitions = make([]TrackLibraryPackageDefinition, MaxTrackLibraryDefinitions+1)
		},
		"duplicate article": func(doc *TrackLibraryPackage) {
			doc.Definitions = append(doc.Definitions, doc.Definitions[0])
		},
		"invalid kind": func(doc *TrackLibraryPackage) { doc.Definitions[0].Kind = "spiral" },
		"non finite length": func(doc *TrackLibraryPackage) {
			doc.Definitions[0].LengthMM = math.NaN()
		},
		"invalid radius": func(doc *TrackLibraryPackage) {
			zero := 0.0
			doc.Definitions[0].MinimumRadiusMM = &zero
		},
		"duplicate port": func(doc *TrackLibraryPackage) {
			doc.Definitions[0].Geometry.Ports[1].ID = "a"
		},
		"invalid direction": func(doc *TrackLibraryPackage) {
			doc.Definitions[0].Geometry.Ports[0].DirectionDegrees = 360
		},
		"short route": func(doc *TrackLibraryPackage) {
			doc.Definitions[0].Geometry.Routes[0].Points = doc.Definitions[0].Geometry.Routes[0].Points[:1]
		},
		"oversized name": func(doc *TrackLibraryPackage) {
			doc.Definitions[0].Name = strings.Repeat("x", MaxTrackLibraryDefinitionName+1)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			doc := validTrackLibraryPackage()
			mutate(&doc)
			if err := ValidateTrackLibraryPackage(doc); err == nil {
				t.Fatal("expected invalid package")
			}
		})
	}
}

func validTrackLibraryPackage() TrackLibraryPackage {
	return TrackLibraryPackage{
		Format: TrackLibraryPackageFormat, SchemaVersion: TrackLibraryPackageSchemaVersion,
		ExportedAt: "2026-08-10T08:00:00Z",
		Library: TrackLibraryPackageMetadata{
			Manufacturer: "Kühn", TrackSystem: "TT", Gauge: "TT", Scale: "1:120",
			Version: "2026.1", SourceURL: "https://example.com/catalogue.pdf",
			Status: TrackGeometryVerified,
		},
		Definitions: []TrackLibraryPackageDefinition{{
			ArticleNumber: "72620", Name: "Gerades Gleis", Kind: TrackGeometryStraight,
			LengthMM: 128, SourceURL: "https://example.com/72620", Status: TrackGeometryVerified,
			Geometry: TrackGeometry{SchemaVersion: 1,
				Ports: []TrackPort{
					{ID: "a", XMM: 0, YMM: 0, DirectionDegrees: 180},
					{ID: "b", XMM: 128, YMM: 0, DirectionDegrees: 0},
				},
				Routes: []TrackRoute{{ID: "main", Points: []TrackPoint{{XMM: 0, YMM: 0}, {XMM: 128, YMM: 0}}}},
			},
		}},
	}
}
