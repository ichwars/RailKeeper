import { ExternalLink, PackageSearch, Pencil, Save, Search, Trash2 } from "lucide-react";
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

export function cleanSparePartDescription(value: string) {
  return String(value || "")
    .replace(/\b(?:Artikelnummer|Artikel-Nr\.?|Art\.?\s*Nr\.?|Nr\.?|Nummer|Number|No\.?|Item number|Item no\.?)\s*[:#-]?\s*/gi, "")
    .replace(/\b(?:Preis|Price)\s*[:#-]?\s*\d+(?:[,.]\d{1,2})?\s*(?:€|EUR)?/gi, "")
    .replace(/\d+(?:[,.]\d{1,2})?\s*(?:€|EUR)/gi, "")
    .replace(/\*\s*In den Warenkorb\b/gi, "")
    .replace(/\bIn den Warenkorb\b/gi, "")
    .replace(/\bAdd to cart\b/gi, "")
    .replace(/\s{2,}/g, " ")
    .replace(/\s*[-·|]\s*$/g, "")
    .trim();
}

export function strictCleanSparePartDescription(value: string) {
  return String(value || "")
    .replace(/^(?:GER|DE|ENG|EN)\s*[:/-]\s*/i, "")
    .replace(/\b(?:Artikelnummer|Artikel-Nr\.?|Art\.?\s*Nr\.?|Nr\.?|Nummer|Number|No\.?|Item number|Item no\.?)\s*[:#-]?\s*/gi, "")
    .replace(/\b(?:Preis|Price)\s*[:#-]?\s*\d+(?:[,.]\d{1,2})?\s*(?:\u20ac|EUR)?/gi, "")
    .replace(/\d+(?:[,.]\d{1,2})?\s*(?:\u20ac|EUR)/gi, "")
    .replace(/\*?\s*\b(?:In den Warenkorb|Zum Warenkorb hinzuf(?:u|ue)gen|In den Einkaufswagen|Add to cart|Add to shopping cart|Add to basket|Add to bag|Ajouter au panier|Anadir al carrito|Aggiungi al carrello|In winkelwagen|Toevoegen aan winkelwagen)\b/gi, "")
    .replace(/\s{2,}/g, " ")
    .replace(/\s*[-\u00b7|]\s*$/g, "")
    .trim();
}

export function isRealSparePartCandidate(part: ArticleSearchSparePart) {
  const description = strictCleanSparePartDescription(part.description);
  const signal = `${part.articleNumber || ""} ${description} ${part.url || ""}`.toLocaleLowerCase("de-DE");
  if (!description || description.length < 3) return false;
  if (/\b(bedienungsanl|bedienungsanleitung|ersatzteilliste|ersatzteilblatt|spare parts list|manual|download|katalog|catalog|et-blatt|explosionszeichnung|serviceblatt)\b/i.test(signal)) {
    return false;
  }
  return Boolean(part.price || /\b(kuppl|lautsprecher|decoder|reifen|haftreifen|radsatz|rad|motor|puffer|schraube|stromabnehmer|getriebe|geh(?:ae|ä)use|leiterplatte|feder|achse|traction tire|loudspeaker|coupler|speaker)\b/i.test(signal));
}

export function isStrictRealSparePartCandidate(part: ArticleSearchSparePart) {
  const description = strictCleanSparePartDescription(part.description);
  const signal = `${part.articleNumber || ""} ${description} ${part.url || ""}`.toLocaleLowerCase("de-DE");
  if (!description || description.length < 3) return false;
  if (/\b(bedienungsanl|bedienungsanleitung|ersatzteilliste|ersatzteilblatt|spare parts list|manual|download|katalog|catalog|et-blatt|explosionszeichnung|serviceblatt)\b/i.test(signal)) {
    return false;
  }
  return Boolean(part.price || /\b(kuppl|lautsprecher|decoder|reifen|haftreifen|radsatz|rad|motor|puffer|schraube|stromabnehmer|getriebe|geh(?:ae|\u00e4)use|leiterplatte|feder|achse|traction tire|loudspeaker|coupler|speaker)\b/i.test(signal));
}

export function shortSparePartSource(value?: string) {
  const source = String(value || "").trim();
  if (!source) return "";
  return source
    .replace(/^www\./i, "")
    .replace(/\.(de|com|eu|net|org)$/i, "")
    .replace(/-shop$/i, "")
    .replace(/shop$/i, "")
    .trim() || source;
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
        <div className="spare-parts-search-head">
          <div>
            <h3>{t("vehicles.spareParts.webTitle")}</h3>
            <p>{t("vehicles.spareParts.webHelp")}</p>
          </div>
          <button type="button" className="secondary-button compact-action" onClick={onSearchSpareParts} disabled={!selected || readonly || searchLoading}>
            <Search size={15} aria-hidden="true" />
            {searchLoading ? t("vehicles.spareParts.searching") : t("vehicles.spareParts.search")}
          </button>
        </div>
        {searchError && <p className="form-message">{searchError}</p>}
        {searchRan && !searchLoading && foundSpareParts.length === 0 && <p className="empty-state compact">{t("vehicles.spareParts.noWebParts")}</p>}
        {foundSpareParts.length > 0 && (
          <div className="spare-parts-web-results">
            <div className="spare-parts-web-actions">
              <span className="selection-count">{t("vehicles.spareParts.selectedCount", { count: selectedFoundCount })}</span>
            </div>
            <div className="compact-table spare-parts-found-table">
            <table>
              <thead>
                <tr>
                  <th className="select-column">
                    <input
                      type="checkbox"
                      checked={allFoundSelected}
                      onChange={(event) => onToggleAllFoundSpareParts(event.target.checked)}
                      disabled={readonly || saving || foundSpareParts.length === 0}
                      aria-label={t("vehicles.spareParts.selectAll")}
                    />
                  </th>
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
                  const cleanDescription = strictCleanSparePartDescription(part.description) || part.description || "-";
                  const source = shortSparePartSource(part.source) || part.source || "-";
                  return (
                    <tr key={key}>
                      <td className="select-column">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={(event) => onToggleFoundSparePart(key, event.target.checked)}
                          disabled={readonly || saving}
                          aria-label={checked ? t("vehicles.spareParts.deselect") : t("vehicles.spareParts.select")}
                        />
                      </td>
                      <td>{part.articleNumber || "-"}</td>
                      <td>{cleanDescription}</td>
                      <td>{part.price || "-"}</td>
                      <td>{part.url ? <a href={part.url} target="_blank" rel="noreferrer"><ExternalLink size={14} /> {source}</a> : source}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            </div>
          </div>
        )}
      </section>

    </section>
  );
}
