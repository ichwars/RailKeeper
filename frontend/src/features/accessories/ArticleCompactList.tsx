import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleActions } from "./ArticleActions";
import { articleSubtypeLabel } from "./articleSubtypes";
import { articleTypeLabel } from "./articleTypes";

type ArticleCompactListProps = {
  items: AccessoryArticleListItem[];
  articleTypeEntries: MasterDataEntry[];
  subtypeEntries: MasterDataEntry[];
  canEdit: boolean;
  onView: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

export function ArticleCompactList({
  items,
  articleTypeEntries,
  subtypeEntries,
  canEdit,
  onView,
  onEdit,
  onArchive,
  onRestore
}: ArticleCompactListProps) {
  const { t } = useI18n();

  return (
    <div className="article-mobile-list" role="list" aria-label={t("accessories.view.mobileLabel")}>
      {items.map((article) => {
        const typeLabel = articleTypeLabel(article.articleType, articleTypeEntries, t);
        const subtypeLabel = article.subtype
          ? articleSubtypeLabel(article.articleType, article.subtype, subtypeEntries, t)
          : typeLabel;

        return (
          <article key={article.id} className="article-mobile-item" role="listitem">
            <button
              type="button"
              className="article-mobile-media"
              onClick={() => onView(article)}
              aria-label={t("accessories.actions.viewNamed", { name: article.name })}
            >
              {article.primaryImageUrl ? (
                <img src={article.primaryImageUrl} alt="" />
              ) : (
                <div className="image-placeholder">{t("exhibition.noPreview")}</div>
              )}
            </button>
            <button type="button" className="article-mobile-main" onClick={() => onView(article)}>
              <span>{article.inventoryNumber}</span>
              <strong>{article.name}</strong>
              <small>{article.manufacturer || t("common.none")} · {article.articleNumber || t("common.none")}</small>
            </button>
            <div className="article-mobile-meta">
              <span>{article.gauges.join(", ") || t("common.none")}</span>
              <small>{subtypeLabel}</small>
              <strong>{t("accessories.table.stockOwned", { count: article.owned })}</strong>
            </div>
            <div className="article-mobile-actions">
              <ArticleActions article={article} canEdit={canEdit} onView={onView} onEdit={onEdit}
                onArchive={onArchive} onRestore={onRestore} />
            </div>
          </article>
        );
      })}
    </div>
  );
}
