import type {
  ArticleSearchInput,
  ArticleSearchResult,
  MasterDataEntry
} from "../../shared/api";
import {
  articleSelectionKey,
  isBadCommonArticleValue,
  type ArticleSearchFieldGroup,
  type Translate
} from "../../shared/articleSearch/articleSearchModel";
import { articleSearchSources } from "../../shared/articleSearch/articleSearchPreferences";
import type { ArticleEditorForm } from "./articleEditorModel";

const directFieldMap = {
  manufacturer: "manufacturer",
  articleNumber: "articleNumber",
  name: "name",
  ean: "ean",
  scale: "scale",
  description: "description",
  articleSourceUrl: "productUrl"
} as const;

type AccessorySearchFieldKey = keyof typeof directFieldMap | "gauge";

const searchFieldKeys = new Set<AccessorySearchFieldKey>([
  ...Object.keys(directFieldMap) as Array<keyof typeof directFieldMap>,
  "gauge"
]);

function compact(value: unknown) {
  return String(value ?? "").trim();
}

function knownActiveLabel(entries: MasterDataEntry[], value: string) {
  const normalized = value.trim().toLocaleLowerCase("de-DE");
  return entries.find((entry) => entry.active && (
    entry.label.toLocaleLowerCase("de-DE") === normalized ||
    entry.key.toLocaleLowerCase("de-DE") === normalized
  ))?.label;
}

export function isSelectableAccessorySearchValue(
  key: string,
  value: string,
  manufacturers: MasterDataEntry[],
  gauges: MasterDataEntry[]
) {
  if (!isUsableAccessorySearchValue(key, value)) return false;
  if (key === "manufacturer") return Boolean(knownActiveLabel(manufacturers, value));
  if (key === "gauge") return Boolean(knownActiveLabel(gauges, value));
  return true;
}

export function accessorySearchInput(form: ArticleEditorForm): ArticleSearchInput {
  const fields = Object.fromEntries(Object.entries({
    ean: form.ean,
    scale: form.scale,
    description: form.description,
    articleSourceUrl: form.productUrl
  }).map(([key, value]) => [key, value.trim()]).filter(([, value]) => value));
  return {
    manufacturer: compact(form.manufacturer) || undefined,
    articleNumber: compact(form.articleNumber) || undefined,
    name: compact(form.name) || undefined,
    gauge: compact(form.gauges[0]) || undefined,
    searchSources: articleSearchSources(),
    fields
  };
}

export function hasAccessorySearchCriteria(input: ArticleSearchInput) {
  const fields = input.fields || {};
  if (compact(fields.ean)) return true;
  const identity = compact(input.articleNumber || fields.articleNumber) || compact(input.name || fields.name);
  const manufacturer = compact(input.manufacturer || fields.manufacturer);
  const gauge = compact(input.gauge || fields.gauge);
  return Boolean(identity && manufacturer && gauge);
}

export function accessorySearchFieldGroups(t: Translate): ArticleSearchFieldGroup[] {
  return [
    {
      key: "article",
      label: t("vehicles.articleSearch.group.model"),
      fields: [
        { key: "manufacturer", label: t("accessories.field.manufacturer") },
        { key: "articleNumber", label: t("accessories.field.articleNumber") },
        { key: "name", label: t("accessories.field.name") },
        { key: "ean", label: t("accessories.editor.fields.ean") },
        { key: "gauge", label: t("accessories.toolbar.gauge") },
        { key: "scale", label: t("accessories.editor.fields.scale") }
      ]
    },
    {
      key: "more",
      label: t("vehicles.articleSearch.group.more"),
      fields: [
        { key: "description", label: t("accessories.field.description") },
        { key: "articleSourceUrl", label: t("accessories.editor.fields.productUrl") }
      ]
    }
  ];
}

export function currentAccessorySearchValue(form: ArticleEditorForm, key: string) {
  if (key === "gauge") return form.gauges.join(", ");
  if (key === "articleSourceUrl") return form.productUrl.trim();
  const target = directFieldMap[key as keyof typeof directFieldMap];
  return target ? compact(form[target]) : "";
}

export function isUsableAccessorySearchValue(key: string, value: string) {
  if (!searchFieldKeys.has(key as AccessorySearchFieldKey) || isBadCommonArticleValue(key, value)) {
    return false;
  }
  if (key !== "articleSourceUrl") return true;
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

export function applyAccessorySearchResult({
  form,
  result,
  resultIndex,
  selectedFields,
  manufacturers,
  gauges
}: {
  form: ArticleEditorForm;
  result: ArticleSearchResult;
  resultIndex: number;
  selectedFields: Record<string, boolean>;
  manufacturers: MasterDataEntry[];
  gauges: MasterDataEntry[];
}): Partial<ArticleEditorForm> {
  const patch: Partial<ArticleEditorForm> = {};
  for (const [key, field] of Object.entries(result.fields)) {
    if (!selectedFields[articleSelectionKey(result, key, resultIndex)] ||
        !isUsableAccessorySearchValue(key, field.value)) {
      continue;
    }
    if (key === "manufacturer") {
      const label = knownActiveLabel(manufacturers, field.value);
      if (label) patch.manufacturer = label;
      continue;
    }
    if (key === "gauge") {
      const label = knownActiveLabel(gauges, field.value);
      if (label) patch.gauges = [...new Set([...form.gauges, label])];
      continue;
    }
    const target = directFieldMap[key as keyof typeof directFieldMap];
    if (target) Object.assign(patch, { [target]: field.value.trim() });
  }
  return patch;
}
