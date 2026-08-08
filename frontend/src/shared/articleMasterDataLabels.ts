import type { AccessoryArticleType, MasterDataEntry } from "./api";

type Translate = (key: string, values?: Record<string, string | number>) => string;

const articleTypeSeedLabels: Record<AccessoryArticleType, string> = {
  track: "Track",
  signal: "Signal",
  decoder: "Decoder",
  electrical_control: "Electrical control",
  building_equipment: "Building equipment",
  landscape_consumable: "Landscape consumable",
  lighting: "Lighting",
  other: "Other"
};

const stockUnitSeedLabels: Record<string, string> = {
  piece: "Piece",
  pack: "Pack",
  meter: "Meter",
  gram: "Gram",
  milliliter: "Milliliter"
};

const articleSubtypeKeys = new Set([
  "track:straight", "track:curve", "track:flex", "track:turnout", "track:crossing",
  "track:double_slip", "track:transition", "track:buffer_stop",
  "signal:light", "signal:semaphore", "signal:main", "signal:distant", "signal:block",
  "signal:entry", "signal:exit", "signal:shunting",
  "decoder:locomotive", "decoder:function", "decoder:accessory", "decoder:switching",
  "decoder:servo", "decoder:feedback",
  "electrical_control:turnout_drive", "electrical_control:feedback", "electrical_control:booster",
  "electrical_control:power_supply", "electrical_control:sensor", "electrical_control:relay",
  "electrical_control:distribution", "electrical_control:control_element",
  "building_equipment:building", "building_equipment:platform", "building_equipment:bridge",
  "building_equipment:tunnel_portal", "building_equipment:road_vehicle", "building_equipment:figure",
  "building_equipment:street_equipment", "building_equipment:interior_equipment",
  "landscape_consumable:grass", "landscape_consumable:scatter", "landscape_consumable:tree",
  "landscape_consumable:water", "landscape_consumable:paint", "landscape_consumable:adhesive",
  "landscape_consumable:ballast", "landscape_consumable:wire", "landscape_consumable:cable",
  "landscape_consumable:fastener",
  "lighting:lamp", "lighting:led", "lighting:light_strip", "lighting:building_lighting",
  "lighting:effect_lighting", "other:other"
]);

function subtypeSeedLabel(key: string): string {
  const value = key.slice(key.indexOf(":") + 1).replaceAll("_", " ");
  if (value === "led") return "LED";
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function standardTranslationKey(entry: MasterDataEntry): string | null {
  if (entry.type === "stock_unit" && stockUnitSeedLabels[entry.key] === entry.label) {
    return `accessories.subject.option.${entry.key}`;
  }
  if (entry.type === "article_type") {
    const articleType = entry.key as AccessoryArticleType;
    if (articleTypeSeedLabels[articleType] === entry.label) {
      return `accessories.articleType.${articleType}`;
    }
  }
  if (entry.type === "accessory_subtype" && articleSubtypeKeys.has(entry.key) &&
      subtypeSeedLabel(entry.key) === entry.label) {
    const [articleType, subtype] = entry.key.split(":", 2);
    return `accessories.subtype.${articleType}.${subtype}`;
  }
  return null;
}

export function isStandardArticleSubtype(key: string): boolean {
  return articleSubtypeKeys.has(key);
}

export function masterDataDisplayLabel(entry: MasterDataEntry, t: Translate): string {
  const translationKey = standardTranslationKey(entry);
  return translationKey ? t(translationKey) : entry.label;
}

export function masterDataPersistedLabel(entry: MasterDataEntry, draftLabel: string, t: Translate): string {
  const trimmed = draftLabel.trim();
  const translationKey = standardTranslationKey(entry);
  if (translationKey && trimmed === t(translationKey)) return entry.label;
  return trimmed;
}
