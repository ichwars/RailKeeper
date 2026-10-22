import type { VehicleCVValueInput } from "../../shared/api";
import { decoderDefinition, decoderIndexUnknown, isECoSReading, resolveDecoderLibrary } from "../../shared/decoderLibrary";
import { bundledDecoderInventory, bundledDecoderProfiles, type DecoderInventory } from "../../shared/decoderLibraryModel";
import { railComPlus } from "../../shared/railComPlus";
import { zimoMNAnalysis } from "./zimoMNAnalysis";
import { lenzStandardAnalysis } from "./lenzStandardAnalysis";
import { piko56500Analysis } from "./piko56500Analysis";
import { roco10882Analysis } from "./roco10882Analysis";
import { uhlenbrock76420Analysis } from "./uhlenbrock76420Analysis";
import { kuehnN45Analysis } from "./kuehnN45Analysis";

export type DecoderCV = VehicleCVValueInput;
export type Translator = (key: string, values?: Record<string, string | number>) => string;

export function decoderProfiles(values: DecoderCV[]) {
  return Array.from(new Set(values.map((value) => value.decoderProfile || ""))).sort();
}

export function decoderAnalysis(values: DecoderCV[], profile: string, catalog = bundledDecoderProfiles,
  inventory: DecoderInventory | null = bundledDecoderInventory, decoderType = "") {
  const rows = values.filter((value) => (value.decoderProfile || "") === profile);
  // Never resolve a base CV from an indexed page, including page 0/0.
  const modelLabel = isECoSReading({ decoderProfile: profile }) && decoderType.trim()
    ? decoderType : profile;
  const libraryMatch = resolveDecoderLibrary(rows, profile, catalog, inventory, modelLabel);
  const libraryProfile = catalog.find(p => p.active && p.id === libraryMatch.profileId);
  const modelRules = libraryProfile?.kind === "model" && ["matched", "family"].includes(libraryMatch.status) && libraryMatch.firmwareCompatible !== false;
  const ownDefinitionsScope = Boolean(modelRules && libraryProfile?.rules?.modelDefinitionsOnly);
  const lenzScope = Boolean(modelRules && libraryProfile?.rules?.lenzStandardV2);
  const pikoScope = Boolean(modelRules && libraryProfile?.rules?.piko56500);
  const rocoScope = Boolean(modelRules && libraryProfile?.rules?.roco10882);
  const uhlenbrockScope = Boolean(modelRules && libraryProfile?.rules?.uhlenbrock76420);
  const kuehnScope = Boolean(modelRules && libraryProfile?.rules?.kuehnN45);
  const base = new Map(rows.filter((value) => value.indexCV31 == null && value.indexCV32 == null)
    .map((value) => [value.cvNumber, value]));
  const get = (number: number) => {
    const value = base.get(number);
    const documented = value && decoderDefinition(catalog, libraryMatch, value);
    if (!documented || documented.definition.reserved || !value) return undefined;
    if ((kuehnScope || uhlenbrockScope || rocoScope || pikoScope || ownDefinitionsScope) && documented.profile.id !== libraryProfile?.id) return undefined;
    if ((lenzScope || kuehnScope || uhlenbrockScope || rocoScope || pikoScope || ownDefinitionsScope) && (!Number.isInteger(value.value) || value.value < documented.definition.minimum ||
      value.value > documented.definition.maximum || (documented.definition.allowedValues &&
        !documented.definition.allowedValues.includes(value.value)))) return undefined;
    return value.value;
  };
  const indexed = (cv: number, cv31: number, cv32: number) => rows.find((row) =>
    row.cvNumber === cv && row.indexCV31 === cv31 && row.indexCV32 === cv32)?.value;
  const config = get(29), high = get(17), low = get(18);
  const decodedLong = high == null || low == null || high < 192 || high > 231
    ? undefined : (high - 192) * 256 + low;
  const longAddress = (lenzScope || kuehnScope || uhlenbrockScope || rocoScope || pikoScope || ownDefinitionsScope) && decodedLong != null && (decodedLong < 1 || decodedLong > 9999) ? undefined : decodedLong;
  const validLongAddress = (uhlenbrockScope || pikoScope) && longAddress != null && longAddress < 128 ? undefined : longAddress;
  const address = config == null ? undefined : config & 32 ? validLongAddress : get(1);
  const esu5 = Boolean(modelRules && libraryProfile?.mapping);
  const consist = get(19);
  const enabledFunctions: string[] = [];
  const mask1 = get(21), mask2 = get(22);
  if (mask1 != null) for (let bit = 0; bit < 8; bit++) if (mask1 & (1 << bit)) enabledFunctions.push(`F${bit + 1}`);
  if (mask2 != null) {
    if (mask2 & 1) enabledFunctions.unshift("F0 →");
    if (mask2 & 2) enabledFunctions.unshift("F0 ←");
    for (let bit = 2; bit < 6; bit++) if (mask2 & (1 << bit)) enabledFunctions.push(`F${bit + 7}`);
  }
  const rules = modelRules ? libraryProfile?.rules : undefined;
  const zimo = rules?.zimoMN ? zimoMNAnalysis(get) : undefined;
  const lenz = lenzScope ? lenzStandardAnalysis(get) : undefined;
  const piko = pikoScope ? piko56500Analysis(get, indexed) : undefined;
  const roco = rocoScope ? roco10882Analysis(get) : undefined;
  const uhlenbrock = uhlenbrockScope ? uhlenbrock76420Analysis(get) : undefined;
  const kuehn = kuehnScope ? kuehnN45Analysis(get) : undefined;
  const brake = rules?.abc ? get(rules.abc.cv) : undefined;
  const forward = rules?.constantBrake ? get(rules.constantBrake.forwardCV) : undefined;
  const reverse = rules?.constantBrake ? get(rules.constantBrake.reverseCV) : undefined;
  const brakeEnabled = rules?.constantBrake?.enableCV ? get(rules.constantBrake.enableCV) : undefined;
  const analogMode = rules?.dcAnalog ? get(rules.dcAnalog.cv) : undefined;
  const dcAnalog = config == null ? undefined : !(config & 4) ? false
    : !rules?.dcAnalog || analogMode == null ? undefined : Boolean(analogMode & rules.dcAnalog.mask);
  const firmware = rows.find((row) => row.cvNumber === 7 && row.indexCV31 == null)?.description
    ?.match(/^LokProgrammer Decoder: .+ \(([\d.]+)\)$/)?.[1];
  return {
    rows, get, indexed, address, longAddress: validLongAddress, esu5, modelRules, rules, config, firmware,
    dcAnalog: roco ? roco.analog : uhlenbrock ? uhlenbrock.dcAnalog : kuehn ? kuehn.dcAnalog : zimo ? zimo.dcAnalog : dcAnalog, zimo, lenz, kuehn, uhlenbrock, roco, piko,
    railcomEnabled: piko ? piko.railcomEnabled : roco ? roco.railcomEnabled : uhlenbrock ? undefined : lenz ? lenz.railcom : config == null ? undefined : Boolean(config & 8),
    libraryMatch, libraryProfile, catalog, railComPlus: railComPlus(rows, libraryProfile, libraryMatch),
    vehicleDecoderType: modelLabel !== profile && libraryMatch.matchedBy === "name" ? modelLabel : undefined,
    definition: (value: DecoderCV) => decoderDefinition(catalog, libraryMatch, { ...value, decoderProfile: profile }),
    indexUnknown: (value: DecoderCV) => decoderIndexUnknown(catalog, libraryMatch, { ...value, decoderProfile: profile }),
    consistAddress: kuehn && consist != null && consist > 127 ? undefined
      : zimo ? zimo.consistAddress : consist == null ? undefined : consist & 127,
    consistReverse: lenz || kuehn ? undefined : zimo ? zimo.consistReverse : consist == null ? undefined : Boolean(consist & 128),
    consistCVs: zimo ? [19, 20] : [19],
    enabledFunctions: lenz || kuehn || uhlenbrock || roco ? [] : zimo ? zimo.enabledFunctions : enabledFunctions,
    functionsComplete: !lenz && !kuehn && !uhlenbrock && !roco && (zimo ? zimo.functionsComplete : mask1 != null && mask2 != null),
    abc: !rules?.abc || brake == null ? undefined : Boolean(brake & rules.abc.mask),
    constantBrakeCVs: piko ? [145, 142, 27] : lenz ? [51] : zimo ? [140] : rules?.constantBrake
      ? [rules.constantBrake.forwardCV, rules.constantBrake.reverseCV] : [],
    constantBrake: piko ? piko.constantBrake : lenz ? lenz.constantBrake : zimo ? zimo.constantBrake : !rules?.constantBrake ? undefined : rules.constantBrake.enableCV
      ? brakeEnabled == null ? undefined : brakeEnabled > 0
      : (forward != null && forward > 0) || (reverse != null && reverse > 0)
      ? true : forward === 0 && reverse === 0 ? false : undefined,
    indexPages: new Set(rows.filter((row) => row.indexCV31 != null)
      .map((row) => `${row.indexCV31}/${row.indexCV32}`)).size
  };
}

export type DecoderAnalysis = ReturnType<typeof decoderAnalysis>;

export function cvSource(analysis: DecoderAnalysis, ...numbers: number[]) {
  return numbers.map((number) => `CV${number}=${analysis.get(number) ?? "?"}`).join(" · ");
}

export type MappingSuggestion = {
  row: number; functionKey: string; direction: "forward" | "reverse" | "both";
  actions: string[]; source: string; warning: boolean;
  sourceCVs?: number[];
};

export function decoderMapping(analysis: DecoderAnalysis, t: Translator): MappingSuggestion[] {
  if (analysis.piko) return analysis.piko.mappings.map((m, i) => ({
    row: i + 1, functionKey: `F${m.key}`, direction: m.direction,
    actions: m.outputs.map(o => o.name < 2 ? t(o.name === 0 ? "vehicles.decoder.frontLight" : "vehicles.decoder.rearLight")
      : `A${o.name - 1}${o.available ? "" : ` (${t("vehicles.decoder.piko.unavailable")})`}`),
    warning: m.outputs.some(o => !o.available), source: cvSource(analysis, 96, m.cv), sourceCVs: [m.cv, 96]
  }));
  if (analysis.roco) {
    const mapped = analysis.roco.mappings.map(m => ({
      row: 0, functionKey: `F${m.key}`, direction: m.direction,
      actions: [...m.outputs.map(o => `${t(`vehicles.decoder.roco.output.${o.name}`)}${o.available ? "" : ` (${t("vehicles.decoder.roco.unavailableOutput")})`}`),
        ...(m.unknownBits ? [t("vehicles.decoder.roco.unknownBits", { bits: m.unknownBits })] : [])],
      source: cvSource(analysis, m.cv), sourceCVs: [m.cv],
      warning: Boolean(m.unknownBits || m.outputs.some(o => !o.available))
    }));
    const logical = [
      { key: analysis.roco.shuntingKey, cv: 155, action: "vehicles.decoder.shunting" },
      { key: analysis.roco.noDelayKey, cv: 156, action: "vehicles.decoder.noDelay" }
    ].flatMap(m => m.key ? [{ row: 0, functionKey: m.key, direction: "both" as const,
      actions: [t(m.action)], source: cvSource(analysis, m.cv), sourceCVs: [m.cv], warning: false }] : []);
    return [...mapped, ...logical].map((m, i) => ({ ...m, row: i + 1 }));
  }
  if (analysis.uhlenbrock) return analysis.uhlenbrock.mappings.map((m, i) => ({
    row: i + 1, functionKey: `F${m.key}`, direction: m.direction,
    actions: m.actions.map(action => t(`vehicles.decoder.uhlenbrock.action.${action}`)),
    source: cvSource(analysis, m.cv), sourceCVs: [m.cv], warning: false
  }));
  if (analysis.kuehn) return kuehnMapping(analysis, t);
  if (!analysis.lenz) return esuMapping(analysis, t);
  return analysis.lenz.mappings.map((mapping, index) => ({
    row: index + 1, functionKey: `F${mapping.key}`, direction: mapping.direction,
    actions: [...mapping.outputs.map(output => t("vehicles.decoder.lenz.output", { output })),
      ...(mapping.unknownBit ? [t("vehicles.decoder.lenz.unknownOutput")] : [])],
    source: cvSource(analysis, mapping.cv), warning: mapping.unknownBit, sourceCVs: [mapping.cv]
  }));
}

function kuehnMapping(a: DecoderAnalysis, t: Translator): MappingSuggestion[] {
  if (!a.kuehn) return [];
  const mapped: MappingSuggestion[] = a.kuehn.mappings.map((m, i) => ({
    row: i + 1, functionKey: `F${m.key}`, direction: m.direction,
    actions: [...m.outputs.map(o => {
      const name = t("vehicles.decoder.kuehn.output", { output: o.output });
      const conditions = [
        ...("EF".includes(o.output) ? [t(o.available == null ? "vehicles.decoder.kuehn.susiUnknown"
          : o.available ? "vehicles.decoder.kuehn.logicLevel" : "vehicles.decoder.kuehn.susiBlocked")] : []),
        ...(o.effect?.offForward ? [t("vehicles.decoder.kuehn.offForward")] : []),
        ...(o.effect?.offReverse ? [t("vehicles.decoder.kuehn.offReverse")] : [])
      ];
      return conditions.length ? `${name} (${conditions.join(" · ")})` : name;
    }), ...(m.unknownBits ? [t("vehicles.decoder.kuehn.unknownBits", { bits: m.unknownBits })] : [])],
    source: cvSource(a, m.cv, ...m.outputs.map(o => [49, 50, 51, 52, 47, 48]["ABCDEF".indexOf(o.output)]),
      ...(m.outputs.some(o => "EF".includes(o.output)) ? [56] : [])),
    warning: Boolean(m.unknownBits || m.outputs.some(o => o.available !== true)), sourceCVs: [m.cv]
  }));
  const logical = [
    { keys: a.kuehn.shuntingKeys, action: "vehicles.decoder.shunting", cv: 58 },
    { keys: a.kuehn.noDelayKeys, action: "vehicles.decoder.noDelay", cv: 58 },
    { keys: a.kuehn.dimmingKeys, action: "vehicles.decoder.kuehn.dimAction", cv: 59 }
  ].flatMap(({ keys, action, cv }) => keys?.map(functionKey => ({
    row: 0, functionKey, direction: "both" as const, actions: [t(action)],
    source: cvSource(a, cv), sourceCVs: [cv], warning: false
  })) ?? []);
  return [...mapped, ...logical].map((row, i) => ({ ...row, row: i + 1 }));
}

// ESU LokPilot 5 manual, mapping table: 72 rows, 16 rows per page, 16-byte stride.
// Only simple positive single-key conditions are named. Missing fields stay unknown.
export function esuMapping(analysis: DecoderAnalysis, t: Translator): MappingSuggestion[] {
  const layout = analysis.libraryProfile?.mapping;
  if (!analysis.esu5 || !layout) return [];
  const result: MappingSuggestion[] = [];
  for (let row = 0; row < layout.rows; row++) {
    const start = layout.startCV + (row % layout.rowsPerPage) * layout.stride;
    const conditionPage = layout.conditionPage + Math.floor(row / layout.rowsPerPage);
    const actionPage = layout.actionPage + Math.floor(row / layout.rowsPerPage);
    const conditions = Array.from({ length: 9 }, (_, i) => analysis.indexed(start + i, layout.indexCV31, conditionPage));
    if (conditions.some((value) => value == null)) continue;
    const extraCondition = analysis.indexed(start + 9, layout.indexCV31, conditionPage);
    if (extraCondition != null && extraCondition !== 0) continue;
    const a = conditions[0] ?? 0;
    // Bits 0/1 also depend on driving/stopping; leave these complex rows unnamed.
    if ((a & 3) || (a & 12) === 12) continue;
    const keys: number[] = [];
    let complex = false;
    for (let key = 0; key <= 28; key++) {
      const position = key + 2, value = conditions[Math.floor(position / 4)] ?? 0;
      const condition = (value >> ((position % 4) * 2)) & 3;
      if (condition === 1) keys.push(key);
      if (condition > 1) complex = true;
    }
    if (complex || keys.length !== 1 || ((conditions[7] ?? 0) & 192) || (conditions[8] ?? 0)) continue;
    const k = analysis.indexed(start, layout.indexCV31, actionPage), m = analysis.indexed(start + 2, layout.indexCV31, actionPage);
    const n = analysis.indexed(start + 3, layout.indexCV31, actionPage), o = analysis.indexed(start + 4, layout.indexCV31, actionPage);
    const q = analysis.indexed(start + 6, layout.indexCV31, actionPage);
    if ([k, m, n, o, q].some((value) => value == null)) continue;
    const actions: string[] = [];
    let warning = false;
    for (let bit = 0; bit < 8; bit++) if ((k ?? 0) & (1 << bit)) {
      if (bit < 2) actions.push(t(bit === 0 ? "vehicles.decoder.frontLight" : "vehicles.decoder.rearLight"));
      else {
        const output = bit - 1;
        const powerPack = output === 3 && analysis.indexed(291, 16, 0) === 31;
        actions.push(`AUX${output}${powerPack ? " · PowerPack Control" : ""}`);
        warning ||= powerPack;
      }
    }
    if ((n ?? 0) & 2) actions.push(t("vehicles.decoder.shunting"));
    if ((o ?? 0) & 8) actions.push(t("vehicles.decoder.noDelay"));
    const l = analysis.indexed(start + 1, layout.indexCV31, actionPage), p = analysis.indexed(start + 5, layout.indexCV31, actionPage);
    if (l || p || m || q || ((n ?? 0) & ~2) || ((o ?? 0) & ~8)) {
      actions.push(t("vehicles.decoder.additionalActions")); warning = true;
    }
    if (!actions.length) continue;
    result.push({ row: row + 1, functionKey: `F${keys[0]}`,
      direction: a & 4 ? "forward" : a & 8 ? "reverse" : "both", actions, warning,
      source: `CV${start}–${start + 8} · CV31=${layout.indexCV31} · CV32=${conditionPage}/${actionPage}` });
  }
  return result;
}
