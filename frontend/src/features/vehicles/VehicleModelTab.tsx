import { Barcode, ChevronDown, ChevronUp, ExternalLink, PackageSearch } from "lucide-react";
import type { ReactNode } from "react";
import { CreateVehicleRequest, MasterDataEntry, MasterDataRelation } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { sourceDisplayName } from "./articleSearch";
import { RequiredLabel, VehicleDetailsFields, VehicleOwnershipFields } from "./VehicleFormFields";

type ECoSRequiredField = "manufacturer" | "name" | "gauge" | "category" | "gattung";

type MasterDataOptions = {
  manufacturers: MasterDataEntry[];
  gauges: MasterDataEntry[];
  epochs: MasterDataEntry[];
  railwayCompanies: MasterDataEntry[];
  categories: MasterDataEntry[];
  gattungen: MasterDataEntry[];
  symbols: MasterDataEntry[];
  categoryRelations: MasterDataRelation[];
};

type OpenSections = {
  model: boolean;
  details: boolean;
  vehicle: boolean;
};

function compactValue(value: unknown) {
  return String(value ?? "").trim();
}

export function VehicleModelTab({
  form,
  readonly,
  articleSearchLoading,
  canRunArticleSearch,
  articleIdentityFilled,
  options,
  filteredGattungen,
  openSections,
  selectOptions,
  ecosFieldClass,
  onToggleSection,
  onOpenBarcodeSearch,
  onRunArticleSearch,
  onUpdate,
  onUpdateCategory,
  onOpenQr,
  canOpenQr,
  onUpdateCouplingFront,
  onUpdateCouplingSame
}: {
  form: CreateVehicleRequest;
  readonly: boolean;
  articleSearchLoading: boolean;
  canRunArticleSearch: boolean;
  articleIdentityFilled: boolean;
  options: MasterDataOptions;
  filteredGattungen: MasterDataEntry[];
  openSections: OpenSections;
  selectOptions: (entries: MasterDataEntry[], emptyLabel?: string) => ReactNode;
  ecosFieldClass: (field: ECoSRequiredField) => string;
  onToggleSection: (section: keyof OpenSections) => void;
  onOpenBarcodeSearch: () => void;
  onRunArticleSearch: () => void;
  onUpdate: (patch: Partial<CreateVehicleRequest>) => void;
  onUpdateCategory: (category: string) => void;
  onOpenQr: () => void;
  canOpenQr: boolean;
  onUpdateCouplingFront: (couplingFront: string) => void;
  onUpdateCouplingSame: (couplingSame: boolean) => void;
}) {
  const { t } = useI18n();

  return (
    <div className="accordion-stack">
      <section className="accordion-section">
        <button type="button" className="accordion-trigger" onClick={() => onToggleSection("model")}>
          {openSections.model ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
          {t("vehicles.tab.model")}
        </button>
        {openSections.model && (
          <div className="accordion-content vehicle-form">
            <div className="article-search-box">
              <div>
                <strong>{t("vehicles.articleSearch.title")}</strong>
                <span>{t("vehicles.articleSearch.subtitle")}</span>
              </div>
              <div className="article-search-actions">
                <button type="button" className="secondary-button" onClick={onOpenBarcodeSearch} disabled={readonly || articleSearchLoading} title={t("vehicles.articleSearch.barcodeTitle")}>
                  <Barcode size={15} aria-hidden="true" />
                  {t("vehicles.articleSearch.barcode")}
                </button>
                <button
                  type="button"
                  className="secondary-button"
                  onClick={onRunArticleSearch}
                  disabled={readonly || articleSearchLoading || !canRunArticleSearch}
                  title={!canRunArticleSearch ? t("vehicles.articleSearch.missingInput") : undefined}
                >
                  <PackageSearch size={15} aria-hidden="true" />
                  {articleSearchLoading ? t("vehicles.articleSearch.searching") : t("vehicles.articleSearch.search")}
                </button>
              </div>
            </div>

            {form.articleSourceUrl && (
              <p className="source-note compact-source-note">
                <ExternalLink size={15} aria-hidden="true" />
                <span>
                  {t("vehicles.source")}: <a href={form.articleSourceUrl} target="_blank" rel="noreferrer">{sourceDisplayName(form.articleSourceUrl)}</a>
                </span>
              </p>
            )}

            <div className="form-row">
              <label>
                {t("vehicle.field.inventoryNumber")}
                <input value={form.inventoryNumber || ""} onChange={(event) => onUpdate({ inventoryNumber: event.target.value })} disabled={readonly} placeholder={t("vehicles.inventoryNumberAuto")} />
              </label>
              <label>
                <RequiredLabel label={t("vehicle.field.articleNumber")} filled={articleIdentityFilled} />
                <input value={form.articleNumber || ""} onChange={(event) => onUpdate({ articleNumber: event.target.value })} disabled={readonly} />
              </label>
            </div>

            <div className="form-row">
              <label className={ecosFieldClass("manufacturer")}>
                <RequiredLabel label={t("vehicle.field.manufacturer")} filled={Boolean(compactValue(form.manufacturer))} />
                <select value={form.manufacturer} onChange={(event) => onUpdate({ manufacturer: event.target.value })} disabled={readonly} required>
                  {selectOptions(options.manufacturers, t("vehicles.select.placeholder"))}
                </select>
              </label>
              <label className={ecosFieldClass("gauge")}>
                <RequiredLabel label={t("vehicle.field.gauge")} filled={Boolean(compactValue(form.gauge))} />
                <select value={form.gauge} onChange={(event) => onUpdate({ gauge: event.target.value })} disabled={readonly} required>
                  {selectOptions(options.gauges, t("vehicles.select.placeholder"))}
                </select>
              </label>
            </div>

            <label className={ecosFieldClass("name")}>
              <RequiredLabel label={t("vehicle.field.name")} filled={articleIdentityFilled} />
              <input value={form.name} onChange={(event) => onUpdate({ name: event.target.value })} disabled={readonly} required />
            </label>

            <div className="form-row">
              <label>
                {t("vehicle.field.railwayCompany")}
                <select value={form.railwayCompany || ""} onChange={(event) => onUpdate({ railwayCompany: event.target.value })} disabled={readonly}>
                  {selectOptions(options.railwayCompanies)}
                </select>
              </label>
              <label>
                {t("vehicle.field.epoch")}
                <select value={form.epoch || ""} onChange={(event) => onUpdate({ epoch: event.target.value })} disabled={readonly}>
                  {selectOptions(options.epochs)}
                </select>
              </label>
            </div>

            <div className="form-row">
              <label className={ecosFieldClass("category")}>
                <RequiredLabel label={t("vehicle.field.category")} filled={Boolean(compactValue(form.category))} />
                <select value={form.category || ""} onChange={(event) => onUpdateCategory(event.target.value)} disabled={readonly} required>
                  {selectOptions(options.categories, t("vehicles.select.placeholder"))}
                </select>
              </label>
              <label className={ecosFieldClass("gattung")}>
                <RequiredLabel label={t("vehicle.field.gattung")} filled={Boolean(compactValue(form.gattung))} />
                <select value={form.gattung || ""} onChange={(event) => onUpdate({ gattung: event.target.value })} disabled={readonly || filteredGattungen.length === 0} required>
                  {selectOptions(filteredGattungen, t("vehicles.select.placeholder"))}
                </select>
              </label>
            </div>

            <label>
              {t("vehicle.field.description")}
              <textarea value={form.description || ""} onChange={(event) => onUpdate({ description: event.target.value })} disabled={readonly} rows={4} />
            </label>

            <div className="form-row">
              <label>
                {t("vehicle.field.series")}
                <input value={form.series || ""} onChange={(event) => onUpdate({ series: event.target.value })} disabled={readonly} />
              </label>
              <label>
                {t("vehicle.field.vehicleNumber")}
                <input value={form.vehicleNumber || ""} onChange={(event) => onUpdate({ vehicleNumber: event.target.value })} disabled={readonly} />
              </label>
            </div>

            <div className="form-row three-columns decoder-row">
              <label>
                {t("vehicle.field.digitalDecoderNumber")}
                <span className="inline-switch-input">
                  <span className="switch-field" aria-label="Digital">
                    <input type="checkbox" checked={Boolean(form.digital)} onChange={(event) => onUpdate({ digital: event.target.checked })} disabled={readonly} />
                    <span />
                  </span>
                  <input value={form.digitalDecoderNumber || ""} onChange={(event) => onUpdate({ digitalDecoderNumber: event.target.value })} disabled={readonly || !form.digital} />
                </span>
              </label>
              <label>
                {t("vehicle.field.dtDecoderNumber")}
                <span className="inline-switch-input">
                  <span className="switch-field" aria-label="DT Decoder">
                    <input type="checkbox" checked={Boolean(form.dtDecoder)} onChange={(event) => onUpdate({ dtDecoder: event.target.checked })} disabled={readonly} />
                    <span />
                  </span>
                  <input value={form.dtDecoderNumber || ""} onChange={(event) => onUpdate({ dtDecoderNumber: event.target.value })} disabled={readonly || !form.dtDecoder} />
                </span>
              </label>
              <label>
                {t("vehicle.field.decoderType")}
                <input value={form.decoderType || ""} onChange={(event) => onUpdate({ decoderType: event.target.value })} disabled={readonly} />
              </label>
            </div>

            <div className="form-row compact-switch-row">
              <label className="switch-label">
                {t("vehicle.field.exhibitionReady")}
                <span className="switch-field">
                  <input type="checkbox" checked={Boolean(form.exhibitionReady)} onChange={(event) => onUpdate({ exhibitionReady: event.target.checked })} disabled={readonly} />
                  <span />
                </span>
              </label>
              <label className="switch-label">
                {t("vehicle.field.exhibition")}
                <span className="switch-field">
                  <input type="checkbox" checked={Boolean(form.exhibition)} onChange={(event) => onUpdate({ exhibition: event.target.checked })} disabled={readonly} />
                  <span />
                </span>
              </label>
              <label className="switch-label">
                ABC-Bremsen
                <span className="switch-field">
                  <input type="checkbox" checked={Boolean(form.abcBrakes)} onChange={(event) => onUpdate({ abcBrakes: event.target.checked })} disabled={readonly} />
                  <span />
                </span>
              </label>
            </div>

            <div className="form-row three-columns">
              <label>
                EAN-Nr.
                <input value={form.ean || ""} onChange={(event) => onUpdate({ ean: event.target.value })} disabled={readonly} />
              </label>
              <label>
                {t("vehicle.field.productionPeriod")}
                <input value={form.productionPeriod || ""} onChange={(event) => onUpdate({ productionPeriod: event.target.value })} disabled={readonly} placeholder="TT. MM. JJJJ" />
              </label>
              <label>
                {t("vehicle.field.listPrice")}
                <input value={form.listPrice || ""} onChange={(event) => onUpdate({ listPrice: event.target.value })} disabled={readonly} inputMode="decimal" />
              </label>
            </div>
          </div>
        )}
      </section>

      <section className="accordion-section">
        <button type="button" className="accordion-trigger" onClick={() => onToggleSection("details")}>
          {openSections.details ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
          {t("vehicles.details.title")}
        </button>
        {openSections.details && (
          <div className="accordion-content vehicle-form">
            <VehicleDetailsFields
              form={form}
              readonly={readonly}
              onOpenQr={onOpenQr}
              canOpenQr={canOpenQr}
              update={onUpdate}
              updateCouplingFront={onUpdateCouplingFront}
              updateCouplingSame={onUpdateCouplingSame}
            />
          </div>
        )}
      </section>

      <section className="accordion-section">
        <button type="button" className="accordion-trigger" onClick={() => onToggleSection("vehicle")}>
          {openSections.vehicle ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
          {t("vehicles.vehicle.title")}
        </button>
        {openSections.vehicle && (
          <div className="accordion-content vehicle-form">
            <VehicleOwnershipFields
              form={form}
              readonly={readonly}
              update={onUpdate}
            />
          </div>
        )}
      </section>
    </div>
  );
}
