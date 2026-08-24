import { CheckCircle2, ChevronRight, CircleAlert, History, Trash2, XCircle } from "lucide-react";
import type { KeyboardEvent } from "react";

import { formatDateTime, type Language } from "../../shared/i18n";
import type { DataTransferArea, DataTransferJob, DataTransferJobState } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;
type TerminalJobState = Extract<
  DataTransferJobState,
  "completed" | "completed_with_warnings" | "failed" | "cancelled"
>;

const terminalStates: TerminalJobState[] = ["completed", "completed_with_warnings", "failed", "cancelled"];

type TransferHistoryTableProps = {
  canDelete: boolean;
  jobs: DataTransferJob[];
  language: Language;
  mutating: boolean;
  onDeleteRequest: (job: DataTransferJob) => void;
  onSelect: (id: string) => void;
  selectedJobId: string | null;
  t: Translate;
};

export function TransferHistoryTable({
  canDelete,
  jobs,
  language,
  mutating,
  onDeleteRequest,
  onSelect,
  selectedJobId,
  t
}: TransferHistoryTableProps) {
  const terminalJobs = jobs.filter((job): job is DataTransferJob & { state: TerminalJobState } =>
    terminalStates.includes(job.state as TerminalJobState)
  );

  return (
    <section className="panel data-transfer-panel transfer-history-panel">
      <header className="data-transfer-panel-head">
        <h2><History size={20} aria-hidden="true" />{t("importExport.dashboard.history.title")}</h2>
      </header>
      <div className="data-transfer-table-wrap">
        <table className="data-transfer-table transfer-history-table">
          <thead>
            <tr>
              <th>{t("importExport.dashboard.history.time")}</th>
              <th>{t("importExport.dashboard.history.operation")}</th>
              <th>{t("importExport.dashboard.history.area")}</th>
              <th>{t("importExport.dashboard.history.records")}</th>
              <th>{t("importExport.dashboard.history.result")}</th>
              <th>{t("importExport.dashboard.history.file")}</th>
              <th aria-label={t("importExport.dashboard.history.action")} />
            </tr>
          </thead>
          <tbody>
            {terminalJobs.length === 0 ? (
              <tr><td className="data-transfer-empty" colSpan={7}>{t("importExport.dashboard.history.empty")}</td></tr>
            ) : terminalJobs.map((job) => (
              <tr
                className="transfer-history-row"
                aria-selected={job.id === selectedJobId}
                key={job.id}
                onClick={() => onSelect(job.id)}
                onKeyDown={(event) => selectFromKeyboard(event, job.id, onSelect)}
                tabIndex={0}
              >
                <td>{formatDateTime(job.completedAt || job.createdAt, language)}</td>
                <td>{t(`importExport.dashboard.direction.${job.direction}`)}</td>
                <td title={areaLabels(job.areas, t)}>
                  <span className="data-transfer-truncate">{areaLabels(job.areas, t)}</span>
                </td>
                <td>{new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB").format(job.totalRecords)}</td>
                <td><HistoryResult state={job.state} t={t} /></td>
                <td><span className="data-transfer-truncate" title={job.sourceName}>{job.sourceName || "–"}</span></td>
                <td><span className="transfer-history-actions">
                  <button
                    type="button"
                    className="icon-button transfer-row-menu"
                    aria-label={`${job.profileName} ${t("importExport.dashboard.history.details")}`}
                    title={t("importExport.dashboard.history.details")}
                    onClick={(event) => {
                      event.stopPropagation();
                      onSelect(job.id);
                    }}
                  >
                    <ChevronRight size={16} aria-hidden="true" />
                  </button>
                  {canDelete && job.state === "cancelled" ? (
                    <button type="button" className="icon-button transfer-row-menu transfer-history-delete"
                      aria-label={t("importExport.dashboard.delete.action", { name: job.profileName })}
                      title={t("importExport.dashboard.delete.action", { name: job.profileName })}
                      disabled={mutating} onClick={(event) => {
                        event.stopPropagation();
                        onDeleteRequest(job);
                      }}>
                      <Trash2 size={16} aria-hidden="true" />
                    </button>
                  ) : null}
                </span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function HistoryResult({ state, t }: { state: TerminalJobState; t: Translate }) {
  const { label, tone } = terminalStatePresentation[state];
  const Icon = tone === "success" ? CheckCircle2 : tone === "warning" ? CircleAlert : XCircle;
  return (
    <span className={`transfer-history-result ${tone}`}>
      <Icon size={14} aria-hidden="true" />{t(`importExport.dashboard.state.${label}`)}
    </span>
  );
}

function selectFromKeyboard(event: KeyboardEvent<HTMLTableRowElement>, id: string, onSelect: (id: string) => void) {
  if (event.key !== "Enter" && event.key !== " ") return;
  event.preventDefault();
  onSelect(id);
}

function areaLabels(areas: DataTransferArea[], t: Translate) {
  return areas.map((area) => t(`importExport.dashboard.area.${area}`)).join(", ");
}

const terminalStatePresentation: Record<TerminalJobState, {
  label: "completed" | "completedWarnings" | "failed" | "cancelled";
  tone: "success" | "warning" | "danger";
}> = {
  completed: { label: "completed", tone: "success" },
  completed_with_warnings: { label: "completedWarnings", tone: "warning" },
  failed: { label: "failed", tone: "danger" },
  cancelled: { label: "cancelled", tone: "danger" }
};
