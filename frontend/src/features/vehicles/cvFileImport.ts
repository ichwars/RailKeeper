import { api, type VehicleCVValueInput } from "../../shared/api";

import { readLanguage, translate } from "../../shared/i18n";

const importError = (key: string) => new Error(translate(readLanguage(), `vehicles.cv.${key}`));

export const maxCVImportValues = 10000;
export const maxCVTextImportBytes = 2 * 1024 * 1024;
// Full JSON exports also contain record identities, timestamps and history, removed before import.
export const maxCVJSONImportBytes = 32 * 1024 * 1024;

export async function readCVImportFile(file: File): Promise<VehicleCVValueInput[]> {
  const json = /\.json$/i.test(file.name);
  if (file.size > (json ? maxCVJSONImportBytes : maxCVTextImportBytes)) {
    throw importError(json ? "jsonFileTooLarge" : "fileTooLarge");
  }
  const text = (await file.text()).replace(/^\uFEFF/, "");
  if (/^(?:Decoder:|Index:|CV\s+\d+\s*=)/m.test(text)) {
    if (file.size > maxCVTextImportBytes) throw importError("fileTooLarge");
    const preview = await api.previewVehicleCVFile(file);
    if (preview.cvImportError) throw importError("invalidFile");
    return (preview.suggestedCvValues || []).map((value) => ({
      ...value, decoderProfile: preview.suggestedDecoderProfile || "", sourceFileId: ""
    }));
  }
  return cvValuesFromImport(text);
}

function record(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) throw importError("invalidRecord");
  return value as Record<string, unknown>;
}

function input(value: unknown): VehicleCVValueInput {
  const row = record(value);
  const number = (value: unknown) => typeof value === "number" || (typeof value === "string" && value.trim()) ? Number(value) : NaN;
  const index = (value: unknown) => value === undefined || value === null || value === "" ? undefined : number(value);
  return {
    cvNumber: number(row.cvNumber), value: number(row.value),
    indexCV31: index(row.indexCV31), indexCV32: index(row.indexCV32),
    sourceLine: index(row.sourceLine), sourceExportedAt: String(row.sourceExportedAt || ""),
    description: String(row.description || ""), category: String(row.category || ""),
    protocol: String(row.protocol || ""), decoderProfile: String(row.decoderProfile || ""),
    // Exported attachment IDs belong to their original vehicle and must be explicitly reassigned.
    sourceFileId: ""
  };
}

function csvCells(line: string, delimiter: string): string[] {
  const cells: string[] = [];
  let current = "";
  let quoted = false;
  for (let i = 0; i < line.length; i++) {
    const char = line[i];
    if (char === '"') {
      if (quoted && line[i + 1] === '"') { current += '"'; i++; }
      else quoted = !quoted;
    } else if (char === delimiter && !quoted) { cells.push(current.trim()); current = ""; }
    else current += char;
  }
  if (quoted) throw importError("invalidCsv");
  return [...cells, current.trim()];
}

export function cvValuesFromImport(text: string): VehicleCVValueInput[] {
  const trimmed = text.trim();
  if (!trimmed) return [];
  let values: VehicleCVValueInput[];
  if (trimmed.startsWith("{") || trimmed.startsWith("[")) {
    const parsed: unknown = JSON.parse(trimmed);
    const rows: unknown = Array.isArray(parsed) ? parsed : record(parsed).cvValues;
    if (!Array.isArray(rows)) throw importError("invalidJson");
    if (rows.length > maxCVImportValues) throw importError("tooMany");
    values = rows.map(input);
  } else {
    const numberedLines = text.split(/\r?\n/).map((text, index) => ({ text, line: index + 1 })).filter(row => row.text.trim());
    const lines = numberedLines.map(row => row.text);
    const sourceLines = numberedLines.map(row => row.line);
    const delimiter = lines[0].includes(";") ? ";" : ",";
    const cells = lines.map((line) => csvCells(line, delimiter));
    const normalize = (key: string) => key.toLowerCase().replace(/[^a-z0-9]/g, "");
    const aliases: Record<string, string> = {
      cv: "cvNumber", cvnr: "cvNumber", cvnumber: "cvNumber", wert: "value", value: "value",
      beschreibung: "description", description: "description", kategorie: "category", category: "category",
      decoderprofile: "decoderProfile", decoderprofil: "decoderProfile", protocol: "protocol", protokoll: "protocol",
      sourceline: "sourceLine", originalzeile: "sourceLine", sourceexportedat: "sourceExportedAt", exportdatum: "sourceExportedAt",
      cv31: "indexCV31", cv32: "indexCV32", indexcv31: "indexCV31", indexcv32: "indexCV32"
    };
    const hasHeader = cells[0].some((cell) => aliases[normalize(cell)] === "cvNumber");
    if (hasHeader) sourceLines.shift();
    const headers = hasHeader ? cells.shift()!.map((cell) => aliases[normalize(cell)] || "")
      : ["cvNumber", "value", "description", "category", "decoderProfile", "indexCV31", "indexCV32"];
    if (cells.length > maxCVImportValues) throw importError("tooMany");
    values = cells.map((cells, index) => {
      const row = Object.fromEntries(headers.filter(Boolean).map(key => [key, cells[headers.indexOf(key)]]));
      if (!row.sourceLine) row.sourceLine = String(sourceLines[index]);
      return input(row);
    });
  }
  if (values.length > maxCVImportValues) throw importError("tooMany");
  return values;
}
