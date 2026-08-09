import type { AccessoryArticleType, MasterDataEntry } from "../../shared/api";
import { isStandardArticleSubtype, masterDataDisplayLabel } from "../../shared/articleMasterDataLabels";

type Translate = (key: string, values?: Record<string, string | number>) => string;

function canonicalKey(entry: MasterDataEntry, articleType: AccessoryArticleType): string | null {
  const prefix = `${articleType}:`;
  return entry.key.startsWith(prefix) ? entry.key.slice(prefix.length) : null;
}

function canonicalValue(articleType: AccessoryArticleType, subtype: string): string {
  const prefix = `${articleType}:`;
  return subtype.startsWith(prefix) ? subtype.slice(prefix.length) : subtype;
}

export type ArticleSubtypeOption = {
  value: string;
  label: string;
  active: boolean;
};

export function articleSubtypeLabel(
  articleType: AccessoryArticleType,
  subtype: string,
  entries: readonly MasterDataEntry[],
  t: Translate
): string {
  const fullKey = subtype.startsWith(`${articleType}:`) ? subtype : `${articleType}:${subtype}`;
  const configured = entries.find((entry) => entry.key === fullKey);
  if (configured) return masterDataDisplayLabel(configured, t);
  if (isStandardArticleSubtype(fullKey)) {
    return t(`accessories.subtype.${articleType}.${canonicalValue(articleType, subtype)}`);
  }
  return subtype;
}

export function articleSubtypeOptions(
  articleType: AccessoryArticleType,
  currentSubtype: string,
  entries: readonly MasterDataEntry[],
  t: Translate
): ArticleSubtypeOption[] {
  const currentValue = canonicalValue(articleType, currentSubtype);
  const options = entries.flatMap((entry) => {
    const value = canonicalKey(entry, articleType);
    if (!value || (!entry.active && value !== currentValue)) return [];
    return [{
      value: value === currentValue ? currentSubtype : value,
      label: articleSubtypeLabel(articleType, value, entries, t),
      active: entry.active
    }];
  });
  if (currentSubtype && !options.some((option) => canonicalValue(articleType, option.value) === currentValue)) {
    options.push({ value: currentSubtype, label: articleSubtypeLabel(
      articleType, currentSubtype, entries, t
    ), active: false });
  }
  return options;
}
