import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleActions } from "./ArticleActions";
import { articleSubtypeLabel } from "./articleSubtypes";
import { articleTypeLabel } from "./articleTypes";

type ArticleCardGridProps = {
  items: AccessoryArticleListItem[];
  articleTypeEntries: MasterDataEntry[];
  subtypeEntries: MasterDataEntry[];
  canEdit: boolean;
  onView: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

export function ArticleCardGrid({
  items,
  articleTypeEntries,
  subtypeEntries,
  canEdit,
  onView,
  onEdit,
  onArchive,
  onRestore
}: ArticleCardGridProps) {
  const { t } = useI18n();

  return (
    <div className="article-card-grid" role="list" aria-label={t("accessories.view.cardsLabel")}>
      {items.map((article) => {
        const primaryLocation = article.locationNames[0] || t("common.none");
        const typeLabel = articleTypeLabel(article.articleType, articleTypeEntries, t);
        const subtypeLabel = article.subtype
          ? articleSubtypeLabel(article.articleType, article.subtype, subtypeEntries, t)
          : t("common.none");

        return (
          <article key={article.id} className="article-card" role="listitem">
            <button
              type="button"
              className="article-card-media"
              onClick={() => onView(article)}
              aria-label={t("accessories.actions.viewNamed", { name: article.name })}
            >
              {article.primaryImageUrl ? (
                <img src={article.primaryImageUrl} alt="" />
              ) : (
                <div className="image-placeholder">{t("exhibition.noPreview")}</div>
              )}
            </button>
            <div className="article-card-body">
              <div className="article-card-title">
                <div>
                  <strong title={article.inventoryNumber}>{article.inventoryNumber}</strong>
                  <span title={article.manufacturer}>{article.manufacturer}</span>
                </div>
                <span className="article-card-gauge">
                  {article.gauges.join(", ") || t("common.none")}
                </span>
              </div>
              <button type="button" className="article-name-button" onClick={() => onView(article)}>
                <strong title={article.name}>{article.name}</strong>
              </button>
              <dl>
                <div>
                  <dt>{t("accessories.table.articleNumber")}</dt>
                  <dd title={article.articleNumber || undefined}>{article.articleNumber || t("common.none")}</dd>
                </div>
                <div>
                  <dt>{t("accessories.table.type")}</dt>
                  <dd><strong>{typeLabel}</strong><small>{subtypeLabel}</small></dd>
                </div>
                <div>
                  <dt>{t("accessories.table.storage")}</dt>
                  <dd title={article.locationNames.join(", ")}>{primaryLocation}</dd>
                  {article.locationNames.length > 1 ? (
                    <small>{t("accessories.table.moreLocations", {
                      count: article.locationNames.length - 1
                    })}</small>
                  ) : null}
                </div>
              </dl>
              <div className="article-card-stock">
                <strong>{t("accessories.table.stockOwned", { count: article.owned })}</strong>
                <small>{t("accessories.table.stockBreakdown", {
                  available: article.available,
                  reserved: article.reserved,
                  installed: article.installed
                })}</small>
              </div>
              <ArticleActions article={article} canEdit={canEdit} onView={onView} onEdit={onEdit}
                onArchive={onArchive} onRestore={onRestore} />
            </div>
          </article>
        );
      })}
    </div>
  );
}
