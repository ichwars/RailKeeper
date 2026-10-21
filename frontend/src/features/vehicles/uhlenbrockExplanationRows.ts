import { cvSource, decoderMapping, type DecoderAnalysis, type Translator } from "./decoderAnalysis";
import type { CVExplanationRow, CVTopic } from "./cvExplanationRows";

export function uhlenbrockExplanationRows(a: DecoderAnalysis, topic: CVTopic, t: Translator, language: "de" | "en"): CVExplanationRow[] {
  const data = a.uhlenbrock;
  if (!data) return [];
  const unknown = t("vehicles.decoder.unknown");
  const status = (value: boolean | undefined) => value == null ? unknown : t(value ? "common.yes" : "common.no");
  const row = (key: string, reading: string | number | undefined, numbers: number[]): CVExplanationRow => ({
    key: `uhlenbrock-${key}`, name: t(`vehicles.decoder.uhlenbrock.${key}.name`),
    help: t(`vehicles.decoder.uhlenbrock.${key}.help`), reading: reading ?? unknown,
    source: cvSource(a, ...numbers), numbers
  });
  const raw = (cv: number): CVExplanationRow => {
    const d = a.definition({ cvNumber: cv, value: 0 })?.definition;
    return { key: `uhlenbrock-${cv}`, name: d?.name[language] ?? t("vehicles.cv.noDescription"),
      help: d?.help[language] ?? t("vehicles.decoder.unassignedHelp"), reading: a.get(cv) ?? unknown,
      source: cvSource(a, cv), numbers: [cv] };
  };
  if (topic === "address") return [row("acAnalog", status(data.acAnalog), [29, 51])];
  if (topic === "motor") return [
    row("loadControl", status(data.loadControl), [49]), row("parameterOrder", status(data.parameterOrder), [2, 6, 5]),
    ...[47, 48, 53, 54, 55, 56, 57, 58].map(raw)
  ];
  if (topic === "braking") return [
    row("dcBrake", status(data.dcBrake), [29, 49]),
    row("brakeTarget", data.brakeTarget, data.targetSelected === false ? [49] : [49, 52])
  ];
  if (topic === "outputs") return [
    ...decoderMapping(a, t).map(m => ({ key: `mapping-${m.row}`, name: `${m.functionKey} · ${t(`vehicles.decoder.${m.direction}`)}`,
      help: t("vehicles.decoder.uhlenbrock.mappingRowHelp"), reading: m.actions.join(" · "), source: m.source, numbers: m.sourceCVs ?? [] })),
    raw(50), row("dimmingAux", status(data.dimmingAux), [49, 50]),
    row("lightsSwapped", status(data.lightsSwapped), [49]),
    row("timedOutputs", data.timedOutputs?.join(" · ") || (data.timedOutputs ? t("vehicles.decoder.noneEnabled") : unknown), [98]),
    row("duration", data.durationSeconds == null ? undefined
      : `${data.durationSeconds.toLocaleString(language, { maximumFractionDigits: 1 })} s`, [99, 98])
  ];
  if (topic === "communication") return [
    row("protocols", data.protocols && t(`vehicles.decoder.uhlenbrock.protocol.${data.protocols}`), [49]),
    row("susi", data.susi && t(`vehicles.decoder.uhlenbrock.susi.${data.susi}`), [49]),
    row("shortCircuit", status(data.shortCircuit), [60])
  ];
  return [];
}
