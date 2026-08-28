import { useEffect, useRef, useState } from "react";
import { ArrowDown, ArrowUp, Columns3, RotateCcw } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import {
  articleTableColumns,
  type ArticleColumnMove,
  type ArticleTableColumn
} from "./articleTableColumns";

type ArticleColumnPickerProps = {
  columns: readonly ArticleTableColumn[];
  loading?: boolean;
  onToggle: (column: ArticleTableColumn) => void;
  onMove: (column: ArticleTableColumn, direction: ArticleColumnMove) => void;
  onReset: () => void;
};

export function ArticleColumnPicker({
  columns,
  loading = false,
  onToggle,
  onMove,
  onReset
}: ArticleColumnPickerProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const { t } = useI18n();

  useEffect(() => {
    if (!open) return;
    const closeOutside = (event: PointerEvent) => {
      if (event.target instanceof Node && rootRef.current?.contains(event.target)) return;
      setOpen(false);
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      event.preventDefault();
      setOpen(false);
      triggerRef.current?.focus();
    };
    document.addEventListener("pointerdown", closeOutside);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOutside);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  const visibleColumns = new Set(columns);
  const identityLocked = (column: ArticleTableColumn) => visibleColumns.has(column) &&
    ((column === "inventoryNumber" && !visibleColumns.has("name")) ||
      (column === "name" && !visibleColumns.has("inventoryNumber")));

  return (
    <div ref={rootRef} className="article-column-picker">
      <button
        ref={triggerRef}
        type="button"
        className={`icon-button${open ? " active" : ""}`}
        aria-label={t("accessories.view.columns")}
        title={t("accessories.view.columns")}
        aria-haspopup="dialog"
        aria-expanded={open}
        disabled={loading}
        onClick={() => setOpen((current) => !current)}
      >
        <Columns3 size={16} aria-hidden="true" />
      </button>
      {open ? (
        <div
          className="article-column-picker-popover"
          role="group"
          aria-label={t("accessories.view.columnsGroup")}
        >
          <span className="article-column-order-label">{t("accessories.columns.visibleOrder")}</span>
          <div className="article-column-order">
            {columns.map((column, index) => {
              const label = t(`accessories.table.${column}`);
              return (
                <div key={column} className="article-column-order-row">
                  <span>{label}</span>
                  <span className="article-column-order-actions">
                    <button type="button" className="icon-button"
                      disabled={index === 0}
                      aria-label={t("accessories.columns.moveUp", { label })}
                      title={t("accessories.columns.moveUp", { label })}
                      onClick={() => onMove(column, "up")}
                    >
                      <ArrowUp size={14} aria-hidden="true" />
                    </button>
                    <button type="button" className="icon-button"
                      disabled={index === columns.length - 1}
                      aria-label={t("accessories.columns.moveDown", { label })}
                      title={t("accessories.columns.moveDown", { label })}
                      onClick={() => onMove(column, "down")}
                    >
                      <ArrowDown size={14} aria-hidden="true" />
                    </button>
                  </span>
                </div>
              );
            })}
          </div>
          {articleTableColumns.map((column) => (
            <label key={column}>
              <input
                type="checkbox"
                checked={visibleColumns.has(column)}
                disabled={identityLocked(column)}
                onChange={() => onToggle(column)}
              />
              <span>{t(`accessories.table.${column}`)}</span>
            </label>
          ))}
          <button type="button" className="article-column-picker-reset" onClick={onReset}>
            <RotateCcw size={15} aria-hidden="true" />
            {t("accessories.view.resetColumns")}
          </button>
        </div>
      ) : null}
    </div>
  );
}
