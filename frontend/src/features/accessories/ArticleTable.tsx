import { type CSSProperties, type ReactNode, useEffect, useRef } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";

import type {
  AccessoryArticleListItem,
  AccessoryArticleSort,
  AccessorySortDirection,
  MasterDataEntry
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  tableColumnWidth,
  tableMinimumWidth,
  type TableColumnWidths
} from "../../shared/tableColumnLayout";
import { TableColumnResizeHandle } from "../../shared/ui/TableColumnResizeHandle";
import { ArticleActions } from "./ArticleActions";
import {
  articleTableColumnWidthDefinitions,
  defaultArticleTableColumns,
  type ArticleTableColumn
} from "./articleTableColumns";
import { articleSubtypeLabel } from "./articleSubtypes";
import { articleTypeLabel } from "./articleTypes";
import { formatAccessoryMoney } from "./accessoryMoney";

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
  columns?: readonly ArticleTableColumn[];
  columnWidths?: TableColumnWidths<ArticleTableColumn>;
  columnWidthsLoading?: boolean;
  onPreviewColumnWidth?: (column: ArticleTableColumn, width: number) => void;
  onCommitColumnWidth?: (column: ArticleTableColumn, width: number) => void;
};

const tableColumns: Record<ArticleTableColumn, { sort?: AccessoryArticleSort }> = {
  image: { sort: "image" },
  inventoryNumber: { sort: "inventoryNumber" },
  manufacturer: { sort: "manufacturer" },
  articleNumber: { sort: "articleNumber" },
  name: { sort: "name" },
  type: { sort: "type" },
  gauge: { sort: "gauge" },
  listPrice: {},
  stock: { sort: "stock" },
  storage: { sort: "storage" }
};

const fixedTableWidth = 44 + 136;

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
  columns = defaultArticleTableColumns,
  columnWidths = {},
  columnWidthsLoading = false,
  onPreviewColumnWidth,
  onCommitColumnWidth
}: ArticleTableProps) {
  const selectAllRef = useRef<HTMLInputElement | null>(null);
  const { language, t } = useI18n();
  const allSelected = items.length > 0 && items.every((item) => selectedIDs.has(item.id));
  const someSelected = items.some((item) => selectedIDs.has(item.id));
  const layout = { columns: [...columns], widths: columnWidths };
  const minimumWidth = tableMinimumWidth(layout, articleTableColumnWidthDefinitions, fixedTableWidth);
  const tableStyle = {
    "--article-table-min-width": `${minimumWidth}px`
  } as CSSProperties;

  useEffect(() => {
    if (selectAllRef.current) selectAllRef.current.indeterminate = someSelected && !allSelected;
  }, [allSelected, someSelected]);

  const resizeHandle = (column: ArticleTableColumn, label: string) => {
    if (columnWidthsLoading || !onPreviewColumnWidth || !onCommitColumnWidth) return null;
    const definition = articleTableColumnWidthDefinitions[column];
    return <TableColumnResizeHandle
      label={t("common.resizeColumn", { label })}
      width={tableColumnWidth(layout, column, articleTableColumnWidthDefinitions)}
      minWidth={definition.minWidth}
      maxWidth={definition.maxWidth}
      defaultWidth={definition.defaultWidth}
      onPreview={(width) => onPreviewColumnWidth(column, width)}
      onCommit={(width) => onCommitColumnWidth(column, width)}
    />;
  };

  const renderHeader = (key: ArticleTableColumn) => {
    const columnSort = tableColumns[key].sort;
    const label = t(`accessories.table.${key}`);
    const active = columnSort !== undefined && sort === columnSort;
    return (
      <th key={key} className={`article-${key}-cell`} aria-label={label}
        aria-sort={columnSort ? (active ? (direction === "asc" ? "ascending" : "descending") : "none") : undefined}>
        {columnSort ? (
          <button type="button" className={active ? "article-sort-button active" : "article-sort-button"}
            onClick={() => onSort(columnSort)}
            aria-label={t("accessories.table.sortBy", { column: label })}>
            <span>{label}</span>
            {active ? direction === "asc"
              ? <ChevronUp size={14} aria-hidden="true" />
              : <ChevronDown size={14} aria-hidden="true" />
              : null}
          </button>
        ) : label}
        {resizeHandle(key, label)}
      </th>
    );
  };

  const renderCell = (article: AccessoryArticleListItem, column: ArticleTableColumn): ReactNode => {
    const storageTitle = article.locationNames.join(", ");
    const primaryLocation = article.locationNames[0] || t("common.none");
    switch (column) {
      case "image":
        return <td key={column} className="article-image-cell">
          {article.primaryImageUrl
            ? <img className="inventory-thumb" src={article.primaryImageUrl} alt="" />
            : <div className="image-placeholder">{t("exhibition.noPreview")}</div>}
        </td>;
      case "inventoryNumber":
        return <td key={column} className="article-inventoryNumber-cell article-inventory-cell">
          <span className="article-truncate" title={article.inventoryNumber}>{article.inventoryNumber}</span>
        </td>;
      case "manufacturer":
        return <td key={column} className="article-manufacturer-cell">
          <span className="article-truncate" title={article.manufacturer}>{article.manufacturer}</span>
        </td>;
      case "articleNumber":
        return <td key={column} className="article-articleNumber-cell article-number-cell">
          <span className="article-truncate" title={article.articleNumber || undefined}>
            {article.articleNumber || t("common.none")}
          </span>
        </td>;
      case "name":
        return <td key={column} className="article-name-cell article-main-cell">
          {onView ? <button type="button" className="article-name-button" onClick={() => onView(article)}>
            <strong className="article-truncate" title={article.name}>{article.name}</strong>
          </button> : <div className="article-name-content">
            <strong className="article-truncate" title={article.name}>{article.name}</strong>
          </div>}
        </td>;
      case "type":
        return <td key={column} className="article-type-cell">
          <strong>{articleTypeLabel(article.articleType, articleTypeEntries, t)}</strong>
          <small>{article.subtype
            ? articleSubtypeLabel(article.articleType, article.subtype, subtypeEntries, t)
            : t("common.none")}</small>
        </td>;
      case "gauge":
        return <td key={column} className="article-gauge-cell">
          {article.gauges.length ? article.gauges.join(", ") : t("common.none")}
        </td>;
      case "listPrice":
        return <td key={column} className="article-listPrice-cell">
          {formatAccessoryMoney(article.listPrice, language) || t("common.none")}
        </td>;
      case "stock":
        return <td key={column} className="article-stock-cell">
          <strong>{t("accessories.table.stockOwned", { count: article.owned })}</strong>
          <small>{t("accessories.table.stockBreakdown", {
            available: article.available,
            reserved: article.reserved,
            installed: article.installed
          })}</small>
        </td>;
      case "storage":
        return <td key={column} className="article-storage-cell">
          <span className="article-truncate" title={storageTitle || primaryLocation}>{primaryLocation}</span>
          {article.locationNames.length > 1
            ? <small>{t("accessories.table.moreLocations", { count: article.locationNames.length - 1 })}</small>
            : null}
        </td>;
    }
  };

  return (
    <div className="table-wrap article-table-wrap">
      <table className="inventory-table article-table" style={tableStyle}>
        <colgroup>
          <col className="select-cell" style={{ width: 44, minWidth: 44, maxWidth: 44 }} />
          {columns.map((column) => <col key={column} data-column={column} style={{
            width: `${tableColumnWidth(layout, column, articleTableColumnWidthDefinitions)}px`
          }} />)}
          <col className="table-fill-cell" style={{
            width: `max(0px, calc(100% - ${minimumWidth}px))`
          }} />
          <col className="actions-cell" style={{ width: 136, minWidth: 136, maxWidth: 136 }} />
        </colgroup>
        <thead>
          <tr>
            <th className="select-cell" aria-label={t("accessories.table.select")}>
              <label className="table-select-field" title={t("accessories.table.selectAll")}>
                <input ref={selectAllRef} type="checkbox" checked={allSelected}
                  disabled={items.length === 0} aria-label={t("accessories.table.selectAll")}
                  onChange={() => onToggleAll?.()} />
              </label>
            </th>
            {columns.map(renderHeader)}
            <th className="table-fill-cell" aria-hidden="true" />
            <th className="actions-cell">{t("accessories.table.actions")}</th>
          </tr>
        </thead>
        <tbody>
          {items.map((article) => (
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
              {columns.map((column) => renderCell(article, column))}
              <td className="table-fill-cell" aria-hidden="true" />
              <td className="actions-cell">
                <ArticleActions article={article} canEdit={canEdit} canDelete={canDelete}
                  onView={onView} onEdit={onEdit} onArchive={onArchive} onRestore={onRestore}
                  onDelete={onDelete} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
