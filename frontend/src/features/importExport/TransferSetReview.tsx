import { AlertTriangle, CheckCircle2, CircleAlert, Layers3 } from "lucide-react";

import type { Language } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import type {
  DataTransferIssue,
  DataTransferIssueResolution,
  DataTransferVehicleSetPreview
} from "./dataTransferModel";

type TransferSetReviewProps = {
  busy: boolean;
  issues: DataTransferIssue[];
  language: Language;
  onResolve: (issue: DataTransferIssue, resolution: DataTransferIssueResolution) => Promise<void>;
  sets: DataTransferVehicleSetPreview[];
};

type SetResolution = Extract<DataTransferIssueResolution, "replace" | "copy" | "skip">;

const setResolutions = {
  duplicate_vehicle_set_inventory_number: ["replace", "copy", "skip"],
  vehicle_set_member_external_conflict: ["copy", "skip"]
} satisfies Record<string, SetResolution[]>;

export function TransferSetReview({ busy, issues, language, onResolve, sets }: TransferSetReviewProps) {
  if (sets.length === 0) return null;
  const copy = setReviewCopy(language);

  return (
    <section className="transfer-set-review" aria-labelledby="transfer-set-review-title">
      <header>
        <Layers3 size={18} aria-hidden="true" />
        <h4 id="transfer-set-review-title">{copy.title}</h4>
      </header>
      <div className="transfer-set-review-list">
        {sets.map((set) => {
          const setIssues = issues.filter((item) => item.recordKey === set.recordKey);
          const issue = setIssues.find(isSetConflictIssue);
          const Icon = set.classification === "error"
            ? CircleAlert
            : set.classification === "warning" ? AlertTriangle : CheckCircle2;
          return (
            <article className={`transfer-set-review-card ${set.classification}`} key={set.recordKey}>
              <div className="transfer-set-review-main">
                <strong>{set.recordKey}</strong>
                <span>{set.data.name || copy.unnamed}</span>
                <small>{memberCount(set.memberRecordKeys.length, language)}</small>
              </div>
              <span className={`transfer-review-status ${set.classification}`}>
                <Icon size={15} aria-hidden="true" />
                {copy[set.classification]}
              </span>
              {issue ? (
                <label className="transfer-set-review-resolution">
                  <span>{issue.message}</span>
                  <AppSelect
                    aria-label={`${copy.resolutionFor} ${set.recordKey}`}
                    disabled={busy}
                    value={issue.selectedResolution}
                    onChange={(event) => {
                      const resolution = event.target.value as DataTransferIssueResolution;
                      if (resolution) void onResolve(issue, resolution);
                    }}
                  >
                    <option value="">{copy.choose}</option>
                    {resolutionsForSetIssue(issue).map((resolution) => (
                      <option key={resolution} value={resolution}>{copy.resolutions[resolution]}</option>
                    ))}
                  </AppSelect>
                </label>
              ) : setIssues.length || set.diagnostics?.length ? (
                <span className="transfer-set-review-diagnostic">
                  {setIssues.length
                    ? setIssues.map((item) => item.message).join(" ")
                    : set.diagnostics?.map((diagnostic) => diagnostic.code).join(", ")}
                </span>
              ) : null}
            </article>
          );
        })}
      </div>
    </section>
  );
}

function isSetConflictIssue(issue: DataTransferIssue) {
  return issue.code === "duplicate_vehicle_set_inventory_number" ||
    issue.code === "vehicle_set_member_external_conflict";
}

function resolutionsForSetIssue(issue: DataTransferIssue): SetResolution[] {
  if (issue.code === "duplicate_vehicle_set_inventory_number") {
    return setResolutions.duplicate_vehicle_set_inventory_number;
  }
  return setResolutions.vehicle_set_member_external_conflict;
}

function memberCount(count: number, language: Language) {
  if (language === "de") return `${count} ${count === 1 ? "Mitglied" : "Mitglieder"}`;
  return `${count} ${count === 1 ? "member" : "members"}`;
}

function setReviewCopy(language: Language) {
  return language === "de" ? {
    title: "Erkannte Fahrzeugsets",
    unnamed: "Unbenanntes Set",
    ready: "Bereit",
    warning: "Hinweis",
    error: "Fehler",
    choose: "Aktion wählen",
    resolutionFor: "Auflösung für Set",
    resolutions: {
      replace: "Bestehendes Set aktualisieren",
      copy: "Als neues Set importieren",
      skip: "Dieses Set nicht importieren"
    }
  } : {
    title: "Detected vehicle sets",
    unnamed: "Unnamed set",
    ready: "Ready",
    warning: "Warning",
    error: "Error",
    choose: "Choose action",
    resolutionFor: "Resolution for set",
    resolutions: {
      replace: "Update existing set",
      copy: "Import as a new set",
      skip: "Do not import this set"
    }
  };
}
