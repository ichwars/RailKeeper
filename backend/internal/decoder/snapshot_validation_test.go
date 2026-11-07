package decoder

import (
	"encoding/json"
	"testing"
)

func TestCapturedHardwareKeepsUnknownDatesAndRequiresConfirmedIdentity(t *testing.T) {
	id := "decoder"
	hardware := DecoderSnapshotHardware{DecoderID: id, DecoderLabel: "Physical decoder", InstallationID: "installation", InstallationState: "installed"}
	data, err := json.Marshal(hardware)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckSnapshotHardware(string(data), "installation", &id); err != nil {
		t.Fatal(err)
	}
	if err := CheckSnapshotHardware("{}", "", nil); err != nil {
		t.Fatal("unknown hardware must remain readable")
	}
	for _, test := range []struct {
		raw, installation string
		confirmed         *string
	}{
		{string(data), "", nil}, {string(data), "installation", nil},
		{string(data), "other-installation", &id},
		{`{"decoderId":"decoder","decoderLabel":"Physical decoder","manufacturer":"","model":"","serialNumber":"","installationId":"installation","installationState":"installed"}`, "installation", &id},
	} {
		if err := CheckSnapshotHardware(test.raw, test.installation, test.confirmed); err == nil {
			t.Fatal("unconfirmed identity or missing explicit unknown dates accepted")
		}
	}
}

func FuzzSnapshotPayload(f *testing.F) {
	f.Add(`{"values":[]}`)
	f.Add(`{"values":[{"id":"cv","vehicleId":"vehicle","cvNumber":257,"indexCV31":0,"indexCV32":0,"value":3,"createdAt":"","updatedAt":""}]}`)
	f.Add(`{"values":null}`)
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 65536 {
			t.Skip()
		}
		if CheckSnapshotPayload(raw, "vehicle", "", nil) != nil {
			return
		}
		var payload DecoderCVSnapshotPayload
		if err := StrictSnapshotJSON(raw, &payload); err != nil {
			t.Fatal("accepted payload cannot be decoded", err)
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		if err := CheckSnapshotPayload(string(encoded), "vehicle", "", nil); err != nil {
			t.Fatal("accepted partial state lost identity on round trip", err)
		}
	})
}
