import { validCVExportTimestamp } from "./cvProvenance";
import { VehicleCVFilePreview, VehicleCVValue, VehicleCVValueInput, VehicleFunctionInput } from "../../shared/api";

export type FunctionMappingImport = VehicleFunctionInput & { functionKey: string };
export type CVImportStatus = "new" | "changed" | "same" | "invalid";
export type CVImportRow = {
  id: string;
  input: VehicleCVValueInput;
  existing?: VehicleCVValue;
  status: CVImportStatus;
  selected: boolean;
  message: string;
};
export type CVImportPreview = {
  fileName: string;
  vehicleId?: string;
  rows: CVImportRow[];
};
export type CVFileUploadPreview = {
  files: File[];
  previews: VehicleCVFilePreview[];
};

export const functionKeys = Array.from({ length: 32 }, (_, index) => `F${index}`);
export const functionTypes = ["standard", "sound", "licht", "kupplung", "rauch", "sonderfunktion"];
export const functionModes = ["dauer", "moment"];
export const commonDecoderProfiles = ["ESU LokPilot 5", "ESU LokSound 5", "Zimo MS", "Zimo MX", "D&H SD", "D&H DH", "Märklin mLD3", "Märklin mSD3", "Lenz Standard+"];
export const cvCategories = ["Adresse", "Fahrverhalten", "Motor", "Licht", "Sound", "Funktion", "Decoder", "Sonstiges"];
export const cvProtocols = ["Motorola 14", "Motorola 27", "Motorola 28", "Motorola FX 14", "DCC 14", "DCC 28", "DCC 128", "LGB", "Selectrix"];

export { cvValuesFromImport } from "./cvFileImport";

export function functionMappingsFromImport(text: string): FunctionMappingImport[] {
  const trimmed = text.trim();
  if (!trimmed) return [];
  const parsed = JSON.parse(trimmed);
  const rows = Array.isArray(parsed) ? parsed : parsed.functions || parsed.functionMappings || [];
  return rows.map((row: Partial<FunctionMappingImport>) => ({
    functionKey: String(row.functionKey || "").toUpperCase(),
    name: String(row.name || ""),
    symbolKey: String(row.symbolKey || ""),
    functionType: String(row.functionType || "standard"),
    mode: String(row.mode || "dauer"),
    directionDependent: Boolean(row.directionDependent),
    notes: String(row.notes || "")
  }));
}

export function isValidFunctionMapping(value: FunctionMappingImport) {
  return functionKeys.includes(value.functionKey) &&
    functionTypes.includes(value.functionType || "standard") &&
    functionModes.includes(value.mode || "dauer");
}

export function cvValueKey(value: Pick<VehicleCVValueInput, "cvNumber" | "decoderProfile" | "indexCV31" | "indexCV32">) {
  return `${Number(value.cvNumber)}::${(value.decoderProfile || "").trim()}::${value.indexCV31 ?? "base"}::${value.indexCV32 ?? "base"}`;
}

function normalizeCVText(value?: string) {
  return (value || "").trim();
}

function cvImportChanges(existing: VehicleCVValue, input: VehicleCVValueInput) {
  const changes = [];
  if (Number(existing.value) !== Number(input.value)) changes.push("Wert");
  if (normalizeCVText(existing.description) !== normalizeCVText(input.description)) changes.push("Beschreibung");
  if (normalizeCVText(existing.category) !== normalizeCVText(input.category)) changes.push("Kategorie");
  if (normalizeCVText(existing.protocol) !== normalizeCVText(input.protocol)) changes.push("Protokoll");
  if (normalizeCVText(existing.sourceFileId) !== normalizeCVText(input.sourceFileId)) changes.push("Quelldatei");
  if (input.sourceLine != null && existing.sourceLine !== input.sourceLine) changes.push("Originalzeile");
  if (input.sourceExportedAt && existing.sourceExportedAt !== input.sourceExportedAt) changes.push("Exportdatum");
  return changes;
}

export function buildCVImportPreview(fileName: string, values: VehicleCVValueInput[], existingValues: VehicleCVValue[]): CVImportPreview {
  const existing = new Map(existingValues.map((entry) => [cvValueKey(entry), entry]));
  const seen = new Set<string>();
  const rows = values.map((input, index) => {
    const key = cvValueKey(input);
    if (!isValidCVValueInput(input)) {
      return {
        id: `${index}-${key}`,
        input,
        status: "invalid" as CVImportStatus,
        selected: false,
        message: "ungültig"
      };
    }
    if (seen.has(key)) {
      return {
        id: `${index}-${key}`,
        input,
        status: "invalid" as CVImportStatus,
        selected: false,
        message: "doppelt im Import"
      };
    }
    seen.add(key);
    const match = existing.get(key);
    if (!match) {
      return {
        id: `${index}-${key}`,
        input,
        status: "new" as CVImportStatus,
        selected: true,
        message: "neu"
      };
    }
    const changes = cvImportChanges(match, input);
    if (changes.length === 0) {
      return {
        id: `${index}-${key}`,
        input,
        existing: match,
        status: "same" as CVImportStatus,
        selected: false,
        message: "bereits gleich"
      };
    }
    return {
      id: `${index}-${key}`,
      input,
      existing: match,
      status: "changed" as CVImportStatus,
      selected: true,
      message: `ändert ${changes.join(", ")}`
    };
  });
  return { fileName, rows };
}

export function isValidCVValueInput(value: VehicleCVValueInput) {
  const hasIndex = value.indexCV31 != null || value.indexCV32 != null;
  const validIndex = !hasIndex || (value.indexCV31 != null && value.indexCV32 != null &&
    [value.indexCV31, value.indexCV32].every((index) => Number.isInteger(index) && index >= 0 && index <= 255) &&
    value.cvNumber >= 257 && value.cvNumber <= 512);
  const validSource = (value.sourceLine == null || (Number.isInteger(value.sourceLine) && value.sourceLine >= 1 && value.sourceLine <= 10000000))
    && validCVExportTimestamp(value.sourceExportedAt);
  return validSource && validIndex && Number.isInteger(Number(value.cvNumber)) &&
    Number(value.cvNumber) >= 1 &&
    Number(value.cvNumber) <= 1024 &&
    Number.isInteger(Number(value.value)) &&
    Number(value.value) >= 0 &&
    Number(value.value) <= 255;
}
