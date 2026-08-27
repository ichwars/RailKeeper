import { AlertTriangle, CheckCircle2, CircleAlert } from "lucide-react";

import type { Language } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
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
    replace: "Bestehenden Datensatz überschreiben",
    merge: "Einträge zusammenführen",
    copy: "Als Kopie anlegen",
    skip: "Diesen Datensatz nicht importieren",
    use_existing: "Vorhandenen Datensatz verwenden",
    create: "Neu anlegen",
    link: "Fahrzeug verknüpfen"
  },
  en: {
    replace: "Overwrite existing record",
    merge: "Merge entries",
    copy: "Create a copy",
    skip: "Do not import this record",
    use_existing: "Use existing record",
    create: "Create new record",
    link: "Link vehicle"
  }
} satisfies Record<Language, Record<DataTransferIssueResolution, string>>;

export function TransferReviewTable({ busy, issues, language, onResolve, records }: TransferReviewTableProps) {
  const copy = language === "de"
    ? { area: "Bereich", record: "Datensatz", status: "Status", action: "Vorschlag", issue: "Prüfung",
      choose: "Aktion wählen", row: "Zeile", ready: "Bereit", warning: "Hinweis", error: "Fehler" }
    : { area: "Area", record: "Record", status: "Status", action: "Proposed action", issue: "Review",
      choose: "Choose action", row: "Row", ready: "Ready", warning: "Warning", error: "Error" };

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
                  <span className="transfer-review-record">
                    <strong>{record.recordKey || "–"}</strong>
                    {record.rowNumber ? <small>{copy.row} {record.rowNumber}</small> : null}
                  </span>
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
                      <AppSelect
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
                          <option key={resolution} value={resolution}>{resolutionLabel(issue, resolution, language)}</option>
                        ))}
                      </AppSelect>
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

function resolutionLabel(
  issue: DataTransferIssue,
  resolution: DataTransferIssueResolution,
  language: Language
) {
  if (issue.code === "duplicate_inventory_number" && resolution === "copy") {
    return language === "de"
      ? "Als neuen Datensatz mit neuer Inventarnummer importieren"
      : "Import as a new record with a new inventory number";
  }
  if (issue.code === "matching_manufacturer_article_number") {
    if (resolution === "use_existing") {
      return language === "de" ? "Vorhandenes Fahrzeug verwenden" : "Use existing vehicle";
    }
    if (resolution === "create") {
      return language === "de"
        ? "Zusätzlich als neues Fahrzeug importieren"
        : "Import as an additional new vehicle";
    }
  }
  return resolutionLabels[language][resolution];
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
