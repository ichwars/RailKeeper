import type { MasterDataEntry, VehicleFunctionInput } from "../../shared/api";
import { exactECoSSymbol } from "../../shared/ecosSymbols";
import { decoderMapping, esuMapping, type DecoderAnalysis, type MappingSuggestion, type Translator } from "./decoderAnalysis";
import { railComPlusFlagAvailability, railComPlusFlagPatch } from "./railComPlusFunctionFlags";
import type { FunctionEdit } from "./vehicleTransforms";

export function functionCVSuggestion(analysis: DecoderAnalysis, key: string, current: FunctionEdit,
  symbols: MasterDataEntry[], t: Translator, mappingRows: MappingSuggestion[] = analysis.esuStandard
    ? decoderMapping(analysis, t) : esuMapping(analysis, t)): { patch?: Partial<VehicleFunctionInput>; notesFull?: boolean } {
  if (current.deleted) return {};
  const data = analysis.railComPlus, row = data?.functions.find(item => item.functionKey === key);
  const flags = data && row ? railComPlusFlagPatch(data, row, current, t) : undefined;
  const patch: Partial<VehicleFunctionInput> = { ...flags };
  const symbol = row && !current.symbolKey ? exactECoSSymbol(row.iconCode, symbols) : undefined;
  const mappings = mappingRows.filter(item => item.functionKey === key);
  const actions = [...new Set(mappings.flatMap(item => item.actions))];
  const knownName = mappings.length > 0 && mappings.every(item => !item.warning) &&
    !actions.some(action => /^AUX\d/.test(action));
  const sources: string[] = [];
  if (!current.name?.trim() && knownName) {
    patch.name = actions.join(" / ");
    const context = analysis.libraryMatch.matchedBy === "ecos-profile" && analysis.ecosDecoderProfile
      ? ` · ECoS-Profil: ${analysis.ecosDecoderProfile.name}` : "";
    sources.push(`CV-Mapping · ${key} · ${patch.name} · ${mappings.map(item => item.source).join("; ")}${context}`);
  }
  if (symbol && row) {
    patch.symbolKey = symbol.key;
    if (!current.persisted && (!current.functionType || current.functionType === "standard") && symbol.metadata?.category === "Licht") {
      patch.functionType = "licht";
    }
    if (flags?.notes) patch.notes = `${flags.notes} · ${t("railcom.symbol.code", { code: row.iconCode })}`;
    else sources.push([
      "RailComPlus", "CV31=1 / CV32=1", `CV${row.firstCV}=${row.firstValue}; CV${row.firstCV + 1}=${row.secondValue}`,
      t("railcom.symbol.code", { code: row.iconCode }),
      ...([row.sourceLine, row.secondSourceLine].some(line => line != null)
        ? [t("railcom.line", { line: [row.sourceLine, row.secondSourceLine].filter(line => line != null).join(" / ") })] : []),
      ...(row.sourceExportedAt ? [`${t("vehicles.decoder.exportedAt")}: ${row.sourceExportedAt}`] : [])
    ].join(" · "));
  }
  if (!Object.keys(patch).length) {
    const available = railComPlusFlagAvailability(current);
    const canFillFlags = data?.flagsVerification && row && (available.mode || available.inverted);
    return { notesFull: Boolean(canFillFlags) };
  }
  patch.notes = [patch.notes || current.notes, ...sources].filter(Boolean).join("\n");
  if (new TextEncoder().encode(patch.notes).length > 1000) return { notesFull: true };
  return { patch };
}
