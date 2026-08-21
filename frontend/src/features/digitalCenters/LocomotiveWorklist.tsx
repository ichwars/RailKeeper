import {
  ChevronLeft, ChevronRight, Circle, Filter, MoreHorizontal, RefreshCw, Search, Settings2
} from "lucide-react";

import type {
  DigitalCenterCompareFilter,
  DigitalCenterCompareStatus,
  DigitalCenterReadSession,
  DigitalCenterWorkItemPage
} from "./digitalCenterModel";

const filterLabels: Array<{ value: DigitalCenterCompareFilter; label: string }> = [
  { value: "all", label: "Alle" },
  { value: "deviation", label: "Prüfen" },
  { value: "new", label: "Neu" }
];

export function LocomotiveWorklist({
  page, search, compareStatus, loading, error, onSearch, onCompareStatus,
  onPage, onPageSize, onRefresh, onCompare
}: {
  page: DigitalCenterWorkItemPage;
  search: string;
  compareStatus: DigitalCenterCompareFilter;
  loading: boolean;
  error: string;
  onSearch: (value: string) => void;
  onCompareStatus: (value: DigitalCenterCompareFilter) => void;
  onPage: (value: number) => void;
  onPageSize: (value: number) => void;
  onRefresh: () => Promise<DigitalCenterReadSession>;
  onCompare: (itemID: string) => void;
}) {
  return (
    <section className="digital-centers-panel digital-centers-worklist" aria-labelledby="worklist-title">
      <header className="digital-centers-panel-head">
        <h2 id="worklist-title">Lok-Arbeitsliste</h2>
        <span className="digital-worklist-total">{page.total} Loks</span>
        <button type="button" className="digital-center-icon-button" aria-label="Arbeitsliste aktualisieren"
          title="Arbeitsliste aktualisieren" onClick={() => void onRefresh().catch(() => undefined)}>
          <RefreshCw size={16} aria-hidden="true" />
        </button>
      </header>
      <div className="digital-worklist-controls">
        <label className="digital-worklist-search">
          <Search size={17} aria-hidden="true" />
          <span className="sr-only">Lok suchen</span>
          <input type="search" value={search} placeholder="Lok suchen..." aria-label="Lok suchen"
            onChange={(event) => onSearch(event.target.value)} />
        </label>
        <div className="digital-worklist-filters" aria-label="Abgleich filtern">
          {filterLabels.map((filter) => (
            <button key={filter.value} type="button" className={compareStatus === filter.value ? "active" : ""}
              aria-label={filter.value === "deviation" ? "Abweichung filtern" : `${filter.label} filtern`}
              aria-pressed={compareStatus === filter.value} onClick={() => onCompareStatus(filter.value)}>
              {filter.label} {filter.value === "all" && <span>{page.total}</span>}
            </button>
          ))}
          <button type="button" className="digital-center-icon-button" aria-label="Weitere Filter" title="Weitere Filter">
            <Filter size={16} aria-hidden="true" />
          </button>
        </div>
      </div>
      <div className="digital-worklist-table-wrap">
        <table className="digital-worklist-table">
          <thead><tr>
            <th>Lokname</th><th>Decoder-Adresse</th><th>Protokoll</th><th>Abgleich</th><th>Status</th>
            <th><Settings2 size={15} aria-label="Aktionen" /></th>
          </tr></thead>
          <tbody>
            {loading && <StateRow text="Lokdaten werden geladen" />}
            {!loading && error && <StateRow text={error} tone="error" />}
            {!loading && !error && page.items.length === 0 && <StateRow text="Noch keine Lokdaten gelesen" />}
            {!loading && !error && page.items.map((item) => (
              <tr key={item.id}>
                <td title={item.name}><span className={`digital-item-dot ${item.compareStatus}`}>
                  <Circle size={9} fill="currentColor" aria-hidden="true" /></span>{item.name}</td>
                <td>{item.decoderAddress}</td><td>{item.protocol}</td>
                <td className={`digital-compare-${item.compareStatus}`}>{compareLabel(item.compareStatus)}</td>
                <td><span className={`digital-item-status ${item.compareStatus}`}>
                  <Circle size={9} fill="currentColor" aria-hidden="true" />{statusLabel(item.compareStatus)}</span></td>
                <td><button type="button" className="digital-center-icon-button"
                  aria-label={`${item.name} vergleichen`} title={`${item.name} vergleichen`}
                  onClick={() => onCompare(item.id)}><MoreHorizontal size={17} aria-hidden="true" /></button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <footer className="digital-worklist-pagination">
        <label>Zeilen pro Seite
          <select aria-label="Zeilen pro Seite" value={page.pageSize}
            onChange={(event) => onPageSize(Number(event.target.value))}>
            <option value={10}>10</option><option value={25}>25</option><option value={50}>50</option>
          </select>
        </label>
        <span>{page.total === 0 ? "0" : `${(page.page - 1) * page.pageSize + 1}–${Math.min(page.page * page.pageSize, page.total)}`} von {page.total}</span>
        <button type="button" aria-label="Vorherige Seite" disabled={page.page <= 1}
          onClick={() => onPage(page.page - 1)}><ChevronLeft size={16} aria-hidden="true" /></button>
        <strong>{page.page}</strong>
        <button type="button" aria-label="Nächste Seite" disabled={page.page >= page.totalPages}
          onClick={() => onPage(page.page + 1)}><ChevronRight size={16} aria-hidden="true" /></button>
      </footer>
    </section>
  );
}

function StateRow({ text, tone = "muted" }: { text: string; tone?: "muted" | "error" }) {
  return <tr><td colSpan={6} className={`digital-worklist-state ${tone}`}>{text}</td></tr>;
}

function compareLabel(status: DigitalCenterCompareStatus) {
  if (status === "ok") return "OK";
  if (status === "deviation") return "Abweichung";
  if (status === "missing") return "Fehlt in Zentrale";
  if (status === "new") return "Neu";
  return "Konflikt";
}

function statusLabel(status: DigitalCenterCompareStatus) {
  if (status === "ok") return "Aktiv";
  if (status === "new") return "Neu";
  if (status === "missing") return "Fehlt";
  return "Prüfen";
}
