import type { AccessoryArticleType, MasterDataEntry } from "../../shared/api";
import { masterDataDisplayLabel } from "../../shared/articleMasterDataLabels";

type Translate = (key: string, values?: Record<string, string | number>) => string;

export const articleTypeOrder = [
  "track",
  "signal",
  "decoder",
  "electrical_control",
  "building_equipment",
  "landscape_consumable",
  "lighting",
  "other"
] as const satisfies readonly AccessoryArticleType[];

export type ArticleTypeOption = {
  value: AccessoryArticleType;
  label: string;
  active: boolean;
};

export function articleTypeLabel(
  articleType: AccessoryArticleType,
  entries: readonly MasterDataEntry[],
  t: Translate
): string {
  const configured = entries.find((entry) => entry.type === "article_type" && entry.key === articleType);
  return configured ? masterDataDisplayLabel(configured, t) : t(`accessories.articleType.${articleType}`);
}

export function articleTypeOptions(
  entries: readonly MasterDataEntry[],
  historicalType: AccessoryArticleType | null,
  t: Translate
): ArticleTypeOption[] {
  const byKey = new Map(entries
    .filter((entry) => entry.type === "article_type")
    .map((entry) => [entry.key, entry]));
  return articleTypeOrder.flatMap((articleType) => {
    const configured = byKey.get(articleType);
    if (!configured) {
      return historicalType === articleType
        ? [{ value: articleType, label: articleTypeLabel(articleType, entries, t), active: false }]
        : [];
    }
    if (!configured.active && historicalType !== articleType) return [];
    return [{
      value: articleType,
      label: articleTypeLabel(articleType, entries, t),
      active: configured.active
    }];
  });
}
