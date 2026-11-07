package decoder

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func CheckSnapshotHardware(raw, installation string, confirmedDecoderID *string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fields); err != nil || fields == nil {
		return fmt.Errorf("invalid hardware snapshot")
	}
	if installation == "" {
		if len(fields) != 0 {
			return fmt.Errorf("unconfirmed hardware must remain unknown")
		}
		return nil
	}
	var hardware DecoderSnapshotHardware
	if err := StrictSnapshotJSON(raw, &hardware); err != nil || hardware.InstallationID != installation ||
		(confirmedDecoderID == nil || hardware.DecoderID != *confirmedDecoderID) || strings.TrimSpace(hardware.DecoderLabel) == "" {
		return fmt.Errorf("invalid captured decoder identity")
	}
	for _, key := range []string{"decoderId", "decoderLabel", "manufacturer", "model", "serialNumber", "installationId", "installationState"} {
		var text string
		if fields[key] == nil || string(fields[key]) == "null" || json.Unmarshal(fields[key], &text) != nil {
			return fmt.Errorf("missing captured hardware field %s", key)
		}
	}
	for _, key := range []string{"installedAt", "removedAt"} {
		if fields[key] == nil {
			return fmt.Errorf("missing nullable hardware date %s", key)
		}
		if string(fields[key]) == "null" {
			continue
		}
		var date string
		if json.Unmarshal(fields[key], &date) != nil {
			return fmt.Errorf("invalid captured hardware date")
		}
		if _, err := time.Parse(time.DateOnly, date); err != nil {
			return fmt.Errorf("invalid captured hardware date")
		}
	}
	if hardware.InstallationState != "installed" && hardware.InstallationState != "removed" && hardware.InstallationState != "unresolved" {
		return fmt.Errorf("invalid captured installation state")
	}
	if hardware.RemovedAt != nil && (hardware.InstallationState != "removed" ||
		(hardware.InstalledAt != nil && *hardware.RemovedAt < *hardware.InstalledAt)) {
		return fmt.Errorf("invalid captured installation date range")
	}
	return nil
}
