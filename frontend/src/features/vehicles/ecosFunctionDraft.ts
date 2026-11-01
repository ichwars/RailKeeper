import type { MasterDataEntry } from "../../shared/api";
import type { DecoderLibrary } from "../../shared/decoderLibraryModel";
import { hasECoSDecoderIdentityConflict } from "../../shared/ecosSourceConflicts";
import { decoderAnalysis, decoderMapping, type Translator } from "./decoderAnalysis";
import { ecosFunctionSuggestion } from "./ecosFunctionSuggestions";
import type { ECoSVehicleDraftPayload } from "./vehicleViewModel";

export function supplementECoSFunctionDraft(draft: ECoSVehicleDraftPayload, symbols: MasterDataEntry[],
  library: DecoderLibrary | null, t: Translator): ECoSVehicleDraftPayload {
  const snapshot = draft.ecosFunctions;
  if (draft.source !== "ecos" || draft.mode !== "create" || !snapshot ||
    String(snapshot.objectId) !== draft.externalMapping.externalId || !snapshot.host.trim() ||
    !Number.isFinite(Date.parse(snapshot.readAt)) || snapshot.functions.length > 32) return draft;
  const indices = new Set<number>();
  for (const value of snapshot.functions) {
    if (!Number.isInteger(value.index) || value.index < 0 || value.index > 31 || indices.has(value.index) ||
      !Number.isInteger(value.description) || value.description < 0 || value.description > 2147483647 ||
      typeof value.momentary !== "boolean") return draft;
    indices.add(value.index);
  }
  const identityConflict = hasECoSDecoderIdentityConflict(draft.cvSourceConflicts, library?.inventory);
  const analysis = decoderAnalysis(draft.cvValues, "ECoS", identityConflict ? [] : library?.profiles || [],
    identityConflict ? null : library?.inventory || null,
    draft.vehicle.decoderType || "", { name: draft.sourceSummary.profile, objectId: String(snapshot.objectId),
      host: snapshot.host, readAt: snapshot.readAt });
  const mappings = decoderMapping(analysis, t);
  const functionValues = [...draft.functionValues];
  for (const value of snapshot.functions) {
    const functionKey = `F${value.index}`;
    if (functionValues.some(row => row.functionKey === functionKey)) continue;
    const proposal = ecosFunctionSuggestion(value, analysis, {}, symbols,
      snapshot.readAt, snapshot.objectId, t, snapshot.host, mappings);
    if (proposal.status === "ready" && proposal.patch) functionValues.push({ functionKey, ...proposal.patch });
  }
  return { ...draft, functionValues };
}
