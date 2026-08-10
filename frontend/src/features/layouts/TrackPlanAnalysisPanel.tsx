import { AlertTriangle, CheckCircle2, PackageSearch } from "lucide-react";

import type { TrackPlanAnalysis, TrackPlanIssueCode } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

const issueSymbols: Record<TrackPlanIssueCode, string> = {
  open_end: "○",
  incompatible_connection: "↯",
  overlap: "!",
  broken_geometry: "×",
  elevation_mismatch: "↕",
  grade_limit_exceeded: "↗",
  insufficient_clearance: "↕",
  flex_radius_below_limit: "⌒"
};

function countIssues(analysis: TrackPlanAnalysis, code: TrackPlanIssueCode): number {
  return analysis.issues.filter((issue) => issue.code === code).length;
}

export function TrackPlanAnalysisPanel({ analysis, selectedObjectId, onSelectObject }: {
  analysis: TrackPlanAnalysis;
  selectedObjectId?: string;
  onSelectObject?: (objectID: string) => void;
}) {
  const { language, t } = useI18n();
  const connections = analysis.connections.length;
  const openEnds = countIssues(analysis, "open_end");
  const overlaps = countIssues(analysis, "overlap");
  const elevationMismatches = countIssues(analysis, "elevation_mismatch");
  const gradeLimitExceedances = countIssues(analysis, "grade_limit_exceeded");
  const clearanceViolations = countIssues(analysis, "insufficient_clearance");
  const flexRadiusWarnings = countIssues(analysis, "flex_radius_below_limit");
  const numberFormat = new Intl.NumberFormat(language === "de" ? "de-DE" : "en-GB", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });

  return <section className="track-plan-analysis" aria-label={t("layouts.trackAnalysis.title")}>
    <header>
      <div><CheckCircle2 size={16} /><h4>{t("layouts.trackAnalysis.title")}</h4></div>
      <div className="track-analysis-summary" aria-label={t("layouts.trackAnalysis.summary")}>
        <span>{t(connections === 1 ? "layouts.trackAnalysis.connectionOne"
          : "layouts.trackAnalysis.connectionMany", { count: connections })}</span>
        <span>{t(openEnds === 1 ? "layouts.trackAnalysis.openEndOne"
          : "layouts.trackAnalysis.openEndMany", { count: openEnds })}</span>
        <span>{t(overlaps === 1 ? "layouts.trackAnalysis.overlapOne"
          : "layouts.trackAnalysis.overlapMany", { count: overlaps })}</span>
        <span>{t(elevationMismatches === 1 ? "layouts.trackAnalysis.elevationMismatchOne"
          : "layouts.trackAnalysis.elevationMismatchMany", { count: elevationMismatches })}</span>
        <span>{t(gradeLimitExceedances === 1 ? "layouts.trackAnalysis.gradeLimitExceededOne"
          : "layouts.trackAnalysis.gradeLimitExceededMany", { count: gradeLimitExceedances })}</span>
        <span>{t(clearanceViolations === 1 ? "layouts.trackAnalysis.clearanceViolationOne"
          : "layouts.trackAnalysis.clearanceViolationMany", { count: clearanceViolations })}</span>
        <span>{t(flexRadiusWarnings === 1 ? "layouts.trackAnalysis.flexRadiusWarningOne"
          : "layouts.trackAnalysis.flexRadiusWarningMany", { count: flexRadiusWarnings })}</span>
      </div>
    </header>
    <div className="track-analysis-grid">
      <div>
        <h5><AlertTriangle size={15} />{t("layouts.trackAnalysis.issues")}</h5>
        {analysis.issues.length === 0
          ? <p className="layout-empty">✓ {t("layouts.trackAnalysis.noIssues")}</p>
          : <ul className="track-issue-list">{analysis.issues.map((issue, index) => {
            let issueText = t(`layouts.trackAnalysis.issue.${issue.code}`);
            if (issue.code === "elevation_mismatch" && issue.elevationDifferenceMm !== undefined) {
              issueText = t("layouts.trackAnalysis.issueElevationDetail", {
                difference: numberFormat.format(issue.elevationDifferenceMm)
              });
            } else if (issue.code === "grade_limit_exceeded" && issue.gradePercent !== undefined &&
              issue.gradeLimitPercent !== undefined) {
              issueText = t("layouts.trackAnalysis.issueGradeLimitDetail", {
                grade: numberFormat.format(issue.gradePercent),
                limit: numberFormat.format(issue.gradeLimitPercent)
              });
            } else if (issue.code === "insufficient_clearance" && issue.clearanceMm !== undefined &&
              issue.clearanceLimitMm !== undefined) {
              issueText = t("layouts.trackAnalysis.issueClearanceDetail", {
                clearance: numberFormat.format(issue.clearanceMm),
                limit: numberFormat.format(issue.clearanceLimitMm)
              });
            } else if (issue.code === "flex_radius_below_limit" && issue.radiusMm !== undefined &&
              issue.radiusLimitMm !== undefined) {
              issueText = t("layouts.trackAnalysis.issueFlexRadiusDetail", {
                radius: numberFormat.format(issue.radiusMm),
                limit: numberFormat.format(issue.radiusLimitMm)
              });
            }
            const label = `${t(`layouts.trackAnalysis.severity.${issue.severity}`)}: ${issueText}`;
            return <li key={`${issue.code}-${issue.objectIds.join("-")}-${index}`}>
              <button type="button" className={`track-issue severity-${issue.severity}`}
                aria-label={label} aria-pressed={issue.objectIds.includes(selectedObjectId ?? "")}
                onClick={() => issue.objectIds[0] && onSelectObject?.(issue.objectIds[0])}>
                <b aria-hidden="true">{issueSymbols[issue.code]}</b>
                <span>{issueText}</span>
              </button>
            </li>;
          })}</ul>}
      </div>
      <div className="track-materials">
        <h5><PackageSearch size={15} />{t("layouts.trackAnalysis.materials")}</h5>
        <div className="table-wrap"><table>
          <thead><tr>
            <th>{t("layouts.trackAnalysis.article")}</th>
            <th>{t("layouts.trackAnalysis.required")}</th>
            <th>{t("layouts.trackAnalysis.physical")}</th>
            <th>{t("layouts.trackAnalysis.reserved")}</th>
            <th>{t("layouts.trackAnalysis.available")}</th>
            <th>{t("layouts.trackAnalysis.missing")}</th>
            <th>{t("layouts.trackAnalysis.inventory")}</th>
          </tr></thead>
          <tbody>{analysis.materials.map((material) => <tr key={material.geometryId}>
            <td><strong>{material.manufacturer} {material.articleNumber}</strong><small>{material.name}</small></td>
            <td>{material.requiredQuantity}</td>
            <td>{material.physicalQuantity}</td>
            <td>{material.reservedQuantity}</td>
            <td>{material.availableQuantity}</td>
            <td className={material.missingQuantity > 0 ? "material-missing" : ""}>
              <span aria-hidden="true">{material.missingQuantity > 0 ? "!" : "✓"}</span>{" "}
              <span>{material.missingQuantity}</span>
            </td>
            <td>{material.inventoryNumbers.length > 0 ? material.inventoryNumbers.join(", ")
              : t("layouts.trackAnalysis.notLinked")}</td>
          </tr>)}</tbody>
        </table></div>
      </div>
    </div>
  </section>;
}
