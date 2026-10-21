import { cvSource, decoderMapping, type DecoderAnalysis, type Translator } from "./decoderAnalysis";
import type { CVExplanationRow, CVTopic } from "./cvExplanationRows";

export function pikoExplanationRows(a: DecoderAnalysis, topic: CVTopic, t: Translator): CVExplanationRow[] {
  const data = a.piko;
  if (!data) return [];
  const unknown = t("vehicles.decoder.unknown");
  const status = (v: boolean | undefined) => v == null ? unknown : t(v ? "common.yes" : "common.no");
  const row = (key: string, reading: string | number | undefined, numbers: number[]): CVExplanationRow => ({
    key: `piko-${key}`, name: t(`vehicles.decoder.piko.${key}.name`), help: t(`vehicles.decoder.piko.${key}.help`),
    reading: reading ?? unknown, source: cvSource(a, ...numbers), numbers
  });
  if (topic === "braking") return [
    row("distance", data.brakeDistanceCm == null ? undefined : `${data.brakeDistanceCm} cm`, [141, 145, 142]),
    row("triggers", data.triggers == null ? undefined : [
      ...["ABC", "DC", "DCC-Broadcast", "DCC 0"].filter((_, bit) => data.triggers! & (1 << bit)),
      ...(data.triggers & 48 ? [unknown] : [])
    ].join(" · ") || t("vehicles.decoder.noneEnabled"), [145, 27]),
    row("testKey", a.get(142) === 0 ? t("vehicles.decoder.inactive") : data.testKey, [142]),
    row("threshold", data.abcThresholdV == null ? undefined : `≈ ${data.abcThresholdV.toFixed(1)} V`, [48]),
    row("slowSpeed", a.get(50), [50]),
    row("shuttleWait", a.get(51) == null ? undefined : `${a.get(51)! * 100} ms`, [51]),
    row("reverseWait", a.get(52) == null ? undefined : `${a.get(52)! * 100} ms`, [52])
  ];
  if (topic === "communication") return [
    row("protocols", data.protocolNames?.join(" · ") || (data.protocolNames ? t("vehicles.decoder.noneEnabled") : undefined), [12, 29]),
    row("acAnalog", status(data.acAnalog), [12, 29]),
    row("railcomPlus", status(data.railcomPlus), [28, 29]),
    row("lock", status(data.lockMatches), [15, 16]),
    row("analogThresholds", unknown, [102, 103, 104])
  ];
  if (topic !== "outputs") return [];
  return [
    row("mappingMode", data.mappingMode == null ? undefined : data.mappingMode === 1
      ? t("vehicles.decoder.piko.simple") : t("vehicles.decoder.piko.advanced"), [96]),
    ...decoderMapping(a, t).map(m => ({ key: `mapping-${m.row}`, name: `${m.functionKey} · ${t(`vehicles.decoder.${m.direction}`)}`,
      help: t("vehicles.decoder.piko.mappingRowHelp"), reading: m.actions.join(" · "), source: m.source, numbers: m.sourceCVs ?? [],
      note: m.warning ? t("vehicles.decoder.piko.unavailable") : undefined })),
    ...data.outputSets.map(o => ({ key: `piko-output-${o.output}-${o.set}`,
      name: `${o.output < 2 ? t(o.output === 0 ? "vehicles.decoder.frontLight" : "vehicles.decoder.rearLight") : `A${o.output - 1}`} · ${o.set}`,
      help: t("vehicles.decoder.piko.outputHelp"), reading: o.effect == null ? unknown : t(`vehicles.decoder.piko.effect.${o.effect}`),
      note: [o.pwm == null ? `PWM: ${unknown}` : `PWM: ${o.pwm}/64`,
        t("vehicles.decoder.piko.autoOff", { value: status(o.autoOff) }),
        ...(o.autoOffSeconds == null ? [] : [`${o.autoOffSeconds} s`]),
        ...(o.smoke ? [t("vehicles.decoder.piko.smoke")] : [t("vehicles.decoder.piko.fade", { on: status(o.fadeIn), off: status(o.fadeOut) })])].join(" · "),
      source: `CV31=18 / CV32=0 · ${o.raw.map((value, offset) => `CV${o.start + offset}=${value ?? "?"}`).join(" · ")}`,
      numbers: [o.sourceCV], mappingSource: { cvNumber: o.sourceCV, indexCV31: 18, indexCV32: 0 }
    }))
  ];
}
