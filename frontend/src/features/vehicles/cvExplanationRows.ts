import { cvSource, decoderMapping, type DecoderAnalysis, type Translator } from "./decoderAnalysis";
import { pikoExplanationRows } from "./pikoExplanationRows";
import { rocoExplanationRows } from "./rocoExplanationRows";
import { uhlenbrockExplanationRows } from "./uhlenbrockExplanationRows";
import { lenzExplanationRows } from "./lenzExplanationRows";
import { kuehnExplanationRows } from "./kuehnExplanationRows";

export type CVTopic = "address" | "motor" | "braking" | "outputs" | "communication";
export type CVExplanationRow = {
  key: string; name: string; help: string; reading: string | number; note?: string; source: string; numbers: number[];
  mappingSource?: { cvNumber: number; indexCV31: number; indexCV32: number };
};

export function cvExplanationRows(a: DecoderAnalysis, topic: CVTopic, t: Translator, language: "de" | "en" = "de"): CVExplanationRow[] {
  const unknown = t("vehicles.decoder.unknown");
  const status = (value: boolean | undefined) => value == null ? unknown : t(value ? "common.yes" : "common.no");
  const bit = (number: number) => a.config == null ? undefined : Boolean(a.config & (1 << number));
  const row = (key: string, reading: string | number | undefined, numbers: number[], note?: string): CVExplanationRow => ({
    key, name: t(`vehicles.decoder.reading.${key}.name`), help: t(`vehicles.decoder.reading.${key}.help`),
    reading: reading ?? unknown, note, source: cvSource(a, ...numbers), numbers
  });
  const raw = (number: number): CVExplanationRow => {
    const documented = a.definition({ cvNumber: number, value: 0 });
    return { key: String(number), name: documented?.definition.name[language] || t("vehicles.cv.noDescription"),
      help: documented?.definition.help[language] || t("vehicles.decoder.unassignedHelp"),
      reading: a.get(number) ?? unknown, source: cvSource(a, number), numbers: [number] };
  };
  const distanceNote = (fine: boolean, numbers: number[]) =>
    `${t(fine ? "vehicles.decoder.zimo.fineParameter" : "vehicles.decoder.zimo.basicParameter")} · ${cvSource(a, ...numbers)}`;
  if (topic === "address") return [
    row("address", a.address, bit(5) ? [17, 18, 29] : [1, 29], a.config == null ? unknown
      : t(bit(5) ? "vehicles.decoder.longDccAddress" : "vehicles.decoder.shortDccAddress")),
    row("longAddress", a.longAddress, [17, 18, 29], a.config == null ? unknown
      : t(bit(5) ? "vehicles.cv.enabled" : "vehicles.decoder.inactive")),
    row("steps", a.config == null ? undefined : bit(1) ? "28/128" : a.lenz || a.kuehn || a.uhlenbrock || a.piko ? "14/27" : "14", [29], a.config == null ? unknown
      : t(`vehicles.decoder.bit29.0.${bit(0) ? "on" : "off"}`)),
    { ...row("analog", a.rules?.dcAnalog || a.kuehn || a.uhlenbrock || a.roco ? a.dcAnalog == null ? unknown
      : t(a.dcAnalog ? "vehicles.decoder.dcEnabled" : "vehicles.decoder.dcDisabled") : status(bit(2)),
    a.uhlenbrock ? [29, 51] : a.kuehn ? [27, 29] : a.rules?.dcAnalog ? [a.rules.dcAnalog.cv, 29] : [29]), ...(a.uhlenbrock ? { help: t("vehicles.decoder.uhlenbrock.analogHelp") } : {}) },
    { ...row("railcom", status(a.railcomEnabled), a.uhlenbrock ? [29] : [28, 29]),
      ...(a.piko ? { help: t("vehicles.decoder.piko.railcom.help") } : a.roco ? { help: t("vehicles.decoder.roco.railcomHelp") } : a.uhlenbrock ? { help: t("vehicles.decoder.uhlenbrock.railcomHelp") } : a.lenz ? { help: t("vehicles.decoder.lenz.railcomHelp") } : {}) },
    row("consist", a.consistAddress == null ? undefined : a.consistAddress === 0
      ? t("vehicles.decoder.inactive") : a.lenz || a.kuehn ? a.consistAddress : `${a.consistAddress} · ${t(a.consistReverse
        ? "vehicles.decoder.reverse" : "vehicles.decoder.forward")}`, a.consistCVs,
    a.zimo?.autoConsist ? `${t("vehicles.decoder.zimo.autoConsist")} · ${cvSource(a, ...a.consistCVs)}` : undefined),
    row("consistFunctions", a.functionsComplete ? a.enabledFunctions.join(" · ") || t("vehicles.decoder.noneEnabled")
      : a.zimo && a.enabledFunctions.length ? a.enabledFunctions.join(" · ") : unknown, [21, 22],
    a.zimo?.ambiguousFunctionBit ? `${t("vehicles.decoder.zimo.ambiguousFunction")} · ${cvSource(a, 21, 22)}`
      : a.roco ? t("vehicles.decoder.roco.consistPermissions") : a.uhlenbrock ? t("vehicles.decoder.uhlenbrock.consistPermissions") : a.kuehn ? t("vehicles.decoder.kuehn.consistPermissions") : a.lenz ? t("vehicles.decoder.lenz.consistPermissions") : !a.functionsComplete ? t("vehicles.decoder.incomplete") : undefined),
    ...pikoExplanationRows(a, topic, t), ...uhlenbrockExplanationRows(a, topic, t, language)
  ];
  if (topic === "motor") return [
    a.modelRules && a.rules?.motorCurveCV ? raw(a.rules.motorCurveCV)
      : row("curve", a.config == null ? undefined : a.lenz && !bit(4) ? t("vehicles.speedCurve.lenzMode")
        : t(`vehicles.decoder.bit29.4.${bit(4) ? "on" : "off"}`), [29]),
    ...[2, 6, 5, 3, 4].map(raw), ...uhlenbrockExplanationRows(a, topic, t, language), ...lenzExplanationRows(a, topic, t, language), ...kuehnExplanationRows(a, topic, t, language)
  ];
  if (topic === "braking") return [
    ...rocoExplanationRows(a, topic, t),
    ...(a.uhlenbrock ? uhlenbrockExplanationRows(a, topic, t, language) : []),
    row("abc", status(a.abc), a.rules?.abc ? [a.rules.abc.cv] : []),
    row("constantBrake", status(a.constantBrake), a.constantBrakeCVs),
    ...pikoExplanationRows(a, topic, t), ...lenzExplanationRows(a, topic, t, language),
    ...kuehnExplanationRows(a, topic, t, language),
    ...(a.zimo ? [
      row("brakeMode", a.zimo.brakeMode == null ? undefined : t(`vehicles.decoder.zimo.mode.${a.zimo.brakeMode}`), [140]),
      row("brakeForward", a.zimo.brakeDistance.forward, a.zimo.brakeDistance.forwardCVs,
        distanceNote(a.zimo.brakeDistance.fine, a.zimo.brakeDistance.forwardCVs)),
      row("brakeReverse", a.zimo.brakeDistance.reverse, a.zimo.brakeDistance.reverseCVs,
        distanceNote(a.zimo.brakeDistance.fine, a.zimo.brakeDistance.reverseCVs))
    ] : []),
    ...Array.from(new Set(a.catalog.flatMap(profile => profile.active && profile.id === a.libraryMatch.profileId
      ? profile.definitions.filter(d => d.topic === "braking" && !a.uhlenbrock && !a.roco && !a.piko).map(d => d.cv) : []))).map(raw)
  ];
  if (topic === "outputs" && a.piko) return pikoExplanationRows(a, topic, t);
  if (topic === "outputs" && a.roco) return rocoExplanationRows(a, topic, t);
  if (topic === "outputs" && a.uhlenbrock) return uhlenbrockExplanationRows(a, topic, t, language);
  if (topic === "outputs" && (a.lenz || a.kuehn)) {
    const expanded = new Set(a.kuehn ? [13, 14, 47, 48, 49, 50, 51, 52, 55, 57, 58, 59, 60, 61, 62, 63, 113]
      : [57, 58, 59, 60, 61, 62, 63, 64, 147, 148, 149]);
    return [
      ...decoderMapping(a, t).map(mapping => ({ key: `mapping-${mapping.row}`,
        name: `${mapping.functionKey} · ${t(`vehicles.decoder.${mapping.direction}`)}`,
        help: t(a.kuehn ? "vehicles.decoder.kuehn.mappingRowHelp" : "vehicles.decoder.lenz.mappingRowHelp"), reading: mapping.actions.join(" · "),
        note: mapping.warning ? t(a.kuehn ? "vehicles.decoder.review" : "vehicles.decoder.lenz.unknownOutput") : undefined,
        source: mapping.source, numbers: mapping.sourceCVs || [] })),
      ...uhlenbrockExplanationRows(a, topic, t, language), ...lenzExplanationRows(a, topic, t, language),
      ...kuehnExplanationRows(a, topic, t, language),
      ...(a.libraryProfile?.definitions.filter(d => d.topic === "outputs" && !expanded.has(d.cv) &&
        !(d.cv >= 33 && d.cv <= (a.kuehn ? 46 : 47)) && !(d.cv >= 129 && d.cv <= 144) && a.get(d.cv) != null).map(d => raw(d.cv)) || [])
    ];
  }
  const layout = a.libraryProfile?.mapping;
  if (topic === "outputs") return layout ? decoderMapping(a, t).map((mapping) => ({
    key: `mapping-${mapping.row}`, name: `${mapping.functionKey} · ${t(`vehicles.decoder.${mapping.direction}`)}`,
    help: mapping.warning ? t("vehicles.decoder.mappingLimits") : t("vehicles.decoder.mappingHelp"),
    reading: mapping.actions.join(" · "), source: mapping.source, numbers: [],
    mappingSource: { cvNumber: layout.startCV + ((mapping.row - 1) % layout.rowsPerPage) * layout.stride,
      indexCV31: layout.indexCV31, indexCV32: layout.conditionPage + Math.floor((mapping.row - 1) / layout.rowsPerPage) }
  })) : (a.modelRules ? a.libraryProfile?.definitions.filter(d => d.topic === "outputs" &&
    d.indexCV31 == null && a.get(d.cv) != null).map(d => raw(d.cv)) || [] : []);
  const communication = a.libraryProfile?.definitions.filter(d => d.topic === "communication" && d.indexCV31 == null).map(d => d.cv) || [];
  return [row("firmware", a.firmware, [7]), ...pikoExplanationRows(a, topic, t), ...rocoExplanationRows(a, topic, t), ...uhlenbrockExplanationRows(a, topic, t, language), ...lenzExplanationRows(a, topic, t, language), ...kuehnExplanationRows(a, topic, t, language),
    ...Array.from(new Set([8, 28, 29, ...communication])).map(raw)];
}
