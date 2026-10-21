import { cvSource, type DecoderAnalysis, type Translator } from "./decoderAnalysis";
import type { CVExplanationRow, CVTopic } from "./cvExplanationRows";

export function kuehnExplanationRows(a: DecoderAnalysis, topic: CVTopic, t: Translator, language: "de" | "en"): CVExplanationRow[] {
  const data = a.kuehn;
  if (!data) return [];
  const unknown = t("vehicles.decoder.unknown");
  const status = (value: boolean | undefined) => value == null ? unknown : t(value ? "common.yes" : "common.no");
  const list = (values: string[] | undefined) => values?.join(" · ") || (values ? t("vehicles.decoder.noneEnabled") : unknown);
  const row = (key: string, reading: string | number | undefined, numbers: number[]): CVExplanationRow => ({
    key: `kuehn-${key}`, name: t(`vehicles.decoder.kuehn.${key}.name`), help: t(`vehicles.decoder.kuehn.${key}.help`),
    reading: reading ?? unknown, source: cvSource(a, ...numbers), numbers
  });
  const documented = (cv: number, reading: string | number | undefined): CVExplanationRow => {
    const d = a.definition({ cvNumber: cv, value: 0 })?.definition;
    return { key: `kuehn-${cv}`, name: d?.name[language] ?? t("vehicles.cv.noDescription"),
      help: d?.help[language] ?? t("vehicles.decoder.unassignedHelp"), reading: reading ?? unknown,
      source: cvSource(a, cv), numbers: [cv] };
  };
  if (topic === "motor") return [
    documented(9, data.frequency == null ? undefined : data.frequency >= 1000
      ? `≈${(data.frequency / 1000).toLocaleString(language)} kHz` : `≈${data.frequency} Hz`),
    row("loadControl", status(data.loadControl), [56, 53, 54]),
    row("emk", status(data.emkDivider), [56]),
    documented(95, data.reverseFactor == null ? undefined : `×${data.reverseFactor.toLocaleString(language, { maximumFractionDigits: 4 })}`)
  ];
  if (topic === "braking") return [
    row("dcOpposite", status(data.dcOppositeRetained), [27]),
    row("dcForward", status(data.dcForward), [27])
  ];
  if (topic === "communication") return [
    row("susi", status(data.susi), [56]), row("memory", status(data.memory), [112])
  ];
  if (topic !== "outputs") return [];
  return [
    row("analogKeys", list(data.analogKeys), [13, 14, 33, 34]),
    row("shuntingKeys", list(data.shuntingKeys), [58]), row("noDelayKeys", list(data.noDelayKeys), [58]),
    row("dimmingKeys", list(data.dimmingKeys), [59, 55]),
    row("shuntingLight", data.shuntingLight == null ? undefined : `${list(data.shuntingLight)} · ${t("vehicles.decoder.kuehn.requiresF0")}: ${status(data.shuntingNeedsF0)}`, [57, 56]),
    ...data.effects.map((effect, i) => {
      const cv = [49, 50, 51, 52, 47, 48][i];
      const reading = effect == null ? undefined : [effect.code == null ? unknown : t(`vehicles.decoder.kuehn.effect.${effect.code}`),
        ...(effect.offForward ? [t("vehicles.decoder.kuehn.offForward")] : []),
        ...(effect.offReverse ? [t("vehicles.decoder.kuehn.offReverse")] : [])].join(" · ");
      return documented(cv, reading);
    }),
    documented(55, data.effectCycle == null ? undefined : t("vehicles.decoder.kuehn.cycleDimming", {
      cycle: data.effectCycle, level: data.effectDimming ?? unknown
    })),
    ...data.dimming.map((digits, i) => documented(60 + i, digits == null ? undefined
      : `${"ACE"[i]}: ${digits.units} · ${"BDF"[i]}: ${digits.tens}`)),
    documented(63, data.pullInSeconds == null ? undefined : `${data.pullInSeconds.toLocaleString(language)} s · ${t("vehicles.decoder.kuehn.holding", {
      level: data.holdingLevel === 8 ? t("vehicles.decoder.kuehn.fullyOff") : data.holdingLevel ?? unknown
    })}`),
    documented(113, data.randomSeconds == null ? undefined : `${data.randomSeconds.toLocaleString(language)} s`)
  ];
}
