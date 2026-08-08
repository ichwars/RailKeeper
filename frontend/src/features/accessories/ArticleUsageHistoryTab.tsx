import type { AccessoryArticle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AccessoryInstallationsPanel } from "./AccessoryInstallationsPanel";
import { AccessoryReservationsPanel } from "./AccessoryReservationsPanel";
import type { ArticleEditorResources, ArticleEditorPermissions } from "./useArticleEditorController";

export function ArticleUsageHistoryTab({ article, resources, permissions, disabled, onChanged }: {
  article: AccessoryArticle;
  resources: ArticleEditorResources;
  permissions: ArticleEditorPermissions;
  disabled: boolean;
  onChanged: () => Promise<void>;
}) {
  const { t } = useI18n();
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
        canReserve={!disabled && permissions.canReserve} onChanged={onChanged} />
      <AccessoryInstallationsPanel article={article} reservations={currentReservations}
        installations={currentInstallations} assets={resources.assets} locations={resources.locations}
        vehicles={resources.vehicles} layouts={resources.layouts} units={resources.units}
        canInstall={!disabled && permissions.canInstall} onChanged={onChanged} />
    </section>
    <section className="article-editor-section">
      <h3>{t("accessories.editor.usage.history")}</h3>
      <div className="table-wrap"><table><thead><tr>
        <th>{t("accessories.editor.stock.date")}</th><th>{t("accessories.editor.usage.event")}</th>
        <th>{t("accessories.field.quantity")}</th><th>{t("accessories.field.target")}</th>
      </tr></thead><tbody>{history.map((event) => <tr key={event.id}>
        <td>{new Date(event.occurredAt).toLocaleString()}</td>
        <td>{t(`accessories.editor.usageEvent.${event.type}`)}</td><td>{event.quantity}</td>
        <td>{event.vehicleId || event.layoutUnitId || event.layoutId || "-"}</td>
      </tr>)}</tbody></table></div>
    </section>
  </section>;
}
