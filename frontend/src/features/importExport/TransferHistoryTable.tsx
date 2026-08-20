import { CheckCircle2, ChevronRight, CircleAlert, History, XCircle } from "lucide-react";
import type { KeyboardEvent } from "react";

import { formatDateTime, type Language } from "../../shared/i18n";
import type { DataTransferArea, DataTransferJob, DataTransferJobState } from "./dataTransferModel";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type TransferHistoryTableProps = {
  jobs: DataTransferJob[];
  language: Language;
  onSelect: (id: string) => void;
  selectedJobId: string | null;
  t: Translate;
};

export function TransferHistoryTable({ jobs, language, onSelect, selectedJobId, t }: TransferHistoryTableProps) {
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
            {jobs.length === 0 ? (
              <tr><td className="data-transfer-empty" colSpan={7}>{t("importExport.dashboard.history.empty")}</td></tr>
            ) : jobs.map((job) => (
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
                <td className="data-transfer-truncate" title={areaLabels(job.areas, t)}>{areaLabels(job.areas, t)}</td>
                <td>{new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB").format(job.totalRecords)}</td>
                <td><HistoryResult state={job.state} t={t} /></td>
                <td><span className="data-transfer-truncate" title={job.sourceName}>{job.sourceName || "–"}</span></td>
                <td>
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
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function HistoryResult({ state, t }: { state: DataTransferJobState; t: Translate }) {
  const tone = stateTone(state);
  const Icon = tone === "success" ? CheckCircle2 : tone === "warning" ? CircleAlert : XCircle;
  return (
    <span className={`transfer-history-result ${tone}`}>
      <Icon size={14} aria-hidden="true" />{stateLabel(state, t)}
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

function stateLabel(state: DataTransferJobState, t: Translate) {
  if (state === "completed") return t("importExport.dashboard.state.completed");
  if (state === "completed_with_warnings") return t("importExport.dashboard.state.completedWarnings");
  if (state === "failed") return t("importExport.dashboard.state.failed");
  if (state === "cancelled") return t("importExport.dashboard.state.cancelled");
  return t("importExport.dashboard.state.running");
}

function stateTone(state: DataTransferJobState) {
  if (state === "completed") return "success";
  if (state === "completed_with_warnings" || state === "review_required") return "warning";
  return state === "failed" || state === "cancelled" ? "danger" : "success";
}
