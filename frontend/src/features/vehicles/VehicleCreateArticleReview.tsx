import {
  AlertTriangle,
  Barcode,
  Building2,
  CalendarRange,
  ChevronDown,
  CircleCheck,
  CircleDot,
  ExternalLink,
  Factory,
  FileText,
  Hash,
  Image as ImageIcon,
  Link,
  ListChecks,
  Pencil,
  Ruler,
  Settings,
  Tag,
  TrainFront,
  type LucideIcon
} from "lucide-react";

import type { ArticleSearchResult, CreateVehicleRequest } from "../../shared/api";
import { articleFieldStatus, articleSelectionKey, imageSelectionKey } from "../../shared/articleSearch/articleSearchModel";
import { useI18n } from "../../shared/i18n";
import { AppSelect } from "../../shared/ui/AppSelect";
import { articleFieldLabels, currentArticleValue, isArticleFieldKey } from "./articleSearch";
import type { ArticleSearchController } from "./useArticleSearchController";

const createReviewGroups = [
  { key: "identification", keys: ["manufacturer", "articleNumber", "ean", "name", "gauge", "category", "gattung", "series", "vehicleNumber", "articleSourceUrl"] },
  { key: "railway", keys: ["railwayCompany", "epoch"] },
  { key: "technical", keys: ["lengthMm", "weightG", "axles", "axleCount", "tractionTireCount", "adapter", "powerPickup", "digital", "decoderType"] },
  { key: "description", keys: ["description", "additionalInfo", "productionPeriod", "listPrice"] },
  { key: "media", keys: [] }
] as const;

const reviewGroupIcons: Record<(typeof createReviewGroups)[number]["key"], LucideIcon> = {
  identification: FileText,
  railway: TrainFront,
  technical: Settings,
  description: Pencil,
  media: ImageIcon
};

const reviewFieldIcons: Record<string, LucideIcon> = {
  manufacturer: Factory,
  articleNumber: Hash,
  ean: Barcode,
  name: Tag,
  gauge: Ruler,
  category: TrainFront,
  gattung: TrainFront,
  series: ListChecks,
  vehicleNumber: Hash,
  articleSourceUrl: Link,
  railwayCompany: Building2,
  epoch: CalendarRange
};

function sourceHost(url: string) {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return url;
  }
}

export function selectedArticleReviewCount(
  result: ArticleSearchResult,
  resultIndex: number,
  selectedFields: Record<string, boolean>,
  selectedImages: Record<string, boolean>
) {
  const fields = Object.keys(result.fields).filter((key) => (
    isArticleFieldKey(key) && selectedFields[articleSelectionKey(result, key, resultIndex)]
  )).length;
  const images = (result.images || []).filter((image) => (
    selectedImages[imageSelectionKey(result, image, resultIndex)]
  )).length;
  return fields + images;
}

export function VehicleCreateArticleReview({
  result,
  resultIndex,
  current,
  controller,
  memberCount = 0,
  imageOwners = {},
  onAssignImage = () => undefined
}: {
  result: ArticleSearchResult;
  resultIndex: number;
  current: CreateVehicleRequest;
  controller: ArticleSearchController;
  memberCount?: number;
  imageOwners?: Record<string, number>;
  onAssignImage?: (imageURL: string, memberIndex: number) => void;
}) {
  const { t } = useI18n();
  const fieldRows = Object.entries(result.fields).filter(([key]) => isArticleFieldKey(key));
  const conflictCount = fieldRows.filter(([key, field]) => articleFieldStatus(
    currentArticleValue(current, key as keyof CreateVehicleRequest), field.value
  ) === "conflict").length;
  const selectedCount = selectedArticleReviewCount(
    result,
    resultIndex,
    controller.state.selectedFields,
    controller.state.selectedImages
  );
  return (
    <section className="vehicle-create-article-review">
      <div className="vehicle-wizard-section-head vehicle-create-review-heading">
        <div><h3>{t("vehicles.wizard.reviewImport")}</h3></div>
      </div>
      <header className="vehicle-create-review-source">
        <span className="vehicle-create-review-source-type">{result.source}</span>
        <span className="vehicle-create-review-source-copy">
          <strong>{result.title}</strong>
          <small>{sourceHost(result.url)}</small>
        </span>
        <a className="vehicle-create-review-source-link" href={result.url} target="_blank" rel="noreferrer">
          <ExternalLink size={14} />{t("vehicles.articleSearch.sourceOpen")}
        </a>
      </header>
      <div className="vehicle-create-review-stats">
        <span><ListChecks size={14} />{t("vehicles.wizard.reviewFound", { count: fieldRows.length })}</span>
        <span><CircleCheck size={14} />{t("vehicles.wizard.reviewSelected", { count: selectedCount })}</span>
        {conflictCount > 0 && <span className="warning"><AlertTriangle size={14} />
          {t(conflictCount === 1 ? "vehicles.wizard.reviewDeviation" : "vehicles.wizard.reviewDeviations", {
            count: conflictCount
          })}</span>}
      </div>
      <div className="vehicle-create-review-groups">
        {createReviewGroups.map((group) => {
          const GroupIcon = reviewGroupIcons[group.key];
          const rows = group.key === "media" ? [] : fieldRows.filter(([key]) => group.keys.includes(key as never));
          const conflicts = rows.filter(([key, field]) => articleFieldStatus(
            currentArticleValue(current, key as keyof CreateVehicleRequest), field.value
          ) === "conflict").length;
          if (rows.length === 0 && (group.key !== "media" || !result.images?.length)) return null;
          return (
            <details className="vehicle-create-review-group" key={group.key} open={group.key === "identification"}>
              <summary><GroupIcon size={14} /><span>{t(`vehicles.wizard.reviewGroup.${group.key}`)}</span>
                <small>{group.key === "media"
                  ? t("vehicles.wizard.reviewFiles", { count: result.images?.length || 0 })
                  : t("vehicles.wizard.reviewFields", { count: rows.length })}</small>
                {conflicts > 0 ? <em>{t(conflicts === 1
                  ? "vehicles.wizard.reviewDeviation" : "vehicles.wizard.reviewDeviations", { count: conflicts })}</em>
                  : <span className="vehicle-create-review-conflict-placeholder" aria-hidden="true" />}
                <ChevronDown className="vehicle-create-review-chevron" size={14} /></summary>
              {group.key === "media" ? (
                <div className="vehicle-create-review-images">
                  {result.images?.map((image, imageIndex) => {
                    const selectionKey = imageSelectionKey(result, image, resultIndex);
                    const defaultMemberIndex = memberCount > 0 ? imageIndex % memberCount : 0;
                    return (
                      <div className="vehicle-create-review-image" key={image.url}>
                        <label><input type="checkbox"
                          checked={Boolean(controller.state.selectedImages[selectionKey])}
                          onChange={(event) => {
                            controller.commands.toggleImage(result, resultIndex, image, event.target.checked);
                            if (event.target.checked && imageOwners[image.url] === undefined && memberCount > 0) {
                              onAssignImage(image.url, defaultMemberIndex);
                            }
                          }} />
                          <img src={image.url} alt="" /></label>
                        {memberCount > 0 && (
                          <AppSelect aria-label={t("vehicles.wizard.assignSetImage")}
                            value={String(imageOwners[image.url] ?? defaultMemberIndex)}
                            onChange={(event) => onAssignImage(image.url, Number(event.target.value))}>
                            {Array.from({ length: memberCount }, (_, memberIndex) => (
                              <option key={memberIndex} value={memberIndex}>
                                {t("vehicles.wizard.memberLabel", { count: memberIndex + 1 })}
                              </option>
                            ))}
                          </AppSelect>
                        )}
                      </div>
                    );
                  })}
                </div>
              ) : (
                <div className="vehicle-create-review-table">
                  <div className="vehicle-create-review-table-head" aria-hidden="true">
                    <span>{t("vehicles.articleSearch.field")}</span>
                    <span>{t("vehicles.wizard.reviewFoundValue")}</span>
                    <span>{t("vehicles.wizard.reviewApply")}</span>
                  </div>
                  {rows.map(([key, field]) => {
                    const selectionKey = articleSelectionKey(result, key, resultIndex);
                    const currentValue = currentArticleValue(current, key as keyof CreateVehicleRequest);
                    const FieldIcon = reviewFieldIcons[key] || CircleDot;
                    const status = articleFieldStatus(currentValue, field.value);
                    return (
                      <label className="vehicle-create-review-row" key={key} data-status={status}
                        title={status === "conflict" ? `${t("vehicles.articleSearch.current")}: ${currentValue}` : undefined}>
                        <span className="vehicle-create-review-field"><FieldIcon size={13} />
                          {articleFieldLabels[key as keyof CreateVehicleRequest] || field.label}</span>
                        <span className="vehicle-create-review-value">{field.value}</span>
                        <input type="checkbox" checked={Boolean(controller.state.selectedFields[selectionKey])}
                          aria-label={`${articleFieldLabels[key as keyof CreateVehicleRequest] || field.label}: ${field.value}`}
                          onChange={(event) => controller.commands.toggleField(result, resultIndex, key, event.target.checked)} />
                      </label>
                    );
                  })}
                </div>
              )}
            </details>
          );
        })}
      </div>
    </section>
  );
}
