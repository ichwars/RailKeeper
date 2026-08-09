import type { AccessoryArticle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AccessoryStockPanel } from "./AccessoryStockPanel";
import { AccessoryReservationsPanel } from "./AccessoryReservationsPanel";
import { AccessoryInstallationsPanel } from "./AccessoryInstallationsPanel";
import type { ArticleEditorFieldErrors, ArticleEditorForm } from "./articleEditorModel";
import type { ArticleEditorResources } from "./useArticleEditorController";

export function ArticleStockTab({ article, form, errors, resources, disabled, canReserve, canInstall, onChange,
  onChanged, onDirtyChange }: {
  article: AccessoryArticle | null;
  form: ArticleEditorForm;
  errors: ArticleEditorFieldErrors;
  resources: ArticleEditorResources;
  disabled: boolean;
  canReserve: boolean;
  canInstall: boolean;
  onChange: (patch: Partial<ArticleEditorForm>) => void;
  onChanged: () => Promise<void>;
  onDirtyChange: (scope: string, dirty: boolean) => void;
}) {
  const { t } = useI18n();
  return <section className="article-editor-tab" aria-label={t("accessories.editor.tabs.stock")}>
    <div className="article-stock-settings">
      <label className="app-field">
        <span className="app-field-label">{t("accessories.editor.fields.inventoryStrategy")}</span>
        <AppSelect value={form.inventoryStrategy} disabled={disabled} aria-label={t("accessories.editor.fields.inventoryStrategy")}
          onChange={(event) => onChange({ inventoryStrategy: event.target.value as ArticleEditorForm["inventoryStrategy"] })}>
          <option value="quantity">{t("accessories.editor.strategy.quantity")}</option>
          <option value="individual">{t("accessories.editor.strategy.individual")}</option>
          <option value="quantity_later_individual">{t("accessories.editor.strategy.hybrid")}</option>
        </AppSelect>
      </label>
      <AppNumberInput label={t("accessories.editor.fields.minimumStock")} min="0" step="1" inputMode="numeric"
        disabled={disabled}
        value={form.minimumStock} error={errors.minimumStock}
        onValueChange={(value) => onChange({ minimumStock: value })} />
    </div>
    {article ? <AccessoryStockPanel article={article} stock={resources.stock} movements={resources.movements}
      assets={resources.assets} locations={resources.locations} canEdit={!disabled} onChanged={onChanged}
      onDirtyChange={(dirty) => onDirtyChange("stockCommands", dirty)} />
      : <p className="article-editor-hint">{t("accessories.editor.saveBeforeStock")}</p>}
    {article ? <>
      <AccessoryReservationsPanel article={article} reservations={resources.reservations} assets={resources.assets}
        locations={resources.locations} vehicles={resources.vehicles} layouts={resources.layouts} units={resources.units}
        canReserve={canReserve} onChanged={onChanged}
        onDirtyChange={(dirty) => onDirtyChange("reservation", dirty)} />
      <AccessoryInstallationsPanel article={article} reservations={resources.reservations}
        installations={resources.installations} assets={resources.assets} locations={resources.locations}
        vehicles={resources.vehicles} layouts={resources.layouts} units={resources.units} canInstall={canInstall}
        onChanged={onChanged} onDirtyChange={(dirty) => onDirtyChange("installation", dirty)} />
    </> : null}
  </section>;
}
