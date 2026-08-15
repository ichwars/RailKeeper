import { useEffect, useState } from "react";

import { api, type AccessoryArticleListItem, type MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleCardGrid } from "./ArticleCardGrid";
import { ArticleCompactList } from "./ArticleCompactList";
import { ArticleEditorDialog } from "./ArticleEditorDialog";
import { ArticleMetrics } from "./ArticleMetrics";
import { ArticleOverviewHeader } from "./ArticleOverviewHeader";
import { ArticleTable } from "./ArticleTable";
import { ArticleToolbar } from "./ArticleToolbar";
import {
  persistArticleTableColumns,
  resetArticleTableColumns,
  storedArticleTableColumns,
  toggleArticleTableColumn,
  type ArticleTableColumn
} from "./articleTableColumns";
import { AccessoryConfirmDialog } from "./AccessoryConfirmDialog";
import { persistArticleViewMode, storedArticleViewMode, type ArticleViewMode } from "./articleViewMode";
import { useArticleOverview } from "./useArticleOverview";
import { useArticleEditorController } from "./useArticleEditorController";
import { useArticleCoreMasterData } from "./useArticleCoreMasterData";
import { useAccessoryArticleSearchController } from "./useAccessoryArticleSearchController";

export type ArticleOpenMode = "view" | "edit";

const compactOverviewQuery = "(max-width: 900px)";

function useCompactArticleOverview() {
  const [compact, setCompact] = useState(() =>
    typeof window.matchMedia === "function" && window.matchMedia(compactOverviewQuery).matches);

  useEffect(() => {
    if (typeof window.matchMedia !== "function") return;
    const mediaQuery = window.matchMedia(compactOverviewQuery);
    const update = () => setCompact(mediaQuery.matches);
    update();
    mediaQuery.addEventListener("change", update);
    return () => mediaQuery.removeEventListener("change", update);
  }, []);

  return compact;
}

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
  const canDelete = roles.includes("Admin");
  const overview = useArticleOverview({ enabled: canRead });
  const editor = useArticleEditorController({ roles, onSaved: overview.reload });
  const coreMasterData = useArticleCoreMasterData(editor.isOpen);
  const articleSearch = useAccessoryArticleSearchController({
    form: editor.form,
    readOnly: editor.isFormReadOnly,
    pendingImageCount: editor.pendingArticleImages.length,
    manufacturers: coreMasterData.manufacturers,
    gauges: coreMasterData.gauges,
    updateForm: editor.changeForm,
    addImages: editor.addPendingArticleImages
  });
  const [subtypeEntries, setSubtypeEntries] = useState<MasterDataEntry[]>([]);
  const [articleTypeEntries, setArticleTypeEntries] = useState<MasterDataEntry[]>([]);
  const [selectedArticleIDs, setSelectedArticleIDs] = useState<Set<string>>(new Set());
  const [pendingDeleteArticle, setPendingDeleteArticle] =
    useState<AccessoryArticleListItem | null>(null);
  const [viewMode, setViewMode] = useState<ArticleViewMode>(storedArticleViewMode);
  const [visibleColumns, setVisibleColumns] = useState(storedArticleTableColumns);
  const compactOverview = useCompactArticleOverview();
  const { t } = useI18n();

  useEffect(() => {
    if (!canRead) return;
    let active = true;
    void api.masterData("accessory_subtype").then((entries) => {
      if (active) setSubtypeEntries(entries);
    }).catch(() => undefined);
    void api.masterData("article_type").then((entries) => {
      if (active) setArticleTypeEntries(entries);
    }).catch(() => undefined);
    return () => { active = false; };
  }, [canRead]);

  useEffect(() => {
    const visibleIDs = new Set(overview.data.items.map((item) => item.id));
    setSelectedArticleIDs((current) => {
      const next = new Set([...current].filter((id) => visibleIDs.has(id)));
      return next.size === current.size ? current : next;
    });
  }, [overview.data.items]);

  if (!canRead) {
    return <section className="panel"><p>{t("accessories.overview.noAccess")}</p></section>;
  }

  const openArticle = (article: AccessoryArticleListItem, mode: ArticleOpenMode) => {
    if (onOpenArticle) onOpenArticle(article.id, mode);
    else editor.openArticle(article.id, mode, article.hasUsageHistory);
  };
  const createArticle = onCreateArticle || editor.openCreate;
  const changeViewMode = (mode: ArticleViewMode) => {
    setViewMode(mode);
    persistArticleViewMode(mode);
  };
  const toggleColumn = (column: ArticleTableColumn) => setVisibleColumns((current) => {
    const next = toggleArticleTableColumn(current, column);
    persistArticleTableColumns(next);
    return next;
  });
  const resetColumns = () => {
    const next = resetArticleTableColumns();
    persistArticleTableColumns(next);
    setVisibleColumns(next);
  };
  const toggleSelection = (id: string) => setSelectedArticleIDs((current) => {
    const next = new Set(current);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    return next;
  });
  const toggleAll = () => setSelectedArticleIDs((current) => {
    const visibleIDs = overview.data.items.map((item) => item.id);
    const allSelected = visibleIDs.length > 0 && visibleIDs.every((id) => current.has(id));
    return allSelected ? new Set() : new Set(visibleIDs);
  });

  const isFirstLoad = overview.loading && overview.data.items.length === 0;
  const hasNoArticles = overview.data.items.length === 0 &&
    overview.data.metrics.articleCount === 0 && !overview.hasActiveFilters;
  const hasNoResults = overview.data.items.length === 0 && !hasNoArticles;
  const showErrorOnly = Boolean(overview.error) && overview.data.items.length === 0;

  return (
    <>
      <ArticleOverviewHeader canEdit={canEdit} onCreate={createArticle} />
      {!canEdit ? <p className="article-read-only-note">{t(roles.includes("Planner")
        ? "accessories.overview.plannerReadOnly" : "accessories.overview.readOnly")}</p> : null}
      <ArticleMetrics
        metrics={overview.data.metrics}
        activeStatus={overview.filters.status}
        onReset={overview.resetFilters}
        onStatusChange={(status) => overview.setFilter("status", status)}
      />
      <section className="panel inventory-panel article-overview-panel"
        aria-label={t("accessories.overview.listTitle")} aria-busy={overview.loading}>
        <div className="panel-head inventory-list-head article-list-head">
          <div>
            <h2>{t("accessories.overview.listTitle")}</h2>
            <p>{t("accessories.overview.listSubtitle")}</p>
          </div>
          <ArticleToolbar
            filters={overview.filters}
            options={overview.data.filters}
            articleTypeEntries={articleTypeEntries}
            resultCount={overview.data.items.length}
            viewMode={viewMode}
            hasActiveFilters={overview.hasActiveFilters}
            onViewModeChange={changeViewMode}
            visibleColumns={visibleColumns}
            onToggleColumn={toggleColumn}
            onResetColumns={resetColumns}
            onFilterChange={overview.setFilter}
            onReset={overview.resetFilters}
          />
        </div>
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
          <>
            {compactOverview ? (
              <ArticleCompactList
                items={overview.data.items}
                articleTypeEntries={articleTypeEntries}
                subtypeEntries={subtypeEntries}
                canEdit={canEdit}
                canDelete={canDelete}
                onView={(article) => openArticle(article, "view")}
                onEdit={(article) => openArticle(article, "edit")}
                onArchive={(article) => overview.archiveArticle(article.id)}
                onRestore={(article) => overview.restoreArticle(article.id)}
                onDelete={setPendingDeleteArticle}
              />
            ) : (
              <div className="article-desktop-content">
                {viewMode === "cards" ? (
                  <ArticleCardGrid
                    items={overview.data.items}
                    articleTypeEntries={articleTypeEntries}
                    subtypeEntries={subtypeEntries}
                    canEdit={canEdit}
                    canDelete={canDelete}
                    onView={(article) => openArticle(article, "view")}
                    onEdit={(article) => openArticle(article, "edit")}
                    onArchive={(article) => overview.archiveArticle(article.id)}
                    onRestore={(article) => overview.restoreArticle(article.id)}
                    onDelete={setPendingDeleteArticle}
                  />
                ) : (
                  <ArticleTable
                    items={overview.data.items}
                    articleTypeEntries={articleTypeEntries}
                    subtypeEntries={subtypeEntries}
                    sort={overview.sort}
                    direction={overview.direction}
                    canEdit={canEdit}
                    canDelete={canDelete}
                    visibleColumns={visibleColumns}
                    selectedIDs={selectedArticleIDs}
                    onToggleSelection={toggleSelection}
                    onToggleAll={toggleAll}
                    onSort={overview.setSort}
                    onView={(article) => openArticle(article, "view")}
                    onEdit={(article) => openArticle(article, "edit")}
                    onArchive={(article) => overview.archiveArticle(article.id)}
                    onRestore={(article) => overview.restoreArticle(article.id)}
                    onDelete={setPendingDeleteArticle}
                  />
                )}
              </div>
            )}
          </>
        )}
      </section>
      {editor.isOpen ? <ArticleEditorDialog
        key={editor.sessionKey}
        mode={editor.mode}
        form={editor.form}
        article={editor.article}
        activeTab={editor.activeTab}
        hasUsageHistory={editor.hasUsageHistory}
        saving={editor.saving}
        loading={editor.loading}
        error={editor.error}
        resourceError={editor.resourceError}
        fieldErrors={editor.fieldErrors}
        tabErrors={editor.tabErrors}
        subjectFieldErrors={editor.subjectFieldErrors}
        customFields={editor.customFields}
        customFieldsLoading={editor.customFieldsLoading}
        customFieldsError={editor.customFieldsError}
        articleTypeEntries={editor.articleTypeEntries}
        articleTypeEntriesLoading={editor.articleTypeEntriesLoading}
        articleTypeEntriesError={editor.articleTypeEntriesError}
        subtypeEntries={editor.subtypeEntries}
        subtypeEntriesLoading={editor.subtypeEntriesLoading}
        subtypeEntriesError={editor.subtypeEntriesError}
        manufacturerEntries={coreMasterData.manufacturers}
        gaugeEntries={coreMasterData.gauges}
        stockUnitEntries={coreMasterData.stockUnits}
        coreMasterDataLoading={coreMasterData.loading}
        coreMasterDataError={coreMasterData.error}
        duplicateCandidates={editor.duplicateCandidates}
        closeConfirmationOpen={editor.closeConfirmationOpen}
        permissions={editor.permissions}
        resources={editor.resources}
        resourcesStale={editor.resourcesStale}
        returnFocusTo={editor.returnFocusTo}
        articleSearch={articleSearch}
        onChange={editor.changeForm}
        onTabChange={editor.setActiveTab}
        onSubmit={editor.submit}
        onRequestClose={editor.requestClose}
        onConfirmClose={editor.confirmClose}
        onCancelClose={editor.cancelClose}
        onConfirmDuplicate={editor.confirmDuplicateSave}
        onCancelDuplicate={editor.cancelDuplicateSave}
        onResourcesChanged={editor.refreshResources}
        onRetryResources={editor.retryResources}
        onRetryCustomFields={editor.retryCustomFields}
        onRetryArticleTypeEntries={editor.retryArticleTypeEntries}
        onRetrySubtypeEntries={editor.retrySubtypeEntries}
        onRetryCoreMasterData={coreMasterData.retry}
        onSubdraftDirty={editor.setSubdraftDirty}
      /> : null}
      <AccessoryConfirmDialog
        action={pendingDeleteArticle ? {
          title: t("accessories.delete.title"),
          body: t("accessories.delete.body", {
            inventoryNumber: pendingDeleteArticle.inventoryNumber,
            name: pendingDeleteArticle.name
          }),
          confirmLabel: t("accessories.delete.confirm"),
          dangerous: true,
          run: () => overview.deleteArticle(pendingDeleteArticle.id)
        } : null}
        onClose={() => setPendingDeleteArticle(null)}
      />
    </>
  );
}
