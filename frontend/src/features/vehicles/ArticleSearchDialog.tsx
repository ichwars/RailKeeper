import { useCallback, useEffect, useState } from "react";
import { AlertTriangle, Check, ExternalLink, X } from "lucide-react";
import {
  ArticleSearchImage,
  ArticleSearchResponse,
  ArticleSearchResult,
  CreateVehicleRequest
} from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import {
  ArticleFieldKey,
  articleFieldGroups,
  articleFieldLabels,
  articleFieldStatus,
  articleResultKey,
  articleSelectionKey,
  currentArticleValue,
  imageSelectionKey,
  isArticleFieldKey,
  sourceDisplayName
} from "./articleSearch";

function previewImageUrl(image?: { url: string; thumbnailUrl?: string }) {
  return image?.thumbnailUrl || image?.url || "";
}

export function ArticleSearchDialog({
  form,
  loading,
  response,
  error,
  selectedFields,
  selectedImages,
  onApply,
  onClose,
  onToggleField,
  onToggleImage
}: {
  form: CreateVehicleRequest;
  loading: boolean;
  response: ArticleSearchResponse | null;
  error: string;
  selectedFields: Record<string, boolean>;
  selectedImages: Record<string, boolean>;
  onApply: (result: ArticleSearchResult) => void;
  onClose: () => void;
  onToggleField: (result: ArticleSearchResult, index: number, key: string, checked: boolean) => void;
  onToggleImage: (result: ArticleSearchResult, index: number, image: ArticleSearchImage, checked: boolean) => void;
}) {
  const { t } = useI18n();
  const [failedImages, setFailedImages] = useState<Record<string, boolean>>({});
  const articleFieldLabel = (key: string, fallback?: string) => {
    const label = t(`vehicle.field.${key}`);
    return label === `vehicle.field.${key}` ? fallback || articleFieldLabels[key as ArticleFieldKey] || key : label;
  };
  const articleGroupTitle = (title: string) => {
    if (title === "Modell") return t("vehicles.articleSearch.group.model");
    if (title === "Masse / Bauart") return t("vehicles.articleSearch.group.mass");
    if (title === "Technik") return t("vehicles.articleSearch.group.technology");
    if (title === "Weitere Daten") return t("vehicles.articleSearch.group.more");
    return title;
  };
  const sourceLabel = (source: string) => t(`settings.articleSearch.source.${source}`);
  const sources = response?.sources || [];
  const manufacturerDomains = response?.manufacturerDomains || [];
  const queries = response?.queries || [];

  useEffect(() => {
    setFailedImages({});
  }, [response?.query]);

  const markImageFailed = useCallback((url: string) => {
    setFailedImages((current) => current[url] ? current : { ...current, [url]: true });
  }, []);

  return (
    <div className="confirm-layer article-search-layer" role="dialog" aria-modal="true" aria-label={t("vehicles.articleSearch.dialogTitle")}>
      <section className="article-search-dialog">
        <div className="panel-head form-head">
          <div>
            <h2>{t("vehicles.articleSearch.dialogTitle")}</h2>
            <p>{response?.query ? t("vehicles.articleSearch.query", { query: response.query }) : t("vehicles.articleSearch.help")}</p>
            {response && (sources.length > 0 || manufacturerDomains.length > 0 || queries.length > 0) && (
              <div className="article-search-trace" aria-label={t("vehicles.articleSearch.trace")}>
                {sources.length > 0 && (
                  <div>
                    <span>{t("vehicles.articleSearch.traceSources")}</span>
                    <strong>{sources.map(sourceLabel).join(" / ")}</strong>
                  </div>
                )}
                {manufacturerDomains.length > 0 && (
                  <div>
                    <span>{t("vehicles.articleSearch.traceDomains")}</span>
                    <strong>{manufacturerDomains.join(" / ")}</strong>
                  </div>
                )}
                {queries.length > 0 && (
                  <details>
                    <summary>{t("vehicles.articleSearch.traceQueries", { count: queries.length })}</summary>
                    <ul>
                      {queries.map((item, index) => (
                        <li key={`${item.source}-${item.query}-${index}`}>
                          <span>{item.source}</span>
                          <code>{item.query}</code>
                        </li>
                      ))}
                    </ul>
                  </details>
                )}
              </div>
            )}
          </div>
          <button type="button" className="icon-button" onClick={onClose} aria-label={t("vehicles.close")} title={t("vehicles.close")}>
            <X size={17} />
          </button>
        </div>

        <div className="article-dialog-state">
          {loading && <p className="empty-state compact">{t("vehicles.articleSearch.loading")}</p>}
          {error && <p className="form-message">{error}</p>}
          {!loading && !error && response && response.results.length === 0 && (
            <p className="empty-state compact">{t("vehicles.articleSearch.empty")}</p>
          )}
        </div>

        <div className="article-result-list">
          {response?.results.map((result, index) => {
            const resultKey = articleResultKey(result, index);
            const selectableKeys = Object.keys(result.fields).filter(isArticleFieldKey);
            const visibleImages = (result.images || []).filter((image) => image.url && !failedImages[image.url]);
            const resultImage = visibleImages[0];
            return (
              <article key={resultKey} className="article-result-card article-result-table-card">
                <header>
                  {resultImage && (
                    <img className="article-result-thumb" src={previewImageUrl(resultImage)} alt="" onError={() => markImageFailed(resultImage.url)} />
                  )}
                  <div>
                    <strong>{result.title}</strong>
                    <span>{result.source} - {Object.keys(result.fields).length} Felder - Trefferwert {result.score}</span>
                    <span className={`article-detail-trace ${result.trace?.detailLoaded ? "loaded" : result.trace?.error ? "failed" : "skipped"}`}>
                      {result.trace?.detailLoaded
                        ? t("vehicles.articleSearch.detailLoaded", { fields: result.trace.detailFields, images: result.trace.detailImages })
                        : result.trace?.error
                          ? t("vehicles.articleSearch.detailFailed")
                          : t("vehicles.articleSearch.detailSkipped")}
                    </span>
                    {result.snippet && <p>{result.snippet}</p>}
                  </div>
                  <a className="secondary-button article-source-button" href={result.url} target="_blank" rel="noreferrer" aria-label={t("vehicles.articleSearch.sourceOpen")} title={t("vehicles.articleSearch.sourceOpen")}>
                    <ExternalLink size={15} />
                    {t("vehicles.articleSearch.sourceOpen")}
                  </a>
                </header>

                {visibleImages.length > 0 && (
                  <div className="article-image-strip" aria-label={t("vehicles.articleSearch.imagesFound")}>
                    {visibleImages.map((image) => {
                      const selectionKey = imageSelectionKey(result, image, index);
                      return (
                        <label key={image.url} className="article-image-option">
                          <input
                            type="checkbox"
                            checked={Boolean(selectedImages[selectionKey])}
                            onChange={(event) => onToggleImage(result, index, image, event.target.checked)}
                          />
                          <img
                            src={previewImageUrl(image)}
                            alt=""
                            onError={() => {
                              markImageFailed(image.url);
                              if (selectedImages[selectionKey]) {
                                onToggleImage(result, index, image, false);
                              }
                            }}
                          />
                        </label>
                      );
                    })}
                  </div>
                )}

                {result.conflicts && result.conflicts.length > 0 && (
                  <div className="conflict-note">
                    <AlertTriangle size={15} aria-hidden="true" />
                    {t("vehicles.articleSearch.conflicts", { fields: result.conflicts.map((key) => articleFieldLabel(key)).join(", ") })}
                  </div>
                )}

                <div className="article-field-groups">
                  {articleFieldGroups.map((group) => {
                    const rows = group.keys
                      .filter((key) => result.fields[key])
                      .map((key) => ({ key, field: result.fields[key] }));
                    if (rows.length === 0) return null;
                    return (
                      <section key={group.title} className="article-field-group">
                        <h3>{articleGroupTitle(group.title)}</h3>
                        <table>
                          <thead>
                            <tr>
                              <th>{t("vehicles.articleSearch.apply")}</th>
                              <th>{t("vehicles.articleSearch.field")}</th>
                              <th>{t("vehicles.articleSearch.current")}</th>
                              <th>{t("vehicles.articleSearch.found")}</th>
                              <th>{t("vehicles.articleSearch.status")}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {rows.map(({ key, field }) => {
                              const current = currentArticleValue(form, key);
                              const status = articleFieldStatus(current, field.value);
                              const foundDisplay = key === "articleSourceUrl" ? sourceDisplayName(field.value) : field.value;
                              const currentDisplay = key === "articleSourceUrl" && current ? sourceDisplayName(current) : current;
                              const selectionKey = articleSelectionKey(result, key, index);
                              return (
                                <tr key={key} className={status === "conflict" ? "conflict" : ""}>
                                  <td>
                                    <input
                                      type="checkbox"
                                      checked={Boolean(selectedFields[selectionKey])}
                                      onChange={(event) => onToggleField(result, index, key, event.target.checked)}
                                    />
                                  </td>
                                  <td><strong>{articleFieldLabel(key, field.label)}</strong></td>
                                  <td>{currentDisplay || "-"}</td>
                                  <td>
                                    {key === "articleSourceUrl" && field.value ? (
                                      <a className="inline-source-link" href={field.value} target="_blank" rel="noreferrer" title={field.value}>
                                        {foundDisplay || t("vehicles.articleSearch.source")}
                                        <ExternalLink size={13} aria-hidden="true" />
                                      </a>
                                    ) : (
                                      foundDisplay || "-"
                                    )}
                                  </td>
                                  <td><span className={`article-status ${status}`}>{t(`vehicles.articleSearch.status.${status}`)}</span></td>
                                </tr>
                              );
                            })}
                          </tbody>
                        </table>
                      </section>
                    );
                  })}
                </div>

                <footer>
                  <span>{t("vehicles.articleSearch.selectableFields", { count: selectableKeys.length })}</span>
                  <button type="button" className="primary-button" onClick={() => onApply(result)}>
                    <Check size={16} aria-hidden="true" />
                    {t("vehicles.articleSearch.applySelected")}
                  </button>
                </footer>
              </article>
            );
          })}
        </div>
      </section>
    </div>
  );
}
