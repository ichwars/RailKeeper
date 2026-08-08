import type { AccessoryArticleType, MasterDataEntry } from "../../shared/api";

const builtInSubtypeKeys = new Set([
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

type Translate = (key: string, values?: Record<string, string | number>) => string;

function canonicalKey(entry: MasterDataEntry, articleType: AccessoryArticleType): string | null {
  const prefix = `${articleType}:`;
  return entry.key.startsWith(prefix) ? entry.key.slice(prefix.length) : null;
}

function canonicalValue(articleType: AccessoryArticleType, subtype: string): string {
  const prefix = `${articleType}:`;
  return subtype.startsWith(prefix) ? subtype.slice(prefix.length) : subtype;
}

function seededSubtypeLabel(fullKey: string): string {
  const value = fullKey.slice(fullKey.indexOf(":") + 1).replaceAll("_", " ");
  if (value === "led") return "LED";
  return value.charAt(0).toUpperCase() + value.slice(1);
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
  if (builtInSubtypeKeys.has(fullKey) && (!configured || configured.label === seededSubtypeLabel(fullKey))) {
    return t(`accessories.subtype.${articleType}.${canonicalValue(articleType, subtype)}`);
  }
  return configured?.label || subtype;
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
