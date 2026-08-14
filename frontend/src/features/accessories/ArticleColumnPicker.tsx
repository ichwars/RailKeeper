import { useEffect, useRef, useState } from "react";
import { Columns3 } from "lucide-react";

import { useI18n } from "../../shared/i18n";
import { articleTableColumns, type ArticleTableColumn } from "./articleTableColumns";

type ArticleColumnPickerProps = {
  visibleColumns: ReadonlySet<ArticleTableColumn>;
  onToggle: (column: ArticleTableColumn) => void;
};

export function ArticleColumnPicker({ visibleColumns, onToggle }: ArticleColumnPickerProps) {
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
        </div>
      ) : null}
    </div>
  );
}
