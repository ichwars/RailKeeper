import { AlertTriangle, Boxes, MapPin, PackageCheck } from "lucide-react";

import type { AccessoryOverviewMetrics } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import type { ArticleOverviewStatusFilter } from "./useArticleOverview";

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
  const cards = [
    {
      key: "articles",
      icon: Boxes,
      label: t("accessories.metrics.articles"),
      value: t("accessories.metrics.articlesValue", {
        articles: metrics.articleCount,
        types: metrics.articleTypeCount
      }),
      active: activeStatus === "",
      action: onReset,
      actionLabel: t("accessories.metrics.filterAll")
    },
    {
      key: "available",
      icon: PackageCheck,
      label: t("accessories.metrics.available"),
      value: t("accessories.metrics.availableValue", {
        available: metrics.available,
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
      value: t("accessories.metrics.allocatedValue", {
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
      value: t("accessories.metrics.careValue", { count: metrics.careHintCount }),
      active: activeStatus === "maintenance_due",
      action: () => onStatusChange("maintenance_due"),
      actionLabel: t("accessories.metrics.filterCare")
    }
  ];

  return (
    <section className="article-metrics" aria-label={t("accessories.metrics.label")}>
      {cards.map(({ key, icon: Icon, label, value, active, action, actionLabel }) => (
        <article key={key} data-testid="article-metric" className={active ? "article-metric active" : "article-metric"}>
          <button type="button" onClick={action} aria-pressed={active} aria-label={actionLabel}>
            <span className="article-metric-icon"><Icon size={17} aria-hidden="true" /></span>
            <span className="article-metric-copy">
              <small>{label}</small>
              <strong>{value}</strong>
            </span>
          </button>
        </article>
      ))}
    </section>
  );
}
