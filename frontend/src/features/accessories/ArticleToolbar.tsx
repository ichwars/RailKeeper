import { Grid2X2, Search, Table2, X } from "lucide-react";

import type { AccessoryArticleFilterOptions, MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { ArticleColumnPicker } from "./ArticleColumnPicker";
import {
  defaultArticleTableColumns,
  type ArticleTableColumn
} from "./articleTableColumns";
import type { ArticleViewMode } from "./articleViewMode";
import type { ArticleOverviewFilters } from "./useArticleOverview";
import { articleTypeLabel, articleTypeOrder } from "./articleTypes";

const statusOptions = [
  "available",
  "allocated",
  "reserved",
  "installed",
  "maintenance_due",
  "defective",
  "archived"
] as const;

function articleTypeFromControl(value: string): ArticleOverviewFilters["articleType"] {
  return articleTypeOrder.find((type) => type === value) || "";
}

function statusFromControl(value: string): ArticleOverviewFilters["status"] {
  return statusOptions.find((status) => status === value) || "";
}

export function ArticleToolbar({
  filters,
  options,
  articleTypeEntries,
  resultCount,
  viewMode,
  hasActiveFilters,
  onViewModeChange,
  visibleColumns = defaultArticleTableColumns,
  onToggleColumn = () => undefined,
  onFilterChange,
  onReset
}: {
  filters: ArticleOverviewFilters;
  options: AccessoryArticleFilterOptions;
  articleTypeEntries: MasterDataEntry[];
  resultCount: number;
  viewMode: ArticleViewMode;
  hasActiveFilters: boolean;
  onViewModeChange: (mode: ArticleViewMode) => void;
  visibleColumns?: ReadonlySet<ArticleTableColumn>;
  onToggleColumn?: (column: ArticleTableColumn) => void;
  onFilterChange: <Key extends keyof ArticleOverviewFilters>(key: Key, value: ArticleOverviewFilters[Key]) => void;
  onReset: () => void;
}) {
  const { t } = useI18n();
  const availableTypes = articleTypeOrder.filter((type) => options.articleTypes.includes(type));

  return (
    <div className="article-toolbar" aria-label={t("accessories.toolbar.label")}>
      <div className="article-toolbar-primary">
        <label className="search-field article-search-field">
          <span>
            <Search size={16} aria-hidden="true" />
            <input
              type="search"
              value={filters.query}
              onChange={(event) => onFilterChange("query", event.target.value)}
              placeholder={t("accessories.toolbar.searchPlaceholder")}
              aria-label={t("accessories.toolbar.search")}
            />
          </span>
        </label>
        <span className="inventory-view-tools article-view-tools" aria-label={t("accessories.view.label")}>
          <button
            type="button"
            className={`icon-button${viewMode === "table" ? " active" : ""}`}
            aria-label={t("accessories.view.table")}
            aria-pressed={viewMode === "table"}
            title={t("accessories.view.table")}
            onClick={() => onViewModeChange("table")}
          >
            <Table2 size={16} aria-hidden="true" />
          </button>
          <button
            type="button"
            className={`icon-button${viewMode === "cards" ? " active" : ""}`}
            aria-label={t("accessories.view.cards")}
            aria-pressed={viewMode === "cards"}
            title={t("accessories.view.cards")}
            onClick={() => onViewModeChange("cards")}
          >
            <Grid2X2 size={16} aria-hidden="true" />
          </button>
          {viewMode === "table" ? (
            <ArticleColumnPicker visibleColumns={visibleColumns} onToggle={onToggleColumn} />
          ) : null}
        </span>
      </div>
      <div className="article-filter-row">
        <AppSelect
          className="article-filter-select"
          value={filters.articleType}
          onChange={(event) => onFilterChange("articleType", articleTypeFromControl(event.target.value))}
          aria-label={t("accessories.toolbar.articleType")}
        >
          <option value="">{t("accessories.toolbar.allArticleTypes")}</option>
          {availableTypes.map((type) => (
            <option key={type} value={type}>{articleTypeLabel(type, articleTypeEntries, t)}</option>
          ))}
        </AppSelect>
        <AppSelect
          className="article-filter-select"
          value={filters.manufacturer}
          onChange={(event) => onFilterChange("manufacturer", event.target.value)}
          aria-label={t("accessories.toolbar.manufacturer")}
        >
          <option value="">{t("accessories.toolbar.allManufacturers")}</option>
          {options.manufacturers.map((manufacturer) => (
            <option key={manufacturer} value={manufacturer}>{manufacturer}</option>
          ))}
        </AppSelect>
        <AppSelect
          className="article-filter-select"
          value={filters.gauge}
          onChange={(event) => onFilterChange("gauge", event.target.value)}
          aria-label={t("accessories.toolbar.gauge")}
        >
          <option value="">{t("accessories.toolbar.allGauges")}</option>
          {options.gauges.map((gauge) => <option key={gauge} value={gauge}>{gauge}</option>)}
        </AppSelect>
        <AppSelect
          className="article-filter-select"
          value={filters.status}
          onChange={(event) => onFilterChange("status", statusFromControl(event.target.value))}
          aria-label={t("accessories.toolbar.status")}
        >
          <option value="">{t("accessories.toolbar.allStatuses")}</option>
          {statusOptions.map((status) => (
            <option key={status} value={status}>{t(`accessories.status.${status}`)}</option>
          ))}
        </AppSelect>
        <AppSelect
          className="article-filter-select article-location-filter"
          value={filters.locationId}
          onChange={(event) => onFilterChange("locationId", event.target.value)}
          aria-label={t("accessories.toolbar.location")}
        >
          <option value="">{t("accessories.toolbar.allLocations")}</option>
          {options.storageLocations.map((location) => (
            <option key={location.id} value={location.id}>{location.name}</option>
          ))}
        </AppSelect>
        {hasActiveFilters ? (
          <button type="button" className="inventory-filter-clear" onClick={onReset}>
            <X size={14} aria-hidden="true" />
            {t("accessories.toolbar.reset")}
          </button>
        ) : null}
        <span className="inventory-filter-result article-result-count" aria-live="polite">
          {t(resultCount === 1 ? "accessories.toolbar.resultOne" : "accessories.toolbar.resultMany", {
            count: resultCount
          })}
        </span>
      </div>
    </div>
  );
}
