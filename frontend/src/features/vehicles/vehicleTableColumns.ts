import type { Vehicle } from "../../shared/api";

export const vehicleTableColumnKeys = [
  "image",
  "inventoryNumber",
  "manufacturer",
  "articleNumber",
  "name",
  "gauge",
  "epoch",
  "exhibition",
  "railwayCompany",
  "category",
  "gattung",
  "series",
  "vehicleNumber",
  "ean",
  "productionPeriod",
  "digital",
  "digitalDecoderNumber",
  "dtDecoder",
  "dtDecoderNumber",
  "decoderType",
  "adapter",
  "abcBrakes",
  "listPrice",
  "acquisitionType",
  "acquiredFrom",
  "purchasePrice",
  "purchaseDate",
  "storageLocation",
  "condition",
  "packaging",
  "lengthMm",
  "weightG",
  "color",
  "lettering",
  "load",
  "interior",
  "axles",
  "axleCount",
  "tractionTireCount",
  "wheelset",
  "couplingFront",
  "couplingRear",
  "powerPickup",
  "driveEnabled",
  "headlightsEnabled",
  "lightingEnabled",
  "soundGeneratorEnabled",
  "smokeGeneratorEnabled",
  "exhibitionReady"
] as const;

export type VehicleTableColumn = typeof vehicleTableColumnKeys[number];
export type VehicleSortableColumn = Exclude<VehicleTableColumn, "image" | "exhibition">;
export type VehicleColumnMove = "up" | "down";
export type VehicleColumnGroup = "identity" | "digital" | "ownership" | "technical" | "equipment";
export type VehicleColumnKind = "text" | "boolean" | "date" | "image" | "exhibition";

export type VehicleColumnDefinition = {
  key: VehicleTableColumn;
  group: VehicleColumnGroup;
  kind: VehicleColumnKind;
};

const booleanColumns = new Set<VehicleTableColumn>([
  "digital",
  "dtDecoder",
  "abcBrakes",
  "driveEnabled",
  "headlightsEnabled",
  "lightingEnabled",
  "soundGeneratorEnabled",
  "smokeGeneratorEnabled",
  "exhibitionReady"
]);

const groupByColumn: Record<VehicleTableColumn, VehicleColumnGroup> = {
  image: "identity",
  inventoryNumber: "identity",
  manufacturer: "identity",
  articleNumber: "identity",
  name: "identity",
  gauge: "identity",
  epoch: "identity",
  exhibition: "equipment",
  railwayCompany: "identity",
  category: "identity",
  gattung: "identity",
  series: "identity",
  vehicleNumber: "identity",
  ean: "identity",
  productionPeriod: "identity",
  digital: "digital",
  digitalDecoderNumber: "digital",
  dtDecoder: "digital",
  dtDecoderNumber: "digital",
  decoderType: "digital",
  adapter: "digital",
  abcBrakes: "digital",
  listPrice: "ownership",
  acquisitionType: "ownership",
  acquiredFrom: "ownership",
  purchasePrice: "ownership",
  purchaseDate: "ownership",
  storageLocation: "ownership",
  condition: "ownership",
  packaging: "ownership",
  lengthMm: "technical",
  weightG: "technical",
  color: "technical",
  lettering: "technical",
  load: "technical",
  interior: "technical",
  axles: "technical",
  axleCount: "technical",
  tractionTireCount: "technical",
  wheelset: "technical",
  couplingFront: "technical",
  couplingRear: "technical",
  powerPickup: "technical",
  driveEnabled: "equipment",
  headlightsEnabled: "equipment",
  lightingEnabled: "equipment",
  soundGeneratorEnabled: "equipment",
  smokeGeneratorEnabled: "equipment",
  exhibitionReady: "equipment"
};

function kindForColumn(key: VehicleTableColumn): VehicleColumnKind {
  if (key === "image") return "image";
  if (key === "exhibition") return "exhibition";
  if (key === "purchaseDate") return "date";
  return booleanColumns.has(key) ? "boolean" : "text";
}

export const vehicleTableColumns: VehicleColumnDefinition[] = vehicleTableColumnKeys.map((key) => ({
  key,
  group: groupByColumn[key],
  kind: kindForColumn(key)
}));

export const defaultVehicleTableColumns: VehicleTableColumn[] = [
  "image",
  "inventoryNumber",
  "manufacturer",
  "articleNumber",
  "name",
  "gauge",
  "epoch",
  "exhibition"
];

export function isVehicleTableColumn(value: unknown): value is VehicleTableColumn {
  return typeof value === "string" && vehicleTableColumnKeys.some((key) => key === value);
}

export function isVehicleDataColumn(column: VehicleTableColumn) {
  return column !== "image" && column !== "exhibition";
}

export function sortableVehicleColumn(column: VehicleTableColumn): column is VehicleSortableColumn {
  return isVehicleDataColumn(column);
}

export function normalizeVehicleTableColumns(values: Iterable<unknown>) {
  const seen = new Set<VehicleTableColumn>();
  const normalized = [...values].flatMap((value) => {
    if (!isVehicleTableColumn(value) || seen.has(value)) return [];
    seen.add(value);
    return [value];
  });

  if (!normalized.some(isVehicleDataColumn)) {
    normalized.push("inventoryNumber");
  }

  return normalized;
}

export function parseVehicleTableColumns(raw: string | undefined) {
  if (!raw) return [...defaultVehicleTableColumns];

  try {
    const parsed: unknown = JSON.parse(raw);
    return Array.isArray(parsed)
      ? normalizeVehicleTableColumns(parsed)
      : [...defaultVehicleTableColumns];
  } catch {
    return [...defaultVehicleTableColumns];
  }
}

export function serializeVehicleTableColumns(columns: Iterable<unknown>) {
  return JSON.stringify(normalizeVehicleTableColumns(columns));
}

export function toggleVehicleTableColumn(
  columns: readonly VehicleTableColumn[],
  column: VehicleTableColumn
) {
  const next = columns.filter((key) => key !== column);
  if (next.length === columns.length) next.push(column);
  return normalizeVehicleTableColumns(next);
}

export function moveVehicleTableColumn(
  columns: readonly VehicleTableColumn[],
  column: VehicleTableColumn,
  direction: VehicleColumnMove
) {
  const next = normalizeVehicleTableColumns(columns);
  const from = next.indexOf(column);
  if (from < 0) return next;

  const to = direction === "up" ? from - 1 : from + 1;
  if (to < 0 || to >= next.length) return next;

  [next[from], next[to]] = [next[to], next[from]];
  return next;
}

export function vehicleColumnSortValue(vehicle: Vehicle, key: VehicleSortableColumn) {
  const value = vehicle[key];
  if (typeof value === "boolean") return value ? "1" : "0";
  return String(value ?? "").trim().toLocaleLowerCase("de-DE");
}
