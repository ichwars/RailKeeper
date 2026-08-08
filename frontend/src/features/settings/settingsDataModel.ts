import type { SettingsTab } from "./settingsModel";

export const generalDataTypes = [
  "manufacturer",
  "vehicle_category",
  "vehicle_gattung",
  "epoch",
  "gauge",
  "railway_company",
  "cv8_manufacturer",
  "symbols"
] as const;

export const articleDataTypes = ["stock_unit", "types", "customFields", "locations"] as const;

export type DataGroup = "general" | "article";
export type GeneralDataType = (typeof generalDataTypes)[number];
export type ArticleDataType = (typeof articleDataTypes)[number];
export type SettingsDataType = GeneralDataType | ArticleDataType;
export type SettingsLocation = {
  tab: SettingsTab;
  group: DataGroup;
  type: SettingsDataType;
};

const settingsTabIds = new Set<SettingsTab>([
  "general",
  "data",
  "digital",
  "importExport",
  "appearance",
  "auth"
]);

function includesValue<T extends string>(values: readonly T[], value: string | null): value is T {
  return value !== null && values.some((candidate) => candidate === value);
}

export function isGeneralDataType(value: string): value is GeneralDataType {
  return includesValue(generalDataTypes, value);
}

export function isArticleDataType(value: string): value is ArticleDataType {
  return includesValue(articleDataTypes, value);
}

export function readSettingsLocation(search: string): SettingsLocation {
  const query = new URLSearchParams(search);
  const requestedTab = query.get("tab");
  if (requestedTab === "articleManagement") {
    return { tab: "data", group: "article", type: "stock_unit" };
  }

  const tab = requestedTab && settingsTabIds.has(requestedTab as SettingsTab)
    ? requestedTab as SettingsTab
    : "general";
  if (tab !== "data") {
    return { tab, group: "general", type: "manufacturer" };
  }

  const group: DataGroup = query.get("group") === "article" ? "article" : "general";
  const requestedType = query.get("type");
  if (group === "article") {
    return {
      tab,
      group,
      type: includesValue(articleDataTypes, requestedType) ? requestedType : "stock_unit"
    };
  }
  return {
    tab,
    group,
    type: includesValue(generalDataTypes, requestedType) ? requestedType : "manufacturer"
  };
}

export function settingsLocationSearch(location: SettingsLocation): string {
  if (location.tab === "general") return "";
  const query = new URLSearchParams({ tab: location.tab });
  if (location.tab === "data") {
    query.set("group", location.group);
    query.set("type", location.type);
  }
  return `?${query.toString()}`;
}
