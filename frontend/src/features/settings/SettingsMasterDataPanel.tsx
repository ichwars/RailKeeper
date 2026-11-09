import { ExternalLink, RefreshCw, Trash2, Upload, X } from "lucide-react";
import { useI18n } from "../../shared/i18n";
import { AppFileButton } from "../../shared/ui/AppFileButton";
import { MasterDataBulkToolbar } from "./MasterDataBulkSelection";
import { normalizeCV8Binary, normalizeCV8Decimal, normalizeCV8Hex } from "./settingsModel";
import { SettingsMasterDataTable } from "./SettingsMasterDataTable";
import type { SettingsMasterDataController } from "./useSettingsMasterData";

export function SettingsMasterDataPanel({ controller }: { controller: SettingsMasterDataController }) {
  const { t } = useI18n();
  const {
    activeType, canEditMasterData, editing, form, message, saving, masterDataBatchBusy, search, setSearch,
    activeDataLabel, activeDataDescription, loading, isSymbolData, isManufacturerData,
    isCV8ManufacturerData, masterDataSelection, reloadActiveType, update, selectSymbolImage, startCreate,
    submit, deactivateSelectedEntries
  } = controller;

  return (
    <section className="master-data-panel">
      <div className="master-data-head">
        <div>
          <h3>{t("settings.data.manage", { label: activeDataLabel })}</h3>
          <p>{activeDataDescription}</p>
        </div>
        <button type="button" className="icon-button" onClick={reloadActiveType} aria-label={t("settings.data.refresh")} title={t("settings.data.refresh")} disabled={loading}>
          <RefreshCw size={16} />
        </button>
      </div>
      {isCV8ManufacturerData && (
        <a className="source-note" href="https://www.nmra.org/sites/default/files/standards/sandrp/DCC/S/appendix_a_s-9_2_2.pdf" target="_blank" rel="noreferrer">
          <ExternalLink size={15} />
          <span>{t("settings.data.cv8Source")}</span>
        </a>
      )}

      <>
        <label className="settings-search">
          {t("settings.data.search")}
          <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder={t("settings.data.searchPlaceholder")} />
        </label>

          <form className={isCV8ManufacturerData ? "master-data-create cv8-create" : activeType === "manufacturer" ? "master-data-create manufacturer-create" : isSymbolData ? "master-data-create symbol-create" : "master-data-create"} onSubmit={submit}>
            <strong>{editing ? t("settings.data.editEntry") : t("settings.data.newEntry")}</strong>
            <input value={form.label} onChange={(event) => update({ label: event.target.value })} placeholder={t("settings.data.entryPlaceholder", { label: activeDataLabel })} required />
            {isManufacturerData && (
              <>
                <input
                  className="master-data-input-compact"
                  value={form.nominalScalesText}
                  onChange={(event) => update({ nominalScalesText: event.target.value })}
                  placeholder={t("settings.data.nominalPlaceholder")}
                />
                <input
                  value={form.website}
                  onChange={(event) => update({ website: event.target.value })}
                  placeholder={t("settings.data.manufacturerWebsitePlaceholder")}
                />
                <input
                  value={form.searchDomainsText}
                  onChange={(event) => update({ searchDomainsText: event.target.value })}
                  placeholder={t("settings.data.searchDomainsPlaceholder")}
                />
                <input
                  value={form.aliasesText}
                  onChange={(event) => update({ aliasesText: event.target.value })}
                  placeholder={t("settings.data.aliasesPlaceholder")}
                />
                <input
                  value={form.sourceUrl}
                  onChange={(event) => update({ sourceUrl: event.target.value })}
                  placeholder={t("settings.data.sourceUrlPlaceholder")}
                />
              </>
            )}
            {isCV8ManufacturerData ? (
              <>
                <input
                  value={form.cvDecimal}
                  onChange={(event) => {
                    const value = event.target.value;
                    const decimal = normalizeCV8Decimal(value);
                    update({
                      cvDecimal: value,
                      cvBinary: decimal ? normalizeCV8Binary("", decimal) : form.cvBinary,
                      cvHex: decimal ? normalizeCV8Hex("", decimal) : form.cvHex
                    });
                  }}
                  placeholder={t("settings.data.cv8DecimalPlaceholder")}
                  inputMode="numeric"
                  required
                />
                <input
                  value={form.cvBinary}
                  onChange={(event) => update({ cvBinary: normalizeCV8Binary(event.target.value, form.cvDecimal) })}
                  placeholder={t("settings.data.cv8BinaryPlaceholder")}
                  inputMode="numeric"
                />
                <input
                  value={form.cvHex}
                  onChange={(event) => update({ cvHex: normalizeCV8Hex(event.target.value, form.cvDecimal) })}
                  placeholder={t("settings.data.cv8HexPlaceholder")}
                />
                <input
                  value={form.cvCountry}
                  onChange={(event) => update({ cvCountry: event.target.value.toUpperCase() })}
                  placeholder={t("settings.data.cv8CountryPlaceholder")}
                  maxLength={8}
                />
              </>
            ) : isSymbolData ? (
              <>
                <textarea
                  value={form.description}
                  onChange={(event) => update({ description: event.target.value })}
                  placeholder={t("settings.data.descriptionPlaceholder")}
                  rows={2}
                />
                <AppFileButton buttonClassName="symbol-upload-field" aria-label={t("settings.data.uploadImage")}
                    accept="image/svg+xml,image/png,image/jpeg,image/webp,.svg"
                    onChange={(event) => {
                      selectSymbolImage(event.target.files?.[0] || null);
                      event.currentTarget.value = "";
                    }}>
                  <Upload size={15} aria-hidden="true" />
                  <span>{t("settings.data.uploadImage")}</span>
                </AppFileButton>
              </>
            ) : isManufacturerData ? null : (
              <input value={form.sourceUrl} onChange={(event) => update({ sourceUrl: event.target.value })} placeholder={t("settings.data.sourceUrlPlaceholder")} />
            )}
            <button className="primary-button" disabled={saving}>
              {saving ? t("vehicles.saving") : editing ? t("vehicles.save") : "+ " + t("settings.data.add")}
            </button>
            {editing && (
              <button type="button" className="icon-button" onClick={startCreate} aria-label={t("vehicles.cancel")} title={t("vehicles.cancel")}>
                <X size={16} />
              </button>
            )}
          </form>

          {isSymbolData && form.imageData && (
            <div className="symbol-preview-card">
              <img src={form.imageData} alt="" />
              <span>{t("settings.data.savedSymbol")}</span>
              <button type="button" className="icon-button danger" onClick={() => update({ imageData: "" })} aria-label={t("settings.data.removeImage")} title={t("settings.data.removeImage")}>
                <Trash2 size={16} />
              </button>
            </div>
          )}

          {canEditMasterData && <MasterDataBulkToolbar
            count={masterDataSelection.selectedCount}
            busy={masterDataBatchBusy}
            onDeactivate={deactivateSelectedEntries}
          />}

          <SettingsMasterDataTable controller={controller} />

          {message && <p className="form-message">{message}</p>}
      </>
    </section>
  );
}
