import { cvSource, decoderMapping, type DecoderAnalysis, type Translator } from "./decoderAnalysis";
import type { CVExplanationRow, CVTopic } from "./cvExplanationRows";

export function rocoExplanationRows(a: DecoderAnalysis, topic: CVTopic, t: Translator): CVExplanationRow[] {
  const data = a.roco;
  if (!data) return [];
  const unknown = t("vehicles.decoder.unknown");
  const status = (v: boolean | undefined) => v == null ? unknown : t(v ? "common.yes" : "common.no");
  const row = (key: string, reading: string | number | undefined, numbers: number[]): CVExplanationRow => ({
    key: `roco-${key}`, name: t(`vehicles.decoder.roco.${key}.name`), help: t(`vehicles.decoder.roco.${key}.help`),
    reading: reading ?? unknown, source: cvSource(a, ...numbers), numbers
  });
  if (topic === "braking") return [
    row("brakeEnabled", status(data.brakeEnabled), [124, 29]),
    row("brakeType", data.brakeType && t(`vehicles.decoder.roco.brakeType.${data.brakeType}`), [112, 124, 29])
  ];
  if (topic === "outputs") return [
    ...decoderMapping(a, t).map(m => ({ key: `mapping-${m.row}`, name: `${m.functionKey} · ${t(`vehicles.decoder.${m.direction}`)}`,
      help: t("vehicles.decoder.roco.mappingRowHelp"), reading: m.actions.join(" · "),
      note: m.warning ? t("vehicles.decoder.roco.mappingWarning") : undefined, source: m.source, numbers: m.sourceCVs ?? [] })),
    row("shuntingKey", data.shuntingKey, [155]), row("noDelayKey", data.noDelayKey, [156]),
    row("dimming", a.get(60) === 0 ? t("vehicles.decoder.roco.noDimming") : a.get(60), [60]),
    row("dimmingExempt", data.dimmingExempt == null ? undefined : [
      data.dimmingExempt.map(name => t(`vehicles.decoder.roco.output.${name}`)).join(" · ") || t("vehicles.decoder.noneEnabled"),
      ...(data.unsupportedDimmingBits ? [t("vehicles.decoder.roco.unsupportedDimming", { bits: data.unsupportedDimmingBits })] : [])
    ].join(" · "), [114, 60]),
    row("delta", a.get(51), [51])
  ];
  if (topic === "communication") return [
    row("software", a.get(7) == null || a.get(65) == null ? undefined : cvSource(a, 7, 65), [7, 65]),
    row("broadcast", status(data.broadcast), [28, 29]), row("cvData", status(data.cvData), [28, 29])
  ];
  return [];
}
