import type {
  AccessoryArticleType,
  AccessoryAttributeValue,
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
import {
  articleTypeFieldRegistry,
  subjectValidationIssues,
  type ArticleSubjectFieldDefinition
} from "./articleTypeFields";

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
  gauges: MasterDataEntry[],
  articleType: AccessoryArticleType = "other"
) {
  if (!isUsableAccessorySearchValue(key, value, articleType)) return false;
  if (key === "manufacturer") return Boolean(knownActiveLabel(manufacturers, value));
  if (key === "gauge") return Boolean(knownActiveLabel(gauges, value));
  return true;
}

function subjectDefinition(articleType: AccessoryArticleType, key: string) {
  return articleTypeFieldRegistry[articleType].find((definition) => definition.key === key);
}

function subjectSearchValue(form: ArticleEditorForm, key: string) {
  const definition = subjectDefinition(form.articleType, key);
  if (!definition) return "";
  if (definition.kind === "number") {
    const draft = form.attributeNumberDrafts[key]?.trim();
    if (draft) return draft.replace(",", ".");
  }
  const attribute = form.attributes.find((candidate) =>
    candidate.key === key && candidate.kind === definition.kind);
  if (!attribute) return "";
  switch (attribute.kind) {
  case "text": return attribute.textValue.trim();
  case "number": return String(attribute.numberValue);
  case "boolean": return String(attribute.booleanValue);
  case "date": return attribute.dateValue.trim();
  case "single_select":
  case "multi_select": return attribute.optionValues.join(", ");
  }
}

export function accessorySearchInput(form: ArticleEditorForm): ArticleSearchInput {
  const subjectFields = Object.fromEntries(articleTypeFieldRegistry[form.articleType]
    .map((definition) => [definition.key, subjectSearchValue(form, definition.key)])
    .filter(([, value]) => value));
  const fields = Object.fromEntries(Object.entries({
    ean: form.ean,
    scale: form.scale,
    description: form.description,
    articleSourceUrl: form.productUrl,
    articleType: form.articleType,
    subtype: form.subtype,
    ...subjectFields
  }).map(([key, value]) => [key, compact(value)]).filter(([, value]) => value));
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

export function accessorySearchFieldGroups(
  t: Translate,
  articleType: AccessoryArticleType = "other"
): ArticleSearchFieldGroup[] {
  const groups: ArticleSearchFieldGroup[] = [
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
  if (articleType === "track") {
    groups.push({
      key: "subject",
      label: t("accessories.editor.tabs.subject", { type: t("accessories.articleType.track") }),
      fields: articleTypeFieldRegistry[articleType].map((definition) => ({
        key: definition.key,
        label: t(definition.labelKey)
      }))
    });
  }
  return groups;
}

export function currentAccessorySearchValue(form: ArticleEditorForm, key: string) {
  if (key === "gauge") return form.gauges.join(", ");
  if (key === "articleSourceUrl") return form.productUrl.trim();
  const target = directFieldMap[key as keyof typeof directFieldMap];
  return target ? compact(form[target]) : subjectSearchValue(form, key);
}

function isUsableSubjectValue(
  definition: ArticleSubjectFieldDefinition,
  value: string,
  articleType: AccessoryArticleType
) {
  const normalized = value.trim();
  if (!normalized) return false;
  if (definition.kind === "number") {
    return !subjectValidationIssues(articleType, [], {
      [definition.key]: normalized.replace(",", ".")
    })[definition.key];
  }
  if (definition.kind === "boolean") return normalized === "true" || normalized === "false";
  if (definition.kind === "single_select") return definition.options?.includes(normalized) ?? false;
  if (definition.kind === "multi_select") {
    const values = normalized.split(",").map((entry) => entry.trim()).filter(Boolean);
    return values.length > 0 && values.every((entry) => definition.options?.includes(entry));
  }
  return true;
}

export function isUsableAccessorySearchValue(
  key: string,
  value: string,
  articleType: AccessoryArticleType = "other"
) {
  const definition = subjectDefinition(articleType, key);
  if (definition) return isUsableSubjectValue(definition, value, articleType);
  if (!searchFieldKeys.has(key as AccessorySearchFieldKey) || isBadCommonArticleValue(key, value)) return false;
  if (key !== "articleSourceUrl") return true;
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:";
  } catch {
    return false;
  }
}

function replaceAttribute(
  attributes: readonly AccessoryAttributeValue[],
  nextAttribute: AccessoryAttributeValue
) {
  return [...attributes.filter((attribute) => attribute.key !== nextAttribute.key), nextAttribute];
}

function subjectAttribute(
  definition: ArticleSubjectFieldDefinition,
  value: string
): AccessoryAttributeValue | undefined {
  const normalized = value.trim();
  switch (definition.kind) {
  case "text": return { key: definition.key, kind: "text", textValue: normalized };
  case "boolean": return { key: definition.key, kind: "boolean", booleanValue: normalized === "true" };
  case "date": return { key: definition.key, kind: "date", dateValue: normalized };
  case "single_select": return { key: definition.key, kind: "single_select", optionValues: [normalized] };
  case "multi_select": return {
    key: definition.key,
    kind: "multi_select",
    optionValues: normalized.split(",").map((entry) => entry.trim()).filter(Boolean)
  };
  case "number": return undefined;
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
  let nextAttributes = form.attributes;
  let nextNumberDrafts = form.attributeNumberDrafts;
  for (const [key, field] of Object.entries(result.fields)) {
    if (!selectedFields[articleSelectionKey(result, key, resultIndex)] ||
        !isUsableAccessorySearchValue(key, field.value, form.articleType)) {
      continue;
    }
    const definition = subjectDefinition(form.articleType, key);
    if (definition?.kind === "number") {
      nextNumberDrafts = {
        ...nextNumberDrafts,
        [key]: field.value.trim().replace(",", ".")
      };
      continue;
    }
    if (definition) {
      const attribute = subjectAttribute(definition, field.value);
      if (attribute) nextAttributes = replaceAttribute(nextAttributes, attribute);
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
  if (nextAttributes !== form.attributes) patch.attributes = nextAttributes;
  if (nextNumberDrafts !== form.attributeNumberDrafts) patch.attributeNumberDrafts = nextNumberDrafts;
  return patch;
}
