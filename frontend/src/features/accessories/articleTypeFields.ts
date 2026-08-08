import type {
  AccessoryArticleType,
  AccessoryAttributeValue,
  MasterDataEntry
} from "../../shared/api";

export type ArticleSubjectFieldKind = AccessoryAttributeValue["kind"];

export type ArticleSubjectFieldDefinition = {
  key: string;
  kind: ArticleSubjectFieldKind;
  labelKey: string;
  helpKey?: string;
  unit?: string;
  options?: readonly string[];
  min?: number;
  max?: number;
  step?: number;
};

export type CustomArticleSubjectFieldDefinition = Omit<ArticleSubjectFieldDefinition, "labelKey"> & {
  label: string;
};

const field = (
  key: string,
  kind: ArticleSubjectFieldKind,
  details: Omit<ArticleSubjectFieldDefinition, "key" | "kind" | "labelKey"> = {}
): ArticleSubjectFieldDefinition => ({
  key,
  kind,
  labelKey: `accessories.subject.field.${key}`,
  ...details
});

const nonnegative = { min: 0 } as const;
const positiveInteger = { min: 0, step: 1 } as const;

export const articleTypeFieldRegistry: Readonly<
  Record<AccessoryArticleType, readonly ArticleSubjectFieldDefinition[]>
> = {
  track: [
    field("trackSystem", "text"),
    field("lengthMm", "number", { ...nonnegative, unit: "mm" }),
    field("radiusMm", "number", { ...nonnegative, unit: "mm" }),
    field("angleDegrees", "number", { min: 0, max: 360, step: 0.1, unit: "°" }),
    field("direction", "single_select", { options: ["left", "right", "symmetric"] }),
    field("frogAngleDegrees", "number", { min: 0, max: 180, step: 0.1, unit: "°" }),
    field("sleeperType", "text"),
    field("railHeightMm", "number", { ...nonnegative, unit: "mm" }),
    field("roadbed", "boolean"),
    field("connectionCount", "number", positiveInteger),
    field("digitalReady", "boolean")
  ],
  signal: [
    field("prototype", "text"),
    field("epoch", "multi_select", { options: ["I", "II", "III", "IV", "V", "VI"] }),
    field("aspects", "multi_select", { options: ["stop", "proceed", "caution", "shunting"] }),
    field("ledCount", "number", positiveInteger),
    field("heightMm", "number", { ...nonnegative, unit: "mm" }),
    field("voltageAC", "number", { ...nonnegative, unit: "V AC" }),
    field("voltageDC", "number", { ...nonnegative, unit: "V DC" }),
    field("mounting", "single_select", { options: ["surface", "flush", "mast", "wall"] }),
    field("driveType", "single_select", { options: ["manual", "motor", "servo", "solenoid"] }),
    field("integratedDecoder", "boolean"),
    field("controlModule", "text")
  ],
  decoder: [
    field("interface", "single_select", { options: ["wired", "nem651", "nem652", "plux16", "plux22", "mtc21", "next18"] }),
    field("protocols", "multi_select", { options: ["dcc", "motorola", "selectrix", "mfx", "railcom"] }),
    field("functionOutputs", "number", positiveInteger),
    field("motorCurrentMa", "number", { ...nonnegative, unit: "mA" }),
    field("outputCurrentMa", "number", { ...nonnegative, unit: "mA" }),
    field("totalCurrentMa", "number", { ...nonnegative, unit: "mA" }),
    field("railCom", "boolean"),
    field("susi", "boolean"),
    field("servoOutputs", "number", positiveInteger),
    field("dimensions", "text"),
    field("firmware", "text")
  ],
  electrical_control: [
    field("inputVoltage", "number", { ...nonnegative, unit: "V" }),
    field("outputVoltage", "number", { ...nonnegative, unit: "V" }),
    field("currentA", "number", { ...nonnegative, unit: "A" }),
    field("powerW", "number", { ...nonnegative, unit: "W" }),
    field("channelCount", "number", positiveInteger),
    field("protocols", "multi_select", { options: ["dcc", "motorola", "selectrix", "loconet", "can", "s88"] }),
    field("connectors", "multi_select", { options: ["screw", "plug", "rj45", "bus", "wire"] }),
    field("protections", "multi_select", { options: ["shortCircuit", "overload", "temperature", "reversePolarity"] }),
    field("compatibleArticles", "multi_select", { options: ["track", "signal", "decoder", "lighting"] })
  ],
  building_equipment: [
    field("epoch", "multi_select", { options: ["I", "II", "III", "IV", "V", "VI"] }),
    field("dimensions", "text"),
    field("footprint", "text"),
    field("material", "text"),
    field("constructionType", "single_select", { options: ["kit", "finished", "semiFinished"] }),
    field("partCount", "number", positiveInteger),
    field("difficulty", "single_select", { options: ["easy", "medium", "advanced"] }),
    field("lightingOptions", "multi_select", { options: ["interior", "platform", "street", "effect"] }),
    field("floorPlanAvailable", "boolean")
  ],
  landscape_consumable: [
    field("material", "text"),
    field("color", "text"),
    field("season", "text"),
    field("content", "number", nonnegative),
    field("contentUnit", "single_select", { options: ["piece", "pack", "meter", "gram", "milliliter"] }),
    field("fiberOrGrainSize", "text"),
    field("coverage", "text"),
    field("suitableScales", "multi_select", { options: ["Z", "N", "TT", "H0", "0", "1", "G"] }),
    field("safetyNotes", "text")
  ],
  lighting: [
    field("lightColor", "text"),
    field("colorTemperatureK", "number", { ...nonnegative, unit: "K" }),
    field("voltage", "number", { ...nonnegative, unit: "V" }),
    field("currentMa", "number", { ...nonnegative, unit: "mA" }),
    field("powerType", "single_select", { options: ["ac", "dc", "acDc"] }),
    field("ledCount", "number", positiveInteger),
    field("dimmable", "boolean"),
    field("dimensions", "text"),
    field("mounting", "single_select", { options: ["surface", "flush", "mast", "wall"] })
  ],
  other: []
};

const attributeKinds: readonly ArticleSubjectFieldKind[] = [
  "text", "number", "boolean", "date", "single_select", "multi_select"
];

function isAttributeKind(value: unknown): value is ArticleSubjectFieldKind {
  return typeof value === "string" && attributeKinds.some((kind) => kind === value);
}

function stringOptions(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.filter((option): option is string => typeof option === "string" && option.trim() !== "")
    .map((option) => option.trim());
}

export function customFieldDefinitions(entries: readonly MasterDataEntry[]): CustomArticleSubjectFieldDefinition[] {
  const seen = new Set<string>();
  return [...entries]
    .filter((entry) => entry.active && entry.type === "accessory_custom_field")
    .sort((left, right) => left.sortOrder - right.sortOrder || left.label.localeCompare(right.label))
    .flatMap((entry) => {
      const kind = entry.metadata.kind;
      if (!isAttributeKind(kind) || seen.has(entry.key)) return [];
      const options = stringOptions(entry.metadata.options);
      if ((kind === "single_select" || kind === "multi_select") && options.length === 0) return [];
      seen.add(entry.key);
      const unit = kind === "number" && typeof entry.metadata.unit === "string" && entry.metadata.unit.trim()
        ? entry.metadata.unit.trim()
        : undefined;
      return [{ key: entry.key, kind, label: entry.label, ...(unit ? { unit } : {}),
        ...(options.length > 0 ? { options } : {}) }];
    });
}

export function fieldDefinitionsForType(
  articleType: AccessoryArticleType,
  customFields: readonly CustomArticleSubjectFieldDefinition[] = []
): readonly (ArticleSubjectFieldDefinition | CustomArticleSubjectFieldDefinition)[] {
  return articleType === "other" ? customFields : articleTypeFieldRegistry[articleType];
}

export function compatibleAttributesForType(
  articleType: AccessoryArticleType,
  attributes: readonly AccessoryAttributeValue[],
  customFields: readonly CustomArticleSubjectFieldDefinition[] = []
): AccessoryAttributeValue[] {
  const definitions = new Map(fieldDefinitionsForType(articleType, customFields)
    .map((definition) => [definition.key, definition.kind]));
  return attributes.filter((attribute) => definitions.get(attribute.key) === attribute.kind);
}

export function compatibleNumberDraftsForType(
  articleType: AccessoryArticleType,
  drafts: Readonly<Record<string, string>>,
  customFields: readonly CustomArticleSubjectFieldDefinition[] = []
): Record<string, string> {
  const numberKeys = new Set(fieldDefinitionsForType(articleType, customFields)
    .filter((definition) => definition.kind === "number").map((definition) => definition.key));
  return Object.fromEntries(Object.entries(drafts).filter(([key]) => numberKeys.has(key)));
}

export function subjectValuesAreValid(
  articleType: AccessoryArticleType,
  attributes: readonly AccessoryAttributeValue[],
  numberDrafts: Readonly<Record<string, string>>
): boolean {
  const definitions = new Map(articleTypeFieldRegistry[articleType].map((definition) => [definition.key, definition]));
  if (articleType !== "other" && attributes.some((attribute) => definitions.get(attribute.key)?.kind !== attribute.kind)) {
    return false;
  }
  return Object.entries(numberDrafts).every(([key, draft]) => {
    if (draft.trim() === "") return true;
    const numberValue = Number(draft.replace(",", "."));
    if (!Number.isFinite(numberValue)) return false;
    const definition = definitions.get(key);
    return (articleType === "other" || definition?.kind === "number") &&
      (definition?.min === undefined || numberValue >= definition.min) &&
      (definition?.max === undefined || numberValue <= definition.max);
  });
}
