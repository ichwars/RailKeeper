package api

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestLokProgrammerCVListFullRoundTrip(t *testing.T) {
	text := lokProgrammerCVSample(t)
	preview := esuxPreviewResponse("decoder.txt", int64(len(text)), "text/plain", []byte(text))
	if preview.CVImportError != "" || len(preview.SuggestedCVValues) != 1367 {
		t.Fatalf("expected all 1367 CVs: count=%d error=%s", len(preview.SuggestedCVValues), preview.CVImportError)
	}
	service := application.NewVehicleService(testRouterDB(t))
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{
		Name: "CV import test", Category: "Lokomotive", Manufacturer: "Test", Gauge: "TT", Gattung: "Diesellok",
	}, "test")
	if err != nil {
		t.Fatal(err)
	}
	input := application.VehicleCVImportInput{}
	for _, row := range preview.SuggestedCVValues {
		input.Values = append(input.Values, application.VehicleCVValueInput{
			CVNumber: row.CVNumber, Value: row.Value, IndexCV31: row.IndexCV31, IndexCV32: row.IndexCV32,
			DecoderProfile: preview.SuggestedDecoderProfile, SourceLine: row.SourceLine, SourceExportedAt: row.SourceExportedAt,
		})
	}
	for range 2 {
		if count, err := service.ImportCVValues(t.Context(), vehicle.ID, input); err != nil || count != 1367 {
			t.Fatalf("full import failed: %d %v", count, err)
		}
	}
	stored, err := service.ListCVValues(t.Context(), vehicle.ID)
	if err != nil || len(stored) != 1367 {
		t.Fatalf("read imported values: %d %v", len(stored), err)
	}
	key := func(n int, high, low *int) string {
		if high == nil && low == nil {
			return fmt.Sprintf("base:%d", n)
		}
		return fmt.Sprintf("%d/%d:%d", *high, *low, n)
	}
	want := map[string]int{}
	pages := map[string]bool{}
	for _, row := range input.Values {
		want[key(row.CVNumber, row.IndexCV31, row.IndexCV32)] = row.Value
	}
	for _, row := range stored {
		id := key(row.CVNumber, row.IndexCV31, row.IndexCV32)
		if value, ok := want[id]; !ok || value != row.Value || len(row.History) != 0 || row.SourceLine == nil || row.SourceExportedAt != "2026-10-17T17:41:59" {
			t.Fatalf("value changed or merged: %s: %+v", id, row)
		}
		if row.IndexCV31 != nil {
			pages[fmt.Sprintf("%d/%d", *row.IndexCV31, *row.IndexCV32)] = true
		}
		delete(want, id)
	}
	if len(want) != 0 || len(pages) != 15 {
		t.Fatalf("missing values=%d pages=%d", len(want), len(pages))
	}
}

func TestLokProgrammerCVListRejectsAmbiguousOrMalformedInput(t *testing.T) {
	for _, text := range []string{
		"CV 257 = 1", "CV 1 = 256", "CV 1 = 3\nCV 1 = 4",
		"Index: 4099 (CV31=16, CV32=4)\nCV 257 = 1",
		"Index: 4099 (CV31=16, CV32=3)\nCV 257 = 1\nCV 257 = 2",
		"Index: broken\nCV 257 = 1", "Decoder: Test\nCV 1 = broken",
		"Index: 4099 (CV31=16, CV32=3)\nCV 513 = 1",
	} {
		rows, _, recognized, err := parseLokProgrammerCVList(text)
		if !recognized || err == nil || len(rows) != 0 {
			t.Fatalf("accepted invalid list: %q", text)
		}
	}
	rows, _, recognized, err := parseLokProgrammerCVList("CV 1 = 3\nIndex: 0 (CV31=0, CV32=0)\nCV 257 = 0")
	if !recognized || err != nil || len(rows) != 2 || rows[0].IndexCV31 != nil || rows[1].IndexCV31 == nil || *rows[1].IndexCV31 != 0 {
		t.Fatalf("zero index is distinct from base: %+v %v", rows, err)
	}
}

func lokProgrammerCVSample(t *testing.T) string {
	t.Helper()
	// The user's complete export can be exercised locally without committing private source files.
	if path := os.Getenv("RAILKEEPER_CV_FIXTURE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	var text strings.Builder
	text.WriteString("Decoder: Test decoder (1.2.3)\r\nFilename: test.esux\r\nCreated: 10/17/2026 17:41:59\r\n")
	for n := 1; n <= 132; n++ {
		fmt.Fprintf(&text, "CV %03d = %03d\r\n", n, n%256)
	}
	for _, page := range []struct{ high, low, count int }{
		{1, 0, 33}, {1, 1, 64}, {1, 2, 20}, {16, 0, 70}, {16, 3, 144}, {16, 4, 144}, {16, 5, 144}, {16, 6, 144},
		{16, 7, 72}, {16, 8, 80}, {16, 9, 80}, {16, 10, 80}, {16, 11, 80}, {16, 12, 40}, {16, 13, 40},
	} {
		fmt.Fprintf(&text, "\r\nIndex: %d (CV31=%d, CV32=%d)\r\n--------------------------------\r\n", page.high*256+page.low, page.high, page.low)
		for n := 0; n < page.count; n++ {
			fmt.Fprintf(&text, "CV %03d = %03d\r\n", 257+n, (n+page.low)%256)
		}
	}
	return text.String()
}

func TestLokProgrammerPreservesFirmwareHeaderWithoutInterpretingCV7(t *testing.T) {
	rows, profile, _, err := parseLokProgrammerCVList("Decoder: LokPilot 5 micro DCC (5.14.183)\nCV 007 = 255")
	if err != nil || len(rows) != 1 || profile != "LokPilot 5 micro DCC" || rows[0].Value != 255 ||
		rows[0].Description != "LokProgrammer Decoder: LokPilot 5 micro DCC (5.14.183)" {
		t.Fatalf("header provenance lost: %+v %q %v", rows, profile, err)
	}
}

func TestLokProgrammerCVProvenanceIncludesHeadersBlankLinesAndIndexes(t *testing.T) {
	text := "\ufeffDecoder: Test\r\nFilename: test.esux\r\nCreated: 10/17/2026 17:41:59\r\nCV 1 = 3\r\n\r\nIndex: 4099 (CV31=16, CV32=3)\r\n-------\r\nCV 257 = 2"
	rows, _, _, err := parseLokProgrammerCVList(text)
	if err != nil || len(rows) != 2 {
		t.Fatalf("parse: %v", err)
	}
	for i, line := range []int{4, 8} {
		if rows[i].SourceLine == nil || *rows[i].SourceLine != line || rows[i].SourceExportedAt != "2026-10-17T17:41:59" {
			t.Fatalf("wrong provenance: %+v", rows[i])
		}
	}
	for _, text := range []string{"Created: test\nCV 1 = 3", "Created: 2/30/2026 17:41:59\nCV 1 = 3", "Created: 10/17/2026 17:41:59\nCreated: 10/18/2026 17:41:59\nCV 1 = 3"} {
		if _, _, _, err := parseLokProgrammerCVList(text); err == nil {
			t.Fatal("accepted invalid export date")
		}
	}
}
