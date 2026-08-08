import type {
  AccessoryArticle,
  AccessoryArticleType,
  AccessoryArticleWriteInput,
  AccessoryInventoryStrategy,
  AccessoryManufacturerStatus
} from "../../shared/api";

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
  attributes: AccessoryArticle["attributes"];
};

export type ArticleEditorFieldErrors = Partial<Record<keyof ArticleEditorForm, string>>;
export type ArticleEditorTabErrors = Partial<Record<ArticleEditorTab, boolean>>;

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
    attributes: []
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
    attributes: article.attributes
  };
}

const splitValues = (value: string) => value.split(/[\n,]/).map((item) => item.trim()).filter(Boolean);
const optional = (value: string) => value.trim() || undefined;

export function articleEditorWriteInput(form: ArticleEditorForm): AccessoryArticleWriteInput {
  const inventoryStrategy = form.inventoryStrategy;
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
    packageQuantity: Number(form.packageQuantity),
    stockUnit: form.stockUnit.trim(),
    minimumStock: Number(form.minimumStock),
    inventoryStrategy,
    manufacturerUrl: optional(form.manufacturerUrl),
    productUrl: optional(form.productUrl),
    alternativeNumbers: splitValues(form.alternativeNumbers),
    keywords: splitValues(form.keywords),
    compatibilityNotes: optional(form.compatibilityNotes),
    internalNotes: optional(form.internalNotes),
    attributes: form.attributes
  };
}

export function validateArticleEditorForm(form: ArticleEditorForm, messages = {
  required: "Pflichtfeld",
  positive: "Muss größer als 0 sein",
  nonnegative: "Darf nicht negativ sein"
}): {
  fieldErrors: ArticleEditorFieldErrors;
  tabErrors: ArticleEditorTabErrors;
} {
  const fieldErrors: ArticleEditorFieldErrors = {};
  const tabErrors: ArticleEditorTabErrors = {};
  if (!form.manufacturer.trim()) fieldErrors.manufacturer = messages.required;
  if (!form.name.trim()) fieldErrors.name = messages.required;
  if (!form.subtype.trim()) fieldErrors.subtype = messages.required;
  if (!form.stockUnit.trim()) fieldErrors.stockUnit = messages.required;
  if (!Number.isFinite(Number(form.packageQuantity)) || Number(form.packageQuantity) <= 0) {
    fieldErrors.packageQuantity = messages.positive;
  }
  if (!Number.isFinite(Number(form.minimumStock)) || Number(form.minimumStock) < 0) {
    fieldErrors.minimumStock = messages.nonnegative;
  }
  if (fieldErrors.manufacturer || fieldErrors.name || fieldErrors.subtype || fieldErrors.stockUnit ||
      fieldErrors.packageQuantity) tabErrors.article = true;
  if (fieldErrors.minimumStock) tabErrors.stock = true;
  return { fieldErrors, tabErrors };
}

export function isArticleEditorDirty(form: ArticleEditorForm, initial: ArticleEditorForm): boolean {
  return JSON.stringify(form) !== JSON.stringify(initial);
}
