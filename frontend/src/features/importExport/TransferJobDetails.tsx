import {
  CheckCircle2,
  Circle,
  CircleAlert,
  ClipboardCheck,
  FileJson,
  FileSpreadsheet,
  Play,
  RotateCcw
} from "lucide-react";

import { formatDateTime, type Language } from "../../shared/i18n";
import type { DataTransferArea, DataTransferJob, DataTransferJobState } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferJobDetailsProps = {
  canImport: boolean;
  canExport: boolean;
  canRetry: (job: DataTransferJob) => boolean;
  detailLoading: boolean;
  job: DataTransferJob | null;
  language: Language;
  mutating: boolean;
  onContinue: (job: DataTransferJob) => void;
  onRetry: (jobId: string) => Promise<unknown>;
  t: Translate;
};

export function TransferJobDetails({
  canImport,
  canExport,
  canRetry,
  detailLoading,
  job,
  language,
  mutating,
  onContinue,
  onRetry,
  t
}: TransferJobDetailsProps) {
  return (
    <section className="panel data-transfer-panel transfer-details-panel">
      <header className="data-transfer-panel-head">
        <h2><ClipboardCheck size={20} aria-hidden="true" />{t("importExport.dashboard.details.title")}</h2>
      </header>
      {!job ? (
        <p className="data-transfer-empty">
          {detailLoading ? t("importExport.dashboard.loading") : t("importExport.dashboard.details.empty")}
        </p>
      ) : <SelectedJobDetails canExport={canExport} canImport={canImport} canRetry={canRetry} job={job} language={language} mutating={mutating}
        onContinue={onContinue} onRetry={onRetry} t={t} />}
    </section>
  );
}

function SelectedJobDetails({
  canExport,
  canImport,
  canRetry,
  job,
  language,
  mutating,
  onContinue,
  onRetry,
  t
}: Omit<TransferJobDetailsProps, "detailLoading" | "job"> & { job: DataTransferJob }) {
  const FileIcon = job.format === "csv" ? FileSpreadsheet : FileJson;
  const progress = jobProgress(job);
  const stages = job.direction === "import"
    ? ["file", "mapping", "review"]
    : ["preparation", "export", "ready"];

  return (
    <div className="transfer-details-content">
      <div className="transfer-detail-title">
        <span className={`transfer-detail-icon ${stateTone(job.state)}`}>
          <FileIcon size={24} aria-hidden="true" />
        </span>
        <div>
          <h3 title={job.profileName}>{job.profileName}</h3>
          <span className={`transfer-state-badge ${stateTone(job.state)}`}>{stateLabel(job.state, t)}</span>
        </div>
      </div>

      <dl className="transfer-detail-list">
        <div><dt>{t("importExport.dashboard.details.source")}</dt><dd title={job.sourceName}>{job.sourceName || "–"}</dd></div>
        <div><dt>{t("importExport.dashboard.details.area")}</dt><dd>{areaLabels(job.areas, t)}</dd></div>
        <div><dt>{t("importExport.dashboard.details.format")}</dt><dd>{job.format === "csv" ? "CSV" : "JSON"}</dd></div>
        <div><dt>{t("importExport.dashboard.details.created")}</dt><dd>{formatDateTime(job.createdAt, language)}</dd></div>
      </dl>

      <div className="transfer-detail-metrics">
        <Metric value={job.totalRecords} label={t("importExport.dashboard.details.rows")} language={language} />
        <Metric value={job.readyRecords} label={t("importExport.dashboard.details.ready")} language={language} tone="success" />
        <Metric value={job.warningRecords} label={t("importExport.dashboard.details.notes")} language={language} tone="warning" />
        <Metric value={job.errorRecords} label={t("importExport.dashboard.details.errors")} language={language} tone="danger" />
      </div>

      <div className="transfer-stage-progress">
        <progress
          max={100}
          value={progress}
          aria-label={`${t("importExport.dashboard.details.progress")} ${job.profileName}`}
          aria-valuenow={progress}
        />
        <ol>
          {stages.map((stage, index) => {
            const completed = progress >= ((index + 1) / stages.length) * 100;
            const current = !completed && progress > (index / stages.length) * 100;
            const Icon = completed ? CheckCircle2 : current ? CircleAlert : Circle;
            return (
              <li className={completed ? "completed" : current ? "current" : ""} key={stage}>
                <Icon size={18} aria-hidden="true" />
                <span>{t(`importExport.dashboard.details.stage.${stage}`)}</span>
              </li>
            );
          })}
        </ol>
      </div>

      <JobAction canExport={canExport} canImport={canImport} canRetry={canRetry} job={job} mutating={mutating}
        onContinue={onContinue} onRetry={onRetry} t={t} />
    </div>
  );
}

function Metric({
  label,
  language,
  tone = "",
  value
}: {
  label: string;
  language: Language;
  tone?: string;
  value: number;
}) {
  return (
    <span className={tone}>
      <strong>{new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB").format(value)}</strong>
      <small>{label}</small>
    </span>
  );
}

function JobAction({
  canExport,
  canImport,
  canRetry,
  job,
  mutating,
  onContinue,
  onRetry,
  t
}: Pick<TransferJobDetailsProps, "canExport" | "canImport" | "canRetry" | "mutating" | "onContinue" | "onRetry" | "t"> & {
  job: DataTransferJob;
}) {
  if (["completed", "completed_with_warnings", "failed", "cancelled"].includes(job.state) && canRetry(job)) {
    return (
      <button type="button" className="primary-button transfer-detail-action" disabled={mutating}
        onClick={() => void onRetry(job.id).catch(() => undefined)}>
        <RotateCcw size={16} aria-hidden="true" />{t("importExport.dashboard.details.retry")}
      </button>
    );
  }
  if (canExport && job.direction === "export" && job.state === "draft") {
    return (
      <button type="button" className="primary-button transfer-detail-action" disabled={mutating}
        onClick={() => onContinue(job)}>
        <Play size={16} fill="currentColor" aria-hidden="true" />
        {job.profileName ? "Export fortsetzen" : t("importExport.dashboard.details.continue")}
      </button>
    );
  }
  if (canImport && job.direction === "import" && job.state === "ready") {
    return (
      <button type="button" className="primary-button transfer-detail-action" disabled={mutating}
        onClick={() => onContinue(job)}>
        <CheckCircle2 size={16} aria-hidden="true" />{t("importExport.dashboard.details.confirm")}
      </button>
    );
  }
  if (canImport && job.direction === "import" && ["draft", "reading", "review_required"].includes(job.state)) {
    return (
      <button type="button" className="primary-button transfer-detail-action" disabled={mutating}
        onClick={() => onContinue(job)}>
        <Play size={16} fill="currentColor" aria-hidden="true" />{t("importExport.dashboard.details.continue")}
      </button>
    );
  }
  return null;
}

function areaLabels(areas: DataTransferArea[], t: Translate) {
  return areas.map((area) => t(`importExport.dashboard.area.${area}`)).join(", ");
}

function stateLabel(state: DataTransferJobState, t: Translate) {
  const keys: Record<DataTransferJobState, string> = {
    draft: "draft",
    reading: "reading",
    review_required: "reviewRequired",
    ready: "ready",
    running: "running",
    completed: "completed",
    completed_with_warnings: "completedWarnings",
    failed: "failed",
    cancelled: "cancelled"
  };
  return t(`importExport.dashboard.state.${keys[state]}`);
}

function stateTone(state: DataTransferJobState) {
  if (state === "failed" || state === "cancelled") return "danger";
  if (state === "review_required" || state === "completed_with_warnings") return "warning";
  if (state === "completed" || state === "ready") return "success";
  return "neutral";
}

function jobProgress(job: DataTransferJob) {
  if (job.state === "completed" || job.state === "completed_with_warnings") return 100;
  const stageProgress: Record<DataTransferJob["stage"], number> = {
    created: 12,
    snapshot: 34,
    preview: 58,
    review: 78,
    completed: 100,
    failed: 78,
    cancelled: 0
  };
  return stageProgress[job.stage];
}
