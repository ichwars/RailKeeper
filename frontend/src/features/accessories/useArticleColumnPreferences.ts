import { useCallback } from "react";

import { translate, useI18n } from "../../shared/i18n";
import { setTableColumnWidth } from "../../shared/tableColumnLayout";
import { useProfileTableLayout } from "../../shared/useProfileTableLayout";
import {
  articleTableColumnSettingKey,
  articleTableColumnWidthDefinitions,
  defaultArticleTableColumns,
  defaultArticleTableLayout,
  moveArticleTableColumn,
  parseArticleTableLayout,
  serializeArticleTableLayout,
  toggleArticleTableColumn,
  type ArticleColumnMove,
  type ArticleTableColumn
} from "./articleTableColumns";

export function useArticleColumnPreferences(onMessage: (message: string) => void) {
  const { language } = useI18n();
  const preferences = useProfileTableLayout({
    settingKey: articleTableColumnSettingKey,
    defaultLayout: defaultArticleTableLayout,
    parse: parseArticleTableLayout,
    serialize: serializeArticleTableLayout,
    legacyValue: () => window.localStorage.getItem(articleTableColumnSettingKey) ?? undefined,
    onLoadError: () => onMessage(translate(language, "accessories.columns.loadError")),
    onSaveError: () => onMessage(translate(language, "accessories.columns.saveError"))
  });

  const applyColumns = useCallback((
    update: (current: ArticleTableColumn[]) => ArticleTableColumn[]
  ) => {
    preferences.commit((current) => ({
      ...current,
      columns: update(current.columns)
    }));
  }, [preferences]);

  const toggleColumn = useCallback((column: ArticleTableColumn) => {
    applyColumns((current) => toggleArticleTableColumn(current, column));
  }, [applyColumns]);

  const moveColumn = useCallback((column: ArticleTableColumn, direction: ArticleColumnMove) => {
    applyColumns((current) => moveArticleTableColumn(current, column, direction));
  }, [applyColumns]);

  const resetColumns = useCallback(() => {
    preferences.commit(() => ({ columns: [...defaultArticleTableColumns], widths: {} }));
  }, [preferences]);

  const previewColumnWidth = useCallback((column: ArticleTableColumn, width: number) => {
    preferences.preview((current) => setTableColumnWidth(
      current,
      column,
      width,
      articleTableColumnWidthDefinitions
    ));
  }, [preferences]);

  const commitColumnWidth = useCallback((column: ArticleTableColumn, width: number) => {
    preferences.commit((current) => setTableColumnWidth(
      current,
      column,
      width,
      articleTableColumnWidthDefinitions
    ));
  }, [preferences]);

  return {
    columns: preferences.layout.columns,
    widths: preferences.layout.widths,
    loading: preferences.loading,
    commitColumnWidth,
    moveColumn,
    previewColumnWidth,
    resetColumns,
    toggleColumn
  };
}
