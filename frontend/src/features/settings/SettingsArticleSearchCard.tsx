import { AppChoiceInput } from "../../shared/ui/AppChoiceInput";
import { useI18n } from "../../shared/i18n";
import { Database } from "lucide-react";
import { articleSearchSourceOptions } from "./settingsModel";
import type { SettingsPreferencesController } from "./useSettingsPreferences";

export function SettingsArticleSearchCard({ preferences }: { preferences: SettingsPreferencesController }) {
  const { t } = useI18n();
  const {
    articleSearchEnabled,
    articleSearchSources,
    updateArticleSearchEnabled,
    updateArticleSearchSource
  } = preferences;

  return (
    <section className="panel settings-card settings-tool-card">
      <div className="settings-card-title">
        <Database size={17} />
        <h2>{t("settings.articleSearch.title")}</h2>
      </div>
      <label className="settings-toggle-row">
        <span>
          <strong>{t("settings.articleSearch.active")}</strong>
          <small>{t("settings.articleSearch.help")}</small>
        </span>
        <span className="switch-field">
          <input type="checkbox" checked={articleSearchEnabled} onChange={(event) => updateArticleSearchEnabled(event.target.checked)} />
          <span />
        </span>
      </label>
      <div className="article-source-grid" aria-label={t("settings.articleSearch.sources")}>
        {articleSearchSourceOptions.map((option) => (
          <label key={option.id} className="article-source-option">
            <AppChoiceInput
              type="checkbox"
              checked={articleSearchSources.includes(option.id)}
              onChange={(event) => updateArticleSearchSource(option.id, event.target.checked)}
              disabled={!articleSearchEnabled}
            />
            <span>
              <strong>{t(option.labelKey)}</strong>
              <small>{t(option.helpKey)}</small>
            </span>
          </label>
        ))}
      </div>
    </section>
  );
}
