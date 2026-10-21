import type { DecoderLibraryMatch, DecoderLibraryProfile, DecoderReading } from "./decoderLibraryModel";

export const railComPlusSource = "https://github.com/JMRI/JMRI/blob/a9e4d386488cb79c0811d479aad3d75b0b33b068/xml/decoders/esu/v4fnDescription.xsl";
export const railComPlusSymbolsSource = "https://esu.eu/download/betriebsanleitungen/digitalsysteme/?tx_esudownloads_pi1%5BdownloadItem%5D=014a5a9436295b8c1130b6eafa0cb993";

export type RailComPlusFunction = {
  functionKey: string; firstCV: number; firstValue: number; secondValue: number;
  iconCode: number; category: number; momentary: boolean; inverted: boolean;
  sourceLine?: number | null; secondSourceLine?: number | null; sourceExportedAt?: string;
};
type Reading = DecoderReading & { sourceLine?: number | null; sourceExportedAt?: string };
export type RailComPlusData = {
  name?: string; nameLastCV?: number; nameIssue?: "incomplete" | "encoding";
  functions: RailComPlusFunction[]; incompleteFunctions: number; unassignedFunctions: number;
  flagsVerification?: { firmware: string; programmerVersion: string };
};

// Reviewed ESU layout only. The name is mutable metadata, never decoder identity evidence.
export function railComPlus(values: Reading[], profile: DecoderLibraryProfile | undefined,
  match: DecoderLibraryMatch): RailComPlusData | undefined {
  if (!profile?.active || profile.id !== "esu-lokpilot-5-micro-dcc" || profile.manufacturerId !== 151 ||
    profile.family !== "LokPilot 5" || match.profileId !== profile.id ||
    match.manufacturerId !== 151 || !["matched", "family"].includes(match.status) ||
    match.firmwareCompatible === false) return undefined;
  if (new Set(values.map(row => (row.decoderProfile || "").trim())).size > 1) return undefined;
  if (values.some(row => row.protocol?.trim() && !/^DCC(?: ?(?:14|28|128))?$/i.test(row.protocol.trim()))) {
    return undefined;
  }
  const rows = values;
  const indexed = (cv: number, page: number) => rows.filter(row => row.cvNumber === cv &&
    row.indexCV31 === 1 && row.indexCV32 === page);
  const read = (cv: number, page: number) => {
    const found = indexed(cv, page);
    return found.length === 1 && Number.isInteger(found[0].value) && found[0].value >= 0 && found[0].value <= 255
      ? found[0] : undefined;
  };
  const result: RailComPlusData = { functions: [], incompleteFunctions: 0, unassignedFunctions: 0 };
  // All four checkbox combinations compared for F0 (Light), F1 (Physical) and F3 (Logical)
  // with LokProgrammer 5.2.18 on 2026-10-20.
  // The descriptor layout for other keys follows the pinned JMRI definition.
  const firmwareRows = values.filter(row => row.cvNumber === 7 && row.indexCV31 == null && row.indexCV32 == null);
  if (firmwareRows.length === 1 && match.firmware === "5.14.183") {
    result.flagsVerification = { firmware: match.firmware, programmerVersion: "5.2.18" };
  }
  if (rows.some(row => row.indexCV31 === 1 && row.indexCV32 === 0)) {
    const text: number[] = [];
    let exportedAt: string | undefined;
    result.nameIssue = "incomplete";
    for (let cv = 261; cv <= 288; cv++) {
      const row = read(cv, 0);
      if (!row) break;
      if (exportedAt && row.sourceExportedAt && exportedAt !== row.sourceExportedAt) break;
      exportedAt ||= row.sourceExportedAt;
      if (row.value === 0) {
        result.name = String.fromCharCode(...text).trim() || undefined;
        result.nameLastCV = cv;
        result.nameIssue = undefined;
        break;
      }
      if (row.value < 32 || row.value > 126) { result.nameIssue = "encoding"; break; }
      text.push(row.value);
    }
  }
  if (!rows.some(row => row.indexCV31 === 1 && row.indexCV32 === 1)) return result;
  for (let n = 0; n < 32; n++) {
    const firstCV = 257 + n * 2, a = read(firstCV, 1), b = read(firstCV + 1, 1);
    if (!a || !b || (a.sourceExportedAt && b.sourceExportedAt && a.sourceExportedAt !== b.sourceExportedAt)) {
      result.incompleteFunctions++;
      continue;
    }
    const iconCode = ((a.value & 15) << 8) | b.value;
    if (iconCode < 2) { result.unassignedFunctions++; continue; }
    result.functions.push({ functionKey: `F${n}`, firstCV, firstValue: a.value, secondValue: b.value,
      iconCode, category: (a.value >> 5) & 3, momentary: Boolean(a.value & 128), inverted: Boolean(a.value & 16),
      sourceLine: a.sourceLine, secondSourceLine: b.sourceLine,
      sourceExportedAt: a.sourceExportedAt || b.sourceExportedAt });
  }
  return result;
}
