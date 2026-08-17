import { ExternalLink } from "lucide-react";

import type { ArticleSearchResponse } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

export function VehicleCreateArticleResults({
  response,
  onSelect,
  onRevise
}: {
  response: ArticleSearchResponse;
  onSelect: (index: number) => void;
  onRevise: () => void;
}) {
  const { t } = useI18n();
  return (
    <section className="vehicle-create-article-results">
      <div className="vehicle-wizard-section-head">
        <div><span>02</span><h3>{t("vehicles.wizard.searchResults")}</h3></div>
        <div className="vehicle-create-results-heading-actions">
          <small>{response.query}</small>
          <button type="button" className="secondary-button" onClick={onRevise}>
            {t("vehicles.wizard.reviseSearch")}
          </button>
        </div>
      </div>
      <div className="vehicle-create-result-list">
        {response.results.map((result, index) => {
          const image = result.images?.[0];
          return (
            <article className={`vehicle-create-result-card ${image ? "has-image" : "without-image"}`}
              key={`${result.url}-${index}`}>
              {image && <img src={image.url} alt="" />}
              <div className="vehicle-create-result-copy">
                <h4>{result.title}</h4>
                <span>{result.source} · {Object.keys(result.fields).length} {t("vehicles.wizard.fields")} · {result.score}</span>
                {result.snippet && <p>{result.snippet}</p>}
                <small>{result.trace?.detailLoaded
                  ? t("vehicles.articleSearch.detailLoaded", { fields: result.trace.detailFields, images: result.trace.detailImages })
                  : t("vehicles.articleSearch.detailSkipped")}</small>
              </div>
              <div className="vehicle-create-result-actions">
                <a className="icon-button" href={result.url} target="_blank" rel="noreferrer"
                  aria-label={t("vehicles.articleSearch.sourceOpen")}><ExternalLink size={16} /></a>
                <button type="button" className="primary-button" onClick={() => onSelect(index)}
                  aria-label={t("vehicles.wizard.selectResult", { title: result.title })}>
                  {t("vehicles.wizard.select")}
                </button>
              </div>
            </article>
          );
        })}
      </div>
    </section>
  );
}
