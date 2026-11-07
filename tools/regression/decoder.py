"""Independent expected decoder-model changes for preservation checks."""

import base64
import copy
import datetime as dt
import json
import re
import zlib

TABLES = ("decoder_instances", "decoder_installations", "decoder_cv_workspaces",
          "decoder_cv_snapshots", "decoder_cv_snapshot_sources")

NOTE = "Übernommener Arbeitsbestand; kein vollständiger Auslesestand. / Adopted working state; not a complete readout."


def _json(raw):
    def pairs(items):
        result = {}
        for key, value in items:
            if key in result:
                raise AssertionError("Duplicate JSON evidence key")
            result[key] = value
        return result
    return json.loads(raw, object_pairs_hook=pairs)


def _equal(left, right, label):
    from preservation import assert_same, digest
    assert_same({"sha256": digest(left)}, {"sha256": digest(right)}, label)


def _timestamp(value):
    parsed = dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
    if parsed.tzinfo is None:
        raise AssertionError("Recording time must include a timezone")
    return parsed


def _identity_time(row, document, observed, fields):
    if re.fullmatch(r"[0-9a-f]{32}", row.get("id", "")) is None:
        raise AssertionError("Unexpected generated decoder identity")
    earliest = _timestamp(document["createdAt"]) - dt.timedelta(seconds=1)
    latest = _timestamp(observed["createdAt"]) + dt.timedelta(seconds=1)
    for field in fields:
        value = row.get(field)
        if not isinstance(value, str) or not value.endswith("Z") or not earliest <= _timestamp(value) <= latest:
            raise AssertionError("Decoder recording time is not a new adoption timestamp")


def _cv(row, histories):
    item = {"id": row["id"], "vehicleId": row["vehicle_id"], "cvNumber": row["cv_number"], "value": row["value"],
            "createdAt": row["created_at"], "updatedAt": row["updated_at"]}
    for source, target in (("index_cv31", "indexCV31"), ("index_cv32", "indexCV32"), ("source_line", "sourceLine")):
        if row.get(source) is not None:
            item[target] = row[source]
    for source, target in (("description", "description"), ("category", "category"), ("protocol", "protocol"),
                           ("decoder_profile", "decoderProfile"), ("source_file_id", "sourceFileId"),
                           ("source_exported_at", "sourceExportedAt")):
        if row.get(source):
            item[target] = row[source]
    history = [{"id": h["id"], "cvValueId": h["cv_value_id"], "vehicleId": h["vehicle_id"],
                "oldValue": h["old_value"], "newValue": h["new_value"], "changedAt": h["changed_at"]}
               for h in histories if h["cv_value_id"] == row["id"]]
    if history:
        item["history"] = sorted(history, key=lambda value: value["id"])
    return item


def _file(row):
    result = {target: row.get(source) or "" for source, target in
              (("id", "id"), ("vehicle_id", "vehicleId"), ("file_name", "fileName"),
               ("original_name", "originalName"), ("created_at", "createdAt"), ("updated_at", "updatedAt"))}
    result["sizeBytes"] = row.get("size_bytes", 0)
    for source, target in (("description", "description"), ("decoder_profile", "decoderProfile"), ("mime_type", "mimeType")):
        if row.get(source):
            result[target] = row[source]
    return result


def _archive_source(tables, vehicle, source):
    snapshots = {row["id"]: row for row in tables["decoder_cv_snapshots"]}
    candidates = [row for row in tables["decoder_cv_snapshot_sources"] if row["source_file_id"] == source
                  and row["content_state"] == "stored" and snapshots[row["snapshot_id"]]["vehicle_reference_id"] == vehicle]
    if not candidates:
        return None
    latest = max(candidates, key=lambda row: (snapshots[row["snapshot_id"]]["recorded_at"], row["snapshot_id"]))
    if len({row["blob_id"] for row in candidates}) == 1:
        return latest
    blobs = {row["id"]: row for row in tables.get("file_blobs", [])}
    contents = set()
    for candidate in candidates:
        blob = blobs[candidate["blob_id"]]
        data = base64.b64decode(blob["data"]) if isinstance(blob["data"], str) else blob["data"]
        try:
            compression = blob.get("compression", "").lower()
            if compression not in ("", "none", "zlib"):
                return None
            contents.add(zlib.decompress(data) if compression == "zlib" else data)
        except zlib.error:
            return None
    if len(contents) != 1:
        return None
    return latest


def expected_decoder_backfill(document, observed):
    """Derive every added group, CV/history payload and source from the old data, not row counts."""
    expected = copy.deepcopy(document)
    tables, actual = expected["tables"], observed["tables"]
    for table in TABLES:
        tables.setdefault(table, [])
    vehicles = {row["id"]: row for row in tables["vehicles"]}
    cvs, files = tables.get("vehicle_cv_values", []), tables.get("vehicle_cv_files", [])
    workspaces = {(row["vehicle_id"], row["profile_text"]): row for row in tables["decoder_cv_workspaces"]}
    pairs = {(row["vehicle_id"], row.get("decoder_profile") or "") for row in cvs + files}
    original_ids = {row["id"] for row in tables["decoder_cv_workspaces"]}
    for vehicle, profile in sorted(pairs - set(workspaces)):
        candidates = [row for row in actual["decoder_cv_workspaces"]
                      if row["vehicle_id"] == vehicle and row["profile_text"] == profile]
        if len(candidates) != 1:
            raise AssertionError("Missing or duplicate expected legacy decoder group")
        row = candidates[0]
        _identity_time(row, document, observed, ("created_at", "updated_at"))
        if row["id"] in original_ids or row["created_at"] != row["updated_at"]:
            raise AssertionError("Legacy decoder group was not recorded independently")
        item = {"id": row["id"], "vehicle_id": vehicle, "profile_text": profile, "origin": "legacy",
                "installation_id": None, "created_at": row["created_at"], "updated_at": row["updated_at"]}
        tables["decoder_cv_workspaces"].append(item)
        original_ids.add(item["id"])
        workspaces[(vehicle, profile)] = item
    source_snapshots = copy.deepcopy(tables["decoder_cv_snapshots"])
    for (vehicle, profile), workspace in workspaces.items():
        if workspace["origin"] != "legacy" or any(row["workspace_id"] == workspace["id"] and row["kind"] == "legacy"
                                                  for row in source_snapshots):
            continue
        candidates = [row for row in actual["decoder_cv_snapshots"] if row["workspace_id"] == workspace["id"]
                      and row["kind"] == "legacy" and row["id"] not in {old["id"] for old in source_snapshots}]
        if len(candidates) != 1:
            raise AssertionError("Missing or duplicate adopted decoder partial state")
        snapshot = candidates[0]
        _identity_time(snapshot, document, observed, ("recorded_at",))
        values = [_cv(row, tables.get("vehicle_cv_value_history", [])) for row in cvs
                  if row["vehicle_id"] == vehicle and (row.get("decoder_profile") or "") == profile]
        payload = _json(snapshot["payload_json"])
        for value in payload.get("values", []):
            if "history" in value:
                value["history"] = sorted(value["history"], key=lambda row: row["id"])
        payload["values"] = sorted(payload.get("values", []), key=lambda row: row["id"])
        _equal({"values": sorted(values, key=lambda row: row["id"])}, payload, "Complete adopted CV/history/provenance payload")
        hardware = {}
        installation_id = workspace["installation_id"]
        if installation_id:
            installation = next(row for row in tables["decoder_installations"] if row["id"] == installation_id)
            decoder = next(row for row in tables["decoder_instances"] if row["id"] == installation["decoder_id"])
            hardware = {"decoderId": decoder["id"], "decoderLabel": decoder["label"], "manufacturer": decoder["manufacturer"],
                        "model": decoder["model"], "serialNumber": decoder["serial_number"], "installationId": installation_id,
                        "installationState": installation["state"], "installedAt": installation["installed_at"], "removedAt": installation["removed_at"]}
        _equal(hardware, _json(snapshot["hardware_json"]), "Confirmed hardware snapshot only")
        item = {"id": snapshot["id"], "workspace_id": workspace["id"], "vehicle_reference_id": vehicle,
                "vehicle_name": vehicles[vehicle]["name"], "vehicle_inventory_number": vehicles[vehicle]["inventory_number"],
                "profile_text": profile, "installation_id": installation_id, "kind": "legacy", "scope": "partial_working_state",
                "payload_json": snapshot["payload_json"], "hardware_json": snapshot["hardware_json"], "note": NOTE,
                "recorded_at": snapshot["recorded_at"]}
        tables["decoder_cv_snapshots"].append(item)
        source_ids = {row["id"] for row in files if row["vehicle_id"] == vehicle and (row.get("decoder_profile") or "") == profile}
        source_ids.update(row["source_file_id"] for row in cvs if row["vehicle_id"] == vehicle
                          and (row.get("decoder_profile") or "") == profile and row.get("source_file_id"))
        for source in sorted(source_ids):
            matches = [row for row in actual["decoder_cv_snapshot_sources"] if row["snapshot_id"] == snapshot["id"] and row["source_file_id"] == source]
            if len(matches) != 1:
                raise AssertionError("Missing or duplicate original decoder source")
            row = matches[0]
            original = next((row for row in files if row["id"] == source), None)
            metadata = _file(original or {"id": source})
            state, blob = "unavailable", None
            if original and original["vehicle_id"] == vehicle and original.get("blob_id"):
                state, blob = "stored", original["blob_id"]
            elif not original or original["vehicle_id"] == vehicle:
                if original and any(file["path"] == original.get("storage_path") for file in document.get("files", [])):
                    raise AssertionError("Filesystem originals require their dedicated byte/rollback probe")
                archive = _archive_source(document["tables"], vehicle, source)
                if archive:
                    metadata, state, blob = _json(archive["metadata_json"]), "stored", archive["blob_id"]
            _equal(metadata, _json(row["metadata_json"]), "Exact frozen original source metadata")
            tables["decoder_cv_snapshot_sources"].append({"snapshot_id": snapshot["id"], "source_file_id": source,
                "metadata_json": row["metadata_json"], "content_state": state, "blob_id": blob})
    return expected
