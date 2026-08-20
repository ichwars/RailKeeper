import { CheckCircle2, Download, FileJson, FileSpreadsheet, Play, RotateCcw, X } from "lucide-react";
import { useRef, useState } from "react";
import { createPortal } from "react-dom";

import { ApiError } from "../../shared/api";
import type { Language } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type {
  DataTransferExportResult,
  DataTransferJob,
  DataTransferJobDetails,
  DataTransferProfile
} from "./dataTransferModel";

type TransferExportDialogProps = {
  downloadUrl: (artifactId: string) => string;
  initialJob?: DataTransferJob;
  initialProfileId?: string;
  language: Language;
  canRetry: boolean;
  onClose: () => void;
  onCreateJob: (profileId: string) => Promise<DataTransferJob>;
  onExecute: (jobId: string) => Promise<DataTransferExportResult>;
  onRefreshJob: (jobId: string) => Promise<DataTransferJobDetails>;
  onRetry: (jobId: string) => Promise<DataTransferJob>;
  profiles: DataTransferProfile[];
};

export function TransferExportDialog({
  canRetry,
  downloadUrl,
  initialJob,
  initialProfileId,
  language,
  onClose,
  onCreateJob,
  onExecute,
  onRefreshJob,
  onRetry,
  profiles
}: TransferExportDialogProps) {
  const copy = exportCopy(language);
  const exportProfiles = profiles.filter((profile) => profile.enabled && profile.direction === "export");
  const [profileId, setProfileId] = useState(initialJob?.profileId || initialProfileId || "");
  const [job, setJob] = useState<DataTransferJob | null>(initialJob || null);
  const [result, setResult] = useState<DataTransferExportResult | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const closeRef = useRef<HTMLButtonElement | null>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(() => {
    if (!busy) onClose();
  }, closeRef);
  const profile = exportProfiles.find((item) => item.id === profileId);
  const snapshot = job || profile;
  const FormatIcon = snapshot?.format === "csv" ? FileSpreadsheet : FileJson;
  const terminalWithoutArtifact = Boolean(job && isTerminal(job) && !result);
  const canExecute = !result && (!job || job.state === "draft");

  async function execute() {
    if (!job && !profileId) return;
    setBusy(true);
    setError("");
    let activeJobId = job?.id || initialJob?.id || "";
    try {
      const activeJob = job || await onCreateJob(profileId);
      activeJobId = activeJob.id;
      setJob(activeJob);
      const nextResult = await onExecute(activeJob.id);
      setResult({
        ...nextResult,
        job: { ...nextResult.job, areas: [...nextResult.job.areas], options: { ...nextResult.job.options } },
        artifact: { ...nextResult.artifact }
      });
    } catch (reason) {
      if (reason instanceof ApiError && reason.status === 409) {
        try {
          const details = await onRefreshJob(activeJobId);
          setJob(details.job);
          setResult(null);
          const artifact = details.artifacts.find((item) => !item.deletedAt);
          if (artifact && ["completed", "completed_with_warnings"].includes(details.job.state)) {
            setResult({ job: details.job, artifact, openFolderAvailable: false });
          }
          setError(details.job.state === "draft" ? copy.conflictRecovery :
            details.job.state === "running" ? copy.runningRecovery : copy.terminalRecovery);
        } catch (refreshError) {
          setError(errorMessage(refreshError, copy.conflictRecovery));
        }
      } else {
        setError(errorMessage(reason, copy.exportError));
      }
    } finally {
      setBusy(false);
    }
  }

  async function retry() {
    if (!job || !terminalWithoutArtifact || !canRetry) return;
    setBusy(true);
    setError("");
    try {
      const draft = await onRetry(job.id);
      setJob(draft);
      setResult(null);
      setError(copy.retryReady);
    } catch (reason) {
      setError(errorMessage(reason, copy.retryError));
    } finally {
      setBusy(false);
    }
  }

  const dialog = (
    <div ref={layerRef} className="confirm-layer data-transfer-dialog-layer" role="dialog" aria-modal="true"
      aria-label={copy.title} onKeyDown={onKeyDown}>
      <section className="panel data-transfer-dialog transfer-export-dialog">
        <header className="data-transfer-dialog-head">
          <div><p className="eyebrow">{copy.eyebrow}</p><h2>{copy.title}</h2></div>
          <button ref={closeRef} type="button" className="icon-button" aria-label={copy.close} disabled={busy}
            onClick={onClose}><X size={19} aria-hidden="true" /></button>
        </header>

        <div className="data-transfer-dialog-body">
          {!initialJob ? (
            <label className="data-transfer-field">
              <span>{copy.profile}</span>
              <select value={profileId} disabled={busy || Boolean(result)} onChange={(event) => {
                setProfileId(event.target.value);
                setJob(null);
                setResult(null);
                setError("");
              }}>
                <option value="">{copy.chooseProfile}</option>
                {exportProfiles.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
          ) : null}

          {snapshot ? (
            <section className="transfer-export-summary" aria-label={copy.configuration}>
              <span className="transfer-export-format-icon"><FormatIcon size={25} aria-hidden="true" /></span>
              <div>
                <strong>{snapshotName(snapshot)}</strong>
                <dl>
                  <div><dt>{copy.areas}</dt><dd>{areaLabels(snapshot.areas, language)}</dd></div>
                  <div><dt>{copy.format}</dt><dd>{snapshot.format === "csv" ? "CSV" : "JSON"}</dd></div>
                </dl>
              </div>
            </section>
          ) : null}

          {snapshot ? (
            <section className="transfer-options-panel">
              <h3>{copy.options}</h3>
              {Object.keys(snapshot.options).length === 0 ? <p>{copy.noOptions}</p> : (
                <ul>{Object.entries(snapshot.options).map(([name, value]) => (
                  <li key={name}>{`${name}: ${optionValue(value, language)}`}</li>
                ))}</ul>
              )}
            </section>
          ) : null}

          {result ? (
            <section className="transfer-export-result" aria-live="polite">
              <CheckCircle2 size={32} aria-hidden="true" />
              <div>
                <h3>{copy.completed}</h3>
                <strong>{exportedCount(result.job.totalRecords, language)}</strong>
                <span>{result.artifact.displayName}</span>
              </div>
              <a className="secondary-button" href={downloadUrl(result.artifact.id)}>
                <Download size={16} aria-hidden="true" />{copy.download}
              </a>
            </section>
          ) : null}

          {job && job.state !== "draft" && !result ? (
            <section className="transfer-export-result" aria-live="polite">
              <div>
                <h3>{copy.currentStatus}</h3>
                <strong>{exportStateLabel(job.state, copy)}</strong>
                <span>{job.state === "running" ? copy.runningHelp : copy.terminalHelp}</span>
              </div>
            </section>
          ) : null}

          {error ? <p className="form-message error" role="alert">{error}</p> : null}
        </div>

        <footer className="data-transfer-dialog-actions">
          <span />
          <span>
            <button type="button" className="secondary-button" disabled={busy} onClick={onClose}>{copy.close}</button>
            {terminalWithoutArtifact && canRetry ? (
              <button type="button" className="primary-button" disabled={busy} onClick={() => void retry()}>
                <RotateCcw size={16} aria-hidden="true" />{copy.retry}
              </button>
            ) : null}
            {canExecute ? (
              <button type="button" className="primary-button" disabled={busy || (!job && !profileId)}
                onClick={() => void execute()}>
                <Play size={16} fill="currentColor" aria-hidden="true" />{copy.execute}
              </button>
            ) : null}
          </span>
        </footer>
      </section>
    </div>
  );

  return <><span ref={anchorRef} hidden aria-hidden="true" />{createPortal(dialog, document.body)}</>;
}

function isTerminal(job: DataTransferJob) {
  return ["failed", "cancelled", "completed", "completed_with_warnings"].includes(job.state);
}

function exportStateLabel(state: DataTransferJob["state"], copy: ReturnType<typeof exportCopy>) {
  const labels = {
    draft: copy.stateDraft,
    reading: copy.stateRunning,
    review_required: copy.stateRunning,
    ready: copy.stateRunning,
    running: copy.stateRunning,
    completed: copy.stateCompleted,
    completed_with_warnings: copy.stateCompletedWarnings,
    failed: copy.stateFailed,
    cancelled: copy.stateCancelled
  };
  return labels[state];
}

function snapshotName(snapshot: DataTransferJob | DataTransferProfile) {
  return "profileName" in snapshot ? snapshot.profileName : snapshot.name;
}

function optionValue(value: unknown, language: Language) {
  if (typeof value === "boolean") return language === "de" ? (value ? "ja" : "nein") : (value ? "yes" : "no");
  if (value === null) return "–";
  if (typeof value === "string" || typeof value === "number") return String(value);
  return JSON.stringify(value);
}

function exportedCount(count: number, language: Language) {
  if (language === "de") return `${count} ${count === 1 ? "Datensatz" : "Datensätze"} exportiert`;
  return `${count} ${count === 1 ? "record" : "records"} exported`;
}

function areaLabels(areas: DataTransferJob["areas"], language: Language) {
  const labels = language === "de"
    ? { vehicles: "Fahrzeuge", accessories: "Zubehör", exhibitionLists: "Ausstellungslisten" }
    : { vehicles: "Vehicles", accessories: "Accessories", exhibitionLists: "Exhibition lists" };
  return areas.map((area) => labels[area]).join(", ");
}

function exportCopy(language: Language) {
  return language === "de" ? {
    title: "Export erstellen", eyebrow: "LOKALER EXPORT", close: "Schließen", profile: "Exportprofil",
    chooseProfile: "Profil wählen", configuration: "Exportkonfiguration", areas: "Bereiche", format: "Format",
    options: "Optionen", noOptions: "Keine zusätzlichen Optionen.", execute: "Export ausführen",
    completed: "Export abgeschlossen", download: "Datei herunterladen", exportError: "Der Export konnte nicht erstellt werden.",
    conflictRecovery: "Der Auftrag wurde zwischenzeitlich geändert und neu gelesen. Bitte Status prüfen und den Export erneut ausführen.",
    runningRecovery: "Der Auftrag wurde zwischenzeitlich geändert und neu gelesen. Der Export läuft bereits.",
    terminalRecovery: "Der Auftrag wurde zwischenzeitlich geändert und neu gelesen. Für diesen Endstatus ist vor einem neuen Export ein Wiederholungsauftrag erforderlich.",
    currentStatus: "Aktueller Status", runningHelp: "Der Server verarbeitet diesen Export. Eine erneute Ausführung ist gesperrt.",
    terminalHelp: "Dieser Auftrag kann nicht direkt erneut ausgeführt werden.", retry: "Erneut versuchen",
    retryReady: "Ein neuer Entwurf wurde angelegt und kann ausgeführt werden.", retryError: "Der Export konnte nicht erneut angelegt werden.",
    stateDraft: "Entwurf", stateRunning: "Export läuft", stateCompleted: "Abgeschlossen",
    stateCompletedWarnings: "Abgeschlossen mit Hinweisen", stateFailed: "Fehlgeschlagen", stateCancelled: "Abgebrochen"
  } : {
    title: "Create export", eyebrow: "LOCAL EXPORT", close: "Close", profile: "Export profile",
    chooseProfile: "Choose profile", configuration: "Export configuration", areas: "Areas", format: "Format",
    options: "Options", noOptions: "No additional options.", execute: "Run export", completed: "Export completed",
    download: "Download file", exportError: "The export could not be created.",
    conflictRecovery: "The job changed in the meantime and was re-read. Review its status and run the export again.",
    runningRecovery: "The job changed in the meantime and was re-read. The export is already running.",
    terminalRecovery: "The job changed in the meantime and was re-read. This terminal state requires a retry job before exporting again.",
    currentStatus: "Current status", runningHelp: "The server is processing this export. Running it again is disabled.",
    terminalHelp: "This job cannot be run again directly.", retry: "Retry",
    retryReady: "A new draft was created and can be run.", retryError: "The export retry could not be created.",
    stateDraft: "Draft", stateRunning: "Export running", stateCompleted: "Completed",
    stateCompletedWarnings: "Completed with warnings", stateFailed: "Failed", stateCancelled: "Cancelled"
  };
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}
