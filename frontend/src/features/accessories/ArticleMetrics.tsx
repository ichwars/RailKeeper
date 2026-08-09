import { AlertTriangle, Boxes, MapPin, PackageCheck } from "lucide-react";

import type { AccessoryOverviewMetrics } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { ArticleOverviewStatusFilter } from "./useArticleOverview";

type MetricCard = {
  key: string;
  icon: typeof Boxes;
  label: string;
  summary: string;
  detail: string;
  active: boolean;
  action?: () => void;
  actionLabel?: string;
};

export function ArticleMetrics({
  metrics,
  activeStatus,
  onReset,
  onStatusChange
}: {
  metrics: AccessoryOverviewMetrics;
  activeStatus: ArticleOverviewStatusFilter;
  onReset: () => void;
  onStatusChange: (status: ArticleOverviewStatusFilter) => void;
}) {
  const { t } = useI18n();
  const cards: MetricCard[] = [
    {
      key: "articles",
      icon: Boxes,
      label: t("accessories.metrics.articles"),
      summary: t("accessories.metrics.articlesSummary", {
        articles: metrics.articleCount,
        articleNoun: t(metrics.articleCount === 1
          ? "accessories.metrics.articleSingular" : "accessories.metrics.articlePlural")
      }),
      detail: t("accessories.metrics.articlesDetail", {
        types: metrics.articleTypeCount,
        typeNoun: t(metrics.articleTypeCount === 1
          ? "accessories.metrics.typeSingular" : "accessories.metrics.typePlural")
      }),
      active: activeStatus === "",
      action: onReset,
      actionLabel: t("accessories.metrics.filterAll")
    },
    {
      key: "available",
      icon: PackageCheck,
      label: t("accessories.metrics.available"),
      summary: t("accessories.metrics.availableSummary", { available: metrics.available }),
      detail: t("accessories.metrics.availableDetail", {
        locations: metrics.locationCount
      }),
      active: activeStatus === "available",
      action: () => onStatusChange("available"),
      actionLabel: t("accessories.metrics.filterAvailable")
    },
    {
      key: "allocated",
      icon: MapPin,
      label: t("accessories.metrics.allocated"),
      summary: t("accessories.metrics.allocatedSummary", {
        count: metrics.reserved + metrics.installed
      }),
      detail: t("accessories.metrics.allocatedDetail", {
        reserved: metrics.reserved,
        installed: metrics.installed
      }),
      active: activeStatus === "allocated",
      action: () => onStatusChange("allocated"),
      actionLabel: t("accessories.metrics.filterAllocated")
    },
    {
      key: "care",
      icon: AlertTriangle,
      label: t("accessories.metrics.care"),
      summary: String(metrics.careHintCount),
      detail: t("accessories.metrics.careDetail"),
      active: false
    }
  ];

  return (
    <section className="inventory-status-row article-metrics" aria-label={t("accessories.metrics.label")}>
      {cards.map(({ key, icon: Icon, label, summary, detail, active, action, actionLabel }) => (
        <article key={key} data-testid="article-metric"
          className={[
            "inventory-status-card",
            "article-metric",
            key === "allocated" ? "wide" : "",
            active ? "active" : ""
          ].filter(Boolean).join(" ")}>
          {action ? <button type="button" onClick={action} aria-pressed={active} aria-label={actionLabel}>
            <span><Icon size={16} aria-hidden="true" /></span>
            <small>{label}</small>
            <strong>{summary}</strong>
            <em>{detail}</em>
          </button> : <>
            <span><Icon size={16} aria-hidden="true" /></span>
            <small>{label}</small>
            <strong>{summary}</strong>
            <em>{detail}</em>
          </>}
        </article>
      ))}
    </section>
  );
}
