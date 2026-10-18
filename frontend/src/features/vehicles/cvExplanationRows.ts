import { cvSource, esuMapping, type DecoderAnalysis, type Translator } from "./decoderAnalysis";

export type CVTopic = "address" | "motor" | "braking" | "outputs" | "communication";
export type CVExplanationRow = {
  key: string; name: string; help: string; reading: string | number; note?: string; source: string; numbers: number[];
  mappingSource?: { cvNumber: number; indexCV31: number; indexCV32: number };
};

export function cvExplanationRows(a: DecoderAnalysis, topic: CVTopic, t: Translator): CVExplanationRow[] {
  const unknown = t("vehicles.decoder.unknown");
  const status = (value: boolean | undefined) => value == null ? unknown : t(value ? "common.yes" : "common.no");
  const bit = (number: number) => a.config == null ? undefined : Boolean(a.config & (1 << number));
  const row = (key: string, reading: string | number | undefined, numbers: number[], note?: string): CVExplanationRow => ({
    key, name: t(`vehicles.decoder.reading.${key}.name`), help: t(`vehicles.decoder.reading.${key}.help`),
    reading: reading ?? unknown, note, source: cvSource(a, ...numbers), numbers
  });
  const raw = (number: number): CVExplanationRow => ({ key: String(number),
    name: t(`vehicles.decoder.cv.${number}.name`), help: t(`vehicles.decoder.cv.${number}.help`),
    reading: a.get(number) ?? unknown, source: cvSource(a, number), numbers: [number]
  });
  if (topic === "address") return [
    row("address", a.address, bit(5) ? [17, 18, 29] : [1, 29], a.config == null ? unknown
      : t(bit(5) ? "vehicles.decoder.longDccAddress" : "vehicles.decoder.shortDccAddress")),
    row("longAddress", a.longAddress, [17, 18, 29], a.config == null ? unknown
      : t(bit(5) ? "vehicles.cv.enabled" : "vehicles.decoder.inactive")),
    row("steps", a.config == null ? undefined : bit(1) ? "28/128" : "14", [29], a.config == null ? unknown
      : t(`vehicles.decoder.bit29.0.${bit(0) ? "on" : "off"}`)),
    row("analog", a.esu5 ? a.dcAnalog == null ? unknown
      : t(a.dcAnalog ? "vehicles.decoder.dcEnabled" : "vehicles.decoder.dcDisabled") : status(bit(2)), a.esu5 ? [50, 29] : [29]),
    row("railcom", status(bit(3)), [28, 29]),
    row("consist", a.consistAddress == null ? undefined : a.consistAddress === 0
      ? t("vehicles.decoder.inactive") : `${a.consistAddress} · ${t(a.consistReverse
        ? "vehicles.decoder.reverse" : "vehicles.decoder.forward")}`, [19]),
    row("consistFunctions", a.functionsComplete ? a.enabledFunctions.join(" · ") || t("vehicles.decoder.noneEnabled")
      : unknown, [21, 22])
  ];
  if (topic === "motor") return [
    row("curve", a.config == null ? undefined : t(`vehicles.decoder.bit29.4.${bit(4) ? "on" : "off"}`), [29]),
    ...[2, 6, 5, 3, 4].map(raw)
  ];
  if (topic === "braking") return [
    row("abc", status(a.abc), [27]), row("constantBrake", status(a.constantBrake), [254, 255]),
    ...[27, 253, 254, 255].map(raw)
  ];
  if (topic === "outputs") return esuMapping(a, t).map((mapping) => ({
    key: `mapping-${mapping.row}`, name: `${mapping.functionKey} · ${t(`vehicles.decoder.${mapping.direction}`)}`,
    help: mapping.warning ? t("vehicles.decoder.mappingLimits") : t("vehicles.decoder.mappingHelp"),
    reading: mapping.actions.join(" · "), source: mapping.source, numbers: [],
    mappingSource: { cvNumber: 257 + ((mapping.row - 1) % 16) * 16, indexCV31: 16, indexCV32: 3 + Math.floor((mapping.row - 1) / 16) }
  }));
  return [row("firmware", a.firmware, [7]), ...[8, 28, 29, 47, 50].map(raw)];
}
