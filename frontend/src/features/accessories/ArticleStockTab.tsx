import type { AccessoryArticle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AccessoryStockPanel } from "./AccessoryStockPanel";
import type { ArticleEditorFieldErrors, ArticleEditorForm } from "./articleEditorModel";
import type { ArticleEditorResources } from "./useArticleEditorController";

export function ArticleStockTab({ article, form, errors, resources, disabled, onChange, onChanged }: {
  article: AccessoryArticle | null;
  form: ArticleEditorForm;
  errors: ArticleEditorFieldErrors;
  resources: ArticleEditorResources;
  disabled: boolean;
  onChange: (patch: Partial<ArticleEditorForm>) => void;
  onChanged: () => Promise<void>;
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
      <AppNumberInput label={t("accessories.editor.fields.minimumStock")} min="0" step="any" disabled={disabled}
        value={form.minimumStock} error={errors.minimumStock}
        onValueChange={(value) => onChange({ minimumStock: value })} />
    </div>
    {article ? <AccessoryStockPanel article={article} stock={resources.stock} movements={resources.movements}
      assets={resources.assets} locations={resources.locations} canEdit={!disabled} onChanged={onChanged} />
      : <p className="article-editor-hint">{t("accessories.editor.saveBeforeStock")}</p>}
  </section>;
}
