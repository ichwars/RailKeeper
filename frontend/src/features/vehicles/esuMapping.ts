import type { DecoderAnalysis, MappingSuggestion, Translator } from "./decoderAnalysis";

// Reviewed ESU V4/V5 layouts, selected only through a matched model profile.
// Only simple positive single-key conditions are named. Missing fields stay unknown.
export function esuMapping(analysis: DecoderAnalysis, t: Translator): MappingSuggestion[] {
  const layout = analysis.libraryProfile?.mapping;
  if ((!analysis.esu5 && !analysis.esuV4) || !layout) return [];
  const result: MappingSuggestion[] = [];
  for (let row = 0; row < layout.rows; row++) {
    const start = layout.startCV + (row % layout.rowsPerPage) * layout.stride;
    const conditionPage = layout.conditionPage + Math.floor(row / layout.rowsPerPage);
    const actionPage = layout.actionPage + Math.floor(row / layout.rowsPerPage);
    const conditions = Array.from({ length: 9 }, (_, i) => analysis.indexed(start + i, layout.indexCV31, conditionPage));
    if (conditions.some((value) => value == null)) continue;
    const extraCondition = layout.format === "esu-v4" ? 0 : analysis.indexed(start + 9, layout.indexCV31, conditionPage);
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
    const v4 = layout.format === "esu-v4", actionStart = v4 ? start + 9 : start;
    const k = analysis.indexed(actionStart, layout.indexCV31, actionPage);
    const m = v4 ? 0 : analysis.indexed(actionStart + 2, layout.indexCV31, actionPage);
    const n = analysis.indexed(actionStart + (v4 ? 2 : 3), layout.indexCV31, actionPage), o = v4 ? 0 : analysis.indexed(actionStart + 4, layout.indexCV31, actionPage);
    const q = analysis.indexed(actionStart + (v4 ? 4 : 6), layout.indexCV31, actionPage);
    if ([k, m, n, o, q].some((value) => value == null)) continue;
    const actions: string[] = [];
    let warning = false;
    for (let bit = 0; bit < 8; bit++) if ((k ?? 0) & (1 << bit)) {
      if (bit < 2) actions.push(t(bit === 0 ? "vehicles.decoder.frontLight" : "vehicles.decoder.rearLight"));
      else {
        const output = bit - 1;
        const powerPack = !v4 && output === 3 && analysis.indexed(291, 16, 0) === 31;
        actions.push(`AUX${output}${powerPack ? " · PowerPack Control" : ""}`);
        warning ||= powerPack;
      }
    }
    if ((n ?? 0) & 2) actions.push(t("vehicles.decoder.shunting"));
    if (v4 ? (n ?? 0) & 1 : (o ?? 0) & 8) actions.push(t("vehicles.decoder.noDelay"));
    const l = analysis.indexed(actionStart + 1, layout.indexCV31, actionPage);
    const p = analysis.indexed(actionStart + (v4 ? 3 : 5), layout.indexCV31, actionPage);
    if (v4 && (l == null || p == null)) continue;
    // Omitted Nano control bytes may contain additional outputs or conditions.
    if (analysis.libraryProfile?.id === "esu-lokpilot-5-nano-dcc" &&
      (extraCondition == null || l == null || p == null)) continue;
    if (l || p || m || q || ((n ?? 0) & ~(v4 ? 3 : 2)) || ((o ?? 0) & ~8)) {
      actions.push(t("vehicles.decoder.additionalActions")); warning = true;
    }
    if (!actions.length) continue;
    result.push({ row: row + 1, functionKey: `F${keys[0]}`,
      direction: a & 4 ? "forward" : a & 8 ? "reverse" : "both", actions, warning,
      source: `CV${start}–${start + 8} · CV31=${layout.indexCV31} · CV32=${conditionPage}/${actionPage}` });
  }
  return result;
}

// LokPilot Standard, ESU instruction manual 2nd edition (2015), function mapping table.
// Directional rows use their own registers; missing reverse registers are never inferred.
export function esuStandardMapping(a: DecoderAnalysis, t: Translator): MappingSuggestion[] {
  if (!a.esuStandard) return [];
  const entries: { cv: number; key: number; direction: MappingSuggestion["direction"] }[] = [
    { cv: 33, key: 0, direction: "forward" }, { cv: 34, key: 0, direction: "reverse" },
    ...Array.from({ length: 12 }, (_, i) => ({ cv: 35 + i, key: i + 1,
      direction: ([1, 4, 5, 6, 7].includes(i + 1) ? "forward" : "both") as MappingSuggestion["direction"] })),
    { cv: 47, key: 1, direction: "reverse" },
    ...Array.from({ length: 6 }, (_, i) => ({ cv: 129 + i, key: 13 + i, direction: "both" as const })),
    ...Array.from({ length: 4 }, (_, i) => ({ cv: 135 + i, key: 4 + i, direction: "reverse" as const }))
  ];
  return entries.flatMap(({ cv, key, direction }, i) => {
    const value = a.get(cv);
    if (value == null || !Number.isInteger(value) || value < 1 || value > 255) return [];
    const actions: string[] = [];
    for (let bit = 0; bit < 8; bit++) if (value & (1 << bit)) {
      actions.push(bit < 2 ? t(bit === 0 ? "vehicles.decoder.frontLight" : "vehicles.decoder.rearLight")
        : bit < 6 ? `AUX${bit - 1}` : t(bit === 6 ? "vehicles.decoder.shunting" : "vehicles.decoder.noDelay"));
    }
    return [{ row: i + 1, functionKey: `F${key}`, direction, actions, warning: false,
      source: `CV${cv}=${value}`, sourceCVs: [cv] }];
  });
}
