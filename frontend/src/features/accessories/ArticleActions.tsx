import { Archive, ArchiveRestore, Eye, Pencil, Trash2 } from "lucide-react";

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
  const { t } = useI18n();
  const archiveLabel = article.archived
    ? "accessories.actions.restore"
    : "accessories.actions.archive";
  const archiveNamedLabel = article.archived
    ? "accessories.actions.restoreNamed"
    : "accessories.actions.archiveNamed";

  return (
    <div className="table-actions article-row-actions">
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
        <button
          type="button"
          className="icon-button article-action-button"
          onClick={() => void (article.archived ? onRestore(article) : onArchive(article))}
          aria-label={t(archiveNamedLabel, { name: article.name })}
          title={t(archiveLabel)}
        >
          {article.archived
            ? <ArchiveRestore size={16} aria-hidden="true" />
            : <Archive size={16} aria-hidden="true" />}
        </button>
      ) : null}
      {canDelete && onDelete ? (
        <button
          type="button"
          className="icon-button article-action-button danger"
          onClick={() => onDelete(article)}
          aria-label={t("accessories.actions.deleteNamed", { name: article.name })}
          title={t("accessories.actions.delete")}
        >
          <Trash2 size={16} aria-hidden="true" />
        </button>
      ) : null}
    </div>
  );
}
