import type { VehicleCVValueInput } from "../../shared/api";

export type DecoderCV = VehicleCVValueInput;
export type Translator = (key: string, values?: Record<string, string | number>) => string;

export function decoderProfiles(values: DecoderCV[]) {
  return Array.from(new Set(values.map((value) => value.decoderProfile || ""))).sort();
}

export function decoderAnalysis(values: DecoderCV[], profile: string) {
  const rows = values.filter((value) => (value.decoderProfile || "") === profile);
  // Never resolve a base CV from an indexed page, including page 0/0.
  const base = new Map(rows.filter((value) => value.indexCV31 == null && value.indexCV32 == null)
    .map((value) => [value.cvNumber, value.value]));
  const get = (number: number) => base.get(number);
  const indexed = (cv: number, cv31: number, cv32: number) => rows.find((row) =>
    row.cvNumber === cv && row.indexCV31 === cv31 && row.indexCV32 === cv32)?.value;
  const config = get(29), high = get(17), low = get(18);
  const longAddress = high == null || low == null || high < 192 || high > 231
    ? undefined : (high - 192) * 256 + low;
  const address = config == null ? undefined : config & 32 ? longAddress : get(1);
  const esu5 = get(8) === 151 && /\bLokPilot\s*5\b/i.test(profile);
  const consist = get(19);
  const enabledFunctions: string[] = [];
  const mask1 = get(21), mask2 = get(22);
  if (mask1 != null) for (let bit = 0; bit < 8; bit++) if (mask1 & (1 << bit)) enabledFunctions.push(`F${bit + 1}`);
  if (mask2 != null) {
    if (mask2 & 1) enabledFunctions.unshift("F0 →");
    if (mask2 & 2) enabledFunctions.unshift("F0 ←");
    for (let bit = 2; bit < 6; bit++) if (mask2 & (1 << bit)) enabledFunctions.push(`F${bit + 7}`);
  }
  const brake = get(27), forward = get(254), reverse = get(255);
  const analogMode = get(50);
  const dcAnalog = config == null ? undefined : !(config & 4) ? false
    : !esu5 || analogMode == null ? undefined : Boolean(analogMode & 2);
  const firmware = rows.find((row) => row.cvNumber === 7 && row.indexCV31 == null)?.description
    ?.match(/^LokProgrammer Decoder: .+ \(([\d.]+)\)$/)?.[1];
  return {
    rows, get, indexed, address, longAddress, esu5, config, firmware, dcAnalog,
    consistAddress: consist == null ? undefined : consist & 127,
    consistReverse: consist == null ? undefined : Boolean(consist & 128),
    enabledFunctions, functionsComplete: mask1 != null && mask2 != null,
    abc: !esu5 || brake == null ? undefined : Boolean(brake & 3),
    constantBrake: !esu5 ? undefined : (forward != null && forward > 0) || (reverse != null && reverse > 0)
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
};

// ESU LokPilot 5 manual, mapping table: 72 rows, 16 rows per page, 16-byte stride.
// Only simple positive single-key conditions are named. Missing fields stay unknown.
export function esuMapping(analysis: DecoderAnalysis, t: Translator): MappingSuggestion[] {
  if (!analysis.esu5) return [];
  const result: MappingSuggestion[] = [];
  for (let row = 0; row < 72; row++) {
    const start = 257 + (row % 16) * 16, conditionPage = 3 + Math.floor(row / 16);
    const actionPage = 8 + Math.floor(row / 16);
    const conditions = Array.from({ length: 9 }, (_, i) => analysis.indexed(start + i, 16, conditionPage));
    if (conditions.some((value) => value == null)) continue;
    const extraCondition = analysis.indexed(start + 9, 16, conditionPage);
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
    const k = analysis.indexed(start, 16, actionPage), m = analysis.indexed(start + 2, 16, actionPage);
    const n = analysis.indexed(start + 3, 16, actionPage), o = analysis.indexed(start + 4, 16, actionPage);
    const q = analysis.indexed(start + 6, 16, actionPage);
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
    const l = analysis.indexed(start + 1, 16, actionPage), p = analysis.indexed(start + 5, 16, actionPage);
    if (l || p || m || q || ((n ?? 0) & ~2) || ((o ?? 0) & ~8)) {
      actions.push(t("vehicles.decoder.additionalActions")); warning = true;
    }
    if (!actions.length) continue;
    result.push({ row: row + 1, functionKey: `F${keys[0]}`,
      direction: a & 4 ? "forward" : a & 8 ? "reverse" : "both", actions, warning,
      source: `CV${start}–${start + 8} · CV31=16 · CV32=${conditionPage}/${actionPage}` });
  }
  return result;
}
