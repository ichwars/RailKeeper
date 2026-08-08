import type { AccessoryArticleListItem } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleMetrics } from "./ArticleMetrics";
import { ArticleEditorDialog } from "./ArticleEditorDialog";
import { ArticleOverviewHeader } from "./ArticleOverviewHeader";
import { ArticleTable } from "./ArticleTable";
import { ArticleToolbar } from "./ArticleToolbar";
import { useArticleOverview } from "./useArticleOverview";
import { useArticleEditorController } from "./useArticleEditorController";

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
  const editor = useArticleEditorController({ roles, onSaved: overview.reload });
  const { t } = useI18n();

  if (!canRead) {
    return <section className="panel"><p>{t("accessories.overview.noAccess")}</p></section>;
  }

  const openArticle = (article: AccessoryArticleListItem, mode: ArticleOpenMode) => {
    if (onOpenArticle) onOpenArticle(article.id, mode);
    else editor.openArticle(article.id, mode, article.hasUsageHistory);
  };
  const createArticle = onCreateArticle || editor.openCreate;

  const isFirstLoad = overview.loading && overview.data.items.length === 0;
  const hasNoArticles = overview.data.items.length === 0 &&
    overview.data.metrics.articleCount === 0 && !overview.hasActiveFilters;
  const hasNoResults = overview.data.items.length === 0 && !hasNoArticles;
  const showErrorOnly = Boolean(overview.error) && overview.data.items.length === 0;

  return (
    <>
      <ArticleOverviewHeader canEdit={canEdit} onCreate={createArticle} />
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
            {canEdit ? (
              <button type="button" className="primary-button" onClick={createArticle}>
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
            onView={(article) => openArticle(article, "view")}
            onEdit={(article) => openArticle(article, "edit")}
            onArchive={(article) => overview.archiveArticle(article.id)}
            onRestore={(article) => overview.restoreArticle(article.id)}
          />
        )}
      </section>
      {editor.isOpen ? <ArticleEditorDialog
        mode={editor.mode}
        form={editor.form}
        article={editor.article}
        activeTab={editor.activeTab}
        hasUsageHistory={editor.hasUsageHistory}
        saving={editor.saving}
        loading={editor.loading}
        error={editor.error}
        fieldErrors={editor.fieldErrors}
        tabErrors={editor.tabErrors}
        duplicateCandidates={editor.duplicateCandidates}
        closeConfirmationOpen={editor.closeConfirmationOpen}
        permissions={editor.permissions}
        resources={editor.resources}
        returnFocusTo={editor.returnFocusTo}
        onChange={editor.changeForm}
        onTabChange={editor.setActiveTab}
        onSubmit={editor.submit}
        onRequestClose={editor.requestClose}
        onConfirmClose={editor.confirmClose}
        onCancelClose={editor.cancelClose}
        onConfirmDuplicate={editor.confirmDuplicateSave}
        onCancelDuplicate={editor.cancelDuplicateSave}
        onResourcesChanged={editor.refreshResources}
      /> : null}
    </>
  );
}
