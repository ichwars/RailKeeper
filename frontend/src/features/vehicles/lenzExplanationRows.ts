import { cvSource, type DecoderAnalysis, type Translator } from "./decoderAnalysis";
import type { CVExplanationRow, CVTopic } from "./cvExplanationRows";

export function lenzExplanationRows(a: DecoderAnalysis, topic: CVTopic, t: Translator, language: "de" | "en"): CVExplanationRow[] {
  const data = a.lenz;
  if (!data) return [];
  const unknown = t("vehicles.decoder.unknown");
  const status = (value: boolean | undefined) => value == null ? unknown : t(value ? "common.yes" : "common.no");
  const list = (keys: string[] | undefined) => keys?.join(" · ") || (keys ? t("vehicles.decoder.noneEnabled") : unknown);
  const row = (key: string, reading: string | number | undefined, numbers: number[]): CVExplanationRow => ({
    key: `lenz-${key}`, name: t(`vehicles.decoder.lenz.${key}.name`), help: t(`vehicles.decoder.lenz.${key}.help`),
    reading: reading ?? unknown, source: cvSource(a, ...numbers), numbers
  });
  if (topic === "motor") return [
    row("motorType", data.motorType, [50, 113, 114]),
    row("functionMode", status(data.functionDecoder), [50]),
    row("loadControl", status(data.loadControl), [50]),
    row("emk", status(data.emkDivider), [50]),
    row("frequency", data.lowFrequency == null ? undefined : data.lowFrequency ? "≈19 Hz" : "≈23 kHz", [50])
  ];
  if (topic === "braking") return [
    row("abcDistance", status(data.constantABC), [51, 52]),
    row("zeroDistance", status(data.constantZero), [51, 52]),
    row("abcDirection", status(data.abcDirectionIndependent), [51]),
    row("shuttle", data.shuttleMode && t(`vehicles.decoder.lenz.shuttle.${data.shuttleMode}`), [51, 54]),
    row("dcStop", status(data.dcStop), [29, 51])
  ];
  if (topic === "outputs") {
    const effects = [data.effectA, data.effectB, data.effectC, data.effectD];
    const seconds = (cv: number) => a.get(cv) == null ? undefined
      : `${(a.get(cv)! * 0.016).toLocaleString(language, { maximumFractionDigits: 3 })} s`;
    return [
      row("dimmingKeys", list(data.dimmingKeys), [57]),
      row("shuntingKeys", list(data.shuntingKeys), [58]),
      row("noDelayKeys", list(data.noDelayKeys), [59]),
      row("effectsABKeys", list(data.effectsABKeys), [61]),
      row("effectsCDKeys", list(data.effectsCDKeys), [64]),
      ...effects.map((effect, i) => row(`effect${"ABCD"[i]}`, effect == null ? undefined
        : t(`vehicles.decoder.lenz.effect.${i < 2 ? "ab" : i === 2 ? "c" : "d"}.${effect}`), [i < 2 ? 60 : 62])),
      row("flash", data.flashHz == null ? undefined
        : `≈${data.flashHz.toLocaleString(language, { maximumFractionDigits: 2 })} Hz`, [63]),
      ...[147, 148, 149].map(cv => row(`time${cv}`, seconds(cv), [cv]))
    ];
  }
  if (topic === "communication") return [row("faults", data.faultBits == null ? undefined
    : data.faultBits.length ? data.faultBits.map(bit => t(`vehicles.decoder.lenz.fault.${bit}`)).join(" · ")
      : t("vehicles.decoder.lenz.fault.none"), [30])];
  return [];
}
