import type { VehicleCVValueInput } from "../../shared/api";
import type { DecoderAnalysis, Translator } from "./decoderAnalysis";

const standardCVs = [1, 2, 3, 4, 5, 6, 7, 8, 17, 18, 19, 21, 22, 28, 29];
const esuCVs = [27, 47, 50, 253, 254, 255];

export function cvValueInfo(value: VehicleCVValueInput, analysis: DecoderAnalysis, t: Translator) {
  const base = value.indexCV31 == null && value.indexCV32 == null;
  const topic = [1, 17, 18, 19, 21, 22, 29].includes(value.cvNumber) ? "address"
    : [2, 3, 4, 5, 6].includes(value.cvNumber) ? "motor"
      : [27, 253, 254, 255].includes(value.cvNumber) ? "braking" : "communication";
  if (base && (standardCVs.includes(value.cvNumber) || analysis.esu5 && esuCVs.includes(value.cvNumber))) {
    return { known: true, category: t(`vehicles.decoder.topic.${topic}`), name: t(`vehicles.decoder.cv.${value.cvNumber}.name`),
      help: t(`vehicles.decoder.cv.${value.cvNumber}.help`) };
  }
  if (base && value.cvNumber >= 67 && value.cvNumber <= 94) {
    return { known: true, category: t("vehicles.decoder.topic.motor"), name: `${t("vehicles.decoder.speedPoint")} ${value.cvNumber - 66}`,
      help: t("vehicles.decoder.speedPointHelp") };
  }
  return { known: false, category: t("vehicles.decoder.cvDetail"), name: t("vehicles.cv.noDescription"), help: t("vehicles.decoder.unassignedHelp") };
}
