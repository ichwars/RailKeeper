import { AlertTriangle, CheckCircle2, PackageSearch } from "lucide-react";

import type { TrackPlanAnalysis, TrackPlanIssueCode } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

const issueSymbols: Record<TrackPlanIssueCode, string> = {
  open_end: "○",
  incompatible_connection: "↯",
  overlap: "!",
  broken_geometry: "×"
};

function countIssues(analysis: TrackPlanAnalysis, code: TrackPlanIssueCode): number {
  return analysis.issues.filter((issue) => issue.code === code).length;
}

export function TrackPlanAnalysisPanel({ analysis, selectedObjectId, onSelectObject }: {
  analysis: TrackPlanAnalysis;
  selectedObjectId?: string;
  onSelectObject?: (objectID: string) => void;
}) {
  const { t } = useI18n();
  const connections = analysis.connections.length;
  const openEnds = countIssues(analysis, "open_end");
  const overlaps = countIssues(analysis, "overlap");

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
      </div>
    </header>
    <div className="track-analysis-grid">
      <div>
        <h5><AlertTriangle size={15} />{t("layouts.trackAnalysis.issues")}</h5>
        {analysis.issues.length === 0
          ? <p className="layout-empty">✓ {t("layouts.trackAnalysis.noIssues")}</p>
          : <ul className="track-issue-list">{analysis.issues.map((issue, index) => {
            const label = `${t(`layouts.trackAnalysis.severity.${issue.severity}`)}: ${t(
              `layouts.trackAnalysis.issue.${issue.code}`
            )}`;
            return <li key={`${issue.code}-${issue.objectIds.join("-")}-${index}`}>
              <button type="button" className={`track-issue severity-${issue.severity}`}
                aria-label={label} aria-pressed={issue.objectIds.includes(selectedObjectId ?? "")}
                onClick={() => issue.objectIds[0] && onSelectObject?.(issue.objectIds[0])}>
                <b aria-hidden="true">{issueSymbols[issue.code]}</b>
                <span>{t(`layouts.trackAnalysis.issue.${issue.code}`)}</span>
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
