import type { AccessoryArticleListItem } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleMetrics } from "./ArticleMetrics";
import { ArticleOverviewHeader } from "./ArticleOverviewHeader";
import { ArticleTable } from "./ArticleTable";
import { ArticleToolbar } from "./ArticleToolbar";
import { useArticleOverview } from "./useArticleOverview";

export type ArticleOpenMode = "view" | "edit";

type AccessoriesViewProps = {
  roles: string[];
  onCreateArticle?: () => void;
  onOpenArticle?: (id: string, mode: ArticleOpenMode) => void;
};

export function AccessoriesView({
  roles,
  onCreateArticle,
  onOpenArticle
}: AccessoriesViewProps) {
  const canRead = roles.some((role) => ["Admin", "Editor", "Viewer", "Planner"].includes(role));
  const canEdit = roles.includes("Admin") || roles.includes("Editor");
  const overview = useArticleOverview({ enabled: canRead });
  const { t } = useI18n();

  if (!canRead) {
    return <section className="panel"><p>{t("accessories.overview.noAccess")}</p></section>;
  }

  const openArticle = (article: AccessoryArticleListItem, mode: ArticleOpenMode) => {
    onOpenArticle?.(article.id, mode);
  };

  const isFirstLoad = overview.loading && overview.data.items.length === 0;
  const hasNoArticles = overview.data.items.length === 0 &&
    overview.data.metrics.articleCount === 0 && !overview.hasActiveFilters;
  const hasNoResults = overview.data.items.length === 0 && !hasNoArticles;
  const showErrorOnly = Boolean(overview.error) && overview.data.items.length === 0;

  return (
    <>
      <ArticleOverviewHeader canEdit={canEdit} onCreate={onCreateArticle} />
      {!canEdit ? <p className="article-read-only-note">{t("accessories.overview.readOnly")}</p> : null}
      <ArticleMetrics
        metrics={overview.data.metrics}
        activeStatus={overview.filters.status}
        onReset={overview.resetFilters}
        onStatusChange={(status) => overview.setFilter("status", status)}
      />
      <section className="panel article-overview-panel" aria-busy={overview.loading}>
        <div className="panel-head article-list-head">
          <div>
            <h2>{t("accessories.overview.listTitle")}</h2>
            <p>{t("accessories.overview.listSubtitle")}</p>
          </div>
        </div>
        <ArticleToolbar
          filters={overview.filters}
          options={overview.data.filters}
          resultCount={overview.data.items.length}
          hasActiveFilters={overview.hasActiveFilters}
          onFilterChange={overview.setFilter}
          onReset={overview.resetFilters}
        />
        {overview.error ? <p className="form-message" role="alert">{overview.error}</p> : null}
        {showErrorOnly ? null : isFirstLoad ? (
          <p className="empty-state">{t("accessories.overview.loading")}</p>
        ) : hasNoArticles ? (
          <div className="empty-state article-empty-state">
            <p>{t("accessories.overview.empty")}</p>
            {canEdit && onCreateArticle ? (
              <button type="button" className="primary-button" onClick={onCreateArticle}>
                {t("accessories.overview.createFirst")}
              </button>
            ) : null}
          </div>
        ) : hasNoResults ? (
          <div className="empty-state article-empty-state">
            <p>{t("accessories.overview.noResults")}</p>
            {overview.hasActiveFilters ? (
              <button type="button" className="secondary-button" onClick={overview.resetFilters}>
                {t("accessories.toolbar.reset")}
              </button>
            ) : null}
          </div>
        ) : (
          <ArticleTable
            items={overview.data.items}
            sort={overview.sort}
            direction={overview.direction}
            canEdit={canEdit}
            onSort={overview.setSort}
            onView={onOpenArticle ? (article) => openArticle(article, "view") : undefined}
            onEdit={onOpenArticle ? (article) => openArticle(article, "edit") : undefined}
            onArchive={(article) => overview.archiveArticle(article.id)}
            onRestore={(article) => overview.restoreArticle(article.id)}
          />
        )}
      </section>
    </>
  );
}
