import { CreateVehicleRequest, ECoSRawLocomotive, ECoSRawProbe, MasterDataEntry, Vehicle, VehicleCVValueInput, VehicleExternalMapping, VehicleExternalMappingInput, VehicleFunctionInput } from "../../shared/api";

export type ImportRow = {
  id: string;
  selected: boolean;
  mode: "create" | "update";
  status: "ok" | "warning" | "error" | "saved";
  issues: string[];
  importedKeys: (keyof CreateVehicleRequest)[];
  duplicateVehicleId?: string;
  externalMapping?: VehicleExternalMappingInput;
  functionSuggestions?: FunctionImportSuggestion[];
  cvSuggestions?: VehicleCVValueInput[];
  vehicle: CreateVehicleRequest;
};

export type FunctionImportSuggestion = VehicleFunctionInput & {
  functionKey: string;
  ecosDescription?: number;
  active?: boolean;
};

export type ECoSMatch = {
  vehicle: Vehicle;
  source: "mapping" | "decoder" | "name";
};

export type ECoSMatchable = Pick<ECoSRawLocomotive, "objectId" | "name" | "address">;

export type ECoSCVImportStatus = "new" | "changed" | "same" | "unmatched";

export type ECoSCVDefinition = {
  number: number;
  label: string;
  category: string;
  description: string;
};

export type ECoSDecoderManufacturer = {
  id: number;
  name: string;
  country?: string;
  binary?: string;
  hex?: string;
};

export type ECoSDecoderManufacturerMap = Record<number, ECoSDecoderManufacturer>;

export type ECoSDecoderProfileHint = {
  locomotive: ECoSRawLocomotive;
  manufacturerId?: number;
  manufacturer?: ECoSDecoderManufacturer;
  version?: number;
  protocol: string;
  profile: string;
};

export type ECoSCVImportRow = {
  id: string;
  locomotive: ECoSRawLocomotive;
  match: ECoSMatch | null;
  input: VehicleCVValueInput;
  definition?: ECoSCVDefinition;
  interpretation: string;
  manufacturer?: ECoSDecoderManufacturer;
  existingValue?: number;
  status: ECoSCVImportStatus;
};

export type ECoSVehicleDraftPayload = {
  source: "ecos";
  sourceSummary: {
    objectId: number;
    name: string;
    address: string;
    protocol: string;
    profile: string;
  };
  vehicle: CreateVehicleRequest;
  externalMapping: VehicleExternalMappingInput;
  cvValues: VehicleCVValueInput[];
  functionValues: FunctionImportSuggestion[];
  unclearFields: VehicleImportField[];
  returnToEcos?: {
    sessionId: string;
    objectId: number;
  };
};

export type ECoSBusyPhase = "idle" | "connecting" | "fetching";

export type ECoSImportSessionStatus = "open" | "editing" | "saved" | "skipped" | "error";

export type ECoSImportSession = {
  id: string;
  createdAt: string;
  updatedAt: string;
  rawProbe: ECoSRawProbe;
  statuses: Record<string, {
    status: ECoSImportSessionStatus;
    vehicleId?: string;
    message?: string;
    updatedAt: string;
  }>;
};

export type Translate = (key: string, values?: Record<string, string | number>) => string;

export type VehicleImportField = keyof CreateVehicleRequest;

export const ecosVehicleDraftStorageKey = "railkeeper.ecosVehicleDraft";
export const ecosImportSessionStorageKey = "railkeeper.ecosImportSession";
export const ecosRequiredFields: VehicleImportField[] = ["manufacturer", "name", "gauge", "category", "gattung"];

export type ColumnMapping = {
  index: number;
  header: string;
  normalized: string;
  key: VehicleImportField | "";
};

export type ImportTablePreview = {
  fileName: string;
  table: string[][];
  mappings: ColumnMapping[];
};

export type ImportChange = {
  key: VehicleImportField;
  label: string;
  current: string;
  incoming: string;
  status: "same" | "fill" | "overwrite";
};

export function optionValue(entry: MasterDataEntry) {
  return entry.label;
}

export const vehicleImportFields: { key: VehicleImportField; label: string }[] = [
  { key: "inventoryNumber", label: "Inventarnummer" },
  { key: "manufacturer", label: "Hersteller" },
  { key: "articleNumber", label: "Artikel-Nr." },
  { key: "articleSourceUrl", label: "Quelle / URL" },
  { key: "name", label: "Bezeichnung" },
  { key: "gauge", label: "Spurweite" },
  { key: "epoch", label: "Epoche" },
  { key: "railwayCompany", label: "Bahngesellschaft" },
  { key: "category", label: "Kategorie" },
  { key: "gattung", label: "Gattung" },
  { key: "description", label: "Beschreibung" },
  { key: "series", label: "Baureihe" },
  { key: "vehicleNumber", label: "Fahrzeug-Nr." },
  { key: "digital", label: "Digital" },
  { key: "digitalDecoderNumber", label: "Digital / Decoder-Nr." },
  { key: "dtDecoder", label: "DT / Decoder" },
  { key: "dtDecoderNumber", label: "DT / Decoder-Nr." },
  { key: "exhibitionReady", label: "Messe tauglich" },
  { key: "exhibition", label: "Ausstellung" },
  { key: "abcBrakes", label: "ABC-Bremsen" },
  { key: "ean", label: "EAN" },
  { key: "productionPeriod", label: "Produktionszeit" },
  { key: "listPrice", label: "Listenpreis" },
  { key: "lengthMm", label: "Länge (mm)" },
  { key: "weightG", label: "Gewicht (g)" },
  { key: "color", label: "Farbe" },
  { key: "lettering", label: "Beschriftung" },
  { key: "load", label: "Beladung" },
  { key: "interior", label: "Inneneinrichtung" },
  { key: "axles", label: "Achsen" },
  { key: "axleCount", label: "Anzahl Achsen" },
  { key: "tractionTireCount", label: "Anzahl Haftreifen" },
  { key: "wheelset", label: "Radsatz" },
  { key: "couplingSame", label: "Kupplung (V=H)" },
  { key: "couplingFront", label: "Kupplung vorne" },
  { key: "couplingRear", label: "Kupplung hinten" },
  { key: "powerPickup", label: "Stromabnahme" },
  { key: "adapter", label: "Adapter / Schnittstelle" },
  { key: "driveEnabled", label: "Antrieb" },
  { key: "driveDescription", label: "Antrieb Beschreibung" },
  { key: "headlightsEnabled", label: "Fahrlicht" },
  { key: "headlightsDescription", label: "Fahrlicht Beschreibung" },
  { key: "lightingEnabled", label: "Beleuchtung" },
  { key: "lightingDescription", label: "Beleuchtung Beschreibung" },
  { key: "soundGeneratorEnabled", label: "Soundgenerator" },
  { key: "soundGeneratorDescription", label: "Soundgenerator Beschreibung" },
  { key: "smokeGeneratorEnabled", label: "Rauchgenerator" },
  { key: "smokeGeneratorDescription", label: "Rauchgenerator Beschreibung" },
  { key: "additionalInfo", label: "Zusatzinformationen" },
  { key: "qrCodeEnabled", label: "QR-Code erstellen" }
];

export const booleanImportFields = new Set<VehicleImportField>([
  "digital",
  "dtDecoder",
  "exhibitionReady",
  "exhibition",
  "abcBrakes",
  "couplingSame",
  "driveEnabled",
  "headlightsEnabled",
  "lightingEnabled",
  "soundGeneratorEnabled",
  "smokeGeneratorEnabled",
  "qrCodeEnabled"
]);

export const columnAliases: Record<string, VehicleImportField> = {
  inventar: "inventoryNumber",
  inventarnummer: "inventoryNumber",
  inventarnr: "inventoryNumber",
  "inventar-nr": "inventoryNumber",
  "inventar-nr-": "inventoryNumber",
  inv: "inventoryNumber",
  invnr: "inventoryNumber",
  "inv-nr": "inventoryNumber",
  "inv-nr-": "inventoryNumber",
  inventarid: "inventoryNumber",
  bestandsnummer: "inventoryNumber",
  nummer: "inventoryNumber",
  hersteller: "manufacturer",
  fabrikat: "manufacturer",
  marke: "manufacturer",
  firma: "manufacturer",
  produzent: "manufacturer",
  artikel: "articleNumber",
  artikelnummer: "articleNumber",
  artikelnr: "articleNumber",
  "artikel-nr": "articleNumber",
  "artikel-nr-": "articleNumber",
  "artikel-nr-alt": "articleNumber",
  artnr: "articleNumber",
  "art-nr": "articleNumber",
  "art-nr-": "articleNumber",
  bestellnummer: "articleNumber",
  bestellnr: "articleNumber",
  "bestell-nr": "articleNumber",
  "bestell-nr-": "articleNumber",
  katalognummer: "articleNumber",
  katalognr: "articleNumber",
  "katalog-nr": "articleNumber",
  "katalog-nr-": "articleNumber",
  url: "articleSourceUrl",
  quelle: "articleSourceUrl",
  source: "articleSourceUrl",
  link: "articleSourceUrl",
  website: "articleSourceUrl",
  webseite: "articleSourceUrl",
  artikelquelle: "articleSourceUrl",
  bezeichnung: "name",
  name: "name",
  modell: "name",
  modellname: "name",
  fahrzeug: "name",
  fahrzeugname: "name",
  titel: "name",
  typ: "name",
  spur: "gauge",
  spurweite: "gauge",
  gauge: "gauge",
  nenngroesse: "gauge",
  nenngrosse: "gauge",
  nenngroessemassstab: "gauge",
  massstab: "gauge",
  masstab: "gauge",
  scale: "gauge",
  epoche: "epoch",
  era: "epoch",
  bahngesellschaft: "railwayCompany",
  bahn: "railwayCompany",
  evu: "railwayCompany",
  verwaltung: "railwayCompany",
  gesellschaft: "railwayCompany",
  kategorie: "category",
  fahrzeugkategorie: "category",
  art: "category",
  gattung: "gattung",
  bauart: "gattung",
  "bauart-gattung": "gattung",
  beschreibung: "description",
  notiz: "description",
  notizen: "description",
  kommentar: "description",
  bemerkung: "description",
  baureihe: "series",
  br: "series",
  reihe: "series",
  fahrzeugnummer: "vehicleNumber",
  fahrzeugnr: "vehicleNumber",
  "fahrzeug-nr": "vehicleNumber",
  digital: "digital",
  decoderja: "digital",
  decoder: "digitalDecoderNumber",
  decodernummer: "digitalDecoderNumber",
  decodernr: "digitalDecoderNumber",
  "decoder-nr": "digitalDecoderNumber",
  digitaldecoder: "digitalDecoderNumber",
  "digital-decoder": "digitalDecoderNumber",
  "digital-decoder-nr": "digitalDecoderNumber",
  digitaldecodernummer: "digitalDecoderNumber",
  dtdecoder: "dtDecoder",
  "dt-decoder": "dtDecoder",
  "dt-decoder-nr": "dtDecoderNumber",
  dtnummer: "dtDecoderNumber",
  dtdecodernummer: "dtDecoderNumber",
  messe: "exhibitionReady",
  messetauglich: "exhibitionReady",
  ausstellung: "exhibition",
  abcbremse: "abcBrakes",
  abcbremsen: "abcBrakes",
  ean: "ean",
  barcode: "ean",
  produktionszeit: "productionPeriod",
  produktion: "productionPeriod",
  baujahr: "productionPeriod",
  bauzeit: "productionPeriod",
  listenpreis: "listPrice",
  herstellerpreis: "listPrice",
  herstellerlistenpreis: "listPrice",
  "herstellerpreis-listenpreis": "listPrice",
  preis: "listPrice",
  uvp: "listPrice",
  laenge: "lengthMm",
  lange: "lengthMm",
  "laenge-mm": "lengthMm",
  "lange-mm": "lengthMm",
  laengemm: "lengthMm",
  langemm: "lengthMm",
  "laenge-in-mm": "lengthMm",
  "lange-in-mm": "lengthMm",
  mass: "lengthMm",
  mas: "lengthMm",
  "mass-mm": "lengthMm",
  "mas-mm": "lengthMm",
  "mass-mm-": "lengthMm",
  "mas-mm-": "lengthMm",
  masse: "lengthMm",
  "masse-mm": "lengthMm",
  gewicht: "weightG",
  "gewicht-g": "weightG",
  gewichtg: "weightG",
  farbe: "color",
  beschriftung: "lettering",
  beladung: "load",
  inneneinrichtung: "interior",
  einrichtung: "interior",
  achsen: "axles",
  anzahl: "axleCount",
  achsanzahl: "axleCount",
  anzahlachsen: "axleCount",
  "anzahl-achsen": "axleCount",
  haftreifen: "tractionTireCount",
  anzahlhaftreifen: "tractionTireCount",
  "anzahl-haftreifen": "tractionTireCount",
  radsatz: "wheelset",
  stromabnahme: "powerPickup",
  stromsystem: "powerPickup",
  strom: "powerPickup",
  adapter: "adapter",
  schnittstelle: "adapter",
  digitaleschnittstelle: "adapter",
  kupplung: "couplingFront",
  kupplungvorne: "couplingFront",
  "kupplung-vorne": "couplingFront",
  kupplunghinten: "couplingRear",
  "kupplung-hinten": "couplingRear",
  kupplungvh: "couplingSame",
  "kupplung-v-h": "couplingSame",
  "kupplung-v=h": "couplingSame",
  antrieb: "driveEnabled",
  antriebbeschreibung: "driveDescription",
  "antrieb-beschreibung": "driveDescription",
  fahrlicht: "headlightsEnabled",
  fahrlichtbeschreibung: "headlightsDescription",
  "fahrlicht-beschreibung": "headlightsDescription",
  beleuchtung: "lightingEnabled",
  licht: "lightingEnabled",
  beleuchtungsbeschreibung: "lightingDescription",
  "beleuchtung-beschreibung": "lightingDescription",
  lichtbeschreibung: "lightingDescription",
  "licht-beschreibung": "lightingDescription",
  sound: "soundGeneratorEnabled",
  soundgenerator: "soundGeneratorEnabled",
  soundmodul: "soundGeneratorEnabled",
  soundbeschreibung: "soundGeneratorDescription",
  "sound-beschreibung": "soundGeneratorDescription",
  soundgeneratorbeschreibung: "soundGeneratorDescription",
  "soundgenerator-beschreibung": "soundGeneratorDescription",
  rauch: "smokeGeneratorEnabled",
  rauchgenerator: "smokeGeneratorEnabled",
  rauchbeschreibung: "smokeGeneratorDescription",
  "rauch-beschreibung": "smokeGeneratorDescription",
  rauchgeneratorbeschreibung: "smokeGeneratorDescription",
  "rauchgenerator-beschreibung": "smokeGeneratorDescription",
  zusatzinfo: "additionalInfo",
  zusatzinformationen: "additionalInfo",
  zusatz: "additionalInfo",
  qrcode: "qrCodeEnabled",
  "qr-code": "qrCodeEnabled"
};

export function normalizeHeader(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/ä/g, "ae")
    .replace(/ö/g, "oe")
    .replace(/ü/g, "ue")
    .replace(/ß/g, "ss")
    .replace(/\s+/g, "")
    .replace(/[._/:()[\]]/g, "-")
    .replace(/^-+|-+$/g, "");
}

export function parseBoolean(value: string) {
  return ["1", "ja", "yes", "true", "wahr", "digital", "d", "x", "vorhanden"].includes(value.trim().toLowerCase());
}

export function comparableECoSName(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/ä/g, "ae")
    .replace(/ö/g, "oe")
    .replace(/ü/g, "ue")
    .replace(/ß/g, "ss")
    .replace(/[^a-z0-9]+/g, "")
    .replace(/^(br|v)(?=\d)/, "");
}

export function defaultColumnMappings(table: string[][]): ColumnMapping[] {
  return (table[0] || []).map((header, index) => {
    const normalized = normalizeHeader(header);
    return {
      index,
      header: header || `Spalte ${index + 1}`,
      normalized,
      key: columnAliases[normalized] || ""
    };
  });
}

export function csvEscape(value: unknown) {
  const text = String(value ?? "");
  return /[;"\n\r]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}

export function parseDelimited(text: string, delimiter: string) {
  const rows: string[][] = [];
  let current = "";
  let row: string[] = [];
  let quoted = false;

  for (let index = 0; index < text.length; index += 1) {
    const char = text[index];
    const next = text[index + 1];
    if (char === '"' && quoted && next === '"') {
      current += '"';
      index += 1;
    } else if (char === '"') {
      quoted = !quoted;
    } else if (char === delimiter && !quoted) {
      row.push(current.trim());
      current = "";
    } else if ((char === "\n" || char === "\r") && !quoted) {
      if (char === "\r" && next === "\n") {
        index += 1;
      }
      row.push(current.trim());
      if (row.some(Boolean)) {
        rows.push(row);
      }
      row = [];
      current = "";
    } else {
      current += char;
    }
  }
  row.push(current.trim());
  if (row.some(Boolean)) {
    rows.push(row);
  }
  return rows;
}

export function parseXMLImport(text: string) {
  const document = new DOMParser().parseFromString(text, "application/xml");
  const parseError = document.querySelector("parsererror");
  if (parseError) {
    throw new Error("XML-Datei konnte nicht gelesen werden.");
  }

  const preferredNames = new Set(["vehicle", "fahrzeug", "locomotive", "lok"]);
  const allElements = Array.from(document.getElementsByTagName("*"));
  let records = allElements.filter((element) => preferredNames.has(element.localName.toLowerCase()));
  if (records.length === 0) {
    const counts = new Map<string, number>();
    allElements.forEach((element) => {
      if (element.children.length > 0 && element !== document.documentElement) {
        const name = element.localName.toLowerCase();
        counts.set(name, (counts.get(name) || 0) + 1);
      }
    });
    const repeated = Array.from(counts.entries()).sort((a, b) => b[1] - a[1])[0]?.[0];
    if (repeated) {
      records = allElements.filter((element) => element.localName.toLowerCase() === repeated);
    }
  }

  const rows = records.map((record) => {
    const values = new Map<string, string>();
    Array.from(record.attributes).forEach((attribute) => values.set(attribute.name, attribute.value.trim()));
    Array.from(record.children).forEach((child) => {
      if (child.children.length === 0) {
        const value = child.textContent?.trim() || "";
        if (value) {
          values.set(child.localName, value);
        }
      }
    });
    return values;
  }).filter((row) => row.size > 0);

  const headers = Array.from(new Set(rows.flatMap((row) => Array.from(row.keys()))));
  if (headers.length === 0 || rows.length === 0) {
    return [];
  }
  return [headers, ...rows.map((row) => headers.map((header) => row.get(header) || ""))];
}

export function detectDelimiter(text: string) {
  const firstLine = text.split(/\r?\n/).find(Boolean) || "";
  const semicolon = (firstLine.match(/;/g) || []).length;
  const comma = (firstLine.match(/,/g) || []).length;
  const tab = (firstLine.match(/\t/g) || []).length;
  if (tab > semicolon && tab > comma) {
    return "\t";
  }
  return semicolon >= comma ? ";" : ",";
}

export function importRowsFromTable(
  table: string[][],
  existing: Vehicle[],
  mappings = defaultColumnMappings(table),
  labels = {
    missingManufacturer: "Hersteller fehlt",
    missingName: "Bezeichnung fehlt",
    missingGauge: "Spur fehlt",
    missingCategory: "Kategorie fehlt",
    missingGattung: "Gattung fehlt",
    duplicate: "Bestehendes Fahrzeug gefunden"
  }
) {
  const existingByInventory = new Map(existing.map((vehicle) => [vehicle.inventoryNumber.toLowerCase(), vehicle]));
  return table.slice(1).map((cells, index) => {
    const vehicle: CreateVehicleRequest = { manufacturer: "", name: "", gauge: "" };
    const importedKeys: (keyof CreateVehicleRequest)[] = [];
    mappings.forEach((mapping) => {
      const key = mapping.key;
      if (!key) {
        return;
      }
      const value = cells[mapping.index]?.trim() || "";
      if (!value) {
        return;
      }
      if (booleanImportFields.has(key)) {
        (vehicle as Record<string, unknown>)[key] = parseBoolean(value);
      } else {
        (vehicle as Record<string, unknown>)[key] = value;
      }
      importedKeys.push(key);
    });

    const issues: string[] = [];
    const duplicate = vehicle.inventoryNumber ? existingByInventory.get(vehicle.inventoryNumber.toLowerCase()) : undefined;
    if (!duplicate) {
      if (!vehicle.manufacturer) issues.push(labels.missingManufacturer);
      if (!vehicle.name) issues.push(labels.missingName);
      if (!vehicle.gauge) issues.push(labels.missingGauge);
      if (!vehicle.category) issues.push(labels.missingCategory);
      if (!vehicle.gattung) issues.push(labels.missingGattung);
    }
    if (duplicate) {
      issues.push(labels.duplicate);
    }

    return {
      id: `row-${index + 1}`,
      selected: !duplicate && issues.length === 0,
      mode: duplicate ? "update" as const : "create" as const,
      status: duplicate ? "warning" as const : issues.length === 0 ? "ok" as const : "error" as const,
      issues,
      importedKeys: Array.from(new Set(importedKeys)),
      duplicateVehicleId: duplicate?.id,
      vehicle
    };
  });
}

export function vehicleToRequest(vehicle: Vehicle): CreateVehicleRequest {
  return {
    inventoryNumber: vehicle.inventoryNumber,
    manufacturer: vehicle.manufacturer,
    articleNumber: vehicle.articleNumber,
    articleSourceUrl: vehicle.articleSourceUrl,
    name: vehicle.name,
    gauge: vehicle.gauge,
    epoch: vehicle.epoch,
    railwayCompany: vehicle.railwayCompany,
    category: vehicle.category,
    gattung: vehicle.gattung,
    description: vehicle.description,
    series: vehicle.series,
    vehicleNumber: vehicle.vehicleNumber,
    digital: vehicle.digital,
    digitalDecoderNumber: vehicle.digitalDecoderNumber,
    dtDecoder: vehicle.dtDecoder,
    dtDecoderNumber: vehicle.dtDecoderNumber,
    exhibitionReady: vehicle.exhibitionReady,
    exhibition: vehicle.exhibition,
    abcBrakes: vehicle.abcBrakes,
    ean: vehicle.ean,
    productionPeriod: vehicle.productionPeriod,
    listPrice: vehicle.listPrice,
    lengthMm: vehicle.lengthMm,
    weightG: vehicle.weightG,
    color: vehicle.color,
    lettering: vehicle.lettering,
    load: vehicle.load,
    interior: vehicle.interior,
    axles: vehicle.axles,
    axleCount: vehicle.axleCount,
    tractionTireCount: vehicle.tractionTireCount,
    wheelset: vehicle.wheelset,
    couplingSame: vehicle.couplingSame,
    couplingFront: vehicle.couplingFront,
    couplingRear: vehicle.couplingRear,
    powerPickup: vehicle.powerPickup,
    adapter: vehicle.adapter,
    driveEnabled: vehicle.driveEnabled,
    driveDescription: vehicle.driveDescription,
    headlightsEnabled: vehicle.headlightsEnabled,
    headlightsDescription: vehicle.headlightsDescription,
    lightingEnabled: vehicle.lightingEnabled,
    lightingDescription: vehicle.lightingDescription,
    soundGeneratorEnabled: vehicle.soundGeneratorEnabled,
    soundGeneratorDescription: vehicle.soundGeneratorDescription,
    smokeGeneratorEnabled: vehicle.smokeGeneratorEnabled,
    smokeGeneratorDescription: vehicle.smokeGeneratorDescription,
    additionalInfo: vehicle.additionalInfo,
    qrCodeEnabled: vehicle.qrCodeEnabled,
    images: vehicle.images?.map((image) => ({
      id: image.id,
      url: image.url,
      title: image.title,
      sourceUrl: image.sourceUrl,
      maintenanceId: image.maintenanceId,
      isPrimary: image.isPrimary,
      sortOrder: image.sortOrder
    }))
  };
}

export function ecosExternalMapping(locomotive: ECoSMatchable & { protocol?: string }): VehicleExternalMappingInput {
  return {
    provider: "ecos",
    externalId: String(locomotive.objectId),
    externalName: locomotive.name || "",
    externalAddress: typeof locomotive.address === "number" ? String(locomotive.address) : "",
    externalProtocol: locomotive.protocol || "",
    syncStatus: "linked"
  };
}

export function formatECoSDirection(direction?: number) {
  if (direction === 0) {
    return "Vorwärts";
  }
  if (direction === 1) {
    return "Rückwärts";
  }
  return "-";
}

export function formatECoSFunctions(locomotive: ECoSRawLocomotive) {
  const functions = locomotive.functions?.filter((fn) => fn.active || fn.description) || [];
  if (functions.length === 0) {
    return "-";
  }
  return functions
    .map((fn) => `F${fn.index}${fn.description ? `:${fn.description}` : ""}${fn.active ? "*" : ""}`)
    .join(", ");
}

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

export function normalizeECoSSymbolCode(code: number) {
  return String(code).padStart(3, "0");
}

export function symbolCodeSignal(symbol: MasterDataEntry) {
  const metadata = symbol.metadata || {};
  return [
    symbol.key,
    symbol.label,
    metadata.code,
    metadata.ecosCode,
    metadata.esuCode,
    metadata.fileName,
    metadata.originalName
  ]
    .map((value) => String(value ?? ""))
    .join(" ");
}

export function findSymbolByECoSCode(code: number, symbols: MasterDataEntry[]) {
  if (!code) return undefined;
  const padded = normalizeECoSSymbolCode(code);
  const codeText = String(code);
  return symbols.find((symbol) => {
    if (!symbol.active) return false;
    const signal = symbolCodeSignal(symbol);
    return new RegExp(`(^|\\D)${padded}(\\D|$)`).test(signal) || new RegExp(`(^|\\D)${codeText}(\\D|$)`).test(signal);
  });
}

export function fallbackECoSFunction(index: number, code?: number): Partial<VehicleFunctionInput> {
  if (index === 0 && (code === 3 || !code)) {
    return {
      name: "Fahrlicht",
      symbolKey: "light",
      functionType: "licht"
    };
  }
  return {};
}

export function ecosFunctionSuggestions(locomotive: ECoSRawLocomotive, symbols: MasterDataEntry[]): FunctionImportSuggestion[] {
  void locomotive;
  void symbols;
  return [];
}

export function ecosCVSuggestions(locomotive: ECoSRawLocomotive): VehicleCVValueInput[] {
  const decoderProfile = locomotive.profile || "";
  const cvs = eCoSCVValuesWithInferredManufacturer(locomotive);
  const inferredManufacturerId = inferECoSDecoderManufacturerId(locomotive);
  return cvs.map((cv) => {
    const definition = ecosStandardCVDefinitions[cv.number];
    const inferredDescription = cv.number === 8 && !findECoSCVValue(locomotive, 8) && typeof inferredManufacturerId === "number"
      ? "Herstellerkennung: aus ECoS-Decoderprofil abgeleitet."
      : "";
    return {
      cvNumber: cv.number,
      value: cv.value,
      category: definition?.category || inferECoSCVCategory(cv.number),
      protocol: normalizeECoSProtocolForCV(locomotive.protocol),
      decoderProfile,
      description: inferredDescription || (definition
        ? `${definition.label}: ${definition.description}`
        : `ECoS ${locomotive.objectId}${locomotive.name ? ` - ${locomotive.name}` : ""}`)
    };
  });
}

export function uniqueECoSValues(values: string[]) {
  return Array.from(new Set(values.filter((value) => value !== "")));
}

export const ecosRawKnownAttributes = new Set([
  "addr",
  "decoder",
  "decodername",
  "decodertype",
  "dir",
  "func",
  "funcdesc",
  "funcset",
  "name",
  "profile",
  "protocol",
  "speed",
  "speedstep",
  "cv",
  "cvs",
  "cvlist"
]);

export function firstECoSAttribute(locomotive: ECoSRawLocomotive, key: string) {
  return uniqueECoSValues(locomotive.attributes?.[key] || [])[0] || "";
}

export function renderECoSRawValue(value: string | number | undefined) {
  if (value === undefined || value === "") {
    return "-";
  }
  return value;
}

export function renderECoSRawDefinitionList(entries: { label: string; value: string | number | undefined }[]) {
  return (
    <dl className="ecos-raw-definition-list">
      {entries.map((entry) => (
        <div key={entry.label}>
          <dt>{entry.label}</dt>
          <dd>{renderECoSRawValue(entry.value)}</dd>
        </div>
      ))}
    </dl>
  );
}

export function rawECoSFunctions(locomotive: ECoSRawLocomotive) {
  const byIndex = new Map<number, { index: number; active: boolean; description?: number }>();
  for (const fn of locomotive.functions || []) {
    byIndex.set(fn.index, fn);
  }
  if (byIndex.size === 0 && locomotive.functionSet) {
    Array.from(locomotive.functionSet).forEach((state, index) => {
      byIndex.set(index, { index, active: state === "1" });
    });
  }
  return Array.from(byIndex.values())
    .filter((fn) => fn.active || typeof fn.description === "number")
    .sort((left, right) => left.index - right.index);
}

export function rawECoSUnknownAttributes(locomotive: ECoSRawLocomotive) {
  return Object.entries(locomotive.attributes || {})
    .filter(([key, values]) => values.length > 0 && !ecosRawKnownAttributes.has(key.toLowerCase()))
    .sort(([left], [right]) => left.localeCompare(right));
}

export function normalizeECoSProtocolForCV(protocol?: string) {
  const value = (protocol || "").trim();
  if (!value) return "";
  const compact = value.replace(/\s+/g, "").toUpperCase();
  const match = compact.match(/^([A-Z]+)(\d+)$/);
  if (match) {
    return `${match[1]} ${match[2]}`;
  }
  return value;
}

export const ecosDecoderManufacturers: ECoSDecoderManufacturerMap = {
  1: { id: 1, name: "CML Electronics" },
  2: { id: 2, name: "Train Technology" },
  3: { id: 3, name: "RamFixx Technologies" },
  11: { id: 11, name: "NCE" },
  12: { id: 12, name: "Wangrow" },
  13: { id: 13, name: "Public Domain / DIY" },
  14: { id: 14, name: "PSI - Dynatrol" },
  17: { id: 17, name: "Advance IC Engineering" },
  18: { id: 18, name: "JLC Enterprises" },
  19: { id: 19, name: "AR Elektronik" },
  20: { id: 20, name: "T4T" },
  21: { id: 21, name: "Kreischer Datentechnik" },
  22: { id: 22, name: "Linn Westcott" },
  23: { id: 23, name: "Computer Dialysis France" },
  24: { id: 24, name: "CTE" },
  25: { id: 25, name: "Throttle-Up" },
  27: { id: 27, name: "LGB" },
  28: { id: 28, name: "Digitrax" },
  29: { id: 29, name: "Lenz" },
  30: { id: 30, name: "Wangrow" },
  31: { id: 31, name: "NCE" },
  32: { id: 32, name: "DCC Supplies" },
  33: { id: 33, name: "TCS" },
  34: { id: 34, name: "ZIMO" },
  35: { id: 35, name: "SoundTraxx" },
  36: { id: 36, name: "NQ Electronics" },
  37: { id: 37, name: "LGB" },
  38: { id: 38, name: "Atlas" },
  39: { id: 39, name: "CVP Products" },
  40: { id: 40, name: "Uhlenbrock" },
  41: { id: 41, name: "Model Rectifier" },
  42: { id: 42, name: "Nagasue" },
  43: { id: 43, name: "KDE" },
  44: { id: 44, name: "Electronic Solutions Ulm" },
  45: { id: 45, name: "Digsight" },
  46: { id: 46, name: "Hornby" },
  47: { id: 47, name: "Viessmann" },
  48: { id: 48, name: "Joka" },
  49: { id: 49, name: "NCE Corporation" },
  50: { id: 50, name: "Fleischmann" },
  51: { id: 51, name: "Kato" },
  52: { id: 52, name: "Digitools" },
  53: { id: 53, name: "Bluecher Elektronik" },
  54: { id: 54, name: "MRC" },
  55: { id: 55, name: "Train Control Systems" },
  56: { id: 56, name: "ZTC Controls" },
  57: { id: 57, name: "CTE" },
  58: { id: 58, name: "MTH" },
  59: { id: 59, name: "Throttle-Up" },
  60: { id: 60, name: "Massoth" },
  65: { id: 65, name: "Throttles Up" },
  85: { id: 85, name: "Uhlenbrock" },
  97: { id: 97, name: "Doehler & Haass" },
  99: { id: 99, name: "Lenz" },
  101: { id: 101, name: "Bachmann" },
  117: { id: 117, name: "QSI" },
  129: { id: 129, name: "Digitrax" },
  141: { id: 141, name: "SoundTraxx" },
  145: { id: 145, name: "ZIMO" },
  151: { id: 151, name: "ESU" },
  153: { id: 153, name: "TCS" },
  157: { id: 157, name: "Kuehn" },
  159: { id: 159, name: "Viessmann" },
  161: { id: 161, name: "Maerklin / Trix" },
  162: { id: 162, name: "PIKO" },
  163: { id: 163, name: "Massoth" },
  167: { id: 167, name: "Maerklin" },
  173: { id: 173, name: "Maerklin / Trix" }
};

const esuCV8ManufacturerId = 151;

function cv8NumberFromMetadataValue(value: unknown) {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const parsed = Number.parseInt(value.trim(), 10);
    return Number.isFinite(parsed) ? parsed : undefined;
  }
  return undefined;
}

function normalizeCV8ManufacturerId(value: number | undefined) {
  return typeof value === "number" && Number.isFinite(value) && value >= 0 && value <= 255 ? value : undefined;
}

function cv8ManufacturerIdFromEntry(entry: MasterDataEntry) {
  const metadataId =
    normalizeCV8ManufacturerId(cv8NumberFromMetadataValue(entry.metadata?.decimal)) ??
    normalizeCV8ManufacturerId(cv8NumberFromMetadataValue(entry.metadata?.cvDecimal)) ??
    normalizeCV8ManufacturerId(cv8NumberFromMetadataValue(entry.metadata?.cv8)) ??
    normalizeCV8ManufacturerId(cv8NumberFromMetadataValue(entry.metadata?.manufacturerId));
  if (typeof metadataId === "number") return metadataId;

  const keyMatch = entry.key.match(/(?:^|-)0*(\d{1,3})$/);
  const keyId = keyMatch ? normalizeCV8ManufacturerId(Number.parseInt(keyMatch[1], 10)) : undefined;
  if (typeof keyId === "number") return keyId;

  const labelMatch = entry.label.match(/^\s*(\d{1,3})\s*-/);
  return labelMatch ? normalizeCV8ManufacturerId(Number.parseInt(labelMatch[1], 10)) : undefined;
}

function cv8StringMetadata(entry: MasterDataEntry, key: string) {
  const value = entry.metadata?.[key];
  return typeof value === "string" && value.trim() ? value.trim() : "";
}

function cv8ManufacturerNameFromEntry(entry: MasterDataEntry) {
  return entry.label.replace(/^\s*\d{1,3}\s*-\s*/, "").trim() || entry.label.trim();
}

function cv8BinaryFromId(id: number) {
  return id.toString(2).padStart(8, "0");
}

function cv8HexFromId(id: number) {
  return `0x${id.toString(16).toUpperCase().padStart(2, "0")}`;
}

export function cv8ManufacturersFromMasterData(entries: MasterDataEntry[]): ECoSDecoderManufacturerMap {
  const next: ECoSDecoderManufacturerMap = { ...ecosDecoderManufacturers };
  for (const entry of entries) {
    if (!entry.active) continue;
    const id = cv8ManufacturerIdFromEntry(entry);
    if (typeof id !== "number") continue;
    const name = cv8ManufacturerNameFromEntry(entry);
    const country = cv8StringMetadata(entry, "country") || cv8StringMetadata(entry, "cvCountry");
    const binary = cv8StringMetadata(entry, "binary") || cv8StringMetadata(entry, "cvBinary") || cv8BinaryFromId(id);
    const hex = cv8StringMetadata(entry, "hex") || cv8StringMetadata(entry, "cvHex") || cv8HexFromId(id);
    next[id] = {
      id,
      name: name || `CV8 ${id}`,
      country: country || undefined,
      binary,
      hex
    };
  }
  return next;
}

export const ecosStandardCVDefinitions: Record<number, ECoSCVDefinition> = {
  1: {
    number: 1,
    label: "Kurzadresse",
    category: "Adresse",
    description: "DCC-Kurzadresse des Decoders."
  },
  2: {
    number: 2,
    label: "Startspannung",
    category: "Fahrverhalten",
    description: "Anfahrspannung beziehungsweise unterste Fahrstufe."
  },
  3: {
    number: 3,
    label: "Beschleunigung",
    category: "Fahrverhalten",
    description: "Beschleunigungsverzoegerung."
  },
  4: {
    number: 4,
    label: "Bremsverzoegerung",
    category: "Fahrverhalten",
    description: "Brems- beziehungsweise Verzoegerungszeit."
  },
  5: {
    number: 5,
    label: "Hoechstgeschwindigkeit",
    category: "Fahrverhalten",
    description: "Maximale Geschwindigkeit, sofern vom Decoder unterstuetzt."
  },
  6: {
    number: 6,
    label: "Mittelgeschwindigkeit",
    category: "Fahrverhalten",
    description: "Mittlere Geschwindigkeit, sofern vom Decoder unterstuetzt."
  },
  7: {
    number: 7,
    label: "Decoder-Version",
    category: "Decoder",
    description: "Versions- oder Profilkennung, herstellerabhängig."
  },
  8: {
    number: 8,
    label: "Herstellerkennung",
    category: "Decoder",
    description: "NMRA-Herstellerkennung des Decoders."
  },
  17: {
    number: 17,
    label: "Lange Adresse High",
    category: "Adresse",
    description: "Oberer Anteil der langen DCC-Adresse."
  },
  18: {
    number: 18,
    label: "Lange Adresse Low",
    category: "Adresse",
    description: "Unterer Anteil der langen DCC-Adresse."
  },
  29: {
    number: 29,
    label: "Basis-Konfiguration",
    category: "Decoder",
    description: "Grundkonfiguration mit Fahrtrichtung, Fahrstufen, Analogbetrieb, Kennlinie und langer Adresse."
  }
};

export function inferECoSCVCategory(cvNumber: number) {
  const definition = ecosStandardCVDefinitions[cvNumber];
  if (definition) return definition.category;
  if (cvNumber === 1 || cvNumber === 17 || cvNumber === 18) return "Adresse";
  if ([2, 3, 4, 5, 6, 29].includes(cvNumber)) return "Fahrverhalten";
  if (cvNumber >= 7 && cvNumber <= 8) return "Decoder";
  if (cvNumber >= 33 && cvNumber <= 46) return "Funktion";
  if (cvNumber >= 49 && cvNumber <= 64) return "Motor";
  return "Sonstiges";
}

export function findECoSCVValue(locomotive: ECoSRawLocomotive, cvNumber: number) {
  return (locomotive.cvs || []).find((cv) => cv.number === cvNumber)?.value;
}

function eCoSDecoderIdentityText(locomotive: ECoSRawLocomotive) {
  const values = [
    locomotive.profile,
    locomotive.protocol,
    ...(locomotive.attributes?.decoder || []),
    ...(locomotive.attributes?.decodertype || []),
    ...(locomotive.attributes?.decodername || [])
  ];
  return values.filter(Boolean).join(" ").toLocaleLowerCase("de-DE");
}

export function inferECoSDecoderManufacturerId(locomotive: ECoSRawLocomotive) {
  const manufacturerId = findECoSCVValue(locomotive, 8);
  if (typeof manufacturerId === "number") return manufacturerId;
  const identity = eCoSDecoderIdentityText(locomotive);
  if (identity.includes("esu") || identity.includes("lokpilot") || identity.includes("loksound")) {
    return esuCV8ManufacturerId;
  }
  return undefined;
}

function eCoSCVValuesWithInferredManufacturer(locomotive: ECoSRawLocomotive) {
  const cvs = [...(locomotive.cvs || [])];
  const inferredManufacturerId = inferECoSDecoderManufacturerId(locomotive);
  if (!cvs.some((cv) => cv.number === 8) && typeof inferredManufacturerId === "number") {
    cvs.push({ number: 8, value: inferredManufacturerId });
  }
  return cvs;
}

export function getECoSDecoderManufacturer(locomotive: ECoSRawLocomotive, manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers) {
  const manufacturerId = inferECoSDecoderManufacturerId(locomotive);
  if (typeof manufacturerId !== "number") return undefined;
  return manufacturers[manufacturerId];
}

function formatECoSDecoderManufacturer(manufacturer: ECoSDecoderManufacturer, manufacturerId: number) {
  const country = manufacturer.country ? ` [${manufacturer.country}]` : "";
  return `${manufacturer.name}${country} (CV8 ${manufacturerId})`;
}

export function buildECoSDecoderProfileHints(raw: ECoSRawProbe | null, manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers): ECoSDecoderProfileHint[] {
  if (!raw) return [];
  return raw.locomotives
    .map((locomotive) => {
      const manufacturerId = inferECoSDecoderManufacturerId(locomotive);
      return {
        locomotive,
        manufacturerId,
        manufacturer: typeof manufacturerId === "number" ? manufacturers[manufacturerId] : undefined,
        version: findECoSCVValue(locomotive, 7),
        protocol: normalizeECoSProtocolForCV(locomotive.protocol),
        profile: locomotive.profile || ""
      };
    });
}

export function interpretECoSCV29(value: number) {
  const flags = [
    value & 1 ? "Fahrtrichtung invertiert" : "Fahrtrichtung normal",
    value & 2 ? "28/128 Fahrstufen" : "14 Fahrstufen",
    value & 4 ? "Analogbetrieb erlaubt" : "nur Digitalbetrieb",
    value & 16 ? "freie Geschwindigkeitstabelle" : "3-Punkt-Kennlinie",
    value & 32 ? "lange Adresse aktiv" : "kurze Adresse aktiv"
  ];
  if (value & 8) {
    flags.push("Bit 3 herstellerabhängig gesetzt");
  }
  return flags.join(" - ");
}

export function interpretECoSLongAddress(locomotive: ECoSRawLocomotive) {
  const high = findECoSCVValue(locomotive, 17);
  const low = findECoSCVValue(locomotive, 18);
  if (typeof high !== "number" || typeof low !== "number" || high < 192 || high > 231) {
    return "";
  }
  return String(((high - 192) * 256) + low);
}

export function interpretECoSCV(cvNumber: number, value: number, locomotive: ECoSRawLocomotive, manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers) {
  if (cvNumber === 8) {
    const manufacturer = getECoSDecoderManufacturer(locomotive, manufacturers);
    return manufacturer ? `${manufacturer.name} (${value})` : `Unbekannte Herstellerkennung (${value})`;
  }
  if (cvNumber === 17 || cvNumber === 18) {
    const longAddress = interpretECoSLongAddress(locomotive);
    return longAddress ? `Lange Adresse ${longAddress}` : "Lange Adresse unvollstaendig";
  }
  if (cvNumber === 29) {
    return interpretECoSCV29(value);
  }
  const definition = ecosStandardCVDefinitions[cvNumber];
  if (definition) {
    return `${definition.label} (${value})`;
  }
  return "-";
}

export function cvPreviewKey(cvNumber: number, decoderProfile?: string) {
  return `${cvNumber}::${(decoderProfile || "").trim().toLocaleLowerCase("de-DE")}`;
}

export function buildECoSCVImportRows(raw: ECoSRawProbe | null, vehicles: Vehicle[], manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers): ECoSCVImportRow[] {
  if (!raw) return [];
  return raw.locomotives.flatMap((locomotive) => {
    const cvs = eCoSCVValuesWithInferredManufacturer(locomotive);
    if (cvs.length === 0) return [];
    const match = findECoSMatch(locomotive, vehicles);
    const decoderProfile = locomotive.profile || "";
    const existingByExactKey = new Map((match?.vehicle.cvValues || []).map((entry) => [cvPreviewKey(entry.cvNumber, entry.decoderProfile), entry]));
    const existingByNumber = new Map((match?.vehicle.cvValues || []).map((entry) => [entry.cvNumber, entry]));
    return cvs.map((cv) => {
      const definition = ecosStandardCVDefinitions[cv.number];
      const input: VehicleCVValueInput = {
        cvNumber: cv.number,
        value: cv.value,
        category: definition?.category || inferECoSCVCategory(cv.number),
        protocol: normalizeECoSProtocolForCV(locomotive.protocol),
        decoderProfile,
        description: definition
          ? `${definition.label}: ${definition.description}`
          : `ECoS ${locomotive.objectId}${locomotive.name ? ` - ${locomotive.name}` : ""}`
      };
      const existing = existingByExactKey.get(cvPreviewKey(input.cvNumber, input.decoderProfile)) || existingByNumber.get(input.cvNumber);
      const status: ECoSCVImportStatus = !match
        ? "unmatched"
        : !existing
          ? "new"
          : existing.value === input.value
            ? "same"
            : "changed";
      return {
        id: `${locomotive.objectId}-${cv.number}`,
        locomotive,
        match,
        input,
        definition,
        interpretation: interpretECoSCV(cv.number, cv.value, locomotive, manufacturers),
        manufacturer: getECoSDecoderManufacturer(locomotive, manufacturers),
        existingValue: existing?.value,
        status
      };
    });
  });
}

export function renderECoSCVImportPreview(
  raw: ECoSRawProbe | null,
  vehicles: Vehicle[],
  t: Translate,
  manufacturers: ECoSDecoderManufacturerMap,
  onOpenVehicleDraft?: (locomotive: ECoSRawLocomotive) => void
) {
  const rows = buildECoSCVImportRows(raw, vehicles, manufacturers);
  const profileHints = buildECoSDecoderProfileHints(raw, manufacturers);
  if (rows.length === 0 && profileHints.length === 0) {
    return null;
  }
  const matchedLocomotives = new Set(rows.filter((row) => row.match).map((row) => row.locomotive.objectId)).size;
  const changedRows = rows.filter((row) => row.status === "new" || row.status === "changed").length;
  return (
    <section className="ecos-cv-review-panel">
      <div className="ecos-preview-toolbar">
        <span>{t("importExport.ecos.cvReview.title")}</span>
        <span>{t("importExport.ecos.cvReview.summary", { count: rows.length, matched: matchedLocomotives, changed: changedRows })}</span>
      </div>
      <p>{t("importExport.ecos.cvReview.subtitle")}</p>
      {profileHints.length > 0 && (
        <div className="ecos-cv-profile-grid" aria-label={t("importExport.ecos.cvReview.decoderProfiles")}>
          {profileHints.map((hint) => (
            <button key={hint.locomotive.objectId} type="button" className="ecos-cv-profile-card" onClick={() => onOpenVehicleDraft?.(hint.locomotive)} title={t("importExport.ecos.openVehicleDraft")}>
              <strong>{hint.locomotive.name || `ECoS ${hint.locomotive.objectId}`}</strong>
              <dl>
                <div>
                  <dt>{t("importExport.ecos.cvReview.manufacturer")}</dt>
                  <dd>
                    {typeof hint.manufacturerId === "number"
                      ? hint.manufacturer
                        ? formatECoSDecoderManufacturer(hint.manufacturer, hint.manufacturerId)
                        : t("importExport.ecos.cvReview.unknownManufacturer", { id: hint.manufacturerId })
                      : t("importExport.ecos.cvReview.missingManufacturerCV")}
                  </dd>
                </div>
                <div>
                  <dt>{t("importExport.ecos.cvReview.version")}</dt>
                  <dd>{typeof hint.version === "number" ? hint.version : "-"}</dd>
                </div>
                <div>
                  <dt>{t("importExport.ecos.cvReview.profileHint")}</dt>
                  <dd>{[hint.manufacturer?.name, hint.profile, hint.protocol].filter(Boolean).join(" - ") || "-"}</dd>
                </div>
              </dl>
            </button>
          ))}
        </div>
      )}
      {rows.length > 0 && <div className="table-wrap ecos-cv-review-table">
        <table>
          <thead>
            <tr>
              <th>{t("importExport.ecos.cvReview.locomotive")}</th>
              <th>{t("importExport.ecos.cvReview.match")}</th>
              <th>CV</th>
              <th>{t("importExport.ecos.cvReview.meaning")}</th>
              <th>{t("importExport.ecos.cvReview.ecosValue")}</th>
              <th>{t("importExport.ecos.cvReview.interpretation")}</th>
              <th>{t("importExport.ecos.cvReview.railkeeperValue")}</th>
              <th>{t("importExport.ecos.cvReview.category")}</th>
              <th>{t("importExport.ecos.cvReview.protocol")}</th>
              <th>{t("importExport.ecos.cvReview.status")}</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.id}>
                <td>
                  <strong>{row.locomotive.name || `ECoS ${row.locomotive.objectId}`}</strong>
                  <small>#{row.locomotive.objectId}{row.locomotive.address ? ` · ${row.locomotive.address}` : ""}</small>
                </td>
                <td>{row.match ? `${row.match.vehicle.inventoryNumber} · ${row.match.vehicle.name}` : t("importExport.ecos.cvReview.noMatch")}</td>
                <td>{row.input.cvNumber}</td>
                <td>{row.definition?.label || "-"}</td>
                <td>{row.input.value}</td>
                <td className="ecos-cv-interpretation">{row.interpretation}</td>
                <td>{row.existingValue ?? "-"}</td>
                <td>{row.input.category || "-"}</td>
                <td>{row.input.protocol || "-"}</td>
                <td><span className={`ecos-cv-status ${row.status}`}>{t(`importExport.ecos.cvReview.status.${row.status}`)}</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>}
      <p className="source-note backup-note">{t("importExport.ecos.cvReview.note")}</p>
    </section>
  );
}

export function renderECoSRawProbe(
  raw: ECoSRawProbe | null,
  vehicles: Vehicle[],
  t: Translate,
  manufacturers: ECoSDecoderManufacturerMap,
  onOpenVehicleDraft?: (locomotive: ECoSRawLocomotive) => void
) {
  if (!raw) {
    return null;
  }
  return (
    <div className="ecos-raw-panel">
      <div className="ecos-preview-toolbar">
        <span>{raw.message}</span>
        <span>{t("importExport.ecos.rawProbeFields", { count: raw.probeFields.length })}</span>
      </div>
      {raw.locomotives.length === 0 ? <p className="empty-state">{t("importExport.ecos.rawEmpty")}</p> : renderECoSCVImportPreview(raw, vehicles, t, manufacturers, onOpenVehicleDraft)}
      <p className="source-note backup-note">{t("importExport.ecos.rawNote")}</p>
    </div>
  );
}

export function mergeExternalMapping(vehicle: Vehicle, mapping: VehicleExternalMapping): Vehicle {
  const mappings = (vehicle.externalMappings || []).filter((item) => !(item.provider === mapping.provider && item.externalId === mapping.externalId));
  return { ...vehicle, externalMappings: [...mappings, mapping] };
}

export function findECoSMatch(locomotive: ECoSMatchable, vehicles: Vehicle[]): ECoSMatch | null {
  const address = typeof locomotive.address === "number" ? String(locomotive.address) : "";
  const objectId = String(locomotive.objectId);
  const name = comparableECoSName(locomotive.name || "");
  for (const vehicle of vehicles) {
    if (vehicle.externalMappings?.some((mapping) => mapping.provider === "ecos" && mapping.externalId === objectId)) {
      return { vehicle, source: "mapping" };
    }
  }
  for (const vehicle of vehicles) {
    if (address && vehicle.digitalDecoderNumber === address) {
      return { vehicle, source: "decoder" };
    }
  }
  for (const vehicle of vehicles) {
    const vehicleName = comparableECoSName(vehicle.name || "");
    const vehicleNumber = comparableECoSName(vehicle.vehicleNumber || "");
    if (name && (
      vehicleName === name ||
      vehicleNumber === name ||
      (vehicleName.length > 5 && name.includes(vehicleName)) ||
      (vehicleNumber.length > 5 && name.includes(vehicleNumber))
    )) {
      return { vehicle, source: "name" };
    }
  }
  return null;
}

export function buildECoSVehicleDraftRow(
  locomotive: ECoSRawLocomotive,
  vehicles: Vehicle[],
  symbols: MasterDataEntry[],
  labels: {
    matched: string;
    missingManufacturer: string;
    missingName: string;
    missingGauge: string;
    missingCategory: string;
    missingGattung: string;
  },
  manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers
): ImportRow {
  const match = findECoSMatch(locomotive, vehicles);
  const manufacturer = getECoSDecoderManufacturer(locomotive, manufacturers);
  const protocol = normalizeECoSProtocolForCV(locomotive.protocol);
  const decoderProfile = [manufacturer?.name, locomotive.profile].filter(Boolean).join(" ");
  const vehicle: CreateVehicleRequest = match
    ? vehicleToRequest(match.vehicle)
    : {
        manufacturer: "",
        name: locomotive.name || `ECoS ${locomotive.objectId}`,
        gauge: "",
        category: "Lokomotive",
        digital: true,
        digitalDecoderNumber: typeof locomotive.address === "number" ? String(locomotive.address) : "",
        decoderType: decoderProfile,
        description: `ECoS-ID ${locomotive.objectId}${protocol ? `, Protokoll ${protocol}` : ""}`
      };

  vehicle.digital = true;
  if (typeof locomotive.address === "number") {
    vehicle.digitalDecoderNumber = String(locomotive.address);
  }
  if (decoderProfile && !vehicle.decoderType) {
    vehicle.decoderType = decoderProfile;
  }
  if (!vehicle.description) {
    vehicle.description = `ECoS-ID ${locomotive.objectId}${protocol ? `, Protokoll ${protocol}` : ""}`;
  }

  const issues: string[] = [];
  if (match) {
    issues.push(labels.matched);
  } else {
    if (!vehicle.manufacturer) issues.push(labels.missingManufacturer);
    if (!vehicle.name) issues.push(labels.missingName);
    if (!vehicle.gauge) issues.push(labels.missingGauge);
    if (!vehicle.category) issues.push(labels.missingCategory);
    if (!vehicle.gattung) issues.push(labels.missingGattung);
  }

  return {
    id: `ecos-raw-${locomotive.objectId}`,
    selected: Boolean(match),
    mode: match ? "update" : "create",
    status: match ? "warning" : issues.length ? "error" : "ok",
    issues,
    importedKeys: ["name", "category", "digital", "digitalDecoderNumber", "decoderType", "description"],
    duplicateVehicleId: match?.vehicle.id,
    externalMapping: ecosExternalMapping(locomotive),
    functionSuggestions: ecosFunctionSuggestions(locomotive, symbols),
    cvSuggestions: ecosCVSuggestions(locomotive),
    vehicle
  };
}

export function mergeImportedVehicle(existing: Vehicle, incoming: CreateVehicleRequest, importedKeys: (keyof CreateVehicleRequest)[]) {
  const merged = vehicleToRequest(existing);
  importedKeys.forEach((key) => {
    const value = incoming[key];
    if (typeof value === "boolean" || (typeof value === "string" && value.trim() !== "")) {
      (merged as Record<string, unknown>)[key] = value;
    }
  });
  return merged;
}

export function displayImportValue(value: unknown, yes = "ja", no = "nein") {
  if (typeof value === "boolean") {
    return value ? yes : no;
  }
  if (typeof value === "string") {
    return value.trim() || "-";
  }
  return "-";
}

export function valuesEqual(current: unknown, incoming: unknown) {
  if (typeof current === "boolean" || typeof incoming === "boolean") {
    return Boolean(current) === Boolean(incoming);
  }
  return String(current ?? "").trim() === String(incoming ?? "").trim();
}

export function getImportChanges(
  row: ImportRow,
  existing: Vehicle | undefined,
  fieldLabel: (key: VehicleImportField) => string,
  yes: string,
  no: string
): ImportChange[] {
  if (!existing) {
    return [];
  }
  return row.importedKeys
    .filter((key) => key !== "images")
    .map((key) => {
      const current = existing[key as keyof Vehicle];
      const incoming = row.vehicle[key];
      const currentText = displayImportValue(current, yes, no);
      const incomingText = displayImportValue(incoming, yes, no);
      return {
        key,
        label: fieldLabel(key),
        current: currentText,
        incoming: incomingText,
        status: valuesEqual(current, incoming) ? "same" : currentText === "-" ? "fill" : "overwrite"
      };
    });
}

export function vehiclesToCSV(vehicles: Vehicle[], fieldLabel: (key: VehicleImportField) => string, yes: string, no: string) {
  const headers = vehicleImportFields.map((field) => fieldLabel(field.key));
  const rows = vehicles.map((vehicle) => {
    const request = vehicleToRequest(vehicle);
    return vehicleImportFields.map((field) => displayImportValue(request[field.key], yes, no).replace(/^-$/, ""));
  });
  return [headers, ...rows].map((row) => row.map(csvEscape).join(";")).join("\n");
}

export function downloadText(fileName: string, content: string, type: string) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  link.click();
  URL.revokeObjectURL(url);
}

export function htmlEscape(value: unknown) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export function printInventory(
  vehicles: Vehicle[],
  fieldLabel: (key: VehicleImportField) => string,
  language: string,
  t: (key: string, values?: Record<string, string | number>) => string
) {
  const printWindow = window.open("", "railkeeper-bestand-druck");
  if (!printWindow) {
    window.alert(t("importExport.print.blocked"));
    return;
  }

  const digital = vehicles.filter((vehicle) => vehicle.digital).length;
  const analog = vehicles.length - digital;
  const rows = vehicles.map((vehicle) => `
    <tr>
      <td>${htmlEscape(vehicle.inventoryNumber)}</td>
      <td>${htmlEscape(vehicle.manufacturer)}</td>
      <td>${htmlEscape(vehicle.articleNumber)}</td>
      <td>${htmlEscape(vehicle.name)}</td>
      <td>${htmlEscape(vehicle.gauge)}</td>
      <td>${htmlEscape(vehicle.epoch)}</td>
      <td>${htmlEscape(vehicle.category)}</td>
      <td>${vehicle.digital ? "digital" : "analog"}</td>
      <td>${htmlEscape(vehicle.listPrice)}</td>
    </tr>
  `).join("");

  printWindow.document.open();
  printWindow.document.write(`<!doctype html>
    <html lang="${htmlEscape(language)}">
      <head>
        <meta charset="utf-8" />
        <title>${htmlEscape(t("importExport.print.title"))}</title>
        <style>
          @page { size: A4 landscape; margin: 14mm; }
          * { box-sizing: border-box; }
          body { margin: 0; color: #0b1e26; font-family: "Segoe UI", Arial, sans-serif; font-size: 11px; }
          header { display: flex; align-items: flex-start; justify-content: space-between; gap: 18px; margin-bottom: 16px; padding-bottom: 10px; border-bottom: 2px solid #1c621b; }
          h1 { margin: 0 0 4px; font-size: 24px; line-height: 1.1; }
          p { margin: 0; color: #4f6869; }
          .stats { display: grid; grid-template-columns: repeat(3, auto); gap: 8px; text-align: right; }
          .stats span { display: block; padding: 6px 8px; border: 1px solid #d5dfdc; border-radius: 6px; background: #f5f8f6; font-weight: 700; }
          table { width: 100%; border-collapse: collapse; }
          th, td { padding: 6px 7px; border-bottom: 1px solid #d5dfdc; text-align: left; vertical-align: top; }
          th { background: #edf2f1; color: #4f6869; font-size: 10px; text-transform: uppercase; }
          tr:nth-child(even) td { background: #f8faf9; }
          footer { margin-top: 12px; color: #4f6869; font-size: 10px; }
        </style>
      </head>
      <body>
        <header>
          <div>
            <h1>${htmlEscape(t("importExport.print.title"))}</h1>
            <p>${htmlEscape(t("importExport.print.footer", { date: new Date().toLocaleString(language === "en" ? "en-US" : "de-DE") }))}</p>
          </div>
          <div class="stats">
            <span>${htmlEscape(t("importExport.print.summary", { total: vehicles.length, digital, analog }))}</span>
            <span>${digital} digital</span>
            <span>${analog} analog</span>
          </div>
        </header>
        <table>
          <thead>
            <tr>
              <th>${htmlEscape(t("importExport.review.inventory"))}</th>
              <th>${htmlEscape(fieldLabel("manufacturer"))}</th>
              <th>${htmlEscape(t("importExport.review.article"))}</th>
              <th>${htmlEscape(fieldLabel("name"))}</th>
              <th>${htmlEscape(fieldLabel("gauge"))}</th>
              <th>${htmlEscape(fieldLabel("epoch"))}</th>
              <th>${htmlEscape(fieldLabel("category"))}</th>
              <th>${htmlEscape(fieldLabel("digital"))}</th>
              <th>${htmlEscape(fieldLabel("listPrice"))}</th>
            </tr>
          </thead>
          <tbody>${rows || `<tr><td colspan="9">${htmlEscape(t("importExport.review.empty"))}</td></tr>`}</tbody>
        </table>
        <footer>${htmlEscape(t("importExport.export.print"))}</footer>
        <script>
          window.addEventListener("load", () => {
            window.focus();
            window.print();
          });
        </script>
      </body>
    </html>`);
  printWindow.document.close();
}
