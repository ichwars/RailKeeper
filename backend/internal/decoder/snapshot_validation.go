package decoder

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func StrictSnapshotJSON(raw string, target any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected one JSON value")
	}
	return nil
}

func CheckSnapshotPayload(raw, vehicle, profile string, sources map[string]bool) error {
	var payload DecoderCVSnapshotPayload
	if err := StrictSnapshotJSON(raw, &payload); err != nil || payload.Values == nil {
		return fmt.Errorf("invalid working-state payload")
	}
	var fields struct {
		Values []map[string]json.RawMessage `json:"values"`
	}
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return fmt.Errorf("invalid working-state fields")
	}
	identities, historyIDs, keys := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for i, value := range payload.Values {
		for _, key := range []string{"id", "vehicleId", "cvNumber", "value", "createdAt", "updatedAt"} {
			if field := fields.Values[i][key]; field == nil || string(field) == "null" {
				return fmt.Errorf("working-state CV is missing %s", key)
			}
		}
		for key, raw := range fields.Values[i] {
			if string(raw) == "null" && key != "indexCV31" && key != "indexCV32" && key != "sourceLine" && key != "history" {
				return fmt.Errorf("CV field %s cannot be null", key)
			}
		}
		if value.ID == "" || identities[value.ID] || value.VehicleID != vehicle || value.DecoderProfile != profile {
			return fmt.Errorf("invalid CV identity or working profile")
		}
		identities[value.ID] = true
		key := CVIdentity(VehicleCVValueInput{CVNumber: value.CVNumber, DecoderProfile: value.DecoderProfile,
			IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32})
		if keys[key] {
			return fmt.Errorf("duplicate CV identity within working snapshot")
		}
		keys[key] = true
		if (value.IndexCV31 == nil) != (value.IndexCV32 == nil) ||
			(value.IndexCV31 != nil && (*value.IndexCV31 < 0 || *value.IndexCV31 > 255 ||
				*value.IndexCV32 < 0 || *value.IndexCV32 > 255 || value.CVNumber < 257 || value.CVNumber > 512)) {
			return fmt.Errorf("invalid CV index page")
		}
		if (value.SourceLine != nil && (*value.SourceLine < 1 || *value.SourceLine > 10000000)) ||
			!ValidCVExportTimestamp(value.SourceExportedAt) || (value.SourceFileID != "" && !sources[value.SourceFileID]) {
			return fmt.Errorf("invalid CV source provenance")
		}
		for _, history := range value.History {
			if history.ID == "" || historyIDs[history.ID] || history.VehicleID != vehicle || history.CVValueID != value.ID {
				return fmt.Errorf("invalid CV history identity")
			}
			historyIDs[history.ID] = true
		}
	}
	return nil
}
