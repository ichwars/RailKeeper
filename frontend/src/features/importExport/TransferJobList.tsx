import { ChevronRight, FileJson, FileSpreadsheet, ListChecks } from "lucide-react";

import type { Language } from "../../shared/i18n";
import type { DataTransferJob, DataTransferJobFilter, DataTransferJobState } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferJobListProps = {
  allJobs: DataTransferJob[];
  filters: DataTransferJobFilter;
  jobs: DataTransferJob[];
  language: Language;
  loading: boolean;
  onFilter: (filter: DataTransferJobFilter) => void;
  onSelect: (id: string) => void;
  selectedJobId: string | null;
  t: Translate;
};

const openStates: DataTransferJobState[] = ["draft", "reading", "review_required", "ready", "running"];
const completedStates: DataTransferJobState[] = ["completed", "completed_with_warnings", "failed", "cancelled"];

export function TransferJobList({
  allJobs,
  filters,
  jobs,
  language,
  loading,
  onFilter,
  onSelect,
  selectedJobId,
  t
}: TransferJobListProps) {
  const selectedFilter = filterKind(filters.states);
  const filterItems = [
    { id: "all", label: t("importExport.dashboard.jobs.all"), count: allJobs.length, states: [] },
    {
      id: "open",
      label: t("importExport.dashboard.jobs.open"),
      count: allJobs.filter((job) => openStates.includes(job.state)).length,
      states: openStates
    },
    {
      id: "completed",
      label: t("importExport.dashboard.jobs.completed"),
      count: allJobs.filter((job) => completedStates.includes(job.state)).length,
      states: completedStates
    }
  ];

  return (
    <section className="panel data-transfer-panel transfer-jobs-panel">
      <header className="data-transfer-panel-head">
        <h2><ListChecks size={20} aria-hidden="true" />{t("importExport.dashboard.jobs.title")}</h2>
      </header>
      <div className="transfer-job-filters" role="group" aria-label={t("importExport.dashboard.jobs.filters")}>
        {filterItems.map((filter) => (
          <button
            type="button"
            className={selectedFilter === filter.id ? "selected" : ""}
            aria-label={filter.label}
            aria-pressed={selectedFilter === filter.id}
            key={filter.id}
            onClick={() => onFilter({ states: filter.states, limit: filters.limit ?? 100 })}
          >
            {filter.label}<span>{formatNumber(filter.count, language)}</span>
          </button>
        ))}
      </div>
      <div className="transfer-job-cards" aria-live="polite">
        {loading && jobs.length === 0 ? (
          <p className="data-transfer-empty">{t("importExport.dashboard.loading")}</p>
        ) : jobs.length === 0 ? (
          <p className="data-transfer-empty">{t("importExport.dashboard.jobs.empty")}</p>
        ) : jobs.map((job) => (
          <JobCard
            job={job}
            key={job.id}
            language={language}
            onSelect={onSelect}
            selected={job.id === selectedJobId}
            t={t}
          />
        ))}
      </div>
    </section>
  );
}

function JobCard({
  job,
  language,
  onSelect,
  selected,
  t
}: {
  job: DataTransferJob;
  language: Language;
  onSelect: (id: string) => void;
  selected: boolean;
  t: Translate;
}) {
  const FileIcon = job.format === "csv" ? FileSpreadsheet : FileJson;
  const maximum = Math.max(job.totalRecords, 1);
  const completed = job.state === "completed" || job.state === "completed_with_warnings";
  const progressValue = completed
    ? maximum
    : Math.min(job.readyRecords, maximum);
  const displayedReady = completed ? job.totalRecords : Math.min(job.readyRecords, job.totalRecords);

  return (
    <button
      type="button"
      className={`transfer-job-card${selected ? " selected" : ""}`}
      aria-pressed={selected}
      onClick={() => onSelect(job.id)}
      title={job.profileName}
    >
      <FileIcon className="transfer-job-file-icon" size={29} aria-hidden="true" />
      <span className="transfer-job-copy">
        <strong>{job.profileName}</strong>
        <small title={job.sourceName}>{job.sourceName || formatLabel(job.format, t)}</small>
        <span className={`transfer-state-badge ${stateTone(job.state)}`}>{stateLabel(job.state, t)}</span>
        <span className="transfer-job-progress-copy">
          {formatNumber(displayedReady, language)}/{formatNumber(job.totalRecords, language)} {t("importExport.readyShort")}
        </span>
        <progress
          max={maximum}
          value={progressValue}
          aria-label={`${t("importExport.dashboard.jobs.progress")} ${job.profileName}`}
        />
      </span>
      <ChevronRight size={18} aria-hidden="true" />
    </button>
  );
}

function filterKind(states: DataTransferJobState[] | undefined) {
  if (!states?.length) return "all";
  if (states.length === openStates.length && states.every((state) => openStates.includes(state))) return "open";
  return "completed";
}

function formatNumber(value: number, language: Language) {
  return new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB").format(value);
}

function formatLabel(format: DataTransferJob["format"], t: Translate) {
  return t(`importExport.dashboard.format.${format === "csv" ? "csv" : "json"}`);
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
