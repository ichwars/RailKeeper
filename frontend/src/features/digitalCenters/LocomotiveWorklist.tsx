import { useState } from "react";
import {
  ChevronLeft, ChevronRight, Circle, Filter, MoreHorizontal, RefreshCw, Search, Settings2
} from "lucide-react";

import { useI18n } from "../../shared/i18n";
import type {
  DigitalCenterCompareFilter,
  DigitalCenterCompareStatus,
  DigitalCenterReadSession,
  DigitalCenterWorkItemPage
} from "./digitalCenterModel";

const quickFilters: DigitalCenterCompareFilter[] = ["all", "deviation", "new"];

const advancedFilters: DigitalCenterCompareStatus[] = ["ok", "deviation", "missing", "new", "conflict"];

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
  const { t } = useI18n();
  const [advancedFiltersOpen, setAdvancedFiltersOpen] = useState(false);
  return (
    <section className="digital-centers-panel digital-centers-worklist" aria-labelledby="worklist-title">
      <header className="digital-centers-panel-head">
        <h2 id="worklist-title">{t("digitalCenters.worklist.title")}</h2>
        <span className="digital-worklist-total">
          {t("digitalCenters.common.locomotiveCount", { count: page.total })}
        </span>
        <button type="button" className="digital-center-icon-button"
          aria-label={t("digitalCenters.worklist.refresh")} title={t("digitalCenters.worklist.refresh")}
          onClick={() => void onRefresh().catch(() => undefined)}>
          <RefreshCw size={16} aria-hidden="true" />
        </button>
      </header>
      <div className="digital-worklist-controls">
        <label className="digital-worklist-search">
          <Search size={17} aria-hidden="true" />
          <span className="sr-only">{t("digitalCenters.worklist.search")}</span>
          <input type="search" value={search} placeholder={t("digitalCenters.worklist.searchPlaceholder")}
            aria-label={t("digitalCenters.worklist.search")}
            onChange={(event) => onSearch(event.target.value)} />
        </label>
        <div className="digital-worklist-filters" aria-label={t("digitalCenters.worklist.filtersLabel")}>
          {quickFilters.map((filter) => {
            const label = quickFilterLabel(filter, t);
            const accessibleLabel = filter === "deviation"
              ? t("digitalCenters.worklist.filterDeviationAction")
              : t("digitalCenters.worklist.filterAction", { label });
            return <button key={filter} type="button" className={compareStatus === filter ? "active" : ""}
              aria-label={accessibleLabel} aria-pressed={compareStatus === filter}
              onClick={() => onCompareStatus(filter)}>
              {label} {filter === "all" && <span>{page.total}</span>}
            </button>
          })}
          <span className="digital-worklist-advanced-wrap">
            <button type="button" className="digital-center-icon-button"
              aria-label={t("digitalCenters.worklist.moreFilters")}
              title={t("digitalCenters.worklist.moreFilters")} aria-expanded={advancedFiltersOpen}
              onClick={() => setAdvancedFiltersOpen((current) => !current)}>
              <Filter size={16} aria-hidden="true" />
            </button>
            {advancedFiltersOpen && <span className="digital-worklist-advanced"
              aria-label={t("digitalCenters.worklist.advancedFilters")}>
              {advancedFilters.map((filter) => {
                const label = compareLabel(filter, t);
                return <button key={filter} type="button"
                  aria-label={t("digitalCenters.worklist.filterAction", { label })}
                  aria-pressed={compareStatus === filter}
                  onClick={() => onCompareStatus(filter)}>{label}</button>;
              })}
            </span>}
          </span>
        </div>
      </div>
      <div className="digital-worklist-table-wrap">
        <table className="digital-worklist-table">
          <thead><tr>
            <th>{t("digitalCenters.worklist.columnName")}</th>
            <th>{t("digitalCenters.worklist.columnAddress")}</th>
            <th>{t("digitalCenters.worklist.columnProtocol")}</th>
            <th>{t("digitalCenters.worklist.columnComparison")}</th>
            <th>{t("digitalCenters.worklist.columnStatus")}</th>
            <th><Settings2 size={15} aria-label={t("digitalCenters.worklist.columnActions")} /></th>
          </tr></thead>
          <tbody>
            {loading && <StateRow text={t("digitalCenters.worklist.loading")} />}
            {!loading && error && <StateRow text={error} tone="error" />}
            {!loading && !error && page.items.length === 0 &&
              <StateRow text={t("digitalCenters.worklist.empty")} />}
            {!loading && !error && page.items.map((item) => (
              <tr key={item.id}>
                <td title={item.name}><span className={`digital-item-dot ${item.compareStatus}`}>
                  <Circle size={9} fill="currentColor" aria-hidden="true" /></span>{item.name}</td>
                <td>{item.decoderAddress}</td><td>{item.protocol}</td>
                <td className={`digital-compare-${item.compareStatus}`}>{compareLabel(item.compareStatus, t)}</td>
                <td><span className={`digital-item-status ${item.compareStatus}`}>
                  <Circle size={9} fill="currentColor" aria-hidden="true" />
                  {statusLabel(item.compareStatus, t)}</span></td>
                <td><button type="button" className="digital-center-icon-button"
                  aria-label={t("digitalCenters.worklist.compareAction", { name: item.name })}
                  title={t("digitalCenters.worklist.compareAction", { name: item.name })}
                  onClick={() => onCompare(item.id)}><MoreHorizontal size={17} aria-hidden="true" /></button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <footer className="digital-worklist-pagination">
        <label>{t("digitalCenters.worklist.pageSize")}
          <select aria-label={t("digitalCenters.worklist.pageSize")} value={page.pageSize}
            onChange={(event) => onPageSize(Number(event.target.value))}>
            <option value={10}>10</option><option value={25}>25</option><option value={50}>50</option>
          </select>
        </label>
        <span>{t("digitalCenters.worklist.range", {
          range: page.total === 0 ? "0" :
            `${(page.page - 1) * page.pageSize + 1}–${Math.min(page.page * page.pageSize, page.total)}`,
          total: page.total
        })}</span>
        <button type="button" aria-label={t("digitalCenters.worklist.previousPage")} disabled={page.page <= 1}
          onClick={() => onPage(page.page - 1)}><ChevronLeft size={16} aria-hidden="true" /></button>
        <strong>{page.page}</strong>
        <button type="button" aria-label={t("digitalCenters.worklist.nextPage")} disabled={page.page >= page.totalPages}
          onClick={() => onPage(page.page + 1)}><ChevronRight size={16} aria-hidden="true" /></button>
      </footer>
    </section>
  );
}

function StateRow({ text, tone = "muted" }: { text: string; tone?: "muted" | "error" }) {
  return <tr><td colSpan={6} className={`digital-worklist-state ${tone}`}>{text}</td></tr>;
}

type Translate = (key: string, values?: Record<string, string | number>) => string;

function quickFilterLabel(filter: DigitalCenterCompareFilter, t: Translate) {
  if (filter === "all") return t("digitalCenters.worklist.filterAll");
  if (filter === "deviation") return t("digitalCenters.worklist.filterReview");
  return t("digitalCenters.worklist.filterNew");
}

function compareLabel(status: DigitalCenterCompareStatus, t: Translate) {
  return t(`digitalCenters.worklist.compare.${status}`);
}

function statusLabel(status: DigitalCenterCompareStatus, t: Translate) {
  if (status === "ok") return t("digitalCenters.common.active");
  if (status === "new") return t("digitalCenters.worklist.filterNew");
  if (status === "missing") return t("digitalCenters.worklist.statusMissing");
  return t("digitalCenters.worklist.filterReview");
}
