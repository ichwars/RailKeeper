package application

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var decoderConfigurationBackupTables = []string{
	"decoder_instances", "decoder_installations", "decoder_cv_workspaces", "decoder_cv_snapshots", "decoder_cv_snapshot_sources",
}

func isDecoderConfigurationBackupTable(table string) bool {
	for _, known := range decoderConfigurationBackupTables {
		if table == known {
			return true
		}
	}
	return false
}

func validateBackupDecoderConfiguration(doc *BackupDocument) []string {
	if err := checkBackupDecoderConfiguration(doc); err != nil {
		return []string{"Decoder-/CV-Stände ungültig. / Invalid decoder/CV records: " + err.Error()}
	}
	return nil
}

func decoderBackupRows(doc *BackupDocument, table string) (map[string]map[string]any, error) {
	rows := map[string]map[string]any{}
	for _, row := range doc.Tables[table] {
		id, ok := backupNonEmptyString(row["id"])
		if !ok || rows[id] != nil {
			return nil, fmt.Errorf("%s: missing or duplicate identity", table)
		}
		rows[id] = row
	}
	return rows, nil
}

func decoderBackupText(row map[string]any, fields ...string) error {
	for _, field := range fields {
		if _, ok := row[field].(string); !ok {
			return fmt.Errorf("missing text field %s", field)
		}
	}
	return nil
}

func decoderBackupTime(row map[string]any, fields ...string) error {
	for _, field := range fields {
		value, ok := row[field].(string)
		if _, err := time.Parse(time.RFC3339Nano, value); !ok || err != nil {
			return fmt.Errorf("invalid timestamp %s", field)
		}
	}
	return nil
}

func decoderBackupDate(row map[string]any, field string) (string, error) {
	value, exists := row[field]
	if !exists {
		return "", fmt.Errorf("missing nullable date %s", field)
	}
	if value == nil {
		return "", nil
	}
	text, ok := value.(string)
	if _, err := time.Parse(time.DateOnly, text); !ok || err != nil {
		return "", fmt.Errorf("invalid date %s", field)
	}
	return text, nil
}

func decoderBackupReference(row map[string]any, field string, targets map[string]map[string]any) (string, error) {
	value, exists := row[field]
	if !exists {
		return "", fmt.Errorf("missing nullable reference %s", field)
	}
	if value == nil {
		return "", nil
	}
	id, ok := backupNonEmptyString(value)
	if !ok || targets[id] == nil {
		return "", fmt.Errorf("invalid reference %s", field)
	}
	return id, nil
}

func checkBackupDecoderConfiguration(doc *BackupDocument) error {
	for _, table := range decoderConfigurationBackupTables {
		if doc.Version < 27 && len(doc.Tables[table]) > 0 {
			return fmt.Errorf("%s requires backup format 27 or newer", table)
		}
	}
	indexes := map[string]map[string]map[string]any{}
	for _, table := range []string{"vehicles", "file_blobs", "decoder_instances", "decoder_installations", "decoder_cv_workspaces", "decoder_cv_snapshots"} {
		rows, err := decoderBackupRows(doc, table)
		if err != nil {
			return err
		}
		indexes[table] = rows
	}
	instances, installations := indexes["decoder_instances"], indexes["decoder_installations"]
	workspaces, snapshots := indexes["decoder_cv_workspaces"], indexes["decoder_cv_snapshots"]
	for _, row := range instances {
		if err := decoderBackupText(row, "label", "manufacturer", "model", "serial_number", "notes"); err != nil {
			return err
		}
		if strings.TrimSpace(row["label"].(string)) == "" {
			return fmt.Errorf("decoder label is empty")
		}
		if err := decoderBackupTime(row, "created_at", "updated_at"); err != nil {
			return err
		}
	}
	installed := map[string]bool{}
	for _, row := range installations {
		decoder, ok := row["decoder_id"].(string)
		if !ok || instances[decoder] == nil {
			return fmt.Errorf("installation has no decoder")
		}
		if err := decoderBackupText(row, "vehicle_reference_id", "vehicle_name", "vehicle_inventory_number", "state", "notes"); err != nil {
			return err
		}
		vehicle, err := decoderBackupReference(row, "vehicle_id", indexes["vehicles"])
		if err != nil || strings.TrimSpace(row["vehicle_reference_id"].(string)) == "" ||
			(vehicle != "" && vehicle != row["vehicle_reference_id"]) {
			return fmt.Errorf("installation vehicle reference is invalid")
		}
		state := row["state"].(string)
		if state != "installed" && state != "removed" && state != "unresolved" {
			return fmt.Errorf("invalid installation state")
		}
		if state == "installed" {
			if vehicle == "" || installed[decoder] {
				return fmt.Errorf("missing vehicle or duplicate active decoder installation")
			}
			installed[decoder] = true
		}
		start, err := decoderBackupDate(row, "installed_at")
		if err != nil {
			return err
		}
		end, err := decoderBackupDate(row, "removed_at")
		if err != nil || (start != "" && end != "" && end < start) || (end != "" && state != "removed") {
			return fmt.Errorf("invalid installation date range")
		}
		if err := decoderBackupTime(row, "recorded_at", "updated_at"); err != nil {
			return err
		}
	}
	profiles := map[string]bool{}
	for _, row := range workspaces {
		if err := decoderBackupText(row, "vehicle_id", "profile_text", "origin"); err != nil {
			return err
		}
		vehicle := row["vehicle_id"].(string)
		key := vehicle + "\x00" + row["profile_text"].(string)
		if indexes["vehicles"][vehicle] == nil || profiles[key] || (row["origin"] != "legacy" && row["origin"] != "manual") {
			return fmt.Errorf("invalid or duplicate working profile")
		}
		profiles[key] = true
		installation, err := decoderBackupReference(row, "installation_id", installations)
		if err != nil || (installation != "" && installations[installation]["vehicle_reference_id"] != vehicle) {
			return fmt.Errorf("working profile installation belongs to another vehicle")
		}
		if err := decoderBackupTime(row, "created_at", "updated_at"); err != nil {
			return err
		}
	}
	sources, err := checkDecoderSnapshotSources(doc.Tables["decoder_cv_snapshot_sources"], snapshots, indexes["file_blobs"])
	if err != nil {
		return err
	}
	for id, row := range snapshots {
		if err := decoderBackupText(row, "workspace_id", "vehicle_reference_id", "vehicle_name", "vehicle_inventory_number", "profile_text", "kind", "scope", "payload_json", "hardware_json", "note"); err != nil {
			return err
		}
		vehicle, workspace := row["vehicle_reference_id"].(string), row["workspace_id"].(string)
		if vehicle == "" || workspace == "" || (indexes["vehicles"][vehicle] != nil && workspaces[workspace] == nil) ||
			(workspaces[workspace] != nil && (workspaces[workspace]["vehicle_id"] != vehicle || workspaces[workspace]["profile_text"] != row["profile_text"])) {
			return fmt.Errorf("invalid historical working profile")
		}
		kind := row["kind"]
		if kind != "legacy" && kind != "manual" && kind != "file_import" && kind != "station_import" && kind != "checkpoint" {
			return fmt.Errorf("invalid snapshot kind")
		}
		if row["scope"] != "partial_working_state" {
			return fmt.Errorf("snapshot cannot claim a complete measurement")
		}
		if err := decoderBackupTime(row, "recorded_at"); err != nil {
			return err
		}
		installation, err := decoderBackupReference(row, "installation_id", installations)
		if err != nil || (installation != "" && installations[installation]["vehicle_reference_id"] != vehicle) {
			return fmt.Errorf("invalid snapshot installation")
		}
		if err := checkDecoderSnapshotHardware(row["hardware_json"].(string), installation, installations); err != nil {
			return err
		}
		if err := checkDecoderSnapshotPayload(row["payload_json"].(string), vehicle, row["profile_text"].(string), sources[id]); err != nil {
			return err
		}
	}
	return nil
}

func checkDecoderSnapshotSources(rows []map[string]any, snapshots, blobs map[string]map[string]any) (map[string]map[string]bool, error) {
	result := map[string]map[string]bool{}
	for _, row := range rows {
		if err := decoderBackupText(row, "snapshot_id", "source_file_id", "metadata_json", "content_state"); err != nil {
			return nil, err
		}
		snapshot, source := row["snapshot_id"].(string), row["source_file_id"].(string)
		if snapshots[snapshot] == nil || source == "" || result[snapshot][source] {
			return nil, fmt.Errorf("missing snapshot or duplicate original source")
		}
		if result[snapshot] == nil {
			result[snapshot] = map[string]bool{}
		}
		result[snapshot][source] = true
		blob, err := decoderBackupReference(row, "blob_id", blobs)
		if err != nil || (row["content_state"] == "stored" && blob == "") ||
			(row["content_state"] == "unavailable" && blob != "") || (row["content_state"] != "stored" && row["content_state"] != "unavailable") {
			return nil, fmt.Errorf("invalid frozen source content reference")
		}
		var metadata VehicleCVFile
		if err := strictDecoderSnapshotJSON(row["metadata_json"].(string), &metadata); err != nil || metadata.ID != source ||
			metadata.SizeBytes < 0 || (row["content_state"] == "stored" && metadata.VehicleID != snapshots[snapshot]["vehicle_reference_id"]) {
			return nil, fmt.Errorf("invalid original source metadata")
		}
	}
	return result, nil
}

func checkDecoderSnapshotHardware(raw, installation string, installations map[string]map[string]any) error {
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
	if err := strictDecoderSnapshotJSON(raw, &hardware); err != nil || hardware.InstallationID != installation ||
		hardware.DecoderID != installations[installation]["decoder_id"] || strings.TrimSpace(hardware.DecoderLabel) == "" {
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
