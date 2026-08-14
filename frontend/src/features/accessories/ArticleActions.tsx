import {
  type KeyboardEvent as ReactKeyboardEvent,
  useEffect,
  useLayoutEffect,
  useRef,
  useState
} from "react";
import { Eye, MoreHorizontal, Pencil } from "lucide-react";

import type { AccessoryArticleListItem } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

type ArticleActionsProps = {
  article: AccessoryArticleListItem;
  canEdit: boolean;
  canDelete?: boolean;
  onView?: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
  onDelete?: (article: AccessoryArticleListItem) => void;
};

export function ArticleActions({
  article,
  canEdit,
  canDelete = false,
  onView,
  onEdit,
  onArchive,
  onRestore,
  onDelete
}: ArticleActionsProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const { t } = useI18n();

  useLayoutEffect(() => {
    if (open) menuRef.current?.querySelector<HTMLButtonElement>("[role='menuitem']")?.focus();
  }, [open]);

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

  const handleMenuKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Tab") {
      setOpen(false);
      return;
    }
    if (!["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) return;
    event.preventDefault();
    const items = Array.from(
      menuRef.current?.querySelectorAll<HTMLButtonElement>("[role='menuitem']") || []
    );
    if (items.length === 0) return;
    const focusedIndex = items.findIndex((item) => item === document.activeElement);
    const current = focusedIndex >= 0 ? focusedIndex : 0;
    let next = current;
    if (event.key === "Home") next = 0;
    else if (event.key === "End") next = items.length - 1;
    else if (event.key === "ArrowDown") next = (current + 1) % items.length;
    else next = (current - 1 + items.length) % items.length;
    items[next]?.focus();
  };

  return (
    <div ref={rootRef} className="table-actions article-row-actions">
      {onView ? (
        <button
          type="button"
          className="icon-button article-action-button"
          onClick={() => onView(article)}
          aria-label={t("accessories.actions.viewNamed", { name: article.name })}
          title={t("accessories.actions.view")}
        >
          <Eye size={16} aria-hidden="true" />
        </button>
      ) : null}
      {canEdit && onEdit ? (
        <button
          type="button"
          className="icon-button article-action-button"
          onClick={() => onEdit(article)}
          aria-label={t("accessories.actions.editNamed", { name: article.name })}
          title={t("accessories.actions.edit")}
        >
          <Pencil size={16} aria-hidden="true" />
        </button>
      ) : null}
      {canEdit || canDelete ? (
        <div className="article-overflow">
          <button
            ref={triggerRef}
            type="button"
            className="icon-button article-action-button"
            onClick={() => setOpen((current) => !current)}
            aria-label={t("accessories.actions.moreNamed", { name: article.name })}
            title={t("accessories.actions.more")}
            aria-haspopup="menu"
            aria-expanded={open}
          >
            <MoreHorizontal size={17} aria-hidden="true" />
          </button>
          {open ? (
            <div
              ref={menuRef}
              className="article-action-menu"
              role="menu"
              onKeyDown={handleMenuKeyDown}
            >
              {canEdit ? (
                <button
                  type="button"
                  role="menuitem"
                  tabIndex={0}
                  onClick={() => {
                    setOpen(false);
                    void (article.archived ? onRestore(article) : onArchive(article));
                  }}
                >
                  {t(article.archived ? "accessories.actions.restore" : "accessories.actions.archive")}
                </button>
              ) : null}
              {canDelete && onDelete ? (
                <button
                  type="button"
                  role="menuitem"
                  tabIndex={-1}
                  className="danger-menu-item"
                  onClick={() => {
                    setOpen(false);
                    onDelete(article);
                  }}
                >
                  {t("accessories.actions.delete")}
                </button>
              ) : null}
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
