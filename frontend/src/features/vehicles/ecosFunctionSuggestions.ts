import type { ECoSFunctionValue, MasterDataEntry, VehicleFunctionInput } from "../../shared/api";
import { exactECoSSymbol } from "../../shared/ecosSymbols";
import type { DecoderAnalysis, MappingSuggestion, Translator } from "./decoderAnalysis";
import { functionCVSuggestion } from "./functionCVSuggestions";
import { railComPlusFlagAvailability } from "./railComPlusFunctionFlags";
import type { FunctionEdit } from "./vehicleTransforms";

export function ecosFunctionSuggestion(value: ECoSFunctionValue, analysis: DecoderAnalysis,
  current: FunctionEdit, symbols: MasterDataEntry[], readAt: string, objectID: number, t: Translator,
  host = "", mappings?: MappingSuggestion[]) {
  const key = `F${value.index}`, stationSymbol = exactECoSSymbol(value.description, symbols);
  // Compare independent evidence, including F0; editor defaults must not hide
  // the library's symbol or mode before checking for conflicts.
  const library = functionCVSuggestion(analysis, key, {}, symbols, t, mappings).patch;
  const mode = value.momentary ? "moment" : "dauer";
  const libraryIcon = analysis.railComPlus?.functions.find(row => row.functionKey === key);
  const conflict = Boolean(libraryIcon && libraryIcon.iconCode !== value.description ||
    library?.mode && library.mode !== mode);
  const patch: Partial<VehicleFunctionInput> = {};
  if (!current.deleted && !conflict) {
    const symbolKey = library?.symbolKey || stationSymbol?.key;
    const name = library?.name || stationSymbol?.label;
    if (!current.name?.trim() && name) patch.name = name;
    if (!current.symbolKey && symbolKey) patch.symbolKey = symbolKey;
    if (!current.persisted && !current.functionType && library?.functionType) patch.functionType = library.functionType;
    const available = railComPlusFlagAvailability(current);
    if (available.mode) patch.mode = mode;
    if (available.inverted && library?.inverted != null) patch.inverted = library.inverted;
  }
  const hasPatch = Object.keys(patch).length > 0;
  if (hasPatch) {
    patch.flagsVersion = 2;
    patch.notes = [current.notes, library?.notes,
      `ECoS · #${objectID} · ${key} · funcexists=${value.description} · ${mode} · ${readAt}${host ? ` · ${host}` : ""}`]
      .filter(Boolean).join("\n");
  }
  const notesFull = Boolean(patch.notes && new TextEncoder().encode(patch.notes).length > 1000);
  return { key, stationSymbol, library, mode,
    status: conflict ? "conflict" : current.deleted || !hasPatch ? "protected" : notesFull ? "notesFull" : "ready",
    patch: hasPatch && !notesFull ? patch : undefined };
}
