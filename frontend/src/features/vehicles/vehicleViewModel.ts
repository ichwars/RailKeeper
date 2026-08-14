import type {
  ArticleSearchInput,
  ArticleSearchResponse,
  CreateVehicleRequest,
  ExhibitionEntry,
  ExhibitionList,
  MasterDataEntry,
  MasterDataRelation,
  Vehicle,
  VehicleCVValueInput,
  VehicleExternalMappingInput,
  VehicleFunctionInput,
  VehicleMaintenance,
  VehicleMaintenanceInput,
  VehicleSparePartInput
} from "../../shared/api";
import { isBadCommonArticleValue } from "../../shared/articleSearch/articleSearchModel";
export {
  articleSearchEnabled,
  articleSearchSources
} from "../../shared/articleSearch/articleSearchPreferences";
import type { ArticleFieldKey } from "./articleSearch";

export const emptyVehicle: CreateVehicleRequest = {
  manufacturer: "",
  articleNumber: "",
  articleSourceUrl: "",
  name: "",
  gauge: "",
  epoch: "",
  railwayCompany: "",
  category: "",
  gattung: "",
  description: "",
  series: "",
  vehicleNumber: "",
  digital: false,
  digitalDecoderNumber: "",
  dtDecoder: false,
  dtDecoderNumber: "",
  decoderType: "",
  exhibitionReady: false,
  exhibition: false,
  abcBrakes: false,
  ean: "",
  productionPeriod: "",
  listPrice: "",
  acquisitionType: "",
  acquiredFrom: "",
  purchasePrice: "",
  purchaseDate: "",
  storageLocation: "",
  storageDetails: "",
  condition: "",
  conditionDetails: "",
  packaging: "",
  lengthMm: "",
  weightG: "",
  color: "",
  lettering: "",
  load: "",
  interior: "",
  axles: "",
  axleCount: "",
  tractionTireCount: "",
  wheelset: "",
  couplingSame: false,
  couplingFront: "",
  couplingRear: "",
  powerPickup: "",
  adapter: "",
  driveEnabled: false,
  driveDescription: "",
  headlightsEnabled: false,
  headlightsDescription: "",
  lightingEnabled: false,
  lightingDescription: "",
  soundGeneratorEnabled: false,
  soundGeneratorDescription: "",
  smokeGeneratorEnabled: false,
  smokeGeneratorDescription: "",
  additionalInfo: "",
  qrCodeEnabled: false
};

export type ModalMode = "create" | "view" | "edit";
export type ModalTab = "model" | "control" | "speedCurve" | "cv" | "uploads" | "maintenance" | "spareParts";
export type SortKey = "inventoryNumber" | "manufacturer" | "articleNumber" | "name" | "gauge" | "epoch" | "category";
export type SortDirection = "asc" | "desc";
export type InventoryViewMode = "table" | "cards";
export type InventoryFilter = "all" | "digital" | "analog" | "withImages" | "withoutImages";
export type MaintenanceFilter = "all" | "due" | "none";
export type InventoryQualityFilter = "none" | "missingArticleNumber" | "missingEan" | "digitalMissingDecoder";
export type InventoryReportMode = "summary" | "details";
export type InventoryReportSelection = "all" | "selected";
export type ECoSRequiredField = "manufacturer" | "name" | "gauge" | "category" | "gattung";
export type ECoSVehicleDraftPayload = {
  source: "ecos";
  mode: "create" | "update";
  targetVehicleId?: string;
  sourceSummary: {
    objectId: number;
    name: string;
    address: string;
    protocol: string;
    profile: string;
  };
  vehicle: CreateVehicleRequest;
  importedKeys: (keyof CreateVehicleRequest)[];
  externalMapping: VehicleExternalMappingInput;
  cvValues: VehicleCVValueInput[];
  functionValues: (VehicleFunctionInput & { functionKey: string })[];
  unclearFields: ECoSRequiredField[];
  returnToEcos?: {
    sessionId: string;
    objectId: number;
  };
};
export const ecosVehicleDraftStorageKey = "railkeeper.ecosVehicleDraft";
export const ecosImportSessionStorageKey = "railkeeper.ecosImportSession";
export const ecosRequiredFields: ECoSRequiredField[] = ["manufacturer", "name", "gauge", "category", "gattung"];
export type MaintenanceReminder = {
  vehicle: Vehicle;
  entry: VehicleMaintenance;
  daysUntilDue: number;
};
export type ExhibitionAssignment = {
  vehicle: Vehicle;
  lists: ExhibitionList[];
  selectedListID: string;
  entries: ExhibitionEntry[];
  loadingEntries: boolean;
  saving: boolean;
  error: string;
};

export const emptyMaintenanceForm: VehicleMaintenanceInput = {
  kind: "Wartung",
  status: "geplant",
  conditionRating: "",
  dueDate: "",
  completedAt: "",
  cost: "",
  notes: ""
};

export const emptySparePartForm: VehicleSparePartInput = {
  articleNumber: "",
  description: "",
  price: "",
  url: ""
};

export const emptyCVForm: VehicleCVValueInput = {
  cvNumber: 1,
  value: 0,
  description: "",
  category: "",
  protocol: "",
  decoderProfile: "",
  sourceFileId: ""
};

export type MasterDataOptions = {
  manufacturers: MasterDataEntry[];
  gauges: MasterDataEntry[];
  epochs: MasterDataEntry[];
  railwayCompanies: MasterDataEntry[];
  categories: MasterDataEntry[];
  gattungen: MasterDataEntry[];
  symbols: MasterDataEntry[];
  categoryRelations: MasterDataRelation[];
};

export const emptyOptions: MasterDataOptions = {
  manufacturers: [],
  gauges: [],
  epochs: [],
  railwayCompanies: [],
  categories: [],
  gattungen: [],
  symbols: [],
  categoryRelations: []
};

export const sortLabels: Record<SortKey, string> = {
  inventoryNumber: "Inventar",
  manufacturer: "Hersteller",
  articleNumber: "Artikel",
  name: "Bezeichnung",
  gauge: "Spur",
  epoch: "Epoche",
  category: "Kategorie"
};

export const inventoryViewSettingKey = "railkeeper.inventoryViewMode";

export function inferFunctionTypeFromSymbol(symbolKey: string, symbols: MasterDataEntry[], fallback = "standard") {
  const symbol = symbols.find((item) => item.active && item.key === symbolKey);
  const signal = `${symbolKey} ${symbol?.label || ""}`.toLocaleLowerCase("de-DE");
  if (!symbolKey) return "standard";
  if (signal.includes("sound") || signal.includes("horn") || signal.includes("pfiff")) return "sound";
  if (signal.includes("licht") || signal.includes("light") || signal.includes("lampe")) return "licht";
  if (signal.includes("kuppl")) return "kupplung";
  if (signal.includes("rauch") || signal.includes("smoke")) return "rauch";
  if (signal.includes("warn") || signal.includes("sonder") || signal.includes("sifa")) return "sonderfunktion";
  return fallback || "standard";
}

export const searchableFieldKeys: ArticleFieldKey[] = [
  "manufacturer",
  "articleNumber",
  "name",
  "gauge",
  "epoch",
  "railwayCompany",
  "category",
  "gattung",
  "description",
  "series",
  "vehicleNumber",
  "digitalDecoderNumber",
  "dtDecoderNumber",
  "ean",
  "productionPeriod",
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
  "adapter",
  "driveDescription",
  "headlightsDescription",
  "lightingDescription",
  "soundGeneratorDescription",
  "smokeGeneratorDescription",
  "additionalInfo"
];

export function optionValue(entry: MasterDataEntry) {
  return entry.label;
}

export function gattungenForCategory(options: MasterDataOptions, category: string | undefined) {
  const categoryKey = options.categories.find((entry) => optionValue(entry) === category)?.key;
  if (!categoryKey) return options.gattungen;

  const allowed = new Set(
    options.categoryRelations
      .filter((relation) => relation.parentKey === categoryKey)
      .map((relation) => relation.childKey)
  );
  return options.gattungen.filter((entry) => allowed.has(entry.key));
}

export function missingVehicleModelFieldLabels(
  form: CreateVehicleRequest,
  t: (key: string) => string
) {
  const fields: Array<{ label: string; value: string | undefined }> = [
    { label: t("vehicle.field.manufacturer"), value: form.manufacturer },
    { label: t("vehicle.field.name"), value: form.name },
    { label: t("vehicle.field.gauge"), value: form.gauge },
    { label: t("vehicle.field.category"), value: form.category },
    { label: t("vehicle.field.gattung"), value: form.gattung }
  ];
  return fields.filter((field) => !compactValue(field.value)).map((field) => field.label);
}

export function valueForSort(vehicle: Vehicle, key: SortKey) {
  return (vehicle[key] || "").toLocaleLowerCase("de-DE");
}

export function compactValue(value: unknown) {
  return String(value ?? "").trim();
}

export function hasArticleSearchCriteria(searchForm: CreateVehicleRequest, searchInput?: ArticleSearchInput) {
  if (searchInput) {
    const fields = searchInput.fields || {};
    const ean = compactValue(fields.ean);
    if (ean) {
      return true;
    }
    const articleNumber = compactValue(searchInput.articleNumber || fields.articleNumber);
    const name = compactValue(searchInput.name || fields.name);
    const manufacturer = compactValue(searchInput.manufacturer || fields.manufacturer);
    const gauge = compactValue(searchInput.gauge || fields.gauge);
    return Boolean((articleNumber || name) && manufacturer && gauge);
  }

  return hasArticleIdentity(searchForm) && Boolean(compactValue(searchForm.manufacturer)) && Boolean(compactValue(searchForm.gauge));
}

export function hasArticleIdentity(form: CreateVehicleRequest) {
  return Boolean(compactValue(form.articleNumber) || compactValue(form.name));
}

export function hasQrPayloadData(vehicle: Vehicle | null, form: CreateVehicleRequest) {
  return Boolean(
    compactValue(form.inventoryNumber || vehicle?.inventoryNumber) ||
    compactValue(form.name || vehicle?.name)
  );
}

export function inventoryViewMode(): InventoryViewMode {
  return window.localStorage.getItem(inventoryViewSettingKey) === "cards" ? "cards" : "table";
}

export function vehicleFieldsForSearch(form: CreateVehicleRequest) {
  return Object.fromEntries(
    searchableFieldKeys
      .map((key) => [key, String(form[key] || "").trim()])
      .filter(([, value]) => value)
  ) as Record<string, string>;
}

export function fieldValue(form: CreateVehicleRequest, key: string) {
  return String(form[key as ArticleFieldKey] || "").trim();
}

export function isBadArticleValue(key: string, value: string) {
  const normalized = value.trim();
  const lower = normalized.toLocaleLowerCase("de-DE");
  if (isBadCommonArticleValue(key, value)) return true;
  if (key === "lengthMm") {
    const number = Number(normalized.replace(",", "."));
    return !Number.isFinite(number) || number < 20 || number > 600;
  }
  if (key === "lightingDescription") {
    return lower.includes("fahrtrichtung") || lower.includes("lichtwechsel") || lower.includes("spitzenlicht") || lower.includes("schlusslicht");
  }
  if (key === "headlightsDescription") {
    return lower.includes("altersempfehlung") || lower.includes("downloads") || lower.includes("bedienungsanleitung");
  }
  if (key === "soundGeneratorDescription") {
    return lower.includes("menu") || lower.includes("menü") || lower.includes("menue") || lower.includes("sprunggröße") || lower.includes("sprunggroesse") || lower.includes("wählen sie") || lower.includes("waehlen sie");
  }
  return false;
}

export function sanitizeArticleSearchResponse(response: ArticleSearchResponse): ArticleSearchResponse {
  return {
    ...response,
    results: response.results.map((result) => {
      const fields = Object.fromEntries(
        Object.entries(result.fields).filter(([key, field]) => !isBadArticleValue(key, field.value))
      );
      return { ...result, fields };
    })
  };
}

export function emptyFunctionEdit(functionKey: string): VehicleFunctionInput & { persisted?: boolean } {
  return {
    name: functionKey === "F0" ? "Fahrlicht" : "",
    symbolKey: functionKey === "F0" ? "light" : "",
    functionType: functionKey === "F0" ? "licht" : "standard",
    mode: "dauer",
    directionDependent: false,
    notes: "",
    persisted: false
  };
}
