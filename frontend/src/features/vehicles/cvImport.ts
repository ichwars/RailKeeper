import { validCVExportTimestamp } from "./cvProvenance";
import { VehicleCVFilePreview, VehicleCVValue, VehicleCVValueInput, VehicleFunctionInput, type DecoderCVWorkspace } from "../../shared/api";
import type { CVWorkingTarget } from "./cvWorkingGroup";

export type FunctionMappingImport = VehicleFunctionInput & { functionKey: string };
export type CVImportStatus = "new" | "changed" | "same" | "invalid";
export type CVImportChangeField = "value" | "description" | "category" | "protocol" | "sourceFileId" | "sourceLine" | "sourceExportedAt";
export type CVImportRow = {
  id: string;
  input: VehicleCVValueInput;
  existing?: VehicleCVValue;
  status: CVImportStatus;
  selected: boolean;
  message: string;
  reason?: "invalid" | "duplicate";
  changedFields?: CVImportChangeField[];
  sourceProfile?: string;
};
export type CVImportPreview = {
  fileName: string;
  vehicleId?: string;
  rows: CVImportRow[];
  workingTarget?: CVWorkingTarget | null;
  workingGroupLabels?: string[];
  previousWorkingGroupLabels?: string[];
};
export type CVFileUploadPreview = {
  files: File[];
  previews: VehicleCVFilePreview[];
  workingTarget?: CVWorkingTarget | null;
  workingGroups?: DecoderCVWorkspace[];
  workingGroupLabels?: Record<string, string>;
};

export const functionKeys = Array.from({ length: 32 }, (_, index) => `F${index}`);
export const functionTypes = ["standard", "sound", "licht", "kupplung", "rauch", "sonderfunktion"];
export const functionModes = ["dauer", "moment"];
export const commonDecoderProfiles = ["ESU LokPilot 5", "ESU LokSound 5", "Zimo MS", "Zimo MX", "D&H SD", "D&H DH", "Märklin mLD3", "Märklin mSD3", "Lenz Standard+"];
export const cvCategories = ["Adresse", "Fahrverhalten", "Motor", "Licht", "Sound", "Funktion", "Decoder", "Sonstiges"];
export const cvProtocols = ["Motorola 14", "Motorola 27", "Motorola 28", "Motorola FX 14", "DCC 14", "DCC 28", "DCC 128", "LGB", "Selectrix"];

export { cvValuesFromImport } from "./cvFileImport";

function nullableFunctionFlag(value: unknown): boolean | null {
  if (value == null) return null;
  if (typeof value !== "boolean") throw new Error("Funktionsangaben müssen true, false oder null sein.");
  return value;
}

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
    flagsVersion: row.flagsVersion ?? parsed.flagsVersion ?? 1,
    directionDependent: (row.flagsVersion ?? parsed.flagsVersion) === 2 ? nullableFunctionFlag(row.directionDependent) : null,
    inverted: (row.flagsVersion ?? parsed.flagsVersion) === 2 ? nullableFunctionFlag(row.inverted) : null,
    legacyDirectionDependent: (row.flagsVersion ?? parsed.flagsVersion) === 2
      ? nullableFunctionFlag(row.legacyDirectionDependent) : nullableFunctionFlag(row.legacyDirectionDependent ?? row.directionDependent),
    notes: String(row.notes || "")
  }));
}

export function isValidFunctionMapping(value: FunctionMappingImport) {
  return Number.isInteger(value.flagsVersion ?? 1) && (value.flagsVersion ?? 1) >= 0 && (value.flagsVersion ?? 1) <= 2 &&
    functionKeys.includes(value.functionKey) &&
    functionTypes.includes(value.functionType || "standard") &&
    functionModes.includes(value.mode || "dauer");
}

export function cvValueKey(value: Pick<VehicleCVValueInput, "cvNumber" | "decoderProfile" | "indexCV31" | "indexCV32">) {
  return `${Number(value.cvNumber)}::${value.decoderProfile || ""}::${value.indexCV31 ?? "base"}::${value.indexCV32 ?? "base"}`;
}

function normalizeCVText(value?: string) {
  return (value || "").trim();
}

function cvImportChanges(existing: VehicleCVValue, input: VehicleCVValueInput) {
  const changes: CVImportChangeField[] = [];
  if (Number(existing.value) !== Number(input.value)) changes.push("value");
  if (normalizeCVText(existing.description) !== normalizeCVText(input.description)) changes.push("description");
  if (normalizeCVText(existing.category) !== normalizeCVText(input.category)) changes.push("category");
  if (normalizeCVText(existing.protocol) !== normalizeCVText(input.protocol)) changes.push("protocol");
  if (normalizeCVText(existing.sourceFileId) !== normalizeCVText(input.sourceFileId)) changes.push("sourceFileId");
  if (input.sourceLine != null && existing.sourceLine !== input.sourceLine) changes.push("sourceLine");
  if (input.sourceExportedAt && existing.sourceExportedAt !== input.sourceExportedAt) changes.push("sourceExportedAt");
  return changes;
}

const legacyChangeLabels: Record<CVImportChangeField, string> = {
  value: "Wert", description: "Beschreibung", category: "Kategorie", protocol: "Protokoll",
  sourceFileId: "Quelldatei", sourceLine: "Originalzeile", sourceExportedAt: "Exportdatum"
};

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
        reason: "invalid" as const,
        selected: false,
        message: "ungültig"
      };
    }
    if (seen.has(key)) {
      return {
        id: `${index}-${key}`,
        input,
        status: "invalid" as CVImportStatus,
        reason: "duplicate" as const,
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
      changedFields: changes,
      selected: true,
      message: `ändert ${changes.map(field => legacyChangeLabels[field]).join(", ")}`
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
