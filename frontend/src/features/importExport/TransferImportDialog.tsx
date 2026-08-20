import { CheckCircle2, FileUp, ShieldCheck, Upload, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState, type ChangeEvent } from "react";
import { createPortal } from "react-dom";

import { ApiError } from "../../shared/api";
import type { Language } from "../../shared/i18n";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type {
  DataTransferIssue,
  DataTransferIssueResolution,
  DataTransferJob,
  DataTransferJobDetails,
  DataTransferPreview,
  DataTransferPreviewRecord,
  DataTransferProfile
} from "./dataTransferModel";
import { TransferReviewTable } from "./TransferReviewTable";

export type ImportDialogStep = "profile" | "file" | "mapping" | "review" | "confirm";

type TransferImportDialogProps = {
  initialDetails?: DataTransferJobDetails | null;
  initialJob?: DataTransferJob;
  initialProfileId?: string;
  initialRequiresReupload: boolean;
  language: Language;
  onCancelJob: (jobId: string) => Promise<unknown>;
  onClose: () => void;
  onConfirm: (jobId: string) => Promise<unknown>;
  onCreateJob: (profileId: string) => Promise<DataTransferJob>;
  onResolve: (jobId: string, issueId: string, resolution: DataTransferIssueResolution) => Promise<DataTransferJob>;
  onRefreshJob: (jobId: string) => Promise<DataTransferJobDetails>;
  onUpload: (jobId: string, file: File) => Promise<DataTransferPreview>;
  profiles: DataTransferProfile[];
};

export function TransferImportDialog({
  initialDetails,
  initialJob,
  initialProfileId,
  initialRequiresReupload,
  language,
  onCancelJob,
  onClose,
  onConfirm,
  onCreateJob,
  onResolve,
  onRefreshJob,
  onUpload,
  profiles
}: TransferImportDialogProps) {
  const copy = importCopy(language);
  const importProfiles = profiles.filter((profile) => profile.enabled && profile.direction === "import");
  const [profileId, setProfileId] = useState(initialJob?.profileId || initialProfileId || "");
  const [job, setJob] = useState<DataTransferJob | null>(initialJob || null);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<DataTransferPreview | null>(() => previewFromDetails(initialJob, initialDetails));
  const [mappingAccepted, setMappingAccepted] = useState(
    Boolean(previewFromDetails(initialJob, initialDetails) && initialJob?.format !== "csv")
  );
  const [requiresReupload, setRequiresReupload] = useState(initialRequiresReupload);
  const requiresReuploadRef = useRef(initialRequiresReupload);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState(initialRequiresReupload ? copy.confirmConflictRecovery : "");
  const closeRef = useRef<HTMLButtonElement | null>(null);
  const fileRef = useRef<HTMLInputElement | null>(null);
  const { anchorRef, layerRef, onKeyDown } = useModalDialogLayer(() => {
    if (!busy) onClose();
  }, closeRef);
  const selectedProfile = importProfiles.find((profile) => profile.id === profileId);
  const snapshot = job || selectedProfile;
  const unresolvedIssues = preview?.issues.filter((issue) => !issue.selectedResolution) || [];
  const approvedRecords = useMemo(() => {
    if (!preview || unresolvedIssues.length > 0) return 0;
    const skipped = new Set(preview.issues
      .filter((issue) => issue.selectedResolution === "skip")
      .map((issue) => `${issue.area}:${issue.recordKey}`));
    return preview.records.filter((record) => !skipped.has(`${record.area}:${record.recordKey}`)).length;
  }, [preview, unresolvedIssues.length]);
  const step = currentStep(profileId, file, preview, job, mappingAccepted, requiresReupload);

  useEffect(() => {
    if (!initialJob) return;
    setProfileId((current) => current || initialJob.profileId);
    setJob((current) => current?.id === initialJob.id && current.revision > initialJob.revision
      ? current
      : initialJob);
    const persistedPreview = previewFromDetails(initialJob, initialDetails);
    if (persistedPreview) {
      setPreview((current) => current?.job.id === persistedPreview.job.id &&
        current.job.revision > persistedPreview.job.revision ? current : persistedPreview);
      if (!requiresReuploadRef.current) setMappingAccepted(initialJob.format !== "csv");
    }
  }, [initialDetails, initialJob]);

  useEffect(() => {
    requiresReuploadRef.current = initialRequiresReupload;
    setRequiresReupload(initialRequiresReupload);
    if (initialRequiresReupload) {
      setMappingAccepted(false);
      setError(copy.confirmConflictRecovery);
    }
  }, [initialJob?.id, initialRequiresReupload, language]);

  async function changeProfile(nextProfileId: string) {
    if (nextProfileId === profileId) return;
    if (preview && !window.confirm(copy.changeWarning)) return;
    setBusy(true);
    setError("");
    try {
      if (job && ["draft", "review_required", "ready"].includes(job.state)) await onCancelJob(job.id);
      setProfileId(nextProfileId);
      setJob(null);
      setFile(null);
      setPreview(null);
      setMappingAccepted(false);
      setRequiresReupload(false);
      requiresReuploadRef.current = false;
    } catch (reason) {
      setError(errorMessage(reason, copy.changeError));
    } finally {
      setBusy(false);
    }
  }

  async function chooseFile(event: ChangeEvent<HTMLInputElement>) {
    const nextFile = event.target.files?.[0];
    event.target.value = "";
    if (!nextFile || (!profileId && !job)) return;
    if (preview && !window.confirm(copy.changeWarning)) return;
    setFile(nextFile);
    setBusy(true);
    setError("");
    let activeJobId = job?.id;
    try {
      const activeJob = job || await onCreateJob(profileId);
      activeJobId = activeJob.id;
      setJob(activeJob);
      const nextPreview = await onUpload(activeJob.id, nextFile);
      setJob(nextPreview.job);
      setPreview(clonePreview(nextPreview));
      setMappingAccepted(nextPreview.job.format !== "csv");
      setRequiresReupload(false);
      requiresReuploadRef.current = false;
    } catch (reason) {
      await recoverConflict(reason, activeJobId, copy.uploadError);
    } finally {
      setBusy(false);
    }
  }

  async function resolve(issue: DataTransferIssue, resolution: DataTransferIssueResolution) {
    if (!preview || !job) return;
    setBusy(true);
    setError("");
    try {
      const nextJob = await onResolve(job.id, issue.id, resolution);
      setJob(nextJob);
      setPreview((current) => current ? {
        ...current,
        job: { ...nextJob },
        issues: current.issues.map((item) => item.id === issue.id
          ? { ...item, selectedResolution: resolution }
          : item)
      } : current);
    } catch (reason) {
      await recoverConflict(reason, job.id, copy.resolveError);
    } finally {
      setBusy(false);
    }
  }

  async function confirmImport() {
    if (!job || unresolvedIssues.length > 0 || approvedRecords === 0 || job.state !== "ready") return;
    setBusy(true);
    setError("");
    try {
      await onConfirm(job.id);
      onClose();
    } catch (reason) {
      await recoverConflict(reason, job.id, copy.confirmError, true);
    } finally {
      setBusy(false);
    }
  }

  async function recoverConflict(reason: unknown, jobId: string | undefined, fallback: string,
    reuploadRequired = false) {
    if (!(reason instanceof ApiError) || reason.status !== 409 || !jobId) {
      setError(errorMessage(reason, fallback));
      return;
    }
    if (reuploadRequired) {
      requiresReuploadRef.current = true;
      setRequiresReupload(true);
      setMappingAccepted(false);
    }
    try {
      const details = await onRefreshJob(jobId);
      setJob(details.job);
      setPreview(previewFromDetails(details.job, details));
      setMappingAccepted(details.job.format !== "csv" && !reuploadRequired);
      setRequiresReupload(reuploadRequired);
      setError(reuploadRequired ? copy.confirmConflictRecovery : copy.conflictRecovery);
    } catch (refreshError) {
      setError(errorMessage(refreshError, copy.conflictRecovery));
    }
  }

  async function cancelJob() {
    if (!job || !window.confirm(copy.cancelJobConfirm)) return;
    setBusy(true);
    setError("");
    try {
      await onCancelJob(job.id);
      onClose();
    } catch (reason) {
      setError(errorMessage(reason, copy.cancelJobError));
    } finally {
      setBusy(false);
    }
  }

  const dialog = (
    <div ref={layerRef} className="confirm-layer data-transfer-dialog-layer" role="dialog" aria-modal="true"
      aria-label={copy.title} onKeyDown={onKeyDown}>
      <section className="panel data-transfer-dialog transfer-import-dialog">
        <header className="data-transfer-dialog-head">
          <div><p className="eyebrow">{copy.eyebrow}</p><h2>{copy.title}</h2></div>
          <button ref={closeRef} type="button" className="icon-button" aria-label={copy.close} disabled={busy}
            onClick={onClose}><X size={19} aria-hidden="true" /></button>
        </header>

        <ol className="data-transfer-wizard-steps" aria-label={copy.progress}>
          {(["profile", "file", "mapping", "review", "confirm"] as ImportDialogStep[]).map((name) => (
            <li key={name} className={stepClass(name, step)}>{copy.steps[name]}</li>
          ))}
        </ol>

        <div className="data-transfer-dialog-body">
          {!initialJob && (
            <label className="data-transfer-field">
              <span>{copy.profile}</span>
              <select value={profileId} disabled={busy} onChange={(event) => void changeProfile(event.target.value)}>
                <option value="">{copy.chooseProfile}</option>
                {importProfiles.map((profile) => <option key={profile.id} value={profile.id}>{profile.name}</option>)}
              </select>
            </label>
          )}

          {snapshot ? (
            <section className="data-transfer-snapshot" aria-label={copy.snapshot}>
              <strong>{snapshotName(snapshot)}</strong>
              <span>{areaLabels(snapshot.areas, language)}</span>
              <span>{snapshot.format === "csv" ? "CSV" : "JSON"}</span>
            </section>
          ) : null}

          {(profileId || initialJob) ? (
            <label className={`data-transfer-file-drop${busy ? " busy" : ""}`}>
              <FileUp size={28} aria-hidden="true" />
              <span><strong>{file?.name || job?.sourceName || copy.chooseFile}</strong><small>{copy.fileHelp}</small></span>
              <input ref={fileRef} type="file" aria-label={copy.fileLabel} disabled={busy}
                accept={snapshot?.format === "csv" ? ".csv,text/csv" : ".json,application/json"}
                onChange={(event) => void chooseFile(event)} />
              <span className="secondary-button"><Upload size={15} aria-hidden="true" />{copy.browse}</span>
            </label>
          ) : null}

          {preview && preview.job.format === "csv" && !mappingAccepted ? (
            <section className="transfer-mapping-section">
              <header><div><p className="eyebrow">{copy.serverPreview}</p><h3>{copy.mappingTitle}</h3></div></header>
              <p>{copy.mappingExplanation}</p>
              <h4>{copy.normalizedFields}</h4>
              <ul>{previewFields(preview).map((field) => <li key={field}><code>{field}</code></li>)}</ul>
              <p>{mappingIssueSummary(preview.issues, language)}</p>
              <div className="transfer-mapping-actions">
                <button type="button" className="secondary-button" disabled={busy}
                  onClick={() => fileRef.current?.click()}>{copy.reupload}</button>
                <button type="button" className="primary-button" disabled={busy || requiresReupload}
                  onClick={() => setMappingAccepted(true)}>{copy.continueReview}</button>
              </div>
            </section>
          ) : null}

          {preview && mappingAccepted && !requiresReupload ? (
            <section className="transfer-preview-section">
              <header>
                <div><p className="eyebrow">{copy.serverPreview}</p><h3>{copy.preview}</h3></div>
                <span className="transfer-preview-revision">{copy.revision} {preview.job.revision}</span>
              </header>
              <div className="transfer-preview-summary">
                <span><strong>{preview.totalRecords}</strong>{copy.records}</span>
                <span className="success"><strong>{preview.readyRecords}</strong>{copy.ready}</span>
                <span className="warning"><strong>{preview.warningRecords}</strong>{copy.warnings}</span>
                <span className="danger"><strong>{preview.errorRecords}</strong>{copy.errors}</span>
              </div>
              <TransferReviewTable busy={busy} issues={preview.issues} language={language} onResolve={resolve}
                records={preview.records} />
              {unresolvedIssues.length > 0 ? (
                <p className="form-message error" role="status">{copy.unresolved.replace("{count}", String(unresolvedIssues.length))}</p>
              ) : (
                <p className="transfer-ready-message"><ShieldCheck size={17} aria-hidden="true" />{copy.readyToConfirm}</p>
              )}
            </section>
          ) : null}

          {error ? <p className="form-message error" role="alert">{error}</p> : null}
        </div>

        <footer className="data-transfer-dialog-actions">
          {job ? <button type="button" className="secondary-button" disabled={busy}
            onClick={() => void cancelJob()}>{copy.cancelJob}</button> : <span />}
          <span>
            <button type="button" className="secondary-button" disabled={busy} onClick={onClose}>{copy.close}</button>
            {preview && mappingAccepted && !requiresReupload ? (
              <button type="button" className="primary-button" disabled={busy || unresolvedIssues.length > 0 ||
                approvedRecords === 0 || job?.state !== "ready"} onClick={() => void confirmImport()}>
                <CheckCircle2 size={16} aria-hidden="true" />{importButtonLabel(approvedRecords, language)}
              </button>
            ) : null}
          </span>
        </footer>
      </section>
    </div>
  );

  return <><span ref={anchorRef} hidden aria-hidden="true" />{createPortal(dialog, document.body)}</>;
}

function currentStep(profileId: string, file: File | null, preview: DataTransferPreview | null, job: DataTransferJob | null,
  mappingAccepted: boolean, requiresReupload: boolean) {
  if (requiresReupload) return "file";
  if (preview) return mappingAccepted
    ? (preview.issues.some((issue) => !issue.selectedResolution) ? "review" : "confirm")
    : "mapping";
  if (file || job?.sourceName) return "mapping";
  return profileId || job ? "file" : "profile";
}

function previewFields(preview: DataTransferPreview) {
  return [...new Set(preview.records.flatMap((record) => Object.keys(record.data)))].sort();
}

function mappingIssueSummary(issues: DataTransferIssue[], language: Language) {
  const count = issues.filter((issue) => issue.code.includes("missing") || issue.code.includes("unmapped")).length;
  return language === "de"
    ? `${count} nicht zugeordnete oder fehlende Felder in der Servervorschau.`
    : `${count} unmapped or missing fields in the server preview.`;
}

function snapshotName(snapshot: DataTransferJob | DataTransferProfile) {
  return "profileName" in snapshot ? snapshot.profileName : snapshot.name;
}

function stepClass(name: ImportDialogStep, current: ImportDialogStep) {
  const order: ImportDialogStep[] = ["profile", "file", "mapping", "review", "confirm"];
  const index = order.indexOf(name);
  const currentIndex = order.indexOf(current);
  return index < currentIndex ? "completed" : index === currentIndex ? "current" : "";
}

function previewFromDetails(job?: DataTransferJob, details?: DataTransferJobDetails | null): DataTransferPreview | null {
  if (!job || !job.sourceName || !details || details.job.id !== job.id) return null;
  const records = previewRecords(job.preview.records);
  return {
    job: { ...job },
    records,
    issues: details.issues.map((issue) => ({ ...issue })),
    totalRecords: job.totalRecords,
    readyRecords: job.readyRecords,
    warningRecords: job.warningRecords,
    errorRecords: job.errorRecords
  };
}

function previewRecords(value: unknown): DataTransferPreviewRecord[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isPreviewRecord).map((record) => ({ ...record, data: { ...record.data } }));
}

function isPreviewRecord(value: unknown): value is DataTransferPreviewRecord {
  if (!value || typeof value !== "object") return false;
  const record = value as Record<string, unknown>;
  return typeof record.area === "string" && typeof record.recordKey === "string" &&
    typeof record.classification === "string" && typeof record.proposedAction === "string" &&
    Boolean(record.data) && typeof record.data === "object";
}

function clonePreview(preview: DataTransferPreview): DataTransferPreview {
  return {
    ...preview,
    job: { ...preview.job, areas: [...preview.job.areas], options: { ...preview.job.options } },
    records: preview.records.map((record) => ({ ...record, data: { ...record.data } })),
    issues: preview.issues.map((issue) => ({ ...issue }))
  };
}

function importButtonLabel(count: number, language: Language) {
  if (language === "de") return `${count} ${count === 1 ? "Datensatz" : "Datensätze"} importieren`;
  return `Import ${count} ${count === 1 ? "record" : "records"}`;
}

function areaLabels(areas: DataTransferJob["areas"], language: Language) {
  const labels = language === "de"
    ? { vehicles: "Fahrzeuge", accessories: "Zubehör", exhibitionLists: "Ausstellungslisten" }
    : { vehicles: "Vehicles", accessories: "Accessories", exhibitionLists: "Exhibition lists" };
  return areas.map((area) => labels[area]).join(", ");
}

function importCopy(language: Language) {
  return language === "de" ? {
    title: "Import prüfen", eyebrow: "SICHERER IMPORT", close: "Schließen", progress: "Importschritte",
    steps: { profile: "Profil", file: "Datei", mapping: "Zuordnung", review: "Prüfung", confirm: "Bestätigung" },
    profile: "Importprofil", chooseProfile: "Profil wählen", snapshot: "Auftragssnapshot", chooseFile: "Datei auswählen",
    fileLabel: "Importdatei", fileHelp: "Die Datei wird erst geprüft. Es werden noch keine Daten geschrieben.",
    browse: "Durchsuchen", serverPreview: "PERSISTENTE SERVERVORSCHAU", preview: "Vorschau", revision: "Revision",
    records: "Datensätze", ready: "bereit", warnings: "Hinweise", errors: "Fehler",
    unresolved: "{count} Konflikt(e) müssen vor dem Import aufgelöst werden.",
    readyToConfirm: "Alle Konflikte sind aufgelöst. Der geprüfte Stand kann importiert werden.",
    cancelJob: "Auftrag abbrechen", cancelJobConfirm: "Diesen Importauftrag abbrechen? Die Vorschau bleibt im Verlauf erhalten.",
    changeWarning: "Die aktuelle Vorschau wird verworfen und als neue Revision erneut geprüft. Fortfahren?",
    uploadError: "Die Datei konnte nicht geprüft werden.", resolveError: "Der Konflikt konnte nicht aufgelöst werden.",
    confirmError: "Der Import konnte nicht bestätigt werden.", cancelJobError: "Der Auftrag konnte nicht abgebrochen werden.",
    changeError: "Der Import konnte nicht auf ein anderes Profil umgestellt werden."
    ,mappingTitle: "Erkannte CSV-Zuordnung", mappingExplanation: "Die Zuordnung wurde serverseitig aus Spalten und Aliasen erkannt. Änderungen erfordern: Quelldatei bearbeiten und erneut hochladen.",
    normalizedFields: "Serverseitig normalisierte Zielfelder (automatische Alias-Erkennung)", reupload: "Datei erneut auswählen", continueReview: "Weiter zur Prüfung",
    conflictRecovery: "Der Auftrag wurde zwischenzeitlich geändert. Die persistente Vorschau wurde neu gelesen. Bitte Zuordnung und Konflikte erneut prüfen.",
    confirmConflictRecovery: "Der Auftrag wurde beim Bestätigen geändert. Bitte die Quelldatei erneut hochladen und die neue Vorschau vollständig prüfen."
  } : {
    title: "Review import", eyebrow: "SAFE IMPORT", close: "Close", progress: "Import steps",
    steps: { profile: "Profile", file: "File", mapping: "Mapping", review: "Review", confirm: "Confirmation" },
    profile: "Import profile", chooseProfile: "Choose profile", snapshot: "Job snapshot", chooseFile: "Choose file",
    fileLabel: "Import file", fileHelp: "The file is reviewed first. No data is written yet.", browse: "Browse",
    serverPreview: "PERSISTENT SERVER PREVIEW", preview: "Preview", revision: "Revision", records: "records",
    ready: "ready", warnings: "warnings", errors: "errors",
    unresolved: "{count} conflict(s) must be resolved before import.",
    readyToConfirm: "All conflicts are resolved. The reviewed revision can be imported.", cancelJob: "Cancel job",
    cancelJobConfirm: "Cancel this import job? The preview remains in history.",
    changeWarning: "The current preview will be discarded and reviewed as a new revision. Continue?",
    uploadError: "The file could not be reviewed.", resolveError: "The conflict could not be resolved.",
    confirmError: "The import could not be confirmed.", cancelJobError: "The job could not be cancelled.",
    changeError: "The import could not be changed to another profile."
    ,mappingTitle: "Detected CSV mapping", mappingExplanation: "The mapping was detected by the server from columns and aliases. To change it, edit the source file and upload it again.",
    normalizedFields: "Server-normalized target fields (automatic alias detection)", reupload: "Choose file again", continueReview: "Continue to review",
    conflictRecovery: "The job changed in the meantime. The persistent preview was re-read. Review the mapping and conflicts again.",
    confirmConflictRecovery: "The job changed during confirmation. Upload the source file again and fully review the new preview."
  };
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}
