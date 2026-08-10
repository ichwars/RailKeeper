import { GitCompareArrows } from "lucide-react";

import type { TrackPlanChangePreview } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function TrackPlanChangePreviewPanel({ preview }: { preview: TrackPlanChangePreview }) {
  const { t } = useI18n();
  const changeCount = preview.objectChanges.length;
  const freeChangeCount = preview.freeObjectChanges.length;
  const addedIssues = preview.issues.added.length;
  const resolvedIssues = preview.issues.resolved.length;

  return <section className="track-change-preview" aria-label={t("layouts.trackChanges.title")}>
    <header><GitCompareArrows size={16} /><h4>{t("layouts.trackChanges.title")}</h4>
      <div className="track-change-counts">
        <span>{t(changeCount === 1 ? "layouts.trackChanges.trackOne"
          : "layouts.trackChanges.trackMany", { count: changeCount })}</span>
        <span>{t(freeChangeCount === 1 ? "layouts.trackChanges.freeOne"
          : "layouts.trackChanges.freeMany", { count: freeChangeCount })}</span>
      </div></header>
    <div className="track-change-grid">
      <div><h5>{t("layouts.trackChanges.materials")}</h5>
        {preview.materialDeltas.length === 0 ? <p>✓ {t("layouts.trackChanges.noMaterialChanges")}</p>
          : <ul>{preview.materialDeltas.map((delta) => <li key={delta.geometryId}>
            <strong>{delta.delta > 0 ? "+" : ""}{delta.delta} Tillig {delta.articleNumber}</strong>
            <span>{delta.baseQuantity} → {delta.currentQuantity}</span>
          </li>)}</ul>}
      </div>
      <div><h5>{t("layouts.trackChanges.issues")}</h5>
        <p>{t(addedIssues === 1 ? "layouts.trackChanges.issueAddedOne"
          : "layouts.trackChanges.issueAddedMany", { count: addedIssues })}</p>
        <p>{t("layouts.trackChanges.issueResolved", { count: resolvedIssues })}</p>
      </div>
      <div><h5>{t("layouts.trackChanges.configurations")}</h5>
        {preview.affectedConfigurations.length === 0
          ? <p>{t("layouts.trackChanges.noConfigurations")}</p>
          : <ul>{preview.affectedConfigurations.map((configuration) =>
            <li key={configuration.id}>{configuration.name}</li>)}</ul>}
      </div>
    </div>
  </section>;
}
