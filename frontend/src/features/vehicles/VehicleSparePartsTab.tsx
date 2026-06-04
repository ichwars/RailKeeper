import { Fragment } from "react";
import { ArrowUpDown, ExternalLink, PackageSearch, Pencil, Save, Search, Trash2 } from "lucide-react";
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

export type SparePartSortKey = "articleNumber" | "description" | "price" | "link";
export type SparePartSortDirection = "asc" | "desc";

function sparePartSortValue(part: VehicleSparePart, key: SparePartSortKey) {
  if (key === "articleNumber") return part.articleNumber || "";
  if (key === "description") return part.description || "";
  if (key === "price") return part.price || "";
  return part.url || "";
}

function availabilityTone(value?: string) {
  const lower = String(value || "").toLocaleLowerCase("de-DE");
  if (!lower) return "unknown";
  if (lower.includes("nicht") || lower.includes("ausverkauft") || lower.includes("derzeit")) return "unavailable";
  if (lower.includes("weniger") || lower.includes("knapp") || lower.includes("gering")) return "limited";
  if (lower.includes("lieferbar") || lower.includes("verf") || lower.includes("lager")) return "available";
  return "unknown";
}

function AvailabilityIcon({ value, label, loading = false }: { value?: string; label: string; loading?: boolean }) {
  const tone = availabilityTone(value);
  const title = value || label;
  return <span className={`availability-icon ${loading ? "loading" : tone}`} role="img" aria-label={title} />;
}

export function VehicleSparePartsTab({
  selected,
  readonly,
  saving,
  sparePartForm,
  editingSparePartID,
  sparePartLookupLoadingID,
  sparePartLookupErrors,
  sparePartLookupResults,
  sparePartStatusLoading,
  sparePartStatuses,
  importAllSparePartsLoading,
  canImportAllSpareParts,
  importAllSparePartsTitle,
  onUpdateSparePartForm,
  onResetSparePartForm,
  onSaveSparePart,
  onEditSparePart,
  onDeleteSparePart,
  onSearchSparePart,
  onApplySparePartLookup,
  onImportAllSpareParts,
  sparePartSort,
  onToggleSparePartSort,
}: {
  selected: Vehicle | null;
  readonly: boolean;
  saving: boolean;
  sparePartForm: VehicleSparePartInput;
  editingSparePartID: string | null;
  sparePartLookupLoadingID: string;
  sparePartLookupErrors: Record<string, string>;
  sparePartLookupResults: Record<string, ArticleSearchSparePart[]>;
  sparePartStatusLoading: Record<string, boolean>;
  sparePartStatuses: Record<string, ArticleSearchSparePart>;
  importAllSparePartsLoading: boolean;
  canImportAllSpareParts: boolean;
  importAllSparePartsTitle: string;
  onUpdateSparePartForm: (patch: Partial<VehicleSparePartInput>) => void;
  onResetSparePartForm: () => void;
  onSaveSparePart: () => void;
  onEditSparePart: (part: VehicleSparePart) => void;
  onDeleteSparePart: (part: VehicleSparePart) => void;
  onSearchSparePart: (part: VehicleSparePart) => void;
  onApplySparePartLookup: (part: VehicleSparePart, result: ArticleSearchSparePart) => void;
  onImportAllSpareParts: () => void;
  sparePartSort: { key: SparePartSortKey; direction: SparePartSortDirection };
  onToggleSparePartSort: (key: SparePartSortKey) => void;
}) {
  const { t } = useI18n();
  const storedParts = selected?.spareParts || [];
  const sortedStoredParts = [...storedParts].sort((left, right) => {
    const result = sparePartSortValue(left, sparePartSort.key).localeCompare(sparePartSortValue(right, sparePartSort.key), "de-DE", {
      numeric: true,
      sensitivity: "base"
    });
    return sparePartSort.direction === "asc" ? result : -result;
  });
  const sortLabel = (key: SparePartSortKey, label: string) => (
    <button
      type="button"
      className={`sort-button ${sparePartSort.key === key ? "active" : ""}`}
      onClick={() => onToggleSparePartSort(key)}
      aria-label={t("vehicles.sortBy", { field: label })}
    >
      {label}
      <ArrowUpDown size={13} aria-hidden="true" />
    </button>
  );

  return (
    <section className="spare-parts-tab">
      <details id="vehicle-spare-parts-editor" className="spare-parts-editor" open>
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
                <input id="vehicle-spare-part-article-number" value={sparePartForm.articleNumber || ""} onChange={(event) => onUpdateSparePartForm({ articleNumber: event.target.value })} disabled={readonly || saving} />
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
          <button
            type="button"
            className="secondary-button compact-action"
            onClick={onImportAllSpareParts}
            disabled={!canImportAllSpareParts || readonly || saving || importAllSparePartsLoading}
            title={importAllSparePartsTitle}
          >
            <PackageSearch size={15} aria-hidden="true" />
            {importAllSparePartsLoading ? t("vehicles.spareParts.importAllLoading") : t("vehicles.spareParts.importAll")}
          </button>
        </div>
        {selected && storedParts.length === 0 && <p className="empty-state compact">{t("vehicles.spareParts.empty")}</p>}
        {storedParts.length > 0 && (
          <div className="compact-table">
            <table>
              <thead>
                <tr>
                  <th>{sortLabel("articleNumber", t("vehicles.spareParts.articleNumber"))}</th>
                  <th>{sortLabel("description", t("vehicles.spareParts.description"))}</th>
                  <th>{sortLabel("price", t("vehicles.spareParts.price"))}</th>
                  <th>{sortLabel("link", t("vehicles.spareParts.link"))}</th>
                  <th>{t("vehicles.actions")}</th>
                </tr>
              </thead>
              <tbody>
                {sortedStoredParts.map((part) => {
                  const visibleUrl = part.url && !part.url.startsWith("/api/v1/vehicles/") ? part.url : "";
                  const lookupResults = sparePartLookupResults[part.id] || [];
                  const lookupError = sparePartLookupErrors[part.id] || "";
                  const lookupRan = Object.prototype.hasOwnProperty.call(sparePartLookupResults, part.id);
                  const lookupLoading = sparePartLookupLoadingID === part.id;
                  const bestLookupResult = lookupResults[0];
                  const statusLoading = Boolean(sparePartStatusLoading[part.id]);
                  const statusResult = bestLookupResult || sparePartStatuses[part.id];
                  const statusTitle = statusLoading ? t("vehicles.spareParts.statusChecking") : statusResult?.availability || t("vehicles.spareParts.availabilityUnknown");
                  return (
                    <Fragment key={part.id}>
                      <tr>
                        <td>{part.articleNumber || "-"}</td>
                        <td>{part.description || "-"}</td>
                        <td>{part.price || "-"}</td>
                        <td>
                          {visibleUrl ? (
                            <a className="icon-button spare-part-link-icon" href={visibleUrl} target="_blank" rel="noreferrer" aria-label={t("vehicles.spareParts.open")} title={t("vehicles.spareParts.open")}>
                              <ExternalLink size={15} />
                            </a>
                          ) : null}
                        </td>
                        <td>
                          <div className="table-actions">
                            <button type="button" className="icon-button" onClick={() => bestLookupResult && onApplySparePartLookup(part, bestLookupResult)} disabled={readonly || saving || !bestLookupResult} aria-label={t("vehicles.spareParts.applyFirstResult")} title={t("vehicles.spareParts.applyFirstResult")}>
                              <Save size={15} />
                            </button>
                            <button type="button" className="icon-button" onClick={() => onSearchSparePart(part)} disabled={readonly || saving || lookupLoading || !part.articleNumber} aria-label={t("vehicles.spareParts.searchOne")} title={t("vehicles.spareParts.searchOne")}>
                              <Search size={15} />
                            </button>
                            <span className="icon-button passive" title={statusTitle}>
                              <AvailabilityIcon value={statusResult?.availability} label={t("vehicles.spareParts.availabilityUnknown")} loading={statusLoading} />
                            </span>
                            <button type="button" className="icon-button" onClick={() => onEditSparePart(part)} disabled={readonly || saving} aria-label={t("vehicles.spareParts.edit")} title={t("vehicles.spareParts.edit")}>
                              <Pencil size={15} />
                            </button>
                            <button type="button" className="icon-button danger" onClick={() => onDeleteSparePart(part)} disabled={readonly || saving} aria-label={t("vehicles.spareParts.delete")} title={t("vehicles.spareParts.delete")}>
                              <Trash2 size={15} />
                            </button>
                          </div>
                        </td>
                      </tr>
                      {(lookupLoading || lookupError || lookupRan) && (
                        <tr className="spare-part-lookup-row">
                          <td colSpan={5}>
                            {lookupLoading && <span className="source-note compact">{t("vehicles.spareParts.lookupLoading")}</span>}
                            {lookupError && <p className="form-message">{lookupError}</p>}
                            {!lookupLoading && !lookupError && lookupRan && lookupResults.length === 0 && <span className="empty-state compact">{t("vehicles.spareParts.lookupEmpty")}</span>}
                            {lookupResults.length > 0 && (
                              <div className="spare-part-lookup-results">
                                {lookupResults.map((result, resultIndex) => {
                                  const source = shortSparePartSource(result.source) || shortSparePartSource(result.url) || t("vehicles.articleSearch.source");
                                  return (
                                    <div key={`${result.url || result.articleNumber}-${resultIndex}`} className="spare-part-lookup-result">
                                      <span>{result.price || "-"}</span>
                                      <span title={result.availability || t("vehicles.spareParts.availabilityUnknown")}>
                                        <AvailabilityIcon value={result.availability} label={t("vehicles.spareParts.availabilityUnknown")} />
                                      </span>
                                      {result.url ? (
                                        <a href={result.url} target="_blank" rel="noreferrer" title={t("vehicles.spareParts.open")}>
                                          <ExternalLink size={14} aria-hidden="true" />
                                          {source}
                                        </a>
                                      ) : (
                                        <span>{source}</span>
                                      )}
                                      <button type="button" className="icon-button" onClick={() => onApplySparePartLookup(part, result)} disabled={readonly || saving} aria-label={t("vehicles.spareParts.applyLookup")} title={t("vehicles.spareParts.applyLookup")}>
                                        <Save size={15} />
                                      </button>
                                    </div>
                                  );
                                })}
                              </div>
                            )}
                          </td>
                        </tr>
                      )}
                    </Fragment>
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
