import type { VehicleCVValueInput } from "../../shared/api";
import type { DecoderAnalysis, Translator } from "./decoderAnalysis";
import type { ManufacturerCVHint } from "./manufacturerCVHints";
import { decoderNumberIsIndexed } from "../../shared/decoderLibrary";

export function cvValueInfo(value: VehicleCVValueInput, analysis: DecoderAnalysis, t: Translator,
  language: "de" | "en" = "de", hint?: ManufacturerCVHint) {
  const confirmedBaseHint = hint?.unindexedAddressConfirmed && analysis.libraryMatch.manufacturerId != null &&
    !decoderNumberIsIndexed(analysis.catalog, analysis.libraryMatch, value.cvNumber);
  const indexUnknown = analysis.indexUnknown(value) && !confirmedBaseHint;
  if (indexUnknown) return { known: false, indexUnknown, category: t("vehicles.decoder.cvDetail"),
    name: t("vehicles.decoder.indexUnresolved"), help: t("vehicles.decoder.indexUnknownHelp"),
    definition: undefined, profile: undefined, manufacturerHint: undefined };
  const documented = analysis.definition(value);
  if (documented) {
    const { definition, profile } = documented;
    const point = definition.cv === 67 && definition.cvEnd === 94 ? ` · ${value.cvNumber - 66}` : "";
    return { known: !definition.reserved, indexUnknown, category: t(`vehicles.decoder.topic.${definition.topic}`),
      name: definition.name[language] + point, help: definition.help[language], definition, profile, manufacturerHint: undefined };
  }
  if (hint) return { known: false, indexUnknown, category: t("decoderLibrary.hintTitle"), name: hint.row.label,
    help: t("decoderLibrary.hintHelp"), definition: undefined, profile: undefined, manufacturerHint: hint };
  return { known: false, indexUnknown, category: t("vehicles.decoder.cvDetail"), name: t("vehicles.cv.noDescription"),
    help: t("vehicles.decoder.unassignedHelp"), definition: undefined, profile: undefined, manufacturerHint: undefined };
}
