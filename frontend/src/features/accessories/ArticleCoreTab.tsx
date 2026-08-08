import type {
  AccessoryArticle,
  AccessoryArticleType,
  AccessoryManufacturerStatus,
  MasterDataEntry
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppMultiSelect } from "../../shared/ui/AppMultiSelect";
import { AppNumberInput } from "../../shared/ui/AppNumberInput";
import { AppSelect } from "../../shared/ui/AppSelect";
import { AppTextInput } from "../../shared/ui/AppTextInput";
import type { ArticleEditorFieldErrors, ArticleEditorForm } from "./articleEditorModel";
import { articleSubtypeOptions } from "./articleSubtypes";
import { articleTypeOptions } from "./articleTypes";
const statuses: AccessoryManufacturerStatus[] = ["announced", "available", "discontinued", "unknown"];
const gauges = ["Z", "N", "TT", "H0", "H0m", "H0e", "0", "1", "G"];

export function ArticleCoreTab({
  form,
  article,
  errors,
  disabled,
  articleTypeDisabled = false,
  typeDependentDisabled = false,
  otherArticleTypeDisabled = false,
  articleTypeEntries,
  subtypeEntries,
  subtypeEntriesLoading = false,
  subtypeEntriesError = "",
  onChange
}: {
  form: ArticleEditorForm;
  article: AccessoryArticle | null;
  errors: ArticleEditorFieldErrors;
  disabled: boolean;
  articleTypeDisabled?: boolean;
  typeDependentDisabled?: boolean;
  otherArticleTypeDisabled?: boolean;
  articleTypeEntries: MasterDataEntry[];
  subtypeEntries: MasterDataEntry[];
  subtypeEntriesLoading?: boolean;
  subtypeEntriesError?: string;
  onChange: (patch: Partial<ArticleEditorForm>) => void;
}) {
  const { t } = useI18n();
  const typeOptions = articleTypeOptions(articleTypeEntries, article?.articleType || null, t);
  const subtypeOptions = articleSubtypeOptions(form.articleType, form.subtype, subtypeEntries, t);
  return (
    <section className="article-editor-tab article-core-tab" aria-label={t("accessories.editor.tabs.article")}>
      <div className="article-editor-image">
        {article?.primaryImageUrl ? <img src={article.primaryImageUrl} alt={t("accessories.editor.fields.productImage")} />
          : <span>{t("accessories.editor.fields.noProductImage")}</span>}
      </div>
      <div className="article-editor-grid">
        <AppTextInput
          autoFocus
          data-article-initial-focus
          label={t("accessories.field.manufacturer")}
          required
          disabled={disabled}
          value={form.manufacturer}
          error={errors.manufacturer}
          onChange={(event) => onChange({ manufacturer: event.target.value })}
        />
        <AppTextInput label={t("accessories.field.articleNumber")} disabled={disabled}
          value={form.articleNumber} onChange={(event) => onChange({ articleNumber: event.target.value })} />
        <AppTextInput label={t("accessories.field.name")} required disabled={disabled}
          value={form.name} error={errors.name} onChange={(event) => onChange({ name: event.target.value })} />
        <AppTextInput label={t("accessories.editor.fields.ean")} disabled={disabled}
          value={form.ean} onChange={(event) => onChange({ ean: event.target.value })} />
        <label className="app-field">
          <span className="app-field-label">{t("accessories.editor.fields.manufacturerStatus")}</span>
          <AppSelect value={form.manufacturerStatus} disabled={disabled} aria-label={t("accessories.editor.fields.manufacturerStatus")}
            onChange={(event) => onChange({ manufacturerStatus: event.target.value as AccessoryManufacturerStatus })}>
            {statuses.map((status) => <option key={status} value={status}>{t(`accessories.editor.manufacturerStatus.${status}`)}</option>)}
          </AppSelect>
        </label>
        <label className="app-field">
          <span className="app-field-label">{t("accessories.toolbar.articleType")}</span>
          <AppSelect value={form.articleType} disabled={disabled || articleTypeDisabled}
            aria-label={t("accessories.toolbar.articleType")}
            onChange={(event) => onChange({ articleType: event.target.value as AccessoryArticleType })}>
            {typeOptions.map((option) => <option key={option.value} value={option.value}
              disabled={!option.active || (option.value === "other" && otherArticleTypeDisabled)}>
              {option.label}</option>)}
          </AppSelect>
        </label>
        <label className="app-field">
          <span className="app-field-label">{t("accessories.editor.fields.subtype")} *</span>
          <AppSelect value={form.subtype} aria-label={t("accessories.editor.fields.subtype")}
            required aria-invalid={Boolean(errors.subtype)}
            aria-describedby={errors.subtype ? "article-editor-subtype-error" : undefined}
            disabled={disabled || typeDependentDisabled || subtypeEntriesLoading || Boolean(subtypeEntriesError)}
            onChange={(event) => onChange({ subtype: event.target.value })}>
            <option value="">{t("accessories.editor.fields.selectSubtype")}</option>
            {subtypeOptions.map((option) => <option key={option.value} value={option.value}>
              {option.label}{option.active ? "" : ` (${t("accessories.editor.fields.inactiveSubtype")})`}
            </option>)}
          </AppSelect>
          {errors.subtype ? <small id="article-editor-subtype-error" className="app-field-error" role="alert">
            {errors.subtype}
          </small> : null}
        </label>
        <AppMultiSelect label={t("accessories.toolbar.gauge")} disabled={disabled}
          options={gauges.map((gauge) => ({ value: gauge, label: gauge }))} value={form.gauges}
          placeholder={t("accessories.editor.fields.noGauge")}
          onValueChange={(value) => onChange({ gauges: value })} />
        <AppTextInput label={t("accessories.editor.fields.scale")} disabled={disabled}
          value={form.scale} onChange={(event) => onChange({ scale: event.target.value })} />
        <AppNumberInput label={t("accessories.editor.fields.packageQuantity")} required min="1" step="1"
          inputMode="numeric"
          disabled={disabled} value={form.packageQuantity} error={errors.packageQuantity}
          onValueChange={(value) => onChange({ packageQuantity: value })} />
        <AppTextInput label={t("accessories.editor.fields.stockUnit")} required disabled={disabled}
          value={form.stockUnit} error={errors.stockUnit}
          onChange={(event) => onChange({ stockUnit: event.target.value })} />
      </div>
      <label className="app-field article-editor-wide-field">
        <span className="app-field-label">{t("accessories.field.description")}</span>
        <textarea disabled={disabled} value={form.description}
          onChange={(event) => onChange({ description: event.target.value })} />
      </label>
      <details className="article-editor-more">
        <summary>{t("accessories.editor.more.title")}</summary>
        <div className="article-editor-grid">
          <AppTextInput label={t("accessories.editor.fields.manufacturerUrl")} disabled={disabled}
            value={form.manufacturerUrl} onChange={(event) => onChange({ manufacturerUrl: event.target.value })} />
          <AppTextInput label={t("accessories.editor.fields.productUrl")} disabled={disabled}
            value={form.productUrl} onChange={(event) => onChange({ productUrl: event.target.value })} />
          <label className="app-field">
            <span className="app-field-label">{t("accessories.editor.fields.alternativeNumbers")}</span>
            <textarea disabled={disabled} value={form.alternativeNumbers}
              onChange={(event) => onChange({ alternativeNumbers: event.target.value })} />
          </label>
          <AppTextInput label={t("accessories.editor.fields.keywords")} disabled={disabled}
            value={form.keywords} onChange={(event) => onChange({ keywords: event.target.value })} />
          <label className="app-field">
            <span className="app-field-label">{t("accessories.editor.fields.compatibility")}</span>
            <textarea disabled={disabled} value={form.compatibilityNotes}
              onChange={(event) => onChange({ compatibilityNotes: event.target.value })} />
          </label>
          <label className="app-field">
            <span className="app-field-label">{t("accessories.editor.fields.internalNotes")}</span>
            <textarea disabled={disabled} value={form.internalNotes}
              onChange={(event) => onChange({ internalNotes: event.target.value })} />
          </label>
        </div>
      </details>
    </section>
  );
}
