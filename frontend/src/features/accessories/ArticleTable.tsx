import { type CSSProperties, useEffect, useRef } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

import type {
  AccessoryArticleListItem,
  AccessoryArticleSort,
  AccessorySortDirection,
  MasterDataEntry
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleActions } from "./ArticleActions";
import {
  defaultArticleTableColumns,
  type ArticleTableColumn
} from "./articleTableColumns";
import { articleSubtypeLabel } from "./articleSubtypes";
import { articleTypeLabel } from "./articleTypes";

type ArticleTableProps = {
  items: AccessoryArticleListItem[];
  subtypeEntries?: MasterDataEntry[];
  articleTypeEntries?: MasterDataEntry[];
  sort: AccessoryArticleSort;
  direction: AccessorySortDirection;
  canEdit: boolean;
  canDelete?: boolean;
  onSort: (sort: AccessoryArticleSort) => void;
  selectedIDs?: Set<string>;
  onToggleSelection?: (id: string) => void;
  onToggleAll?: () => void;
  onView?: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
  onDelete?: (article: AccessoryArticleListItem) => void;
  visibleColumns?: ReadonlySet<ArticleTableColumn>;
};

const sortableColumns: Array<{ sort: AccessoryArticleSort; key: ArticleTableColumn }> = [
  { sort: "image", key: "image" },
  { sort: "inventoryNumber", key: "inventoryNumber" },
  { sort: "manufacturer", key: "manufacturer" },
  { sort: "articleNumber", key: "articleNumber" },
  { sort: "name", key: "name" },
  { sort: "type", key: "type" },
  { sort: "gauge", key: "gauge" },
  { sort: "stock", key: "stock" },
  { sort: "storage", key: "storage" }
];

const articleColumnWidths: Record<ArticleTableColumn, number> = {
  image: 86,
  inventoryNumber: 142,
  manufacturer: 150,
  articleNumber: 126,
  name: 210,
  type: 175,
  gauge: 80,
  stock: 210,
  storage: 170
};

function articleTableStyle(visibleColumns: ReadonlySet<ArticleTableColumn>) {
  const dataWidth = [...visibleColumns]
    .reduce((total, column) => total + articleColumnWidths[column], 0);
  const minimumWidth = Math.max(420, 44 + 128 + dataWidth);
  return { "--article-table-min-width": `${minimumWidth}px` } as CSSProperties;
}

export function ArticleTable({
  items,
  subtypeEntries = [],
  articleTypeEntries = [],
  sort,
  direction,
  canEdit,
  canDelete = false,
  onSort,
  selectedIDs = new Set<string>(),
  onToggleSelection,
  onToggleAll,
  onView,
  onEdit,
  onArchive,
  onRestore,
  onDelete,
  visibleColumns = defaultArticleTableColumns
}: ArticleTableProps) {
  const selectAllRef = useRef<HTMLInputElement | null>(null);
  const { t } = useI18n();
  const allSelected = items.length > 0 && items.every((item) => selectedIDs.has(item.id));
  const someSelected = items.some((item) => selectedIDs.has(item.id));

  useEffect(() => {
    if (selectAllRef.current) selectAllRef.current.indeterminate = someSelected && !allSelected;
  }, [allSelected, someSelected]);

  const renderSortHeader = ({
    sort: columnSort,
    key
  }: { sort: AccessoryArticleSort; key: ArticleTableColumn }) => {
    const active = sort === columnSort;
    const label = t(`accessories.table.${key}`);
    return (
      <th key={columnSort} className={`article-${key}-cell`}
        aria-sort={active ? (direction === "asc" ? "ascending" : "descending") : "none"}>
        <button
          type="button"
          className={active ? "article-sort-button active" : "article-sort-button"}
          onClick={() => onSort(columnSort)}
          aria-label={t("accessories.table.sortBy", { column: label })}
        >
          <span>{label}</span>
          {active ? direction === "asc"
            ? <ChevronUp size={14} aria-hidden="true" />
            : <ChevronDown size={14} aria-hidden="true" />
            : null}
        </button>
      </th>
    );
  };

  return (
    <div className="table-wrap article-table-wrap">
      <table className="inventory-table article-table" style={articleTableStyle(visibleColumns)}>
        <thead>
          <tr>
            <th className="select-cell" aria-label={t("accessories.table.select")}>
              <label className="table-select-field" title={t("accessories.table.selectAll")}>
                <input ref={selectAllRef} type="checkbox" checked={allSelected}
                  disabled={items.length === 0} aria-label={t("accessories.table.selectAll")}
                  onChange={() => onToggleAll?.()} />
              </label>
            </th>
            {sortableColumns.filter(({ key }) => visibleColumns.has(key)).map(renderSortHeader)}
            <th className="actions-cell">{t("accessories.table.actions")}</th>
          </tr>
        </thead>
        <tbody>
          {items.map((article) => {
            const storageTitle = article.locationNames.join(", ");
            const primaryLocation = article.locationNames[0] || t("common.none");
            return (
              <tr key={article.id} className={[
                article.archived ? "archived" : "",
                selectedIDs.has(article.id) ? "selected-row" : ""
              ].filter(Boolean).join(" ")}>
                <td className="select-cell">
                  <label className="table-select-field" title={t("accessories.table.selectNamed", { name: article.name })}>
                    <input type="checkbox" checked={selectedIDs.has(article.id)}
                      aria-label={t("accessories.table.selectNamed", { name: article.name })}
                      onChange={() => onToggleSelection?.(article.id)} />
                  </label>
                </td>
                {visibleColumns.has("image") ? <td className="article-image-cell">
                  {article.primaryImageUrl
                    ? <img className="inventory-thumb" src={article.primaryImageUrl} alt="" />
                    : <div className="image-placeholder">{t("exhibition.noPreview")}</div>}
                </td> : null}
                {visibleColumns.has("inventoryNumber") ? <td className="article-inventoryNumber-cell article-inventory-cell">
                  <span className="article-truncate" title={article.inventoryNumber}>{article.inventoryNumber}</span>
                </td> : null}
                {visibleColumns.has("manufacturer") ? <td className="article-manufacturer-cell">
                  <span className="article-truncate" title={article.manufacturer}>{article.manufacturer}</span>
                </td> : null}
                {visibleColumns.has("articleNumber") ? <td className="article-articleNumber-cell article-number-cell">
                  <span className="article-truncate" title={article.articleNumber || undefined}>
                    {article.articleNumber || t("common.none")}
                  </span>
                </td> : null}
                {visibleColumns.has("name") ? <td className="article-name-cell article-main-cell">
                  {onView ? <button type="button" className="article-name-button" onClick={() => onView(article)}>
                    <strong className="article-truncate" title={article.name}>{article.name}</strong>
                  </button> : <div className="article-name-content">
                    <strong className="article-truncate" title={article.name}>{article.name}</strong>
                  </div>}
                </td> : null}
                {visibleColumns.has("type") ? <td className="article-type-cell">
                  <strong>{articleTypeLabel(article.articleType, articleTypeEntries, t)}</strong>
                  <small>{article.subtype
                    ? articleSubtypeLabel(article.articleType, article.subtype, subtypeEntries, t)
                    : t("common.none")}</small>
                </td> : null}
                {visibleColumns.has("gauge") ? <td className="article-gauge-cell">
                  {article.gauges.length ? article.gauges.join(", ") : t("common.none")}
                </td> : null}
                {visibleColumns.has("stock") ? <td className="article-stock-cell">
                  <strong>{t("accessories.table.stockOwned", { count: article.owned })}</strong>
                  <small>{t("accessories.table.stockBreakdown", {
                    available: article.available,
                    reserved: article.reserved,
                    installed: article.installed
                  })}</small>
                </td> : null}
                {visibleColumns.has("storage") ? <td className="article-storage-cell">
                  <span className="article-truncate" title={storageTitle || primaryLocation}>{primaryLocation}</span>
                  {article.locationNames.length > 1 ? (
                    <small>{t("accessories.table.moreLocations", { count: article.locationNames.length - 1 })}</small>
                  ) : null}
                </td> : null}
                <td className="actions-cell">
                  <ArticleActions article={article} canEdit={canEdit} canDelete={canDelete}
                    onView={onView} onEdit={onEdit} onArchive={onArchive} onRestore={onRestore}
                    onDelete={onDelete} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
