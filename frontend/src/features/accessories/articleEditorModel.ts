import type {
  AccessoryArticle,
  AccessoryAttributeValue,
  AccessoryArticleType,
  AccessoryArticleWriteInput,
  AccessoryInventoryStrategy,
  AccessoryManufacturerStatus,
  AccessoryPurchaseInput
} from "../../shared/api";
import {
  fieldDefinitionsForType,
  subjectValidationIssues,
  type CustomArticleSubjectFieldDefinition
} from "./articleTypeFields";
import { normalizeAccessoryMoney } from "./accessoryMoney";

export type ArticleEditorMode = "create" | "view" | "edit";
export type ArticleEditorTab = "article" | "stock" | "purchaseDocuments" | "subject" | "usageHistory";

export type ArticleEditorForm = {
  manufacturer: string;
  articleNumber: string;
  name: string;
  ean: string;
  manufacturerStatus: AccessoryManufacturerStatus;
  articleType: AccessoryArticleType;
  subtype: string;
  gauges: string[];
  scale: string;
  listPrice: string;
  packageQuantity: string;
  stockUnit: string;
  minimumStock: string;
  inventoryStrategy: AccessoryInventoryStrategy;
  description: string;
  manufacturerUrl: string;
  productUrl: string;
  alternativeNumbers: string;
  keywords: string;
  compatibilityNotes: string;
  internalNotes: string;
  archived: boolean;
  attributes: AccessoryArticle["attributes"];
  attributeNumberDrafts: Record<string, string>;
};

export type ArticleEditorFieldErrors = Partial<Record<keyof ArticleEditorForm, string>>;
export type ArticleEditorTabErrors = Partial<Record<ArticleEditorTab, boolean>>;
export type ArticleSubjectFieldErrors = Record<string, string>;

export function emptyArticleEditorForm(): ArticleEditorForm {
  return {
    manufacturer: "",
    articleNumber: "",
    name: "",
    ean: "",
    manufacturerStatus: "unknown",
    articleType: "other",
    subtype: "",
    gauges: [],
    scale: "",
    listPrice: "",
    packageQuantity: "1",
    stockUnit: "piece",
    minimumStock: "0",
    inventoryStrategy: "quantity",
    description: "",
    manufacturerUrl: "",
    productUrl: "",
    alternativeNumbers: "",
    keywords: "",
    compatibilityNotes: "",
    internalNotes: "",
    archived: false,
    attributes: [],
    attributeNumberDrafts: {}
  };
}

export function articleToEditorForm(article: AccessoryArticle): ArticleEditorForm {
  return {
    manufacturer: article.manufacturer,
    articleNumber: article.articleNumber || "",
    name: article.name,
    ean: article.ean || "",
    manufacturerStatus: article.manufacturerStatus,
    articleType: article.articleType,
    subtype: article.subtype,
    gauges: article.gauges,
    scale: article.scale || "",
    listPrice: article.listPrice || "",
    packageQuantity: String(article.packageQuantity),
    stockUnit: article.stockUnit,
    minimumStock: String(article.minimumStock),
    inventoryStrategy: article.inventoryStrategy,
    description: article.description || "",
    manufacturerUrl: article.manufacturerUrl || "",
    productUrl: article.productUrl || "",
    alternativeNumbers: article.alternativeNumbers.join("\n"),
    keywords: article.keywords.join(", "),
    compatibilityNotes: article.compatibilityNotes || "",
    internalNotes: article.internalNotes || "",
    archived: article.archived,
    attributes: article.attributes,
    attributeNumberDrafts: Object.fromEntries(article.attributes.flatMap((attribute) =>
      attribute.kind === "number" ? [[attribute.key, String(attribute.numberValue)]] : []))
  };
}

const splitValues = (value: string) => value.split(/[\n,]/).map((item) => item.trim()).filter(Boolean);
const optional = (value: string) => value.trim() || undefined;
const editorNumber = (value: string) => Number(value.replace(",", "."));

export function articleEditorWriteInput(
  form: ArticleEditorForm,
  customFields: readonly CustomArticleSubjectFieldDefinition[] = [],
  historicalAttributes: readonly AccessoryAttributeValue[] = []
): AccessoryArticleWriteInput {
  if (Object.keys(subjectValidationIssues(
    form.articleType, form.attributes, form.attributeNumberDrafts, customFields, historicalAttributes
  )).length > 0) throw new Error("invalid subject values");
  const packageQuantity = editorNumber(form.packageQuantity);
  const minimumStock = editorNumber(form.minimumStock);
  const listPrice = normalizeAccessoryMoney(form.listPrice);
  if (!Number.isInteger(packageQuantity) || packageQuantity <= 0 ||
      !Number.isInteger(minimumStock) || minimumStock < 0) {
    throw new Error("invalid article quantities");
  }
  if (listPrice === undefined) throw new Error("invalid list price");
  const inventoryStrategy = form.inventoryStrategy;
  const numberDefinitions = new Map(fieldDefinitionsForType(form.articleType, customFields)
    .filter((definition) => definition.kind === "number").map((definition) => [definition.key, definition]));
  const numberDraftKeys = new Set(Object.keys(form.attributeNumberDrafts));
  const attributes: AccessoryAttributeValue[] = form.attributes.filter((attribute) =>
    attribute.kind !== "number" || !numberDraftKeys.has(attribute.key));
  for (const [key, draft] of Object.entries(form.attributeNumberDrafts)) {
    if (draft.trim() === "") continue;
    const numberValue = Number(draft.replace(",", "."));
    if (!Number.isFinite(numberValue)) continue;
    const existing = form.attributes.find((attribute) => attribute.key === key && attribute.kind === "number");
    const unit = existing?.kind === "number" ? existing.unit : numberDefinitions.get(key)?.unit;
    attributes.push({ key, kind: "number", numberValue, ...(unit ? { unit } : {}) });
  }
  return {
    manufacturer: form.manufacturer.trim(),
    articleNumber: optional(form.articleNumber),
    name: form.name.trim(),
    category: form.subtype.trim(),
    trackingMode: inventoryStrategy === "individual" ? "individual" : "quantity",
    description: optional(form.description),
    ean: optional(form.ean),
    manufacturerStatus: form.manufacturerStatus,
    articleType: form.articleType,
    subtype: form.subtype.trim(),
    gauges: form.gauges,
    scale: optional(form.scale),
    listPrice,
    packageQuantity,
    stockUnit: form.stockUnit.trim(),
    minimumStock,
    inventoryStrategy,
    manufacturerUrl: optional(form.manufacturerUrl),
    productUrl: optional(form.productUrl),
    alternativeNumbers: splitValues(form.alternativeNumbers),
    keywords: splitValues(form.keywords),
    compatibilityNotes: optional(form.compatibilityNotes),
    internalNotes: optional(form.internalNotes),
    archived: form.archived,
    attributes
  };
}

export function articlePurchaseWriteInput(
  purchase: AccessoryPurchaseInput,
  quantityDraft: string,
  locationId: string
): AccessoryPurchaseInput {
  return {
    ...purchase,
    quantity: Number(quantityDraft),
    storageLocationId: purchase.bookToStock ? locationId : undefined
  };
}

export function validateArticleEditorForm(
  form: ArticleEditorForm,
  messages = {
    required: "Pflichtfeld",
    positive: "Muss größer als 0 sein",
    nonnegative: "Darf nicht negativ sein",
    integer: "Nur ganze Zahlen sind zulässig",
    invalidSubject: "Fachwert ist ungültig",
    invalidOption: "Auswahl ist ungültig",
    invalidStep: "Wert entspricht nicht der Schrittweite",
    invalidMoney: "Preis ist ungültig"
  },
  customFields: readonly CustomArticleSubjectFieldDefinition[] = [],
  historicalAttributes: readonly AccessoryAttributeValue[] = []
): {
  fieldErrors: ArticleEditorFieldErrors;
  tabErrors: ArticleEditorTabErrors;
  subjectFieldErrors: ArticleSubjectFieldErrors;
} {
  const fieldErrors: ArticleEditorFieldErrors = {};
  const tabErrors: ArticleEditorTabErrors = {};
  const subjectFieldErrors: ArticleSubjectFieldErrors = {};
  if (!form.manufacturer.trim()) fieldErrors.manufacturer = messages.required;
  if (!form.name.trim()) fieldErrors.name = messages.required;
  if (!form.subtype.trim()) fieldErrors.subtype = messages.required;
  if (!form.stockUnit.trim()) fieldErrors.stockUnit = messages.required;
  const packageQuantity = editorNumber(form.packageQuantity);
  const minimumStock = editorNumber(form.minimumStock);
  if (normalizeAccessoryMoney(form.listPrice) === undefined) {
    fieldErrors.listPrice = messages.invalidMoney;
  }
  if (!Number.isFinite(packageQuantity) || packageQuantity <= 0) {
    fieldErrors.packageQuantity = messages.positive;
  } else if (!Number.isInteger(packageQuantity)) {
    fieldErrors.packageQuantity = messages.integer;
  }
  if (!Number.isFinite(minimumStock) || minimumStock < 0) {
    fieldErrors.minimumStock = messages.nonnegative;
  } else if (!Number.isInteger(minimumStock)) {
    fieldErrors.minimumStock = messages.integer;
  }
  const subjectIssues = subjectValidationIssues(
    form.articleType, form.attributes, form.attributeNumberDrafts, customFields, historicalAttributes
  );
  for (const [key, issue] of Object.entries(subjectIssues)) {
    subjectFieldErrors[key] = issue === "invalidOption" ? messages.invalidOption
      : issue === "invalidStep" ? messages.invalidStep : messages.invalidSubject;
  }
  if (Object.keys(subjectFieldErrors).length > 0) {
    fieldErrors.attributes = messages.invalidSubject;
  }
  if (fieldErrors.manufacturer || fieldErrors.name || fieldErrors.subtype || fieldErrors.stockUnit ||
      fieldErrors.packageQuantity || fieldErrors.listPrice) tabErrors.article = true;
  if (fieldErrors.minimumStock) tabErrors.stock = true;
  if (fieldErrors.attributes) tabErrors.subject = true;
  return { fieldErrors, tabErrors, subjectFieldErrors };
}

export function isArticleEditorDirty(form: ArticleEditorForm, initial: ArticleEditorForm): boolean {
  return JSON.stringify(form) !== JSON.stringify(initial);
}
