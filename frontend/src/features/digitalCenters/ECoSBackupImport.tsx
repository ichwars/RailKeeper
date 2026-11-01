import { useEffect, useId, useRef, useState } from "react";

import { api, ApiError } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import type { ECoSBackupImportResult } from "./digitalCenterModel";

const maximumBackupBytes = 8 * 1024 * 1024;

export function ECoSBackupImport({ stationHost, disabled = false, onImported }: {
  stationHost: string;
  disabled?: boolean;
  onImported: (result: ECoSBackupImportResult) => Promise<void>;
}) {
  const { t } = useI18n();
  const headingId = useId();
  const hintId = useId();
  const [file, setFile] = useState<File | null>(null);
  const [confirmStation, setConfirmStation] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [summary, setSummary] = useState<ECoSBackupImportResult["summary"] | null>(null);
  const alive = useRef(true);
  const running = useRef(false);
  useEffect(() => {
    alive.current = true;
    return () => { alive.current = false; };
  }, []);

  const inspect = async () => {
    if (!file || disabled || running.current) return;
    running.current = true;
    setBusy(true);
    setError("");
    setSummary(null);
    try {
      const result = await api.importECoSBackup(file, { confirmStation, stationHost });
      if (!alive.current) return;
      setSummary(result.summary);
      await onImported(result);
    } catch (cause) {
      if (alive.current) setError(t(backupErrorKey(cause)));
    } finally {
      running.current = false;
      if (alive.current) setBusy(false);
    }
  };

  return <section className="ecos-backup-import" aria-labelledby={headingId}>
    <h2 id={headingId}>{t("digitalCenters.backup.importTitle")}</h2>
    <p>{t("digitalCenters.backup.importExplanation")}</p>
    <div className="ecos-backup-import-controls">
      <label className="ecos-backup-import-file">
        <span>{t("digitalCenters.backup.file")}</span>
        <input type="file" accept=".eco" disabled={busy || disabled} aria-describedby={hintId}
          onChange={(event) => {
            const selected = event.target.files?.[0] ?? null;
            setSummary(null);
            setConfirmStation(false);
            setError("");
            if (selected && (selected.size === 0 || !/\.eco$/i.test(selected.name))) {
              setFile(null);
              setError(t("digitalCenters.backup.invalidFile"));
            } else if (selected && selected.size > maximumBackupBytes) {
              setFile(null);
              setError(t("digitalCenters.backup.tooLarge"));
            } else {
              setFile(selected);
            }
          }} />
      </label>
      <button type="button" className="digital-center-button" disabled={!file || busy || disabled}
        onClick={() => void inspect()}>
        {t(busy ? "digitalCenters.backup.inspecting" : "digitalCenters.backup.inspect")}
      </button>
    </div>
    <p id={hintId} className="ecos-backup-import-hint">{t("digitalCenters.backup.fileHint")}</p>
    <AppCheckbox disabled={busy || disabled} checked={confirmStation}
      label={t("digitalCenters.backup.confirmStation", { host: stationHost })}
      onChange={(event) => setConfirmStation(event.target.checked)} />
    <p className="ecos-backup-import-hint">{t("digitalCenters.backup.confirmHint")}</p>
    <div role="status" aria-live="polite">
      {busy && <p>{t("digitalCenters.backup.inspectingHint")}</p>}
      {summary && <p>{t("digitalCenters.backup.summary", {
        profiles: summary.profiles, cvs: summary.cvValues, linked: summary.linkedVehicles
      })}</p>}
      {summary && <p className="ecos-backup-import-hint">{t("digitalCenters.backup.metadata", {
        serial: `0x${summary.stationSerial.toString(16).padStart(8, "0")}`, time: summary.importedAt
      })}</p>}
    </div>
    {error && <p className="digital-workspace-operation-error" role="alert">{error}</p>}
  </section>;
}

function backupErrorKey(cause: unknown): string {
  if (cause instanceof ApiError) {
    switch (cause.code) {
      case "ecos_backup_unavailable": return "digitalCenters.backup.unavailable";
      case "ecos_backup_invalid": return "digitalCenters.backup.invalidFile";
      case "ecos_backup_too_large": return "digitalCenters.backup.tooLarge";
      case "unsupported_media_type": return "digitalCenters.backup.invalidFile";
      case "digital_center_backup_station_mismatch": return "digitalCenters.backup.stationMismatch";
    }
  }
  return "digitalCenters.backup.importError";
}
