import { Clock3, TriangleAlert } from "lucide-react";
import { useEffect, useState } from "react";

import { api, type AccessoryUsageEvent } from "../../shared/api";
import { formatDateTime, useI18n } from "../../shared/i18n";

export function LayoutTwinHistory({ positionID, productID }: {
  positionID: string;
  productID?: string;
}) {
  const { t, language } = useI18n();
  const [events, setEvents] = useState<AccessoryUsageEvent[]>([]);
  const [loading, setLoading] = useState(Boolean(productID));
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let active = true;
    setEvents([]);
    setFailed(false);
    setLoading(Boolean(productID));
    if (!productID) return () => { active = false; };
    api.accessoryUsageHistory(productID)
      .then((history) => {
        if (!active) return;
        setEvents(history.events
          .filter((event) => event.technicalPositionId === positionID)
          .sort((left, right) => right.occurredAt.localeCompare(left.occurredAt)));
      })
      .catch(() => {
        if (active) setFailed(true);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => { active = false; };
  }, [positionID, productID]);

  if (!productID) return <p className="layout-empty">{t("layouts.twin.history.noArticle")}</p>;
  if (loading) return <p className="layout-empty" role="status">{t("layouts.twin.history.loading")}</p>;
  if (failed) return <p className="layout-twin-history-error" role="alert"><TriangleAlert size={15} />
    {t("layouts.twin.history.error")}</p>;
  if (!events.length) return <p className="layout-empty">{t("layouts.twin.history.empty")}</p>;

  return <ol className="layout-twin-history-list">{events.map((event) => <li key={event.id}>
    <Clock3 size={14} />
    <div><strong>{t(`accessories.editor.usageEvent.${event.type}`)}</strong>
      <span>{historyDetail(event, t)}</span>
      <small>{formatDateTime(event.occurredAt, language)}</small></div>
  </li>)}</ol>;
}

function historyDetail(event: AccessoryUsageEvent, t: (key: string, values?: Record<string, string | number>) => string) {
  const quantity = t("layouts.twin.history.quantity", { quantity: event.quantity });
  if (event.type === "condition_changed" && event.condition) {
    const previous = event.previousCondition ? t(`accessories.condition.${event.previousCondition}`) : "–";
    return `${quantity} · ${previous} → ${t(`accessories.condition.${event.condition}`)}`;
  }
  if (event.type === "installation" && event.condition) {
    return `${quantity} · ${t(`accessories.condition.${event.condition}`)}`;
  }
  if (event.type === "removal" && event.removalDisposition) {
    return `${quantity} · ${t(`accessories.disposition.${event.removalDisposition}`)}`;
  }
  return quantity;
}
