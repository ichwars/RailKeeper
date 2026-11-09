import { AppNumberControl } from "../../shared/ui/AppNumberControl";
import { useI18n } from "../../shared/i18n";
import { Plus, RefreshCw } from "lucide-react";
import type { SettingsGeneralController } from "./useSettingsGeneral";

export function SettingsInventoryCard({ general }: { general: SettingsGeneralController }) {
  const { t } = useI18n();
  const {
    inventorySchemes,
    inventorySchemesLoading,
    inventorySchemesMessage,
    newInventoryScheme,
    loadInventorySchemes,
    updateInventoryScheme,
    updateNewInventoryScheme,
    inventorySchemePreview,
    createInventoryScheme,
    saveInventoryScheme
  } = general;
  const inventoryCategoryLabel = (category: string) => {
    const label = t(`settings.category.${category}`);
    return label === `settings.category.${category}` ? category : label;
  };

  return (
    <section className="panel settings-card settings-tool-card">
      <div className="settings-section-head">
        <div>
          <h2>{t("settings.inventoryNumbers.title")}</h2>
          <p>{t("settings.inventoryNumbers.subtitle")}</p>
        </div>
        <button type="button" className="icon-button" onClick={loadInventorySchemes} aria-label={t("settings.inventoryNumbers.refresh")} title={t("settings.inventoryNumbers.refresh")} disabled={inventorySchemesLoading}>
          <RefreshCw size={16} />
        </button>
      </div>

      <div className="table-wrap settings-inline-table">
        <table>
          <thead>
            <tr>
              <th>{t("settings.inventoryNumbers.category")}</th>
              <th>{t("settings.inventoryNumbers.prefix")}</th>
              <th>{t("settings.inventoryNumbers.next")}</th>
              <th>{t("settings.inventoryNumbers.padding")}</th>
              <th>{t("common.active")}</th>
              <th>{t("settings.inventoryNumbers.preview")}</th>
              <th>{t("exhibition.actions")}</th>
            </tr>
          </thead>
          <tbody>
            <tr className="inventory-scheme-create-row">
              <td>
                <input
                  value={newInventoryScheme.category}
                  onChange={(event) => updateNewInventoryScheme({ category: event.target.value })}
                  placeholder={t("settings.inventoryNumbers.newCategory")}
                />
              </td>
              <td>
                <input
                  value={newInventoryScheme.prefix}
                  onChange={(event) => updateNewInventoryScheme({ prefix: event.target.value })}
                  placeholder={t("settings.inventoryNumbers.newPrefix")}
                />
              </td>
              <td>
                <AppNumberControl min={1} value={newInventoryScheme.nextNumber} onChange={(event) => updateNewInventoryScheme({ nextNumber: Number(event.target.value) })} />
              </td>
              <td>
                <AppNumberControl min={1} max={12} value={newInventoryScheme.padding} onChange={(event) => updateNewInventoryScheme({ padding: Number(event.target.value) })} />
              </td>
              <td>
                <label className="switch-field" aria-label={t("settings.inventoryNumbers.newActiveAria")}>
                  <input type="checkbox" checked={newInventoryScheme.active} onChange={(event) => updateNewInventoryScheme({ active: event.target.checked })} />
                  <span />
                </label>
              </td>
              <td><code>{inventorySchemePreview(newInventoryScheme.prefix, newInventoryScheme.nextNumber, newInventoryScheme.padding)}</code></td>
              <td>
                <button type="button" className="secondary-button compact-action" onClick={createInventoryScheme} disabled={!newInventoryScheme.category.trim() || !newInventoryScheme.prefix.trim()}>
                  <Plus size={15} />
                  {t("settings.inventoryNumbers.create")}
                </button>
              </td>
            </tr>
            {inventorySchemesLoading && inventorySchemes.length === 0 ? (
              <tr>
                <td colSpan={7} className="loading-cell">{t("settings.inventoryNumbers.loading")}</td>
              </tr>
            ) : inventorySchemes.length === 0 ? (
              <tr>
                <td colSpan={7} className="loading-cell">{t("settings.inventoryNumbers.empty")}</td>
              </tr>
            ) : (
              inventorySchemes.map((scheme) => (
                <tr key={scheme.id}>
                  <td><strong>{inventoryCategoryLabel(scheme.category)}</strong></td>
                  <td>
                    <input value={scheme.prefix} onChange={(event) => updateInventoryScheme(scheme.category, { prefix: event.target.value })} />
                  </td>
                  <td>
                    <AppNumberControl min={1} value={scheme.nextNumber} onChange={(event) => updateInventoryScheme(scheme.category, { nextNumber: Number(event.target.value) })} />
                  </td>
                  <td>
                    <AppNumberControl min={1} max={12} value={scheme.padding} onChange={(event) => updateInventoryScheme(scheme.category, { padding: Number(event.target.value) })} />
                  </td>
                  <td>
                    <label className="switch-field" aria-label={t("settings.inventoryNumbers.activeAria", { category: inventoryCategoryLabel(scheme.category) })}>
                      <input type="checkbox" checked={scheme.active} onChange={(event) => updateInventoryScheme(scheme.category, { active: event.target.checked })} />
                      <span />
                    </label>
                  </td>
                  <td><code>{scheme.preview}</code></td>
                  <td>
                    <button type="button" className="secondary-button compact-action" onClick={() => saveInventoryScheme(scheme)}>
                      {t("vehicles.save")}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      {inventorySchemesMessage && <p className="form-message">{inventorySchemesMessage}</p>}
    </section>
  );
}
