import { CheckCircle2, FileUp, ShieldCheck, Upload, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState, type ChangeEvent } from "react";
import { createPortal } from "react-dom";

import { ApiError } from "../../shared/api";
import type { Language } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppSelect } from "../../shared/ui/AppSelect";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import type {
  DataTransferCSVColumnMapping,
  DataTransferCSVMappingInput,
  DataTransferIssue,
  DataTransferIssueResolution,
  DataTransferJob,
  DataTransferJobDetails,
  DataTransferPreview,
  DataTransferPreviewRecord,
  DataTransferProfile,
  DataTransferVehicleSet,
  DataTransferVehicleSetPreview
} from "./dataTransferModel";
import { TransferConfirmDialog, type TransferPendingAction } from "./TransferConfirmDialog";
import { TransferReviewTable } from "./TransferReviewTable";
import { TransferSetReview } from "./TransferSetReview";

export type ImportDialogStep = "profile" | "file" | "mapping" | "review" | "confirm";

type PendingImportChange =
  | { kind: "profile"; profileId: string }
  | { file: File; kind: "file" }
  | { kind: "cancel" };

type TransferImportDialogProps = {
  initialDetails?: DataTransferJobDetails | null;
  initialJob?: DataTransferJob;
  initialProfileId?: string;
  initialRequiresReupload: boolean;
  language: Language;
  onCancelJob: (jobId: string) => Promise<unknown>;
  onClose: () => void;
  onConfirm: (jobId: string, expectedRevision: number) => Promise<unknown>;
  onCreateJob: (profileId: string) => Promise<DataTransferJob>;
  onResolve: (jobId: string, issueId: string, resolution: DataTransferIssueResolution) => Promise<DataTransferJob>;
  onRefreshJob: (jobId: string) => Promise<DataTransferJobDetails>;
  onUpload: (jobId: string, file: File, mapping?: DataTransferCSVMappingInput) => Promise<DataTransferPreview>;
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
  const [csvMapping, setCSVMapping] = useState<DataTransferCSVColumnMapping[]>(() =>
    previewFromDetails(initialJob, initialDetails)?.csvMapping ?? []
  );
  const [saveMappingToProfile, setSaveMappingToProfile] = useState(false);
  const [mappingDirty, setMappingDirty] = useState(false);
  const [mappingAccepted, setMappingAccepted] = useState(
    Boolean(previewFromDetails(initialJob, initialDetails) && !vehicleCSVMappingRequired(initialJob))
  );
  const acceptedMappingIdentityRef = useRef<string | null>(null);
  const [requiresReupload, setRequiresReupload] = useState(initialRequiresReupload);
  const requiresReuploadRef = useRef(initialRequiresReupload);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState(initialRequiresReupload ? copy.confirmConflictRecovery : "");
  const [pendingChange, setPendingChange] = useState<PendingImportChange | null>(null);
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
    const skippedSetMembers = new Set<string>();
    for (const set of preview.vehicleSets ?? []) {
      if (skipped.has(`vehicles:${set.recordKey}`)) {
        for (const memberKey of set.memberRecordKeys) skippedSetMembers.add(memberKey);
      }
    }
    return preview.records.filter((record) =>
      !skipped.has(`${record.area}:${record.recordKey}`) &&
      !(record.area === "vehicles" && skippedSetMembers.has(record.recordKey))
    ).length;
  }, [preview, unresolvedIssues.length]);
  const step = currentStep(profileId, file, preview, job, mappingAccepted, requiresReupload);
  const mappingReady = csvMapping.length === 0 || csvMapping.every((column) => column.origin !== "unmapped");

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
      setCSVMapping(persistedPreview.csvMapping ?? []);
      if (!requiresReuploadRef.current) {
        const persistedIdentity = csvMappingIdentity(initialJob, persistedPreview.csvMapping ?? []);
        const keepAccepted = acceptedMappingIdentityRef.current === persistedIdentity;
        if (!keepAccepted) acceptedMappingIdentityRef.current = null;
        setMappingAccepted(!vehicleCSVMappingRequired(initialJob) || keepAccepted);
      }
    }
  }, [initialDetails, initialJob]);

  useEffect(() => {
    requiresReuploadRef.current = initialRequiresReupload;
    setRequiresReupload(initialRequiresReupload);
    if (initialRequiresReupload) {
      acceptedMappingIdentityRef.current = null;
      setMappingAccepted(false);
      setError(copy.confirmConflictRecovery);
    }
  }, [initialJob?.id, initialRequiresReupload, language]);

  function requestProfileChange(nextProfileId: string) {
    if (nextProfileId === profileId) return;
    if (preview) {
      setPendingChange({ kind: "profile", profileId: nextProfileId });
      return;
    }
    void changeProfile(nextProfileId);
  }

  async function changeProfile(nextProfileId: string) {
    setBusy(true);
    setError("");
    try {
      if (job && ["draft", "review_required", "ready"].includes(job.state)) await onCancelJob(job.id);
      setProfileId(nextProfileId);
      setJob(null);
      setFile(null);
      setPreview(null);
      setCSVMapping([]);
      setSaveMappingToProfile(false);
      setMappingDirty(false);
      acceptedMappingIdentityRef.current = null;
      setMappingAccepted(false);
      setRequiresReupload(false);
      requiresReuploadRef.current = false;
    } catch (reason) {
      setError(errorMessage(reason, copy.changeError));
    } finally {
      setBusy(false);
    }
  }

  function chooseFile(event: ChangeEvent<HTMLInputElement>) {
    const nextFile = event.target.files?.[0];
    event.target.value = "";
    if (!nextFile || (!profileId && !job)) return;
    if (preview) {
      setPendingChange({ kind: "file", file: nextFile });
      return;
    }
    void uploadFile(nextFile);
  }

  async function uploadFile(nextFile: File) {
    acceptedMappingIdentityRef.current = null;
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
      setCSVMapping(cloneCSVMapping(nextPreview.csvMapping));
      setSaveMappingToProfile(false);
      setMappingDirty(false);
      setMappingAccepted(!vehicleCSVMappingRequired(nextPreview.job));
      setRequiresReupload(false);
      requiresReuploadRef.current = false;
    } catch (reason) {
      await recoverConflict(reason, activeJobId, copy.uploadError);
    } finally {
      setBusy(false);
    }
  }

  function changeMapping(columnIndex: number, value: string) {
    setCSVMapping((current) => current.map((column) => {
      if (column.index !== columnIndex) return column;
      if (value === "__ignore__") return { ...column, targetField: "", origin: "ignored" };
      if (value === "") return { ...column, targetField: "", origin: "unmapped" };
      return { ...column, targetField: value, origin: "manual" };
    }));
    setMappingDirty(true);
  }

  async function applyMapping() {
    if (!preview || !job || !mappingReady) return;
    if (!mappingDirty && !saveMappingToProfile) {
      acceptedMappingIdentityRef.current = csvMappingIdentity(job, csvMapping);
      setMappingAccepted(true);
      return;
    }
    if (!file) {
      setError(copy.mappingFileRequired);
      return;
    }
    setBusy(true);
    setError("");
    try {
      const nextPreview = await onUpload(job.id, file, {
        columns: cloneCSVMapping(csvMapping),
        saveToProfile: saveMappingToProfile
      });
      setJob(nextPreview.job);
      setPreview(clonePreview(nextPreview));
      const nextMapping = cloneCSVMapping(nextPreview.csvMapping);
      setCSVMapping(nextMapping);
      setMappingDirty(false);
      acceptedMappingIdentityRef.current = csvMappingIdentity(nextPreview.job, nextMapping);
      setMappingAccepted(true);
    } catch (reason) {
      await recoverConflict(reason, job.id, copy.mappingError);
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
      await onConfirm(job.id, job.revision);
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
      acceptedMappingIdentityRef.current = null;
      setMappingAccepted(false);
    }
    try {
      const details = await onRefreshJob(jobId);
      const refreshedPreview = previewFromDetails(details.job, details);
      const refreshedMapping = refreshedPreview?.csvMapping ?? [];
      const refreshedIdentity = csvMappingIdentity(details.job, refreshedMapping);
      const keepAccepted = acceptedMappingIdentityRef.current === refreshedIdentity;
      if (!keepAccepted) acceptedMappingIdentityRef.current = null;
      setJob(details.job);
      setPreview(refreshedPreview);
      setCSVMapping(refreshedMapping);
      setMappingAccepted(!reuploadRequired && (!vehicleCSVMappingRequired(details.job) || keepAccepted));
      setRequiresReupload(reuploadRequired);
      setError(reuploadRequired ? copy.confirmConflictRecovery : copy.conflictRecovery);
    } catch (refreshError) {
      setError(errorMessage(refreshError, copy.conflictRecovery));
    }
  }

  async function cancelJob() {
    if (!job) return;
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

  const pendingAction: TransferPendingAction | null = pendingChange ? {
    title: pendingChange.kind === "cancel" ? copy.cancelJobTitle : copy.changeTitle,
    body: pendingChange.kind === "cancel" ? copy.cancelJobConfirm : copy.changeWarning,
    confirmLabel: pendingChange.kind === "cancel" ? copy.cancelJob : copy.applyChange,
    dangerous: pendingChange.kind === "cancel",
    errorMessage: pendingChange.kind === "cancel" ? copy.cancelJobError : copy.changeError,
    run: () => pendingChange.kind === "profile"
      ? changeProfile(pendingChange.profileId)
      : pendingChange.kind === "file" ? uploadFile(pendingChange.file) : cancelJob()
  } : null;

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
              <AppSelect aria-label={copy.profile} value={profileId} disabled={busy}
                onChange={(event) => requestProfileChange(event.target.value)}>
                <option value="">{copy.chooseProfile}</option>
                {importProfiles.map((profile) => <option key={profile.id} value={profile.id}>{profile.name}</option>)}
              </AppSelect>
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
                onChange={chooseFile} />
              <span className="secondary-button"><Upload size={15} aria-hidden="true" />{copy.browse}</span>
            </label>
          ) : null}

          {preview && preview.job.format === "csv" && !mappingAccepted ? (
            <section className="transfer-mapping-section">
              <header><div><p className="eyebrow">{copy.serverPreview}</p><h3>{copy.mappingTitle}</h3></div></header>
              <p>{copy.mappingExplanation}</p>
              <div className="transfer-mapping-table-wrap">
                <table className="data-transfer-table transfer-mapping-table">
                  <thead><tr><th>{copy.sourceColumn}</th><th>{copy.targetField}</th><th>{copy.mappingSource}</th></tr></thead>
                  <tbody>{csvMapping.map((column) => (
                    <tr key={`${column.index}:${column.sourceHeader}`}>
                      <td><strong>{column.sourceHeader}</strong></td>
                      <td>
                        <AppSelect
                          aria-label={copy.targetFor.replace("{source}", column.sourceHeader)}
                          value={column.origin === "ignored" ? "__ignore__" : column.targetField}
                          disabled={busy || !file}
                          onChange={(event) => changeMapping(column.index, event.target.value)}
                        >
                          <option value="">{copy.unmapped}</option>
                          <option value="__ignore__">{copy.ignore}</option>
                          {(preview.vehicleFields ?? []).map((field) => (
                            <option key={field.key} value={field.key}
                              disabled={csvMapping.some((item) => item.index !== column.index &&
                                item.targetField === field.key)}>
                              {language === "de" ? field.labelDE : field.labelEN}
                            </option>
                          ))}
                        </AppSelect>
                      </td>
                      <td>{mappingOriginLabel(column.origin, language)}</td>
                    </tr>
                  ))}</tbody>
                </table>
              </div>
              {csvMapping.length === 0 ? <p>{copy.legacyMapping}</p> : null}
              {!mappingReady ? <p className="form-message error" role="status">{copy.mappingIncomplete}</p> : null}
              {job?.profileId && csvMapping.length > 0 ? (
                <AppCheckbox label={copy.saveMapping} checked={saveMappingToProfile}
                  onChange={() => setSaveMappingToProfile((current) => !current)} />
              ) : null}
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
              <TransferSetReview busy={busy} issues={preview.issues} language={language} onResolve={resolve}
                sets={preview.vehicleSets ?? []} />
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
            onClick={() => setPendingChange({ kind: "cancel" })}>{copy.cancelJob}</button> : <span />}
          <span>
            <button type="button" className="secondary-button" disabled={busy} onClick={onClose}>{copy.close}</button>
            {preview && preview.job.format === "csv" && !mappingAccepted ? (
              <button type="button" className="primary-button"
                disabled={busy || requiresReupload || !mappingReady} onClick={() => void applyMapping()}>
                {mappingDirty || saveMappingToProfile || !mappingReady ? copy.validateMapping : copy.continueReview}
              </button>
            ) : null}
            {preview && mappingAccepted && !requiresReupload ? (
              <button type="button" className="primary-button" disabled={busy || unresolvedIssues.length > 0 ||
                approvedRecords === 0 || job?.state !== "ready"} onClick={() => void confirmImport()}>
                <CheckCircle2 size={16} aria-hidden="true" />{importButtonLabel(approvedRecords, language)}
              </button>
            ) : null}
          </span>
        </footer>
        <TransferConfirmDialog action={pendingAction} cancelLabel={copy.cancel}
          onClose={() => setPendingChange(null)} />
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

function vehicleCSVMappingRequired(job: Pick<DataTransferJob, "format" | "areas"> | undefined) {
  return job?.format === "csv" && job.areas.includes("vehicles");
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
    errorRecords: job.errorRecords,
    csvMapping: csvMappingFromPreview(job.preview.csvMapping),
    vehicleFields: vehicleFieldsFromPreview(job.preview.vehicleFields),
    vehicleSets: vehicleSetPreviewsFromPreview(job.preview.vehicleSets)
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
    issues: preview.issues.map((issue) => ({ ...issue })),
    csvMapping: cloneCSVMapping(preview.csvMapping),
    vehicleFields: preview.vehicleFields?.map((field) => ({
      ...field,
      aliases: field.aliases ? [...field.aliases] : undefined
    })),
    vehicleSets: preview.vehicleSets?.map(cloneVehicleSetPreview)
  };
}

function cloneVehicleSetPreview(set: DataTransferVehicleSetPreview): DataTransferVehicleSetPreview {
  return {
    ...set,
    memberRecordKeys: [...set.memberRecordKeys],
    rowNumbers: set.rowNumbers ? [...set.rowNumbers] : undefined,
    diagnostics: set.diagnostics?.map((diagnostic) => ({ ...diagnostic })),
    data: {
      ...set.data,
      members: set.data.members.map((member) => ({ ...member }))
    }
  };
}

function cloneCSVMapping(mapping: DataTransferCSVColumnMapping[] | undefined) {
  return mapping?.map((column) => ({ ...column })) ?? [];
}

function csvMappingIdentity(job: DataTransferJob, mapping: DataTransferCSVColumnMapping[]) {
  const columns = [...mapping]
    .sort((left, right) => left.index - right.index)
    .map(({ index, sourceHeader, normalizedHeader, targetField }) =>
      [index, sourceHeader, normalizedHeader, targetField]
    );
  return JSON.stringify([job.id, job.sourceSha256, columns]);
}

function csvMappingFromPreview(value: unknown): DataTransferCSVColumnMapping[] {
  if (!Array.isArray(value)) return [];
  return value.filter((item): item is DataTransferCSVColumnMapping => {
    if (!item || typeof item !== "object") return false;
    const column = item as Record<string, unknown>;
    return typeof column.index === "number" && typeof column.sourceHeader === "string" &&
      typeof column.normalizedHeader === "string" && typeof column.targetField === "string" &&
      typeof column.origin === "string";
  }).map((column) => ({ ...column }));
}

function vehicleFieldsFromPreview(value: unknown): DataTransferPreview["vehicleFields"] {
  if (!Array.isArray(value)) return [];
  return value.filter((item): item is NonNullable<DataTransferPreview["vehicleFields"]>[number] => {
    if (!item || typeof item !== "object") return false;
    const field = item as Record<string, unknown>;
    return typeof field.key === "string" && typeof field.labelDE === "string" &&
      typeof field.labelEN === "string" && ["string", "integer", "boolean"].includes(String(field.kind));
  }).map((field) => ({ ...field }));
}

const vehicleSetStringFields = [
  "inventoryNumber", "name", "manufacturer", "articleNumber", "articleSourceUrl", "gauge", "epoch",
  "railwayCompany", "category", "gattung", "description", "ean", "productionPeriod", "listPrice",
  "acquisitionType", "acquiredFrom", "purchasePrice", "purchaseDate", "storageLocation", "storageDetails",
  "condition", "conditionDetails", "packaging"
] as const satisfies readonly (keyof DataTransferVehicleSet)[];
const vehicleSetOptionalStringFields = ["id", "createdAt", "updatedAt"] as const;
const vehicleSetPreviewOptionalStringFields = ["targetId", "targetUpdatedAt", "targetFingerprint"] as const;

function vehicleSetPreviewsFromPreview(value: unknown): DataTransferVehicleSetPreview[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isVehicleSetPreview).map(cloneVehicleSetPreview);
}

function isVehicleSetPreview(value: unknown): value is DataTransferVehicleSetPreview {
  if (!isRecord(value)) return false;
  return typeof value.recordKey === "string" &&
    typeof value.classification === "string" &&
    ["ready", "warning", "error"].includes(value.classification) &&
    typeof value.proposedAction === "string" &&
    ["create", "replace", "use_existing", "copy"].includes(value.proposedAction) &&
    hasOptionalStrings(value, vehicleSetPreviewOptionalStringFields) &&
    isStringArray(value.memberRecordKeys) &&
    (value.rowNumbers === undefined || isNumberArray(value.rowNumbers)) &&
    (value.diagnostics === undefined || isVehicleSetDiagnostics(value.diagnostics)) &&
    isVehicleSet(value.data);
}

function isVehicleSet(value: unknown): value is DataTransferVehicleSet {
  if (!isRecord(value) || !vehicleSetStringFields.every((field) => typeof value[field] === "string")) return false;
  if (!hasOptionalStrings(value, vehicleSetOptionalStringFields)) return false;
  if (!Array.isArray(value.members)) return false;
  return value.members.every((member) => isRecord(member) &&
    typeof member.vehicleId === "string" &&
    typeof member.vehicleInventoryNumber === "string" &&
    typeof member.position === "number" &&
    (member.label === undefined || typeof member.label === "string")
  );
}

function isVehicleSetDiagnostics(value: unknown): boolean {
  return Array.isArray(value) && value.every((diagnostic) => isRecord(diagnostic) &&
    typeof diagnostic.rowNumber === "number" &&
    typeof diagnostic.field === "string" &&
    typeof diagnostic.code === "string"
  );
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === "string");
}

function isNumberArray(value: unknown): value is number[] {
  return Array.isArray(value) && value.every((item) => typeof item === "number");
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object";
}

function hasOptionalStrings(value: Record<string, unknown>, fields: readonly string[]) {
  return fields.every((field) => value[field] === undefined || typeof value[field] === "string");
}

function mappingOriginLabel(origin: DataTransferCSVColumnMapping["origin"], language: Language) {
  const labels = language === "de"
    ? { alias: "Automatisch", profile: "Profil", manual: "Manuell", ignored: "Ignoriert", unmapped: "Offen" }
    : { alias: "Automatic", profile: "Profile", manual: "Manual", ignored: "Ignored", unmapped: "Open" };
  return labels[origin];
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
    title: "Import prüfen", eyebrow: "SICHERER IMPORT", close: "Schließen", cancel: "Abbrechen",
    progress: "Importschritte",
    steps: { profile: "Profil", file: "Datei", mapping: "Zuordnung", review: "Prüfung", confirm: "Bestätigung" },
    profile: "Importprofil", chooseProfile: "Profil wählen", snapshot: "Auftragssnapshot", chooseFile: "Datei auswählen",
    fileLabel: "Importdatei", fileHelp: "Die Datei wird erst geprüft. Es werden noch keine Daten geschrieben.",
    browse: "Durchsuchen", serverPreview: "PERSISTENTE SERVERVORSCHAU", preview: "Vorschau", revision: "Revision",
    records: "Datensätze", ready: "bereit", warnings: "Hinweise", errors: "Fehler",
    unresolved: "{count} Konflikt(e) müssen vor dem Import aufgelöst werden.",
    readyToConfirm: "Alle Konflikte sind aufgelöst. Der geprüfte Stand kann importiert werden.",
    cancelJob: "Auftrag abbrechen", cancelJobTitle: "Importauftrag abbrechen?",
    cancelJobConfirm: "Diesen Importauftrag abbrechen? Die Vorschau bleibt im Verlauf erhalten.",
    changeTitle: "Auswahl ändern?", applyChange: "Änderung übernehmen",
    changeWarning: "Die aktuelle Vorschau wird verworfen und als neue Revision erneut geprüft. Fortfahren?",
    uploadError: "Die Datei konnte nicht geprüft werden.", resolveError: "Der Konflikt konnte nicht aufgelöst werden.",
    confirmError: "Der Import konnte nicht bestätigt werden.", cancelJobError: "Der Auftrag konnte nicht abgebrochen werden.",
    changeError: "Der Import konnte nicht auf ein anderes Profil umgestellt werden.",
    mappingTitle: "Erkannte CSV-Zuordnung", mappingExplanation: "Jede CSV-Spalte wird einem RailKeeper-Feld zugeordnet. Nicht benötigte Spalten können ausdrücklich ignoriert werden.",
    sourceColumn: "CSV-Spalte", targetField: "RailKeeper-Feld", mappingSource: "Erkennung",
    targetFor: "Zielfeld für {source}", unmapped: "Nicht zugeordnet", ignore: "Spalte ignorieren",
    mappingIncomplete: "Alle CSV-Spalten müssen zugeordnet oder ignoriert werden.",
    legacyMapping: "Diese ältere Vorschau enthält noch keine Spaltenzuordnung. Sie kann unverändert geprüft werden.",
    saveMapping: "Zuordnung im Profil speichern", validateMapping: "Zuordnung prüfen",
    continueReview: "Weiter zur Prüfung", mappingFileRequired: "Für Änderungen muss die CSV-Datei erneut ausgewählt werden.",
    mappingError: "Die geänderte CSV-Zuordnung konnte nicht geprüft werden.",
    conflictRecovery: "Der Auftrag wurde zwischenzeitlich geändert. Die persistente Vorschau wurde neu gelesen. Bitte Zuordnung und Konflikte erneut prüfen.",
    confirmConflictRecovery: "Der Auftrag wurde beim Bestätigen geändert. Bitte die Quelldatei erneut hochladen und die neue Vorschau vollständig prüfen."
  } : {
    title: "Review import", eyebrow: "SAFE IMPORT", close: "Close", cancel: "Cancel", progress: "Import steps",
    steps: { profile: "Profile", file: "File", mapping: "Mapping", review: "Review", confirm: "Confirmation" },
    profile: "Import profile", chooseProfile: "Choose profile", snapshot: "Job snapshot", chooseFile: "Choose file",
    fileLabel: "Import file", fileHelp: "The file is reviewed first. No data is written yet.", browse: "Browse",
    serverPreview: "PERSISTENT SERVER PREVIEW", preview: "Preview", revision: "Revision", records: "records",
    ready: "ready", warnings: "warnings", errors: "errors",
    unresolved: "{count} conflict(s) must be resolved before import.",
    readyToConfirm: "All conflicts are resolved. The reviewed revision can be imported.", cancelJob: "Cancel job",
    cancelJobTitle: "Cancel import job?", changeTitle: "Change selection?", applyChange: "Apply change",
    cancelJobConfirm: "Cancel this import job? The preview remains in history.",
    changeWarning: "The current preview will be discarded and reviewed as a new revision. Continue?",
    uploadError: "The file could not be reviewed.", resolveError: "The conflict could not be resolved.",
    confirmError: "The import could not be confirmed.", cancelJobError: "The job could not be cancelled.",
    changeError: "The import could not be changed to another profile.",
    mappingTitle: "Detected CSV mapping", mappingExplanation: "Each CSV column is assigned to a RailKeeper field. Unneeded columns can be explicitly ignored.",
    sourceColumn: "CSV column", targetField: "RailKeeper field", mappingSource: "Detection",
    targetFor: "Target field for {source}", unmapped: "Not mapped", ignore: "Ignore column",
    mappingIncomplete: "Every CSV column must be mapped or ignored.", saveMapping: "Save mapping to profile",
    legacyMapping: "This older preview does not contain column mapping data. It can be reviewed unchanged.",
    validateMapping: "Validate mapping", continueReview: "Continue to review",
    mappingFileRequired: "Select the CSV file again to change its mapping.",
    mappingError: "The changed CSV mapping could not be validated.",
    conflictRecovery: "The job changed in the meantime. The persistent preview was re-read. Review the mapping and conflicts again.",
    confirmConflictRecovery: "The job changed during confirmation. Upload the source file again and fully review the new preview."
  };
}

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}
