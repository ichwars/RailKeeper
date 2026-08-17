import { AlertTriangle, ExternalLink } from "lucide-react";

import type { ArticleSearchResult, CreateVehicleRequest } from "../../shared/api";
import { articleFieldStatus, articleSelectionKey, imageSelectionKey } from "../../shared/articleSearch/articleSearchModel";
import { useI18n } from "../../shared/i18n";
import { articleFieldLabels, currentArticleValue, isArticleFieldKey } from "./articleSearch";
import type { ArticleSearchController } from "./useArticleSearchController";

const createReviewGroups = [
  { key: "identification", keys: ["manufacturer", "articleNumber", "ean", "name", "gauge", "category", "gattung", "series", "vehicleNumber", "articleSourceUrl"] },
  { key: "railway", keys: ["railwayCompany", "epoch"] },
  { key: "technical", keys: ["lengthMm", "weightG", "axles", "axleCount", "tractionTireCount", "adapter", "powerPickup", "digital", "decoderType"] },
  { key: "description", keys: ["description", "additionalInfo", "productionPeriod", "listPrice"] },
  { key: "media", keys: [] }
] as const;

export function VehicleCreateArticleReview({
  result,
  resultIndex,
  current,
  controller,
  onApply,
  onBack,
  onContinue
}: {
  result: ArticleSearchResult;
  resultIndex: number;
  current: CreateVehicleRequest;
  controller: ArticleSearchController;
  onApply: () => void;
  onBack: () => void;
  onContinue: () => void;
}) {
  const { t } = useI18n();
  const fieldRows = Object.entries(result.fields).filter(([key]) => isArticleFieldKey(key));
  return (
    <section className="vehicle-create-article-review">
      <div className="vehicle-wizard-section-head">
        <div><span>02</span><h3>{t("vehicles.wizard.reviewImport")}</h3></div>
        <a className="secondary-button" href={result.url} target="_blank" rel="noreferrer">
          <ExternalLink size={15} />{t("vehicles.articleSearch.sourceOpen")}
        </a>
      </div>
      <header className="vehicle-create-review-source">
        <strong>{result.title}</strong><span>{result.source}</span>
        <small>{fieldRows.length} {t("vehicles.wizard.fields")}</small>
      </header>
      <div className="vehicle-create-review-groups">
        {createReviewGroups.map((group) => {
          const rows = group.key === "media" ? [] : fieldRows.filter(([key]) => group.keys.includes(key as never));
          const conflicts = rows.filter(([key, field]) => articleFieldStatus(
            currentArticleValue(current, key as keyof CreateVehicleRequest), field.value
          ) === "conflict").length;
          if (rows.length === 0 && (group.key !== "media" || !result.images?.length)) return null;
          return (
            <details className="vehicle-create-review-group" key={group.key} open={group.key === "identification"}>
              <summary><span>{t(`vehicles.wizard.reviewGroup.${group.key}`)}</span>
                <small>{group.key === "media" ? result.images?.length : rows.length} {t("vehicles.wizard.fields")}</small>
                {conflicts > 0 && <em><AlertTriangle size={13} />{conflicts}</em>}</summary>
              {group.key === "media" ? (
                <div className="vehicle-create-review-images">
                  {result.images?.map((image) => (
                    <label key={image.url}><input type="checkbox"
                      checked={Boolean(controller.state.selectedImages[imageSelectionKey(result, image, resultIndex)])}
                      onChange={(event) => controller.commands.toggleImage(result, resultIndex, image, event.target.checked)} />
                      <img src={image.url} alt="" /></label>
                  ))}
                </div>
              ) : (
                <div className="vehicle-create-review-table">
                  {rows.map(([key, field]) => {
                    const selectionKey = articleSelectionKey(result, key, resultIndex);
                    const currentValue = currentArticleValue(current, key as keyof CreateVehicleRequest);
                    return (
                      <label className="vehicle-create-review-row" key={key}>
                        <input type="checkbox" checked={Boolean(controller.state.selectedFields[selectionKey])}
                          onChange={(event) => controller.commands.toggleField(result, resultIndex, key, event.target.checked)} />
                        <strong>{articleFieldLabels[key as keyof CreateVehicleRequest] || field.label}</strong>
                        <span><small>{t("vehicles.articleSearch.current")}</small>{currentValue || "–"}</span>
                        <span><small>{t("vehicles.articleSearch.found")}</small>{field.value}</span>
                      </label>
                    );
                  })}
                </div>
              )}
            </details>
          );
        })}
      </div>
      <div className="vehicle-create-review-actions">
        <button type="button" className="secondary-button" onClick={onBack}>{t("vehicles.wizard.backResults")}</button>
        <button type="button" className="secondary-button" onClick={onContinue}>{t("vehicles.wizard.continueWithoutImport")}</button>
        <button type="button" className="primary-button" onClick={onApply}>{t("vehicles.articleSearch.applySelected")}</button>
      </div>
    </section>
  );
}
