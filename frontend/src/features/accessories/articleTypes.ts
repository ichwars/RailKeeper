import type { AccessoryArticleType, MasterDataEntry } from "../../shared/api";

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

const seededLabels: Record<AccessoryArticleType, string> = {
  track: "Track",
  signal: "Signal",
  decoder: "Decoder",
  electrical_control: "Electrical control",
  building_equipment: "Building equipment",
  landscape_consumable: "Landscape consumable",
  lighting: "Lighting",
  other: "Other"
};

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
  const configuredLabel = configured?.label.trim();
  if (!configuredLabel || configuredLabel === seededLabels[articleType]) {
    return t(`accessories.articleType.${articleType}`);
  }
  return configuredLabel;
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
