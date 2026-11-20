package api

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestAttachmentArchiveKeepsSelectedBytesAndEnforcesVehicleScope(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	base := "/api/v1/vehicles/" + id
	upload := func(contents string) string {
		response := f.upload(t, base+"/attachments", "Manual.txt", []byte(contents), nil)
		assertStatus(t, response, 201)
		return responseID(t, response)
	}
	first, second := upload("first"), upload("second")
	upload("omitted")
	query := "?file=" + second + "&file=" + first
	result := authedJSONRequest(t, f.router, f.viewer, "GET", base+"/attachment-archive"+query, nil, true)
	assertStatus(t, result, 200)
	raw := result.Body.Bytes()
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil || len(reader.File) != 2 {
		t.Fatalf("unexpected ZIP result: %v", err)
	}
	for i, want := range []string{"second", "first"} {
		file, err := reader.File[i].Open()
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil || string(got) != want || reader.File[0].Name == reader.File[1].Name {
			t.Fatalf("entry %d: %q, %v", i, got, err)
		}
	}
	assertStatus(t, f.call(t, "GET", base+"/attachment-archive?file="+first+"&file="+first, nil), 400)
	assertStatus(t, f.call(t, "GET", "/api/v1/vehicles/"+f.vehicle(t)+"/attachment-archive"+query, nil), 404)
}
