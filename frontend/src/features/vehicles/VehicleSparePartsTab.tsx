import { Check, ExternalLink, PackageSearch, Pencil, Save, Search, Trash2 } from "lucide-react";
import type {
  ArticleSearchSparePart,
  Vehicle,
  VehicleSparePart,
  VehicleSparePartInput
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function sparePartResultKey(part: ArticleSearchSparePart, index: number) {
  return `${part.articleNumber || ""}|${part.description || ""}|${part.url || ""}|${index}`.toLocaleLowerCase();
}

export function VehicleSparePartsTab({
  selected,
  readonly,
  saving,
  searchLoading,
  searchError,
  searchRan,
  sparePartForm,
  editingSparePartID,
  foundSpareParts,
  onUpdateSparePartForm,
  onResetSparePartForm,
  onSaveSparePart,
  onEditSparePart,
  onDeleteSparePart,
  onSearchSpareParts,
  selectedFoundSpareParts,
  onToggleFoundSparePart,
  onToggleAllFoundSpareParts,
  onAdoptSelectedSpareParts,
}: {
  selected: Vehicle | null;
  readonly: boolean;
  saving: boolean;
  searchLoading: boolean;
  searchError: string;
  searchRan: boolean;
  sparePartForm: VehicleSparePartInput;
  editingSparePartID: string | null;
  foundSpareParts: ArticleSearchSparePart[];
  onUpdateSparePartForm: (patch: Partial<VehicleSparePartInput>) => void;
  onResetSparePartForm: () => void;
  onSaveSparePart: () => void;
  onEditSparePart: (part: VehicleSparePart) => void;
  onDeleteSparePart: (part: VehicleSparePart) => void;
  onSearchSpareParts: () => void;
  selectedFoundSpareParts: Record<string, boolean>;
  onToggleFoundSparePart: (key: string, checked: boolean) => void;
  onToggleAllFoundSpareParts: (checked: boolean) => void;
  onAdoptSelectedSpareParts: () => void;
}) {
  const { t } = useI18n();
  const storedParts = selected?.spareParts || [];
  const selectedFoundCount = foundSpareParts.filter((part, index) => selectedFoundSpareParts[sparePartResultKey(part, index)]).length;
  const allFoundSelected = foundSpareParts.length > 0 && selectedFoundCount === foundSpareParts.length;

  return (
    <section className="spare-parts-tab">
      <details className="spare-parts-editor" open>
        <summary>
          <span>
            <PackageSearch size={18} aria-hidden="true" />
            {t("vehicles.spareParts.editor")}
          </span>
        </summary>
        {!selected && <p className="empty-state compact">{t("vehicles.spareParts.emptyUntilSave")}</p>}
        {selected && (
          <>
            <div className="spare-part-form">
              <label>
                {t("vehicles.spareParts.articleNumber")}
                <input value={sparePartForm.articleNumber || ""} onChange={(event) => onUpdateSparePartForm({ articleNumber: event.target.value })} disabled={readonly || saving} />
              </label>
              <label className="spare-part-description">
                {t("vehicles.spareParts.description")}
                <input value={sparePartForm.description || ""} onChange={(event) => onUpdateSparePartForm({ description: event.target.value })} disabled={readonly || saving} />
              </label>
              <label>
                {t("vehicles.spareParts.price")}
                <input value={sparePartForm.price || ""} onChange={(event) => onUpdateSparePartForm({ price: event.target.value })} disabled={readonly || saving} inputMode="decimal" />
              </label>
              <label className="spare-part-link">
                {t("vehicles.spareParts.link")}
                <input value={sparePartForm.url || ""} onChange={(event) => onUpdateSparePartForm({ url: event.target.value })} disabled={readonly || saving} />
              </label>
            </div>
            <div className="maintenance-actions">
              {editingSparePartID && (
                <button type="button" className="secondary-button" onClick={onResetSparePartForm} disabled={readonly || saving}>
                  {t("vehicles.cancel")}
                </button>
              )}
              <button type="button" className="primary-button" onClick={onSaveSparePart} disabled={readonly || saving}>
                <Save size={15} aria-hidden="true" />
                {editingSparePartID ? t("vehicles.spareParts.save") : t("vehicles.spareParts.add")}
              </button>
            </div>
          </>
        )}
      </details>

      <section className="spare-parts-section">
        <div className="upload-head compact">
          <div>
            <h3>{t("vehicles.spareParts.saved")}</h3>
            <p>{t("vehicles.spareParts.savedHelp")}</p>
          </div>
        </div>
        {selected && storedParts.length === 0 && <p className="empty-state compact">{t("vehicles.spareParts.empty")}</p>}
        {storedParts.length > 0 && (
          <div className="compact-table">
            <table>
              <thead>
                <tr>
                  <th>{t("vehicles.spareParts.articleNumber")}</th>
                  <th>{t("vehicles.spareParts.description")}</th>
                  <th>{t("vehicles.spareParts.price")}</th>
                  <th>{t("vehicles.spareParts.link")}</th>
                  <th>{t("vehicles.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {storedParts.map((part) => (
                  <tr key={part.id}>
                    <td>{part.articleNumber || "-"}</td>
                    <td>{part.description || "-"}</td>
                    <td>{part.price || "-"}</td>
                    <td>
                      {part.url ? <a href={part.url} target="_blank" rel="noreferrer"><ExternalLink size={14} /> {t("vehicles.spareParts.open")}</a> : "-"}
                    </td>
                    <td>
                      <div className="table-actions">
                        <button type="button" className="icon-button" onClick={() => onEditSparePart(part)} disabled={readonly || saving} aria-label={t("vehicles.spareParts.edit")} title={t("vehicles.spareParts.edit")}>
                          <Pencil size={15} />
                        </button>
                        <button type="button" className="icon-button danger" onClick={() => onDeleteSparePart(part)} disabled={readonly || saving} aria-label={t("vehicles.spareParts.delete")} title={t("vehicles.spareParts.delete")}>
                          <Trash2 size={15} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="spare-parts-section">
        <div className="upload-head compact">
          <div>
            <h3>{t("vehicles.spareParts.webTitle")}</h3>
            <p>{t("vehicles.spareParts.webHelp")}</p>
          </div>
          <button type="button" className="secondary-button" onClick={onSearchSpareParts} disabled={!selected || readonly || searchLoading}>
            <Search size={15} aria-hidden="true" />
            {searchLoading ? t("vehicles.spareParts.searching") : t("vehicles.spareParts.search")}
          </button>
          {foundSpareParts.length > 0 && (
            <>
              <span className="selection-count">{t("vehicles.spareParts.selectedCount", { count: selectedFoundCount })}</span>
              <button type="button" className="secondary-button" onClick={() => onToggleAllFoundSpareParts(!allFoundSelected)} disabled={readonly || saving}>
                {allFoundSelected ? t("vehicles.spareParts.selectNone") : t("vehicles.spareParts.selectAll")}
              </button>
              <button type="button" className="primary-button" onClick={onAdoptSelectedSpareParts} disabled={readonly || saving || selectedFoundCount === 0}>
                <Save size={15} aria-hidden="true" />
                {t("vehicles.spareParts.adoptSelected")}
              </button>
            </>
          )}
        </div>
        {searchError && <p className="form-message">{searchError}</p>}
        {searchRan && !searchLoading && foundSpareParts.length === 0 && <p className="empty-state compact">{t("vehicles.spareParts.noWebParts")}</p>}
        {foundSpareParts.length > 0 && (
          <div className="compact-table">
            <table>
              <thead>
                <tr>
                  <th>{t("vehicles.spareParts.select")}</th>
                  <th>{t("vehicles.spareParts.articleNumber")}</th>
                  <th>{t("vehicles.spareParts.description")}</th>
                  <th>{t("vehicles.spareParts.price")}</th>
                  <th>{t("vehicles.articleSearch.source")}</th>
                </tr>
              </thead>
              <tbody>
                {foundSpareParts.map((part, index) => {
                  const key = sparePartResultKey(part, index);
                  const checked = Boolean(selectedFoundSpareParts[key]);
                  return (
                    <tr key={key}>
                      <td>
                        <button
                          type="button"
                          className={checked ? "selection-toggle active" : "selection-toggle"}
                          aria-pressed={checked}
                          aria-label={checked ? t("vehicles.spareParts.deselect") : t("vehicles.spareParts.select")}
                          onClick={() => onToggleFoundSparePart(key, !checked)}
                          disabled={readonly || saving}
                        >
                          <Check size={15} aria-hidden="true" />
                        </button>
                      </td>
                      <td>{part.articleNumber || "-"}</td>
                      <td>{part.description || "-"}</td>
                      <td>{part.price || "-"}</td>
                      <td>{part.url ? <a href={part.url} target="_blank" rel="noreferrer"><ExternalLink size={14} /> {part.source || t("vehicles.spareParts.open")}</a> : part.source || "-"}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>

    </section>
  );
}
