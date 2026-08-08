import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { ChevronDown, ChevronUp, Eye, MoreHorizontal, Pencil } from "lucide-react";

import type {
  AccessoryArticleListItem,
  AccessoryArticleSort,
  AccessorySortDirection,
  MasterDataEntry
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
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
  onView?: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

const sortableColumns: Array<{ sort: AccessoryArticleSort; key: string }> = [
  { sort: "article", key: "article" },
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
  onView,
  onEdit,
  onArchive,
  onRestore
}: ArticleTableProps) {
  const [openMenuID, setOpenMenuID] = useState("");
  const [activeMenuIndex, setActiveMenuIndex] = useState(0);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const triggerRefs = useRef(new Map<string, HTMLButtonElement>());
  const { t } = useI18n();

  const restoreTriggerFocus = (articleID: string) => {
    triggerRefs.current.get(articleID)?.focus();
  };

  const closeMenu = (restoreFocus: boolean) => {
    const articleID = openMenuID;
    setOpenMenuID("");
    if (restoreFocus && articleID) restoreTriggerFocus(articleID);
  };

  const focusMenuItem = (index: number) => {
    const items = Array.from(menuRef.current?.querySelectorAll<HTMLButtonElement>("[role='menuitem']") || []);
    if (items.length === 0) return;
    const nextIndex = (index + items.length) % items.length;
    setActiveMenuIndex(nextIndex);
    items[nextIndex]?.focus();
  };

  useLayoutEffect(() => {
    if (openMenuID) focusMenuItem(0);
  }, [openMenuID]);

  useEffect(() => {
    if (!openMenuID) return;
    const closeOnOutsidePointer = (event: PointerEvent) => {
      if (event.target instanceof Element && event.target.closest(".article-overflow")) return;
      closeMenu(false);
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeMenu(true);
      }
    };
    document.addEventListener("pointerdown", closeOnOutsidePointer);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOnOutsidePointer);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [openMenuID]);

  const handleMenuKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    const itemCount = menuRef.current?.querySelectorAll("[role='menuitem']").length || 0;
    if (itemCount === 0) return;
    if (event.key === "ArrowDown") {
      event.preventDefault();
      focusMenuItem(activeMenuIndex + 1);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      focusMenuItem(activeMenuIndex - 1);
    } else if (event.key === "Home") {
      event.preventDefault();
      focusMenuItem(0);
    } else if (event.key === "End") {
      event.preventDefault();
      focusMenuItem(itemCount - 1);
    } else if (event.key === "Tab") {
      closeMenu(false);
    }
  };

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
    <div className={openMenuID ? "table-wrap article-table-wrap menu-open" : "table-wrap article-table-wrap"}>
      <table className="article-table">
        <thead>
          <tr>
            {sortableColumns.map(renderSortHeader)}
            <th className="actions-cell">{t("accessories.table.actions")}</th>
          </tr>
        </thead>
        <tbody>
          {items.map((article) => {
            const storageTitle = article.locationNames.join(", ");
            const primaryLocation = article.locationNames[0] || t("common.none");
            const viewLabel = t("accessories.actions.viewNamed", { name: article.name });
            const editLabel = t("accessories.actions.editNamed", { name: article.name });
            const moreLabel = t("accessories.actions.moreNamed", { name: article.name });
            return (
              <tr key={article.id} className={article.archived ? "archived" : ""}>
                <td className="article-main-cell">
                  {onView ? <button type="button" className="article-name-button" onClick={() => onView(article)}>
                    <strong className="article-truncate" title={article.name}>{article.name}</strong>
                    <span className="article-truncate" title={article.manufacturer}>{article.manufacturer}</span>
                    <small>{article.articleNumber || t("common.none")}</small>
                  </button> : <div className="article-name-content">
                    <strong className="article-truncate" title={article.name}>{article.name}</strong>
                    <span className="article-truncate" title={article.manufacturer}>{article.manufacturer}</span>
                    <small>{article.articleNumber || t("common.none")}</small>
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
                  <div className="table-actions article-row-actions">
                    {onView ? <button type="button" className="icon-button article-action-button" onClick={() => onView(article)}
                      aria-label={viewLabel} title={t("accessories.actions.view")}>
                      <Eye size={16} aria-hidden="true" />
                    </button> : null}
                    {canEdit ? (
                      <>
                        {onEdit ? <button type="button" className="icon-button article-action-button" onClick={() => onEdit(article)}
                          aria-label={editLabel} title={t("accessories.actions.edit")}>
                          <Pencil size={16} aria-hidden="true" />
                        </button> : null}
                        <div className="article-overflow">
                          <button
                            ref={(node) => {
                              if (node) triggerRefs.current.set(article.id, node);
                              else triggerRefs.current.delete(article.id);
                            }}
                            type="button"
                            className="icon-button article-action-button"
                            onClick={() => {
                              setActiveMenuIndex(0);
                              setOpenMenuID((current) => current === article.id ? "" : article.id);
                            }}
                            aria-label={moreLabel}
                            title={t("accessories.actions.more")}
                            aria-haspopup="menu"
                            aria-expanded={openMenuID === article.id}
                          >
                            <MoreHorizontal size={17} aria-hidden="true" />
                          </button>
                          {openMenuID === article.id ? (
                            <div ref={menuRef} className="article-action-menu" role="menu"
                              onKeyDown={handleMenuKeyDown}>
                              <button
                                type="button"
                                role="menuitem"
                                tabIndex={activeMenuIndex === 0 ? 0 : -1}
                                onClick={() => {
                                  closeMenu(false);
                                  void (article.archived ? onRestore(article) : onArchive(article));
                                }}
                              >
                                {t(article.archived ? "accessories.actions.restore" : "accessories.actions.archive")}
                              </button>
                            </div>
                          ) : null}
                        </div>
                      </>
                    ) : null}
                  </div>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
