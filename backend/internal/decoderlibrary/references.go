package decoderlibrary

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const ReferenceMasterDataType = "decoder_reference"

//go:embed references/*.json
var referenceFiles embed.FS

// CVReferenceRow is a source-language register fact, not an executable interpretation rule.
// Unknown index pages and manual-wide applicability stay explicit.
type CVReferenceRow struct {
	CV            int    `json:"cv"`
	CVEnd         int    `json:"cvEnd,omitempty"`
	Label         string `json:"label"`
	ValueRange    string `json:"valueRange,omitempty"`
	FactoryValue  string `json:"factoryValue,omitempty"`
	Page          int    `json:"page"`
	SourceIndex   int    `json:"sourceIndex"`
	IndexScope    string `json:"indexScope"`
	IndexCV31     *int   `json:"indexCV31,omitempty"`
	IndexCV32     *int   `json:"indexCV32,omitempty"`
	Applicability string `json:"applicability"`
}

type CVReference struct {
	ID       string           `json:"id"`
	Revision string           `json:"revision"`
	ModelIDs []string         `json:"modelIds"`
	Title    Text             `json:"title"`
	Note     Text             `json:"note"`
	Status   string           `json:"status"`
	Sources  []Source         `json:"sources"`
	Rows     []CVReferenceRow `json:"rows"`
	Active   bool             `json:"active"`
}

// ReferenceCoverage keeps the frequently loaded decoder catalog small.
// Full source rows are retrieved only for the selected model.
type ReferenceCoverage struct {
	ID       string   `json:"id"`
	ModelIDs []string `json:"modelIds"`
	Status   string   `json:"status"`
	RowCount int      `json:"rowCount"`
	Active   bool     `json:"active"`
}

func (r CVReference) Coverage() ReferenceCoverage {
	return ReferenceCoverage{ID: r.ID, ModelIDs: r.ModelIDs, Status: r.Status, RowCount: len(r.Rows), Active: r.Active}
}

func BundledReferences() []CVReference {
	files, err := referenceFiles.ReadDir("references")
	if err != nil {
		panic(err)
	}
	result := []CVReference{}
	ids := map[string]bool{}
	for _, file := range files {
		body, err := referenceFiles.ReadFile("references/" + file.Name())
		if err != nil {
			panic(err)
		}
		var references []CVReference
		if err := json.Unmarshal(body, &references); err != nil {
			panic(err)
		}
		for _, reference := range references {
			reference.Active = true
			if err := reference.Validate(); err != nil {
				panic(fmt.Errorf("reference %s: %w", reference.ID, err))
			}
			if ids[reference.ID] {
				panic("duplicate bundled CV reference")
			}
			ids[reference.ID] = true
			result = append(result, reference)
		}
	}
	return result
}

func ParseReference(metadata map[string]any) (CVReference, error) {
	var reference CVReference
	body, err := json.Marshal(metadata)
	if err != nil || len(body) > 512*1024 {
		return reference, errors.New("invalid decoder CV reference metadata")
	}
	if err := json.Unmarshal(body, &reference); err != nil {
		return reference, err
	}
	return reference, reference.Validate()
}

func (r CVReference) Validate() error {
	if !identifier.MatchString(r.ID) || !validLabel(r.Revision) || !validText(r.Title) || !validText(r.Note) ||
		len(r.ModelIDs) < 1 || len(r.ModelIDs) > 2048 || len(r.Sources) < 1 || len(r.Sources) > 10 || len(r.Rows) > 1024 {
		return errors.New("invalid CV reference identity or limits")
	}
	if r.Status != "table" && r.Status != "partial" && r.Status != "unavailable" {
		return errors.New("invalid CV reference status")
	}
	if (r.Status == "unavailable") != (len(r.Rows) == 0) {
		return errors.New("CV reference coverage contradicts rows")
	}
	models := map[string]bool{}
	for _, id := range r.ModelIDs {
		if !identifier.MatchString(id) || models[id] {
			return errors.New("invalid CV reference model")
		}
		models[id] = true
	}
	for _, source := range r.Sources {
		if !validSource(source) {
			return errors.New("invalid CV reference source")
		}
	}
	seen := map[string]bool{}
	for _, row := range r.Rows {
		end := row.CVEnd
		if end == 0 {
			end = row.CV
		}
		if row.CV < 1 || end < row.CV || end > 1024 || !validLabel(row.Label) || row.Page < 0 || row.Page > 1000 ||
			row.SourceIndex < 0 || row.SourceIndex >= len(r.Sources) || len(row.ValueRange) > 160 || len(row.FactoryValue) > 160 ||
			strings.ContainsAny(row.ValueRange+row.FactoryValue, "\x00\r\n") {
			return errors.New("invalid CV reference row")
		}
		if row.Applicability != "model" && row.Applicability != "manual" {
			return errors.New("invalid CV reference applicability")
		}
		switch row.IndexScope {
		case "base", "unspecified":
			if row.IndexCV31 != nil || row.IndexCV32 != nil {
				return errors.New("unexpected CV reference index")
			}
		case "indexed":
			if row.IndexCV31 == nil || row.IndexCV32 == nil || *row.IndexCV31 < 0 || *row.IndexCV31 > 255 || *row.IndexCV32 < 0 || *row.IndexCV32 > 255 || row.CV < 257 || end > 512 {
				return errors.New("invalid CV reference index")
			}
		default:
			return errors.New("invalid CV reference index scope")
		}
		key := fmt.Sprintf("%d/%d/%s/%v/%v/%d/%d", row.CV, end, row.IndexScope, pointerKey(row.IndexCV31), pointerKey(row.IndexCV32), row.SourceIndex, row.Page)
		if seen[key] {
			return errors.New("duplicate CV reference row")
		}
		seen[key] = true
	}
	return nil
}

func pointerKey(value *int) int {
	if value == nil {
		return -1
	}
	return *value
}
