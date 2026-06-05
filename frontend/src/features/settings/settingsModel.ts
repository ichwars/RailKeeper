import type { AppView } from "../../app/App";
import type { AuditLogEntry, MasterDataEntry } from "../../shared/api";

export type SettingsTab = "general" | "data" | "digital" | "importExport" | "appearance" | "auth";
export type MasterDataType = {
  type: string;
};

export const settingsTabs: { id: SettingsTab; labelKey: string }[] = [
  { id: "general", labelKey: "settings.tabs.general" },
  { id: "data", labelKey: "settings.tabs.data" },
  { id: "digital", labelKey: "settings.tabs.digital" },
  { id: "importExport", labelKey: "settings.tabs.importExport" },
  { id: "appearance", labelKey: "settings.tabs.appearance" },
  { id: "auth", labelKey: "settings.tabs.auth" }
];

export const masterDataTypes: MasterDataType[] = [
  { type: "manufacturer" },
  { type: "vehicle_category" },
  { type: "vehicle_gattung" },
  { type: "epoch" },
  { type: "gauge" },
  { type: "railway_company" },
  { type: "cv8_manufacturer" },
  { type: "symbols" }
];

export const loadableMasterDataTypes = masterDataTypes;
export const articleSearchSettingKey = "railkeeper.articleSearchEnabled";
export const articleSearchSourcesSettingKey = "railkeeper.articleSearchSources";
export const defaultArticleSearchSources = ["manufacturer", "catalogs", "dealers", "web"];
const legacyArticleSearchSources = ["web", "manufacturer", "dealers", "wiki"];
const previousArticleSearchSources = ["manufacturer", "dealers", "web"];
export const articleSearchSourceOptions = [
  { id: "web", labelKey: "settings.articleSearch.source.web", helpKey: "settings.articleSearch.source.webHelp" },
  { id: "manufacturer", labelKey: "settings.articleSearch.source.manufacturer", helpKey: "settings.articleSearch.source.manufacturerHelp" },
  { id: "catalogs", labelKey: "settings.articleSearch.source.catalogs", helpKey: "settings.articleSearch.source.catalogsHelp" },
  { id: "dealers", labelKey: "settings.articleSearch.source.dealers", helpKey: "settings.articleSearch.source.dealersHelp" },
  { id: "wiki", labelKey: "settings.articleSearch.source.wiki", helpKey: "settings.articleSearch.source.wikiHelp" }
];

export const localSettingKeys = {
  language: "railkeeper.settings.language",
  defaultView: "railkeeper.settings.defaultView",
  dateFormat: "railkeeper.settings.dateFormat",
  timeFormat: "railkeeper.settings.timeFormat",
  defaultPrinter: "railkeeper.settings.defaultPrinter",
  updateChecks: "railkeeper.settings.updateChecks",
  betaUpdates: "railkeeper.settings.betaUpdates",
  ignoredUpdate: "railkeeper.settings.ignoredUpdate",
  darkBackground: "railkeeper.settings.darkBackground",
  darkAccent: "railkeeper.settings.darkAccent",
  darkStyle: "railkeeper.settings.darkStyle",
  lightBackground: "railkeeper.settings.lightBackground",
  lightAccent: "railkeeper.settings.lightAccent",
  lightStyle: "railkeeper.settings.lightStyle",
  digitalProvider: "railkeeper.settings.digitalProvider",
  digitalEcosEnabled: "railkeeper.ecos.enabled",
  digitalEcosHost: "railkeeper.ecos.host",
  digitalEcosPort: "railkeeper.ecos.port",
  digitalZ21Enabled: "railkeeper.z21.enabled",
  digitalZ21Host: "railkeeper.z21.host",
  digitalZ21Port: "railkeeper.z21.port",
  digitalIntellibox3Enabled: "railkeeper.intellibox3.enabled",
  digitalIntellibox3Host: "railkeeper.intellibox3.host",
  digitalIntellibox3Port: "railkeeper.intellibox3.port",
  digitalCS3Enabled: "railkeeper.cs3.enabled",
  digitalCS3Host: "railkeeper.cs3.host",
  digitalCS3Port: "railkeeper.cs3.port",
  twoFactorPrepared: "railkeeper.settings.twoFactorPrepared"
};

export const sidebarOrderChangedEvent = "railkeeper-sidebar-order-changed";
export const legacySidebarOrderKey = "railkeeper.settings.sidebarOrder";
export const sidebarPrefsBaseKey = "railkeeper.settings.sidebarPrefs";
export const defaultSidebarOrder: AppView[] = ["overview", "vehicles", "exhibition", "importExport", "settings"];
export type SidebarPrefs = {
  order: AppView[];
  hidden: AppView[];
};
export function formatBytes(value: number) {
  if (!Number.isFinite(value) || value <= 0) return "0 KB";
  if (value >= 1024 * 1024) return `${(value / 1024 / 1024).toLocaleString("de-DE", { maximumFractionDigits: 1 })} MB`;
  return `${Math.max(1, Math.round(value / 1024)).toLocaleString("de-DE")} KB`;
}

export function readLocalSetting(key: string, fallback: string) {
  return window.localStorage.getItem(key) || fallback;
}

export function readLocalBool(key: string, fallback: boolean) {
  const value = window.localStorage.getItem(key);
  if (value === null) return fallback;
  return value === "true";
}

function isLegacyArticleSearchDefault(sources: string[]) {
  return (
    sources.length === legacyArticleSearchSources.length && legacyArticleSearchSources.every((source) => sources.includes(source))
  ) || (
    sources.length === previousArticleSearchSources.length && previousArticleSearchSources.every((source) => sources.includes(source))
  );
}

export function readArticleSearchSources() {
  try {
    const stored = JSON.parse(window.localStorage.getItem(articleSearchSourcesSettingKey) || "[]") as string[];
    const allowed = new Set(articleSearchSourceOptions.map((option) => option.id));
    const sources = stored.filter((source) => allowed.has(source));
    if (isLegacyArticleSearchDefault(sources)) {
      window.localStorage.setItem(articleSearchSourcesSettingKey, JSON.stringify(defaultArticleSearchSources));
      return defaultArticleSearchSources;
    }
    return sources.length > 0 ? sources : defaultArticleSearchSources;
  } catch {
    return defaultArticleSearchSources;
  }
}

export function sidebarPrefsKey(username: string) {
  return `${sidebarPrefsBaseKey}:${username || "local"}`;
}

export function normalizeSidebarOrder(order: AppView[]) {
  const ordered = order.filter((view): view is AppView => defaultSidebarOrder.includes(view));
  const missing = defaultSidebarOrder.filter((view) => !ordered.includes(view));
  return [...ordered, ...missing];
}

export function readSidebarPrefs(username: string): SidebarPrefs {
  try {
    const stored = JSON.parse(window.localStorage.getItem(sidebarPrefsKey(username)) || "null") as Partial<SidebarPrefs> | null;
    if (stored) {
      return {
        order: normalizeSidebarOrder(Array.isArray(stored.order) ? stored.order : []),
        hidden: Array.isArray(stored.hidden) ? stored.hidden.filter((view): view is AppView => defaultSidebarOrder.includes(view) && view !== "settings") : []
      };
    }
  } catch {
    return { order: defaultSidebarOrder, hidden: [] };
  }
  try {
    const legacyOrder = JSON.parse(window.localStorage.getItem(legacySidebarOrderKey) || "[]") as AppView[];
    return { order: normalizeSidebarOrder(legacyOrder), hidden: [] };
  } catch {
    return { order: defaultSidebarOrder, hidden: [] };
  }
}

export function formatDateTime(value: string) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return date.toLocaleString("de-DE", { dateStyle: "short", timeStyle: "short" });
}

export const emptyForm = {
  key: "",
  label: "",
  active: true,
  sortOrder: 0,
  sourceUrl: "",
  website: "",
  searchDomainsText: "",
  aliasesText: "",
  nominalScalesText: "",
  cvDecimal: "",
  cvBinary: "",
  cvHex: "",
  cvCountry: "",
  description: "",
  imageData: "",
  metadataText: "{}"
};

export type FormState = typeof emptyForm;

export const emptyUserForm = {
  username: "",
  email: "",
  password: "",
  confirmPassword: "",
  roles: ["Viewer"]
};

export type UserFormState = typeof emptyUserForm;

export const emptyPasswordForm = {
  currentPassword: "",
  newPassword: "",
  confirmPassword: ""
};

export type PasswordFormState = typeof emptyPasswordForm;

export function auditActor(entry: AuditLogEntry) {
  return entry.actorUsername || entry.actorUserId || (entry.action === "LoginFailed" ? "Unbekannt" : "System");
}

export function auditTarget(entry: AuditLogEntry) {
  if (!entry.targetType && !entry.targetId) return "-";
  return [entry.targetType, entry.targetId].filter(Boolean).join(" ");
}

export function entryToForm(entry: MasterDataEntry): FormState {
  const isCV8Manufacturer = entry.type === "cv8_manufacturer";
  return {
    key: entry.key,
    label: isCV8Manufacturer ? cv8NameText(entry) : entry.label,
    active: entry.active,
    sortOrder: entry.sortOrder,
    sourceUrl: entry.sourceUrl || "",
    website: metadataString(entry, "website"),
    searchDomainsText: manufacturerSearchDomainsText(entry),
    aliasesText: metadataList(entry, "aliases").join(", "),
    nominalScalesText: nominalScalesText(entry),
    cvDecimal: cv8DecimalText(entry),
    cvBinary: cv8BinaryText(entry),
    cvHex: cv8HexText(entry),
    cvCountry: cv8CountryText(entry),
    description: metadataString(entry, "description"),
    imageData: masterDataImage(entry),
    metadataText: JSON.stringify(entry.metadata || {}, null, 2)
  };
}

function numberFromMetadataValue(value: unknown) {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const decimal = Number.parseInt(value.trim(), 10);
    return Number.isFinite(decimal) ? decimal : undefined;
  }
  return undefined;
}

function validCV8Value(value: number | undefined) {
  return typeof value === "number" && Number.isFinite(value) && value >= 0 && value <= 255 ? value : undefined;
}

export function metadataNumberString(entry: MasterDataEntry, key: string) {
  const value = entry.metadata?.[key];
  if (typeof value === "number" && Number.isFinite(value)) return String(value);
  if (typeof value === "string" && value.trim()) return value.trim();
  return "";
}

export function metadataString(entry: MasterDataEntry, key: string) {
  const value = entry.metadata?.[key];
  return typeof value === "string" && value.trim() ? value.trim() : "";
}

export function metadataList(entry: MasterDataEntry, key: string) {
  const value = entry.metadata?.[key];
  if (Array.isArray(value)) {
    return value.filter((item): item is string => typeof item === "string" && item.trim().length > 0);
  }
  if (typeof value === "string" && value.trim()) {
    return value
      .split(/[,;\n]/)
      .map((item) => item.trim())
      .filter(Boolean);
  }
  return [];
}

export function nominalScalesText(entry: MasterDataEntry) {
  return metadataList(entry, "nominalScales").join(", ");
}

export function manufacturerWebsite(entry: MasterDataEntry) {
  return metadataString(entry, "website");
}

export function domainFromUrl(value: string | undefined) {
  const clean = (value || "").trim();
  if (!clean) return "";
  try {
    const url = new URL(clean.startsWith("http://") || clean.startsWith("https://") ? clean : `https://${clean}`);
    return url.hostname.replace(/^www\./, "");
  } catch {
    return "";
  }
}

export function manufacturerSearchDomains(entry: MasterDataEntry) {
  const explicit = metadataList(entry, "searchDomains");
  if (explicit.length > 0) return explicit;
  const derived = domainFromUrl(manufacturerWebsite(entry));
  return derived ? [derived] : [];
}

export function manufacturerSearchDomainsText(entry: MasterDataEntry) {
  return manufacturerSearchDomains(entry).join(", ");
}

export function manufacturerAliasesText(entry: MasterDataEntry) {
  return metadataList(entry, "aliases").join(", ");
}

export function isWikiLikeUrl(value: string | undefined) {
  const normalized = (value || "").toLocaleLowerCase("de-DE");
  return normalized.includes("modellbau-wiki") || normalized.includes("wikipedia.org") || normalized.includes("wikimedia.org");
}

export function manufacturerNeedsWebsiteReview(entry: MasterDataEntry) {
  const website = manufacturerWebsite(entry);
  return isWikiLikeUrl(website) || (!website && isWikiLikeUrl(entry.sourceUrl));
}

export function cv8NameText(entry: MasterDataEntry) {
  return entry.label.replace(/^\s*\d{1,3}\s*-\s*/, "").trim() || entry.label.trim();
}

export function cv8DecimalValue(entry: MasterDataEntry) {
  const metadataValue =
    validCV8Value(numberFromMetadataValue(entry.metadata?.decimal)) ??
    validCV8Value(numberFromMetadataValue(entry.metadata?.cvDecimal)) ??
    validCV8Value(numberFromMetadataValue(entry.metadata?.cv8)) ??
    validCV8Value(numberFromMetadataValue(entry.metadata?.manufacturerId));
  if (typeof metadataValue === "number") return metadataValue;

  const keyMatch = entry.key.match(/(?:^|-)0*(\d{1,3})$/);
  const keyValue = keyMatch ? validCV8Value(Number.parseInt(keyMatch[1], 10)) : undefined;
  if (typeof keyValue === "number") return keyValue;

  const labelMatch = entry.label.match(/^\s*(\d{1,3})\s*-/);
  return labelMatch ? validCV8Value(Number.parseInt(labelMatch[1], 10)) : undefined;
}

export function cv8DecimalText(entry: MasterDataEntry) {
  const decimal = cv8DecimalValue(entry);
  return typeof decimal === "number" ? String(decimal) : "";
}

export function cv8BinaryText(entry: MasterDataEntry) {
  const binary = metadataString(entry, "binary") || metadataString(entry, "cvBinary");
  if (binary) return normalizeCV8Binary(binary, "");
  const decimal = cv8DecimalValue(entry);
  return typeof decimal === "number" ? decimal.toString(2).padStart(8, "0") : "";
}

export function cv8HexText(entry: MasterDataEntry) {
  const hex = metadataString(entry, "hex") || metadataString(entry, "cvHex");
  if (hex) return normalizeCV8Hex(hex, "");
  const decimal = cv8DecimalValue(entry);
  return typeof decimal === "number" ? `0x${decimal.toString(16).toUpperCase().padStart(2, "0")}` : "";
}

export function cv8CountryText(entry: MasterDataEntry) {
  return metadataString(entry, "country") || metadataString(entry, "cvCountry");
}

export function normalizeCV8Decimal(value: string) {
  const decimal = Number.parseInt(value.trim(), 10);
  return Number.isFinite(decimal) && decimal >= 0 && decimal <= 255 ? String(decimal) : "";
}

export function normalizeCV8Binary(value: string, decimalValue: string) {
  const cleaned = value.replace(/^0b/i, "").replace(/[^01]/g, "");
  if (cleaned) return cleaned.padStart(8, "0").slice(-8);
  const decimal = Number.parseInt(decimalValue, 10);
  return Number.isFinite(decimal) && decimal >= 0 && decimal <= 255 ? decimal.toString(2).padStart(8, "0") : "";
}

export function normalizeCV8Hex(value: string, decimalValue: string) {
  const cleaned = value.trim().replace(/^0x/i, "");
  if (/^[0-9a-f]{1,2}$/i.test(cleaned)) return `0x${cleaned.toUpperCase().padStart(2, "0")}`;
  const decimal = Number.parseInt(decimalValue, 10);
  return Number.isFinite(decimal) && decimal >= 0 && decimal <= 255 ? `0x${decimal.toString(16).toUpperCase().padStart(2, "0")}` : "";
}

export function masterDataImage(entry: MasterDataEntry) {
  return metadataString(entry, "imageData") || metadataString(entry, "activeImageData") || metadataString(entry, "svgData");
}

export function parseList(text: string) {
  return text
    .split(/[,;\n]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

export function applyVisibleMetadata(type: string, metadata: Record<string, unknown>, form: FormState) {
  const next = { ...metadata };
  if (type === "manufacturer") {
    const nominalScales = parseList(form.nominalScalesText);
    const aliases = parseList(form.aliasesText);
    const website = form.website.trim();
    const searchDomains = parseList(form.searchDomainsText);
    const derivedDomain = domainFromUrl(website);
    if (searchDomains.length === 0 && derivedDomain) {
      searchDomains.push(derivedDomain);
    }
    if (nominalScales.length > 0) {
      next.nominalScales = nominalScales;
    } else {
      delete next.nominalScales;
    }
    if (website) {
      next.website = website;
    } else {
      delete next.website;
    }
    if (searchDomains.length > 0) {
      next.searchDomains = searchDomains;
    } else {
      delete next.searchDomains;
    }
    if (aliases.length > 0) {
      next.aliases = aliases;
    } else {
      delete next.aliases;
    }
  }
  if (type === "cv8_manufacturer") {
    const decimal = normalizeCV8Decimal(form.cvDecimal);
    const binary = normalizeCV8Binary(form.cvBinary, decimal);
    const hex = normalizeCV8Hex(form.cvHex, decimal);
    const country = form.cvCountry.trim().toUpperCase();
    if (decimal) {
      next.decimal = Number.parseInt(decimal, 10);
      next.cvDecimal = decimal;
      next.cv8 = Number.parseInt(decimal, 10);
    } else {
      delete next.decimal;
      delete next.cvDecimal;
      delete next.cv8;
    }
    if (binary) {
      next.binary = binary;
      next.cvBinary = binary;
    } else {
      delete next.binary;
      delete next.cvBinary;
    }
    if (hex) {
      next.hex = hex;
      next.cvHex = hex;
    } else {
      delete next.hex;
      delete next.cvHex;
    }
    if (country) {
      next.country = country;
      next.cvCountry = country;
    } else {
      delete next.country;
      delete next.cvCountry;
    }
  }
  if (type === "symbols") {
    if (form.description.trim()) {
      next.description = form.description.trim();
    } else {
      delete next.description;
    }
    if (form.imageData) {
      next.imageData = form.imageData;
      next.activeImageData = form.imageData;
      next.imageMime = form.imageData.startsWith("data:image/svg+xml") ? "image/svg+xml" : "image";
    } else {
      delete next.imageData;
      delete next.activeImageData;
      delete next.svgData;
      delete next.imageMime;
    }
  }
  return next;
}

export function externalLink(entry: MasterDataEntry) {
  const website = metadataString(entry, "website");
  if (website) {
    return { href: website, title: "Website oeffnen" };
  }
  if (entry.sourceUrl) {
    return { href: entry.sourceUrl, title: "Quelle oeffnen" };
  }
  return null;
}
