import { AlertTriangle, CheckCircle2, CircleAlert } from "lucide-react";

import type { Language } from "../../shared/i18n";
import type {
  DataTransferIssue,
  DataTransferIssueResolution,
  DataTransferPreviewRecord
} from "./dataTransferModel";

type TransferReviewTableProps = {
  busy: boolean;
  issues: DataTransferIssue[];
  language: Language;
  onResolve: (issue: DataTransferIssue, resolution: DataTransferIssueResolution) => Promise<void>;
  records: DataTransferPreviewRecord[];
};

const resolutionLabels = {
  de: {
    replace: "Vorhandenen Datensatz ersetzen",
    merge: "Einträge zusammenführen",
    copy: "Als Kopie anlegen",
    skip: "Überspringen",
    use_existing: "Vorhandenen Datensatz verwenden",
    create: "Neu anlegen",
    link: "Fahrzeug verknüpfen"
  },
  en: {
    replace: "Replace existing record",
    merge: "Merge entries",
    copy: "Create a copy",
    skip: "Skip",
    use_existing: "Use existing record",
    create: "Create new record",
    link: "Link vehicle"
  }
} satisfies Record<Language, Record<DataTransferIssueResolution, string>>;

export function TransferReviewTable({ busy, issues, language, onResolve, records }: TransferReviewTableProps) {
  const copy = language === "de"
    ? { area: "Bereich", record: "Datensatz", status: "Status", action: "Vorschlag", issue: "Prüfung",
      choose: "Auflösung wählen", row: "Zeile", ready: "Bereit", warning: "Hinweis", error: "Fehler" }
    : { area: "Area", record: "Record", status: "Status", action: "Proposed action", issue: "Review",
      choose: "Choose resolution", row: "Row", ready: "Ready", warning: "Warning", error: "Error" };

  return (
    <div className="data-transfer-table-wrap transfer-review-wrap">
      <table className="data-transfer-table transfer-review-table">
        <thead>
          <tr>
            <th>{copy.area}</th>
            <th>{copy.record}</th>
            <th>{copy.status}</th>
            <th>{copy.action}</th>
            <th>{copy.issue}</th>
          </tr>
        </thead>
        <tbody>
          {records.map((record, index) => {
            const recordIssues = issues.filter((issue) =>
              issue.area === record.area && issue.recordKey === record.recordKey
            );
            const Icon = record.classification === "error"
              ? CircleAlert
              : record.classification === "warning" ? AlertTriangle : CheckCircle2;
            return (
              <tr className={`transfer-review-${record.classification}`} key={`${record.area}-${record.recordKey}-${index}`}>
                <td>{areaLabel(record.area, language)}</td>
                <td>
                  <strong>{record.recordKey || "–"}</strong>
                  {record.rowNumber ? <small>{copy.row} {record.rowNumber}</small> : null}
                </td>
                <td>
                  <span className={`transfer-review-status ${record.classification}`}>
                    <Icon size={15} aria-hidden="true" />
                    {copy[record.classification]}
                  </span>
                </td>
                <td>{actionLabel(record.proposedAction, language)}</td>
                <td>
                  {recordIssues.length === 0 ? "–" : recordIssues.map((issue) => (
                    <label className="transfer-review-resolution" key={issue.id}>
                      <span>{issue.message}</span>
                      <select
                        aria-label={`${language === "de" ? "Auflösung für" : "Resolution for"} ${issue.recordKey}`}
                        disabled={busy}
                        value={issue.selectedResolution}
                        onChange={(event) => {
                          const resolution = event.target.value as DataTransferIssueResolution;
                          if (resolution) void onResolve(issue, resolution);
                        }}
                      >
                        <option value="">{copy.choose}</option>
                        {resolutionsFor(issue).map((resolution) => (
                          <option key={resolution} value={resolution}>{resolutionLabels[language][resolution]}</option>
                        ))}
                      </select>
                    </label>
                  ))}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function resolutionsFor(issue: DataTransferIssue): DataTransferIssueResolution[] {
  const byCode: Record<string, DataTransferIssueResolution[]> = {
    missing_inventory_number: ["skip"],
    missing_manufacturer: ["skip"],
    missing_name: ["skip"],
    missing_gauge: ["skip"],
    missing_category: ["skip"],
    missing_gattung: ["skip"],
    invalid_vehicle: ["skip"],
    invalid_accessory: ["skip"],
    duplicate_inventory_number: ["replace", "copy", "skip"],
    matching_manufacturer_article_number: ["use_existing", "create", "skip"],
    duplicate_exhibition_list: ["replace", "merge", "copy", "skip"],
    locked_exhibition_list: ["copy", "skip"],
    exhibition_vehicle_reference: ["link", "skip"],
    missing_vehicle_reference: ["skip"],
    duplicate_input_inventory_number: ["skip"]
  };
  return byCode[issue.code] || proposedResolutions(issue);
}

function proposedResolutions(issue: DataTransferIssue): DataTransferIssueResolution[] {
  if (issue.proposedResolution === "replace_or_copy") return ["replace", "copy", "skip"];
  if (issue.proposedResolution) return [issue.proposedResolution, "skip"]
    .filter((value, index, values) => values.indexOf(value) === index) as DataTransferIssueResolution[];
  return ["skip"];
}

function areaLabel(area: DataTransferPreviewRecord["area"], language: Language) {
  const labels = language === "de"
    ? { vehicles: "Fahrzeuge", accessories: "Zubehör", exhibitionLists: "Ausstellungslisten" }
    : { vehicles: "Vehicles", accessories: "Accessories", exhibitionLists: "Exhibition lists" };
  return labels[area];
}

function actionLabel(action: DataTransferPreviewRecord["proposedAction"], language: Language) {
  const labels = language === "de"
    ? { create: "Neu anlegen", replace: "Ersetzen", use_existing: "Vorhandenen verwenden", copy: "Kopie" }
    : { create: "Create", replace: "Replace", use_existing: "Use existing", copy: "Copy" };
  return labels[action];
}
