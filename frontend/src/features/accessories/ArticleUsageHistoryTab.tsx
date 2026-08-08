import type { AccessoryArticle } from "../../shared/api";
import { formatDateTime, useI18n } from "../../shared/i18n";
import { AccessoryInstallationsPanel } from "./AccessoryInstallationsPanel";
import { AccessoryReservationsPanel } from "./AccessoryReservationsPanel";
import { accessoryTargetLabel } from "./AccessoryTargetFields";
import type { ArticleEditorResources } from "./useArticleEditorController";

export function ArticleUsageHistoryTab({ article, resources }: {
  article: AccessoryArticle;
  resources: ArticleEditorResources;
}) {
  const { t, language } = useI18n();
  const currentReservations = resources.reservations.filter((reservation) => reservation.status === "active");
  const currentInstallations = resources.installations.filter((installation) => !installation.removedAt);
  const history = [...(resources.usageHistory?.events || [])]
    .sort((left, right) => left.occurredAt.localeCompare(right.occurredAt));

  return <section className="article-editor-tab article-usage-tab"
    aria-label={t("accessories.editor.tabs.usageHistory")}>
    <section className="article-editor-section">
      <h3>{t("accessories.editor.usage.current")}</h3>
      <AccessoryReservationsPanel article={article} reservations={currentReservations} assets={resources.assets}
        locations={resources.locations} vehicles={resources.vehicles} layouts={resources.layouts} units={resources.units}
        canReserve={false} onChanged={async () => undefined} onDirtyChange={() => undefined} />
      <AccessoryInstallationsPanel article={article} reservations={currentReservations}
        installations={currentInstallations} assets={resources.assets} locations={resources.locations}
        vehicles={resources.vehicles} layouts={resources.layouts} units={resources.units}
        canInstall={false} onChanged={async () => undefined} onDirtyChange={() => undefined} />
    </section>
    <section className="article-editor-section">
      <h3>{t("accessories.editor.usage.history")}</h3>
      <div className="table-wrap"><table><thead><tr>
        <th>{t("accessories.editor.stock.date")}</th><th>{t("accessories.editor.usage.event")}</th>
        <th>{t("accessories.field.quantity")}</th><th>{t("accessories.field.target")}</th>
        <th>{t("accessories.field.condition")}</th>
      </tr></thead><tbody>{history.map((event) => <tr key={event.id}>
        <td>{formatDateTime(event.occurredAt, language)}</td>
        <td>{t(`accessories.editor.usageEvent.${event.type}`)}</td><td>{event.quantity}</td>
        <td>{accessoryTargetLabel(event, resources.vehicles, resources.layouts, resources.units) || "-"}</td>
        <td>{event.type === "removal" && event.removalDisposition
          ? t(`accessories.disposition.${event.removalDisposition}`)
          : "condition" in event && event.condition ? t(`accessories.condition.${event.condition}`) : "-"}</td>
      </tr>)}</tbody></table></div>
    </section>
  </section>;
}
