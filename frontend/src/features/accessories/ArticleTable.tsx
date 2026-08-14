import { useEffect, useRef } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

import type {
  AccessoryArticleListItem,
  AccessoryArticleSort,
  AccessorySortDirection,
  MasterDataEntry
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleActions } from "./ArticleActions";
import { articleSubtypeLabel } from "./articleSubtypes";
import { articleTypeLabel } from "./articleTypes";

type ArticleTableProps = {
  items: AccessoryArticleListItem[];
  subtypeEntries?: MasterDataEntry[];
  articleTypeEntries?: MasterDataEntry[];
  sort: AccessoryArticleSort;
  direction: AccessorySortDirection;
  canEdit: boolean;
  onSort: (sort: AccessoryArticleSort) => void;
  selectedIDs?: Set<string>;
  onToggleSelection?: (id: string) => void;
  onToggleAll?: () => void;
  onView?: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

const sortableColumns: Array<{ sort: AccessoryArticleSort; key: string }> = [
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

export function ArticleTable({
  items,
  subtypeEntries = [],
  articleTypeEntries = [],
  sort,
  direction,
  canEdit,
  onSort,
  selectedIDs = new Set<string>(),
  onToggleSelection,
  onToggleAll,
  onView,
  onEdit,
  onArchive,
  onRestore
}: ArticleTableProps) {
  const selectAllRef = useRef<HTMLInputElement | null>(null);
  const { t } = useI18n();
  const allSelected = items.length > 0 && items.every((item) => selectedIDs.has(item.id));
  const someSelected = items.some((item) => selectedIDs.has(item.id));

  useEffect(() => {
    if (selectAllRef.current) selectAllRef.current.indeterminate = someSelected && !allSelected;
  }, [allSelected, someSelected]);

  const renderSortHeader = ({ sort: columnSort, key }: { sort: AccessoryArticleSort; key: string }) => {
    const active = sort === columnSort;
    const label = t(`accessories.table.${key}`);
    return (
      <th key={columnSort} aria-sort={active ? (direction === "asc" ? "ascending" : "descending") : "none"}>
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
      <table className="inventory-table article-table">
        <thead>
          <tr>
            <th className="select-cell" aria-label={t("accessories.table.select")}>
              <label className="table-select-field" title={t("accessories.table.selectAll")}>
                <input ref={selectAllRef} type="checkbox" checked={allSelected}
                  disabled={items.length === 0} aria-label={t("accessories.table.selectAll")}
                  onChange={() => onToggleAll?.()} />
              </label>
            </th>
            {sortableColumns.map(renderSortHeader)}
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
                <td className="article-image-cell">
                  {article.primaryImageUrl
                    ? <img className="inventory-thumb" src={article.primaryImageUrl} alt="" />
                    : <div className="image-placeholder">{t("exhibition.noPreview")}</div>}
                </td>
                <td className="article-inventory-cell">
                  <span className="article-truncate" title={article.inventoryNumber}>{article.inventoryNumber}</span>
                </td>
                <td className="article-manufacturer-cell">
                  <span className="article-truncate" title={article.manufacturer}>{article.manufacturer}</span>
                </td>
                <td className="article-number-cell">
                  <span className="article-truncate" title={article.articleNumber || undefined}>
                    {article.articleNumber || t("common.none")}
                  </span>
                </td>
                <td className="article-main-cell">
                  {onView ? <button type="button" className="article-name-button" onClick={() => onView(article)}>
                    <strong className="article-truncate" title={article.name}>{article.name}</strong>
                  </button> : <div className="article-name-content">
                    <strong className="article-truncate" title={article.name}>{article.name}</strong>
                  </div>}
                </td>
                <td>
                  <strong>{articleTypeLabel(article.articleType, articleTypeEntries, t)}</strong>
                  <small>{article.subtype
                    ? articleSubtypeLabel(article.articleType, article.subtype, subtypeEntries, t)
                    : t("common.none")}</small>
                </td>
                <td>{article.gauges.length ? article.gauges.join(", ") : t("common.none")}</td>
                <td className="article-stock-cell">
                  <strong>{t("accessories.table.stockOwned", { count: article.owned })}</strong>
                  <small>{t("accessories.table.stockBreakdown", {
                    available: article.available,
                    reserved: article.reserved,
                    installed: article.installed
                  })}</small>
                </td>
                <td className="article-storage-cell">
                  <span className="article-truncate" title={storageTitle || primaryLocation}>{primaryLocation}</span>
                  {article.locationNames.length > 1 ? (
                    <small>{t("accessories.table.moreLocations", { count: article.locationNames.length - 1 })}</small>
                  ) : null}
                </td>
                <td className="actions-cell">
                  <ArticleActions article={article} canEdit={canEdit} onView={onView} onEdit={onEdit}
                    onArchive={onArchive} onRestore={onRestore} />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
