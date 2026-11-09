import type { Language } from "../../shared/i18n";
import type {
  DataTransferCSVColumnMapping,
  DataTransferJob,
  DataTransferJobDetails,
  DataTransferPreview,
  DataTransferPreviewRecord,
  DataTransferProfile,
  DataTransferVehicleSet,
  DataTransferVehicleSetPreview
} from "./dataTransferModel";

export type ImportDialogStep = "profile" | "file" | "mapping" | "review" | "confirm";

export function currentStep(profileId: string, file: File | null, preview: DataTransferPreview | null, job: DataTransferJob | null,
  mappingAccepted: boolean, requiresReupload: boolean) {
  if (requiresReupload) return "file";
  if (preview) return mappingAccepted
    ? (preview.issues.some((issue) => !issue.selectedResolution) ? "review" : "confirm")
    : "mapping";
  if (file || job?.sourceName) return "mapping";
  return profileId || job ? "file" : "profile";
}

export function vehicleCSVMappingRequired(job: Pick<DataTransferJob, "format" | "areas"> | undefined) {
  return job?.format === "csv" && job.areas.includes("vehicles");
}

export function snapshotName(snapshot: DataTransferJob | DataTransferProfile) {
  return "profileName" in snapshot ? snapshot.profileName : snapshot.name;
}

export function stepClass(name: ImportDialogStep, current: ImportDialogStep) {
  const order: ImportDialogStep[] = ["profile", "file", "mapping", "review", "confirm"];
  const index = order.indexOf(name);
  const currentIndex = order.indexOf(current);
  return index < currentIndex ? "completed" : index === currentIndex ? "current" : "";
}

export function previewFromDetails(job?: DataTransferJob, details?: DataTransferJobDetails | null): DataTransferPreview | null {
  if (!job || !job.sourceName || !details || details.job.id !== job.id) return null;
  const records = previewRecords(job.preview.records);
  return {
    job: { ...job },
    records,
    issues: details.issues.map((issue) => ({ ...issue })),
    totalRecords: job.totalRecords,
    readyRecords: job.readyRecords,
    warningRecords: job.warningRecords,
    errorRecords: job.errorRecords,
    csvMapping: csvMappingFromPreview(job.preview.csvMapping),
    vehicleFields: vehicleFieldsFromPreview(job.preview.vehicleFields),
    vehicleSets: vehicleSetPreviewsFromPreview(job.preview.vehicleSets)
  };
}

export function previewRecords(value: unknown): DataTransferPreviewRecord[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isPreviewRecord).map((record) => ({ ...record, data: { ...record.data } }));
}

export function isPreviewRecord(value: unknown): value is DataTransferPreviewRecord {
  if (!value || typeof value !== "object") return false;
  const record = value as Record<string, unknown>;
  return typeof record.area === "string" && typeof record.recordKey === "string" &&
    typeof record.classification === "string" && typeof record.proposedAction === "string" &&
    Boolean(record.data) && typeof record.data === "object";
}

export function clonePreview(preview: DataTransferPreview): DataTransferPreview {
  return {
    ...preview,
    job: { ...preview.job, areas: [...preview.job.areas], options: { ...preview.job.options } },
    records: preview.records.map((record) => ({ ...record, data: { ...record.data } })),
    issues: preview.issues.map((issue) => ({ ...issue })),
    csvMapping: cloneCSVMapping(preview.csvMapping),
    vehicleFields: preview.vehicleFields?.map((field) => ({
      ...field,
      aliases: field.aliases ? [...field.aliases] : undefined
    })),
    vehicleSets: preview.vehicleSets?.map(cloneVehicleSetPreview)
  };
}

export function cloneVehicleSetPreview(set: DataTransferVehicleSetPreview): DataTransferVehicleSetPreview {
  return {
    ...set,
    memberRecordKeys: [...set.memberRecordKeys],
    rowNumbers: set.rowNumbers ? [...set.rowNumbers] : undefined,
    diagnostics: set.diagnostics?.map((diagnostic) => ({ ...diagnostic })),
    data: {
      ...set.data,
      members: set.data.members.map((member) => ({ ...member }))
    }
  };
}

export function cloneCSVMapping(mapping: DataTransferCSVColumnMapping[] | undefined) {
  return mapping?.map((column) => ({ ...column })) ?? [];
}

export function csvMappingIdentity(job: DataTransferJob, mapping: DataTransferCSVColumnMapping[]) {
  const columns = [...mapping]
    .sort((left, right) => left.index - right.index)
    .map(({ index, sourceHeader, normalizedHeader, targetField }) =>
      [index, sourceHeader, normalizedHeader, targetField]
    );
  return JSON.stringify([job.id, job.sourceSha256, columns]);
}

export function csvMappingFromPreview(value: unknown): DataTransferCSVColumnMapping[] {
  if (!Array.isArray(value)) return [];
  return value.filter((item): item is DataTransferCSVColumnMapping => {
    if (!item || typeof item !== "object") return false;
    const column = item as Record<string, unknown>;
    return typeof column.index === "number" && typeof column.sourceHeader === "string" &&
      typeof column.normalizedHeader === "string" && typeof column.targetField === "string" &&
      typeof column.origin === "string";
  }).map((column) => ({ ...column }));
}

export function vehicleFieldsFromPreview(value: unknown): DataTransferPreview["vehicleFields"] {
  if (!Array.isArray(value)) return [];
  return value.filter((item): item is NonNullable<DataTransferPreview["vehicleFields"]>[number] => {
    if (!item || typeof item !== "object") return false;
    const field = item as Record<string, unknown>;
    return typeof field.key === "string" && typeof field.labelDE === "string" &&
      typeof field.labelEN === "string" && ["string", "integer", "boolean"].includes(String(field.kind));
  }).map((field) => ({ ...field }));
}

export function vehicleSetPreviewsFromPreview(value: unknown): DataTransferVehicleSetPreview[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isVehicleSetPreview).map(cloneVehicleSetPreview);
}

export function isVehicleSetPreview(value: unknown): value is DataTransferVehicleSetPreview {
  if (!isRecord(value)) return false;
  return typeof value.recordKey === "string" &&
    typeof value.classification === "string" &&
    ["ready", "warning", "error"].includes(value.classification) &&
    typeof value.proposedAction === "string" &&
    ["create", "replace", "use_existing", "copy"].includes(value.proposedAction) &&
    hasOptionalStrings(value, vehicleSetPreviewOptionalStringFields) &&
    isStringArray(value.memberRecordKeys) &&
    (value.rowNumbers === undefined || isNumberArray(value.rowNumbers)) &&
    (value.diagnostics === undefined || isVehicleSetDiagnostics(value.diagnostics)) &&
    isVehicleSet(value.data);
}

export function isVehicleSet(value: unknown): value is DataTransferVehicleSet {
  if (!isRecord(value) || !vehicleSetStringFields.every((field) => typeof value[field] === "string")) return false;
  if (!hasOptionalStrings(value, vehicleSetOptionalStringFields)) return false;
  if (!Array.isArray(value.members)) return false;
  return value.members.every((member) => isRecord(member) &&
    typeof member.vehicleId === "string" &&
    typeof member.vehicleInventoryNumber === "string" &&
    typeof member.position === "number" &&
    (member.label === undefined || typeof member.label === "string")
  );
}

export function areaLabels(areas: DataTransferJob["areas"], language: Language) {
  const labels = language === "de"
    ? { vehicles: "Fahrzeuge", accessories: "Zubehör", exhibitionLists: "Ausstellungslisten" }
    : { vehicles: "Vehicles", accessories: "Accessories", exhibitionLists: "Exhibition lists" };
  return areas.map((area) => labels[area]).join(", ");
}

export function importCopy(language: Language) {
  return language === "de" ? {
    title: "Import prüfen", eyebrow: "SICHERER IMPORT", close: "Schließen", cancel: "Abbrechen",
    progress: "Importschritte",
    steps: { profile: "Profil", file: "Datei", mapping: "Zuordnung", review: "Prüfung", confirm: "Bestätigung" },
    profile: "Importprofil", chooseProfile: "Profil wählen", snapshot: "Auftragssnapshot", chooseFile: "Datei auswählen",
    fileLabel: "Importdatei", fileHelp: "Die Datei wird erst geprüft. Es werden noch keine Daten geschrieben.",
    browse: "Durchsuchen", serverPreview: "PERSISTENTE SERVERVORSCHAU", preview: "Vorschau", revision: "Revision",
    records: "Datensätze", ready: "bereit", warnings: "Hinweise", errors: "Fehler",
    unresolved: "{count} Konflikt(e) müssen vor dem Import aufgelöst werden.",
    readyToConfirm: "Alle Konflikte sind aufgelöst. Der geprüfte Stand kann importiert werden.",
    cancelJob: "Auftrag abbrechen", cancelJobTitle: "Importauftrag abbrechen?",
    cancelJobConfirm: "Diesen Importauftrag abbrechen? Die Vorschau bleibt im Verlauf erhalten.",
    changeTitle: "Auswahl ändern?", applyChange: "Änderung übernehmen",
    changeWarning: "Die aktuelle Vorschau wird verworfen und als neue Revision erneut geprüft. Fortfahren?",
    uploadError: "Die Datei konnte nicht geprüft werden.", resolveError: "Der Konflikt konnte nicht aufgelöst werden.",
    confirmError: "Der Import konnte nicht bestätigt werden.", cancelJobError: "Der Auftrag konnte nicht abgebrochen werden.",
    changeError: "Der Import konnte nicht auf ein anderes Profil umgestellt werden.",
    mappingTitle: "Erkannte CSV-Zuordnung", mappingExplanation: "Jede CSV-Spalte wird einem RailKeeper-Feld zugeordnet. Nicht benötigte Spalten können ausdrücklich ignoriert werden.",
    sourceColumn: "CSV-Spalte", targetField: "RailKeeper-Feld", mappingSource: "Erkennung",
    targetFor: "Zielfeld für {source}", unmapped: "Nicht zugeordnet", ignore: "Spalte ignorieren",
    mappingIncomplete: "Alle CSV-Spalten müssen zugeordnet oder ignoriert werden.",
    legacyMapping: "Diese ältere Vorschau enthält noch keine Spaltenzuordnung. Sie kann unverändert geprüft werden.",
    saveMapping: "Zuordnung im Profil speichern", validateMapping: "Zuordnung prüfen",
    continueReview: "Weiter zur Prüfung", mappingFileRequired: "Für Änderungen muss die CSV-Datei erneut ausgewählt werden.",
    mappingError: "Die geänderte CSV-Zuordnung konnte nicht geprüft werden.",
    conflictRecovery: "Der Auftrag wurde zwischenzeitlich geändert. Die persistente Vorschau wurde neu gelesen. Bitte Zuordnung und Konflikte erneut prüfen.",
    confirmConflictRecovery: "Der Auftrag wurde beim Bestätigen geändert. Bitte die Quelldatei erneut hochladen und die neue Vorschau vollständig prüfen."
  } : {
    title: "Review import", eyebrow: "SAFE IMPORT", close: "Close", cancel: "Cancel", progress: "Import steps",
    steps: { profile: "Profile", file: "File", mapping: "Mapping", review: "Review", confirm: "Confirmation" },
    profile: "Import profile", chooseProfile: "Choose profile", snapshot: "Job snapshot", chooseFile: "Choose file",
    fileLabel: "Import file", fileHelp: "The file is reviewed first. No data is written yet.", browse: "Browse",
    serverPreview: "PERSISTENT SERVER PREVIEW", preview: "Preview", revision: "Revision", records: "records",
    ready: "ready", warnings: "warnings", errors: "errors",
    unresolved: "{count} conflict(s) must be resolved before import.",
    readyToConfirm: "All conflicts are resolved. The reviewed revision can be imported.", cancelJob: "Cancel job",
    cancelJobTitle: "Cancel import job?", changeTitle: "Change selection?", applyChange: "Apply change",
    cancelJobConfirm: "Cancel this import job? The preview remains in history.",
    changeWarning: "The current preview will be discarded and reviewed as a new revision. Continue?",
    uploadError: "The file could not be reviewed.", resolveError: "The conflict could not be resolved.",
    confirmError: "The import could not be confirmed.", cancelJobError: "The job could not be cancelled.",
    changeError: "The import could not be changed to another profile.",
    mappingTitle: "Detected CSV mapping", mappingExplanation: "Each CSV column is assigned to a RailKeeper field. Unneeded columns can be explicitly ignored.",
    sourceColumn: "CSV column", targetField: "RailKeeper field", mappingSource: "Detection",
    targetFor: "Target field for {source}", unmapped: "Not mapped", ignore: "Ignore column",
    mappingIncomplete: "Every CSV column must be mapped or ignored.", saveMapping: "Save mapping to profile",
    legacyMapping: "This older preview does not contain column mapping data. It can be reviewed unchanged.",
    validateMapping: "Validate mapping", continueReview: "Continue to review",
    mappingFileRequired: "Select the CSV file again to change its mapping.",
    mappingError: "The changed CSV mapping could not be validated.",
    conflictRecovery: "The job changed in the meantime. The persistent preview was re-read. Review the mapping and conflicts again.",
    confirmConflictRecovery: "The job changed during confirmation. Upload the source file again and fully review the new preview."
  };
}

export function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function isVehicleSetDiagnostics(value: unknown): boolean {
  return Array.isArray(value) && value.every((diagnostic) => isRecord(diagnostic) &&
    typeof diagnostic.rowNumber === "number" &&
    typeof diagnostic.field === "string" &&
    typeof diagnostic.code === "string"
  );
}

export function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === "string");
}

export function isNumberArray(value: unknown): value is number[] {
  return Array.isArray(value) && value.every((item) => typeof item === "number");
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object";
}

export function hasOptionalStrings(value: Record<string, unknown>, fields: readonly string[]) {
  return fields.every((field) => value[field] === undefined || typeof value[field] === "string");
}

export function mappingOriginLabel(origin: DataTransferCSVColumnMapping["origin"], language: Language) {
  const labels = language === "de"
    ? { alias: "Automatisch", profile: "Profil", manual: "Manuell", ignored: "Ignoriert", unmapped: "Offen" }
    : { alias: "Automatic", profile: "Profile", manual: "Manual", ignored: "Ignored", unmapped: "Open" };
  return labels[origin];
}

export function importButtonLabel(count: number, language: Language) {
  if (language === "de") return `${count} ${count === 1 ? "Datensatz" : "Datensätze"} importieren`;
  return `Import ${count} ${count === 1 ? "record" : "records"}`;
}

const vehicleSetStringFields = [
  "inventoryNumber", "name", "manufacturer", "articleNumber", "articleSourceUrl", "gauge", "epoch",
  "railwayCompany", "category", "gattung", "description", "ean", "productionPeriod", "listPrice",
  "acquisitionType", "acquiredFrom", "purchasePrice", "purchaseDate", "storageLocation", "storageDetails",
  "condition", "conditionDetails", "packaging"
] as const satisfies readonly (keyof DataTransferVehicleSet)[];
const vehicleSetOptionalStringFields = ["id", "createdAt", "updatedAt"] as const;
const vehicleSetPreviewOptionalStringFields = ["targetId", "targetUpdatedAt", "targetFingerprint"] as const;
