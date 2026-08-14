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
  onView?: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

export function ArticleActions({
  article,
  canEdit,
  onView,
  onEdit,
  onArchive,
  onRestore
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
    if (["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) {
      event.preventDefault();
      menuRef.current?.querySelector<HTMLButtonElement>("[role='menuitem']")?.focus();
    } else if (event.key === "Tab") {
      setOpen(false);
    }
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
      {canEdit ? (
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
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
