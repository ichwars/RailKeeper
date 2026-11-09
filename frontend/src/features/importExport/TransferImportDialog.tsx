import { CheckCircle2, FileUp, ShieldCheck, Upload, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState, type ChangeEvent } from "react";
import { createPortal } from "react-dom";
import { observeAction } from "../../shared/asyncAction";
import type { ImportDialogStep } from "./transferImportModel";
import { areaLabels, cloneCSVMapping, clonePreview, csvMappingIdentity, currentStep, errorMessage, importButtonLabel, importCopy, mappingOriginLabel, previewFromDetails, snapshotName, stepClass, vehicleCSVMappingRequired } from "./transferImportModel";
export type { ImportDialogStep } from "./transferImportModel";

import { ImportReviewContext } from "../../shared/ImportReview";
import { ApiError } from "../../shared/api";
import type { Language } from "../../shared/i18n";
import { translate } from "../../shared/i18n";
import { AppCheckbox } from "../../shared/ui/AppCheckbox";
import { AppFileButton } from "../../shared/ui/AppFileButton";
import { AppSelect } from "../../shared/ui/AppSelect";
import { useModalDialogLayer } from "../../shared/ui/useModalDialogLayer";
import { TransferConfirmDialog, type TransferPendingAction } from "./TransferConfirmDialog";
import { TransferReviewTable } from "./TransferReviewTable";
import { TransferSetReview } from "./TransferSetReview";
import type {
  DataTransferCSVColumnMapping,
  DataTransferCSVMappingInput,
  DataTransferIssue,
  DataTransferIssueResolution,
  DataTransferJob,
  DataTransferJobDetails,
  DataTransferPreview,
  DataTransferProfile
} from "./dataTransferModel";
import { useTransferImportConfirmation } from "./useTransferImportConfirmation";

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
  const confirmation = useTransferImportConfirmation({ job, language, confirm: onConfirm, refresh: onRefreshJob,
    recovered: adoptRecoveredDetails, close: onClose, setBusy, setError });

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
  }, [initialJob?.id, initialRequiresReupload, language, copy.confirmConflictRecovery]);

  function requestProfileChange(nextProfileId: string) {
    if (confirmation.blocked) return;
    if (nextProfileId === profileId) return;
    if (preview) {
      setPendingChange({ kind: "profile", profileId: nextProfileId });
      return;
    }
    observeAction(changeProfile(nextProfileId));
  }

  async function changeProfile(nextProfileId: string) {
    if (confirmation.blocked) return;
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
    if (confirmation.blocked) return;
    const nextFile = event.target.files?.[0];
    event.target.value = "";
    if (!nextFile || (!profileId && !job)) return;
    if (preview) {
      setPendingChange({ kind: "file", file: nextFile });
      return;
    }
    observeAction(uploadFile(nextFile));
  }

  async function uploadFile(nextFile: File) {
    if (confirmation.blocked) return;
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
    if (confirmation.blocked) return;
    setCSVMapping((current) => current.map((column) => {
      if (column.index !== columnIndex) return column;
      if (value === "__ignore__") return { ...column, targetField: "", origin: "ignored" };
      if (value === "") return { ...column, targetField: "", origin: "unmapped" };
      return { ...column, targetField: value, origin: "manual" };
    }));
    setMappingDirty(true);
  }

  async function applyMapping() {
    if (confirmation.blocked) return;
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
    if (confirmation.blocked) return;
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
    await confirmation.confirm();
  }

  function adoptRecoveredDetails(details: DataTransferJobDetails, reuploadRequired: boolean) {
    requiresReuploadRef.current = reuploadRequired;
    const refreshedPreview = previewFromDetails(details.job, details);
    const refreshedMapping = refreshedPreview?.csvMapping ?? [];
    const refreshedIdentity = csvMappingIdentity(details.job, refreshedMapping);
    const keepAccepted = !reuploadRequired && acceptedMappingIdentityRef.current === refreshedIdentity;
    if (!keepAccepted) acceptedMappingIdentityRef.current = null;
    setJob(details.job);
    setPreview(refreshedPreview);
    setCSVMapping(refreshedMapping);
    setMappingAccepted(!reuploadRequired && (!vehicleCSVMappingRequired(details.job) || keepAccepted));
    setRequiresReupload(reuploadRequired);
    setError(details.job.state === "completed" ? "" : reuploadRequired ? copy.confirmConflictRecovery : copy.conflictRecovery);
  }

  async function recoverConflict(reason: unknown, jobId: string | undefined, fallback: string) {
    if (!(reason instanceof ApiError) || reason.status !== 409 || !jobId) {
      setError(errorMessage(reason, fallback));
      return;
    }
    try {
      const details = await onRefreshJob(jobId);
      adoptRecoveredDetails(details, false);
    } catch (refreshError) {
      setError(errorMessage(refreshError, copy.conflictRecovery));
    }
  }

  async function cancelJob() {
    if (!job || confirmation.blocked) return;
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
              <AppSelect aria-label={copy.profile} value={profileId} disabled={busy || confirmation.blocked}
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
            <div className={`data-transfer-file-drop${busy ? " busy" : ""}`}>
              <FileUp size={28} aria-hidden="true" />
              <span><strong>{file?.name || job?.sourceName || copy.chooseFile}</strong><small>{copy.fileHelp}</small></span>
              <AppFileButton ref={fileRef} aria-label={copy.fileLabel} disabled={busy || confirmation.blocked}
                accept={snapshot?.format === "csv" ? ".csv,text/csv" : ".json,application/json"}
                onChange={chooseFile}><Upload size={15} aria-hidden="true" />{copy.browse}</AppFileButton>
            </div>
          ) : null}

          {preview && !confirmation.blocked && preview.job.format === "csv" && !mappingAccepted ? (
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

          {preview && !confirmation.blocked && mappingAccepted && !requiresReupload ? (
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
              <ImportReviewContext language={language} source={preview.job.sourceName || file?.name || translate(language, "importReview.unknownSource")}
                target={translate(language, "importReview.target.inventory")} effect={translate(language, "importReview.effect.transfer")} />
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

          {confirmation.blocked ? <p className="form-message" role={confirmation.outcome === "unknown" ? "alert" : "status"}>
            {confirmation.message}</p> : null}
          {error && !confirmation.blocked ? <p className="form-message error" role="alert">{error}</p> : null}
        </div>

        <footer className="data-transfer-dialog-actions">
          {job && !confirmation.blocked ? <button type="button" className="secondary-button" disabled={busy}
            onClick={() => setPendingChange({ kind: "cancel" })}>{copy.cancelJob}</button> : <span />}
          <span>
            <button type="button" className="secondary-button" disabled={busy} onClick={onClose}>{copy.close}</button>
            {confirmation.outcome === "unknown" ? <button type="button" className="primary-button" disabled={busy}
              onClick={() => observeAction(confirmation.check())}>{confirmation.checkLabel}</button> : null}
            {preview && !confirmation.blocked && preview.job.format === "csv" && !mappingAccepted ? (
              <button type="button" className="primary-button"
                disabled={busy || requiresReupload || !mappingReady} onClick={() => observeAction(applyMapping())}>
                {mappingDirty || saveMappingToProfile || !mappingReady ? copy.validateMapping : copy.continueReview}
              </button>
            ) : null}
            {preview && !confirmation.blocked && mappingAccepted && !requiresReupload ? (
              <button type="button" className="primary-button" disabled={busy || unresolvedIssues.length > 0 ||
                approvedRecords === 0 || job?.state !== "ready"} onClick={() => observeAction(confirmImport())}>
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
