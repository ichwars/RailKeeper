import { ChangeEvent, Fragment, useEffect, useMemo, useState } from "react";
import { AlertTriangle, Check, ClipboardCheck, Database, Download, FileInput, Printer, Save, Upload } from "lucide-react";
import { api, CreateVehicleRequest, ECoSConnectionResult, ECoSRawLocomotive, ECoSRawProbe, MasterDataEntry, Vehicle, VehicleCVValueInput, VehicleExternalMapping, VehicleExternalMappingInput, VehicleFunctionInput } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

type ImportRow = {
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

type FunctionImportSuggestion = VehicleFunctionInput & {
  functionKey: string;
  ecosDescription?: number;
  active?: boolean;
};

type ECoSMatch = {
  vehicle: Vehicle;
  source: "mapping" | "decoder" | "name";
};

type ECoSMatchable = Pick<ECoSRawLocomotive, "objectId" | "name" | "address">;

type ECoSCVImportStatus = "new" | "changed" | "same" | "unmatched";

type ECoSCVDefinition = {
  number: number;
  label: string;
  category: string;
  description: string;
};

type ECoSDecoderManufacturer = {
  id: number;
  name: string;
};

type ECoSDecoderManufacturerMap = Record<number, ECoSDecoderManufacturer>;

type ECoSDecoderProfileHint = {
  locomotive: ECoSRawLocomotive;
  manufacturerId?: number;
  manufacturer?: ECoSDecoderManufacturer;
  version?: number;
  protocol: string;
  profile: string;
};

type ECoSCVImportRow = {
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

type ECoSVehicleDraftPayload = {
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
  unclearFields: VehicleImportField[];
};

type ECoSBusyPhase = "idle" | "connecting" | "fetching";

type Translate = (key: string, values?: Record<string, string | number>) => string;

type VehicleImportField = keyof CreateVehicleRequest;

const ecosVehicleDraftStorageKey = "railkeeper.ecosVehicleDraft";
const ecosRequiredFields: VehicleImportField[] = ["manufacturer", "name", "gauge", "category", "gattung"];

type ColumnMapping = {
  index: number;
  header: string;
  normalized: string;
  key: VehicleImportField | "";
};

type ImportTablePreview = {
  fileName: string;
  table: string[][];
  mappings: ColumnMapping[];
};

type ImportChange = {
  key: VehicleImportField;
  label: string;
  current: string;
  incoming: string;
  status: "same" | "fill" | "overwrite";
};

function optionValue(entry: MasterDataEntry) {
  return entry.label;
}

const vehicleImportFields: { key: VehicleImportField; label: string }[] = [
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

const booleanImportFields = new Set<VehicleImportField>([
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

const columnAliases: Record<string, VehicleImportField> = {
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

function normalizeHeader(value: string) {
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

function parseBoolean(value: string) {
  return ["1", "ja", "yes", "true", "wahr", "digital", "d", "x", "vorhanden"].includes(value.trim().toLowerCase());
}

function comparableECoSName(value: string) {
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

function defaultColumnMappings(table: string[][]): ColumnMapping[] {
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

function csvEscape(value: unknown) {
  const text = String(value ?? "");
  return /[;"\n\r]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}

function parseDelimited(text: string, delimiter: string) {
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

function parseXMLImport(text: string) {
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

function detectDelimiter(text: string) {
  const firstLine = text.split(/\r?\n/).find(Boolean) || "";
  const semicolon = (firstLine.match(/;/g) || []).length;
  const comma = (firstLine.match(/,/g) || []).length;
  const tab = (firstLine.match(/\t/g) || []).length;
  if (tab > semicolon && tab > comma) {
    return "\t";
  }
  return semicolon >= comma ? ";" : ",";
}

function importRowsFromTable(
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

function vehicleToRequest(vehicle: Vehicle): CreateVehicleRequest {
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

function ecosExternalMapping(locomotive: ECoSMatchable & { protocol?: string }): VehicleExternalMappingInput {
  return {
    provider: "ecos",
    externalId: String(locomotive.objectId),
    externalName: locomotive.name || "",
    externalAddress: typeof locomotive.address === "number" ? String(locomotive.address) : "",
    externalProtocol: locomotive.protocol || "",
    syncStatus: "linked"
  };
}

function formatECoSDirection(direction?: number) {
  if (direction === 0) {
    return "Vorwärts";
  }
  if (direction === 1) {
    return "Rückwärts";
  }
  return "-";
}

function formatECoSFunctions(locomotive: ECoSRawLocomotive) {
  const functions = locomotive.functions?.filter((fn) => fn.active || fn.description) || [];
  if (functions.length === 0) {
    return "-";
  }
  return functions
    .map((fn) => `F${fn.index}${fn.description ? `:${fn.description}` : ""}${fn.active ? "*" : ""}`)
    .join(", ");
}

function inferFunctionTypeFromSymbol(symbolKey: string, symbols: MasterDataEntry[], fallback = "standard") {
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

function normalizeECoSSymbolCode(code: number) {
  return String(code).padStart(3, "0");
}

function symbolCodeSignal(symbol: MasterDataEntry) {
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

function findSymbolByECoSCode(code: number, symbols: MasterDataEntry[]) {
  if (!code) return undefined;
  const padded = normalizeECoSSymbolCode(code);
  const codeText = String(code);
  return symbols.find((symbol) => {
    if (!symbol.active) return false;
    const signal = symbolCodeSignal(symbol);
    return new RegExp(`(^|\\D)${padded}(\\D|$)`).test(signal) || new RegExp(`(^|\\D)${codeText}(\\D|$)`).test(signal);
  });
}

function fallbackECoSFunction(index: number, code?: number): Partial<VehicleFunctionInput> {
  if (index === 0 && (code === 3 || !code)) {
    return {
      name: "Fahrlicht",
      symbolKey: "light",
      functionType: "licht"
    };
  }
  return {};
}

function ecosFunctionSuggestions(locomotive: ECoSRawLocomotive, symbols: MasterDataEntry[]): FunctionImportSuggestion[] {
  return (locomotive.functions || [])
    .filter((fn) => fn.active || fn.description)
    .map((fn) => {
      const symbol = typeof fn.description === "number" ? findSymbolByECoSCode(fn.description, symbols) : undefined;
      const fallback = fallbackECoSFunction(fn.index, fn.description);
      const symbolKey = symbol?.key || fallback.symbolKey || "";
      const name = symbol?.label || fallback.name || "";
      const functionType = inferFunctionTypeFromSymbol(symbolKey, symbols, fallback.functionType || "standard");
      const notes = [
        typeof fn.description === "number" ? `ECoS funcdesc ${fn.description}` : "",
        fn.active ? "ECoS aktiv" : ""
      ].filter(Boolean).join(" · ");
      return {
        functionKey: `F${fn.index}`,
        name,
        symbolKey,
        functionType,
        mode: "dauer",
        directionDependent: false,
        notes,
        ecosDescription: fn.description,
        active: fn.active
      };
    });
}

function ecosCVSuggestions(locomotive: ECoSRawLocomotive): VehicleCVValueInput[] {
  const decoderProfile = locomotive.profile || "";
  return (locomotive.cvs || []).map((cv) => {
    const definition = ecosStandardCVDefinitions[cv.number];
    return {
      cvNumber: cv.number,
      value: cv.value,
      category: definition?.category || inferECoSCVCategory(cv.number),
      protocol: normalizeECoSProtocolForCV(locomotive.protocol),
      decoderProfile,
      description: definition
        ? `${definition.label}: ${definition.description}`
        : `ECoS ${locomotive.objectId}${locomotive.name ? ` - ${locomotive.name}` : ""}`
    };
  });
}

function uniqueECoSValues(values: string[]) {
  return Array.from(new Set(values.filter((value) => value !== "")));
}

const ecosRawKnownAttributes = new Set([
  "addr",
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

function firstECoSAttribute(locomotive: ECoSRawLocomotive, key: string) {
  return uniqueECoSValues(locomotive.attributes?.[key] || [])[0] || "";
}

function renderECoSRawValue(value: string | number | undefined) {
  if (value === undefined || value === "") {
    return "-";
  }
  return value;
}

function renderECoSRawDefinitionList(entries: { label: string; value: string | number | undefined }[]) {
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

function rawECoSFunctions(locomotive: ECoSRawLocomotive) {
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

function rawECoSUnknownAttributes(locomotive: ECoSRawLocomotive) {
  return Object.entries(locomotive.attributes || {})
    .filter(([key, values]) => values.length > 0 && !ecosRawKnownAttributes.has(key.toLowerCase()))
    .sort(([left], [right]) => left.localeCompare(right));
}

function normalizeECoSProtocolForCV(protocol?: string) {
  const value = (protocol || "").trim();
  if (!value) return "";
  const compact = value.replace(/\s+/g, "").toUpperCase();
  const match = compact.match(/^([A-Z]+)(\d+)$/);
  if (match) {
    return `${match[1]} ${match[2]}`;
  }
  return value;
}

const ecosDecoderManufacturers: ECoSDecoderManufacturerMap = {
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

function cv8ManufacturersFromMasterData(entries: MasterDataEntry[]): ECoSDecoderManufacturerMap {
  const next: ECoSDecoderManufacturerMap = { ...ecosDecoderManufacturers };
  for (const entry of entries) {
    if (!entry.active) continue;
    const metadataCV8 = typeof entry.metadata?.cv8 === "number" ? entry.metadata.cv8 : Number(entry.metadata?.cv8);
    const keyCV8 = Number((entry.key || "").replace(/\D+/g, ""));
    const id = Number.isFinite(metadataCV8) && metadataCV8 > 0 ? metadataCV8 : keyCV8;
    if (!Number.isFinite(id) || id <= 0) continue;
    const name = entry.label.replace(/^\s*\d+\s*-\s*/, "").trim();
    next[id] = { id, name: name || entry.label };
  }
  return next;
}

const ecosStandardCVDefinitions: Record<number, ECoSCVDefinition> = {
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
    description: "Versions- oder Profilkennung, herstellerabhaengig."
  },
  8: {
    number: 8,
    label: "Herstellerkennung",
    category: "Decoder",
    description: "NMRA-Herstellerkennung des Decoders."
  },
  29: {
    number: 29,
    label: "Basis-Konfiguration",
    category: "Decoder",
    description: "Grundkonfiguration mit Fahrtrichtung, Fahrstufen, Analogbetrieb, Kennlinie und langer Adresse."
  }
};

function inferECoSCVCategory(cvNumber: number) {
  const definition = ecosStandardCVDefinitions[cvNumber];
  if (definition) return definition.category;
  if (cvNumber === 1 || cvNumber === 17 || cvNumber === 18) return "Adresse";
  if ([2, 3, 4, 5, 6, 29].includes(cvNumber)) return "Fahrverhalten";
  if (cvNumber >= 7 && cvNumber <= 8) return "Decoder";
  if (cvNumber >= 33 && cvNumber <= 46) return "Funktion";
  if (cvNumber >= 49 && cvNumber <= 64) return "Motor";
  return "Sonstiges";
}

function findECoSCVValue(locomotive: ECoSRawLocomotive, cvNumber: number) {
  return (locomotive.cvs || []).find((cv) => cv.number === cvNumber)?.value;
}

function getECoSDecoderManufacturer(locomotive: ECoSRawLocomotive, manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers) {
  const manufacturerId = findECoSCVValue(locomotive, 8);
  if (typeof manufacturerId !== "number") return undefined;
  return manufacturers[manufacturerId];
}

function buildECoSDecoderProfileHints(raw: ECoSRawProbe | null, manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers): ECoSDecoderProfileHint[] {
  if (!raw) return [];
  return raw.locomotives
    .map((locomotive) => {
      const manufacturerId = findECoSCVValue(locomotive, 8);
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

function interpretECoSCV29(value: number) {
  const flags = [
    value & 1 ? "Fahrtrichtung invertiert" : "Fahrtrichtung normal",
    value & 2 ? "28/128 Fahrstufen" : "14 Fahrstufen",
    value & 4 ? "Analogbetrieb erlaubt" : "nur Digitalbetrieb",
    value & 16 ? "freie Geschwindigkeitstabelle" : "3-Punkt-Kennlinie",
    value & 32 ? "lange Adresse aktiv" : "kurze Adresse aktiv"
  ];
  if (value & 8) {
    flags.push("Bit 3 herstellerabhaengig gesetzt");
  }
  return flags.join(" - ");
}

function interpretECoSCV(cvNumber: number, value: number, locomotive: ECoSRawLocomotive, manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers) {
  if (cvNumber === 8) {
    const manufacturer = getECoSDecoderManufacturer(locomotive, manufacturers);
    return manufacturer ? `${manufacturer.name} (${value})` : `Unbekannte Herstellerkennung (${value})`;
  }
  if (cvNumber === 29) {
    return interpretECoSCV29(value);
  }
  const definition = ecosStandardCVDefinitions[cvNumber];
  if (definition) {
    return definition.label;
  }
  return "-";
}

function cvPreviewKey(cvNumber: number, decoderProfile?: string) {
  return `${cvNumber}::${(decoderProfile || "").trim().toLocaleLowerCase("de-DE")}`;
}

function buildECoSCVImportRows(raw: ECoSRawProbe | null, vehicles: Vehicle[], manufacturers: ECoSDecoderManufacturerMap = ecosDecoderManufacturers): ECoSCVImportRow[] {
  if (!raw) return [];
  return raw.locomotives.flatMap((locomotive) => {
    const cvs = locomotive.cvs || [];
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

function renderECoSCVImportPreview(
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
                        ? `${hint.manufacturer.name} (CV8 ${hint.manufacturerId})`
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

function renderECoSRawProbe(
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

function mergeExternalMapping(vehicle: Vehicle, mapping: VehicleExternalMapping): Vehicle {
  const mappings = (vehicle.externalMappings || []).filter((item) => !(item.provider === mapping.provider && item.externalId === mapping.externalId));
  return { ...vehicle, externalMappings: [...mappings, mapping] };
}

function findECoSMatch(locomotive: ECoSMatchable, vehicles: Vehicle[]): ECoSMatch | null {
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

function buildECoSVehicleDraftRow(
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

function mergeImportedVehicle(existing: Vehicle, incoming: CreateVehicleRequest, importedKeys: (keyof CreateVehicleRequest)[]) {
  const merged = vehicleToRequest(existing);
  importedKeys.forEach((key) => {
    const value = incoming[key];
    if (typeof value === "boolean" || (typeof value === "string" && value.trim() !== "")) {
      (merged as Record<string, unknown>)[key] = value;
    }
  });
  return merged;
}

function displayImportValue(value: unknown, yes = "ja", no = "nein") {
  if (typeof value === "boolean") {
    return value ? yes : no;
  }
  if (typeof value === "string") {
    return value.trim() || "-";
  }
  return "-";
}

function valuesEqual(current: unknown, incoming: unknown) {
  if (typeof current === "boolean" || typeof incoming === "boolean") {
    return Boolean(current) === Boolean(incoming);
  }
  return String(current ?? "").trim() === String(incoming ?? "").trim();
}

function getImportChanges(
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

function vehiclesToCSV(vehicles: Vehicle[], fieldLabel: (key: VehicleImportField) => string, yes: string, no: string) {
  const headers = vehicleImportFields.map((field) => fieldLabel(field.key));
  const rows = vehicles.map((vehicle) => {
    const request = vehicleToRequest(vehicle);
    return vehicleImportFields.map((field) => displayImportValue(request[field.key], yes, no).replace(/^-$/, ""));
  });
  return [headers, ...rows].map((row) => row.map(csvEscape).join(";")).join("\n");
}

function downloadText(fileName: string, content: string, type: string) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  link.click();
  URL.revokeObjectURL(url);
}

function htmlEscape(value: unknown) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function printInventory(
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

export function ImportExportView() {
  const { language, t } = useI18n();
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [symbols, setSymbols] = useState<MasterDataEntry[]>([]);
  const [masterOptions, setMasterOptions] = useState<Record<string, MasterDataEntry[]>>({});
  const [rows, setRows] = useState<ImportRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [importTable, setImportTable] = useState<ImportTablePreview | null>(null);
  const [ecosHost, setEcosHost] = useState(window.localStorage.getItem("railkeeper.ecos.host") || "");
  const [ecosPort, setEcosPort] = useState(window.localStorage.getItem("railkeeper.ecos.port") || "15471");
  const [ecosBusy, setEcosBusy] = useState<ECoSBusyPhase>("idle");
  const [ecosResult, setEcosResult] = useState<ECoSConnectionResult | null>(null);
  const [ecosRawProbe, setEcosRawProbe] = useState<ECoSRawProbe | null>(null);
  const [ecosMessage, setEcosMessage] = useState("");
  const fieldLabel = (key: VehicleImportField) => t(`vehicle.field.${key}`);
  const issueLabels = {
    missingManufacturer: t("importExport.issue.missingManufacturer"),
    missingName: t("importExport.issue.missingName"),
    missingGauge: t("importExport.issue.missingGauge"),
    missingCategory: t("importExport.issue.missingCategory"),
    missingGattung: t("importExport.issue.missingGattung"),
    duplicate: t("importExport.issue.duplicate"),
    ecosMatched: t("importExport.ecos.matched")
  };
  const cv8Manufacturers = useMemo(
    () => cv8ManufacturersFromMasterData(masterOptions.cv8_manufacturer || []),
    [masterOptions]
  );

  useEffect(() => {
    api.vehicles().then(setVehicles).catch((error: Error) => setMessage(error.message)).finally(() => setLoading(false));
    api.masterDataAll(true)
      .then((entriesByType) => {
        setMasterOptions(entriesByType);
        setSymbols(entriesByType.symbols || []);
      })
      .catch(() => {
        setMasterOptions({});
        setSymbols([]);
      });
  }, []);

  const importSummary = useMemo(() => ({
    total: rows.length,
    selected: rows.filter((row) => row.selected && row.status !== "saved").length,
    errors: rows.filter((row) => row.status === "error").length,
    updates: rows.filter((row) => row.mode === "update" && row.status !== "saved").length,
    saved: rows.filter((row) => row.status === "saved").length
  }), [rows]);

  const mappingSummary = useMemo(() => {
    if (!importTable) {
      return { mapped: 0, unmapped: 0 };
    }
    const visibleMappings = importTable.mappings.filter((mapping) => mapping.header.trim());
    return {
      mapped: visibleMappings.filter((mapping) => mapping.key).length,
      unmapped: visibleMappings.filter((mapping) => !mapping.key).length
    };
  }, [importTable]);

  const loadImportTable = (table: string[][], fileName: string) => {
    if (table.length === 0) {
      setImportTable(null);
      setRows([]);
      setMessage(t("importExport.error.emptyFile"));
      return;
    }
    const mappings = defaultColumnMappings(table);
    const importedRows = importRowsFromTable(table, vehicles, mappings, issueLabels);
    const unmapped = mappings.filter((mapping) => !mapping.key && mapping.header.trim()).length;
    setImportTable({ fileName, table, mappings });
    setRows(importedRows);
    setMessage(unmapped > 0 ? t("importExport.message.unmapped", { count: unmapped }) : "");
  };

  const setColumnMapping = (columnIndex: number, key: VehicleImportField | "") => {
    if (!importTable) {
      return;
    }
    const mappings: ColumnMapping[] = importTable.mappings.map((mapping) => {
      if (mapping.index === columnIndex) {
        return { ...mapping, key };
      }
      return key && mapping.key === key ? { ...mapping, key: "" } : mapping;
    });
    setImportTable({ ...importTable, mappings });
    setRows(importRowsFromTable(importTable.table, vehicles, mappings, issueLabels));
  };

  const updateRow = (rowID: string, patch: Partial<ImportRow["vehicle"]>) => {
    setRows((current) => current.map((row) => {
      if (row.id !== rowID) {
        return row;
      }
      const vehicle = { ...row.vehicle, ...patch };
      const issues: string[] = [];
      const duplicate = vehicle.inventoryNumber ? vehicles.find((existing) => existing.inventoryNumber.toLowerCase() === vehicle.inventoryNumber?.toLowerCase()) : undefined;
      const importedKeys = Array.from(new Set([...row.importedKeys, ...Object.keys(patch) as (keyof CreateVehicleRequest)[]]));
      if (duplicate) {
        issues.push(t("importExport.issue.duplicate"));
        return {
          ...row,
          vehicle,
          importedKeys,
          duplicateVehicleId: duplicate.id,
          mode: "update",
          issues,
          status: "warning",
          selected: row.selected
        };
      }
      if (!vehicle.manufacturer) issues.push(t("importExport.issue.missingManufacturer"));
      if (!vehicle.name) issues.push(t("importExport.issue.missingName"));
      if (!vehicle.gauge) issues.push(t("importExport.issue.missingGauge"));
      if (!vehicle.category) issues.push(t("importExport.issue.missingCategory"));
      if (!vehicle.gattung) issues.push(t("importExport.issue.missingGattung"));
      return {
        ...row,
        vehicle,
        importedKeys,
        duplicateVehicleId: undefined,
        mode: "create",
        issues,
        status: issues.length ? "error" : "ok",
        selected: issues.length ? false : row.selected
      };
    }));
  };

  const setRowSelected = (rowID: string, selected: boolean) => {
    setRows((current) => current.map((item) => item.id === rowID ? { ...item, selected } : item));
  };

  const setRowMode = (rowID: string, mode: ImportRow["mode"]) => {
    setRows((current) => current.map((row) => {
      if (row.id !== rowID) return row;
      if (mode === "create" && row.duplicateVehicleId) {
        return { ...row, mode, selected: false, status: "error", issues: [t("importExport.issue.inventoryExists")] };
      }
      if (mode === "update" && row.duplicateVehicleId) {
        return { ...row, mode, status: "warning", issues: [t("importExport.issue.duplicate")] };
      }
      return { ...row, mode };
    }));
  };

  const handleFile = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    setMessage("");
    const extension = file.name.split(".").pop()?.toLowerCase() || "";
    const text = await file.text();
    if (extension === "json") {
      const parsed = JSON.parse(text) as Vehicle[] | { vehicles?: Vehicle[] };
      const source = Array.isArray(parsed) ? parsed : parsed.vehicles || [];
      const table = [
        [
          fieldLabel("inventoryNumber"),
          fieldLabel("manufacturer"),
          fieldLabel("articleNumber"),
          fieldLabel("name"),
          fieldLabel("gauge"),
          fieldLabel("epoch"),
          fieldLabel("railwayCompany"),
          fieldLabel("category"),
          fieldLabel("gattung"),
          fieldLabel("digital"),
          fieldLabel("digitalDecoderNumber"),
          fieldLabel("listPrice")
        ],
        ...source.map((vehicle) => [vehicle.inventoryNumber, vehicle.manufacturer, vehicle.articleNumber || "", vehicle.name, vehicle.gauge, vehicle.epoch || "", vehicle.railwayCompany || "", vehicle.category || "", vehicle.gattung || "", vehicle.digital ? t("common.yes") : t("common.no"), vehicle.digitalDecoderNumber || "", vehicle.listPrice || ""])
      ];
      loadImportTable(table, file.name);
      return;
    }
    if (extension === "xml") {
      const table = parseXMLImport(text);
      if (table.length) {
        loadImportTable(table, file.name);
      } else {
        setRows([]);
        setImportTable(null);
        setMessage(t("importExport.error.emptyTable"));
      }
      return;
    }

    const delimiter = extension === "tsv" ? "\t" : detectDelimiter(text);
    loadImportTable(parseDelimited(text, delimiter), file.name);
  };

  const saveFunctionSuggestions = async (vehicleID: string, suggestions: FunctionImportSuggestion[] = []) => {
    for (const suggestion of suggestions) {
      await api.updateVehicleFunction(vehicleID, suggestion.functionKey, {
        name: suggestion.name || "",
        symbolKey: suggestion.symbolKey || "",
        functionType: suggestion.functionType || "standard",
        mode: suggestion.mode || "dauer",
        directionDependent: Boolean(suggestion.directionDependent),
        notes: suggestion.notes || ""
      });
    }
  };

  const saveCVSuggestions = async (vehicleID: string, suggestions: VehicleCVValueInput[] = []) => {
    if (suggestions.length === 0) {
      return;
    }
    const vehicle = await api.vehicle(vehicleID);
    for (const suggestion of suggestions) {
      const existing = (vehicle.cvValues || []).find((value) =>
        value.cvNumber === suggestion.cvNumber &&
        (value.decoderProfile || "") === (suggestion.decoderProfile || "")
      );
      if (existing) {
        await api.updateVehicleCVValue(vehicleID, existing.id, suggestion);
      } else {
        await api.createVehicleCVValue(vehicleID, suggestion);
      }
    }
  };

  const saveSelected = async () => {
    setSaving(true);
    setMessage("");
    for (const row of rows) {
      if (!row.selected || row.status === "saved" || row.status === "error") {
        continue;
      }
      try {
        const existing = row.duplicateVehicleId ? vehicles.find((vehicle) => vehicle.id === row.duplicateVehicleId) : undefined;
        let saved = row.mode === "update" && existing
          ? await api.updateVehicle(existing.id, mergeImportedVehicle(existing, row.vehicle, row.importedKeys))
          : await api.createVehicle(row.vehicle);
        if (row.externalMapping) {
          const mapping = await api.upsertVehicleExternalMapping(saved.id, row.externalMapping);
          saved = mergeExternalMapping(saved, mapping);
        }
        await saveFunctionSuggestions(saved.id, row.functionSuggestions);
        await saveCVSuggestions(saved.id, row.cvSuggestions);
        setVehicles((current) => {
          if (row.mode === "update") {
            return current.map((vehicle) => vehicle.id === saved.id ? saved : vehicle);
          }
          return [...current, saved];
        });
        setRows((current) => current.map((item) => item.id === row.id ? { ...item, selected: false, status: "saved", issues: [] } : item));
      } catch (error) {
        const message = error instanceof Error ? error.message : t("importExport.error.importFailed");
        setRows((current) => current.map((item) => item.id === row.id ? { ...item, status: "error", issues: [message] } : item));
      }
    }
    setSaving(false);
  };

  const ecosInput = () => ({
    host: ecosHost.trim(),
    port: Number(ecosPort) || 15471
  });

  const rememberECoSSettings = () => {
    window.localStorage.setItem("railkeeper.ecos.host", ecosHost.trim());
    window.localStorage.setItem("railkeeper.ecos.port", ecosPort.trim() || "15471");
  };

  const testECoSConnection = async () => {
    setEcosBusy("connecting");
    setEcosMessage("");
    setEcosResult(null);
    setEcosRawProbe(null);
    try {
      rememberECoSSettings();
      const result = await api.testECoSConnection(ecosInput());
      setEcosResult(result);
      setEcosMessage(result.message);
    } catch (error) {
      setEcosMessage(error instanceof Error ? error.message : t("importExport.ecos.error"));
    } finally {
      setEcosBusy("idle");
    }
  };

  const probeECoSRawData = async () => {
    if (!ecosResult?.connected) {
      setEcosMessage(t("importExport.ecos.connectFirst"));
      return;
    }
    setEcosBusy("fetching");
    setEcosMessage("");
    setEcosRawProbe(null);
    try {
      rememberECoSSettings();
      const probe = await api.probeECoSLocomotiveRaw(ecosInput());
      setEcosRawProbe(probe);
      setEcosMessage(probe.message);
    } catch (error) {
      setEcosMessage(error instanceof Error ? error.message : t("importExport.ecos.error"));
    } finally {
      setEcosBusy("idle");
    }
  };

  const openECoSVehicleDraft = (locomotive: ECoSRawLocomotive) => {
    const row = buildECoSVehicleDraftRow(locomotive, vehicles, symbols, {
      matched: issueLabels.ecosMatched,
      missingManufacturer: issueLabels.missingManufacturer,
      missingName: issueLabels.missingName,
      missingGauge: issueLabels.missingGauge,
      missingCategory: issueLabels.missingCategory,
      missingGattung: issueLabels.missingGattung
    }, cv8Manufacturers);
    const unclearFields = ecosRequiredFields.filter((field) => !String(row.vehicle[field] ?? "").trim());
    const payload: ECoSVehicleDraftPayload = {
      source: "ecos",
      sourceSummary: {
        objectId: locomotive.objectId,
        name: locomotive.name || `ECoS ${locomotive.objectId}`,
        address: typeof locomotive.address === "number" ? String(locomotive.address) : "",
        protocol: normalizeECoSProtocolForCV(locomotive.protocol),
        profile: locomotive.profile || ""
      },
      vehicle: row.vehicle,
      externalMapping: row.externalMapping || ecosExternalMapping(locomotive),
      cvValues: row.cvSuggestions || [],
      unclearFields
    };
    window.sessionStorage.setItem(ecosVehicleDraftStorageKey, JSON.stringify(payload));
    window.history.pushState(null, "", "/vehicles?source=ecos");
    window.dispatchEvent(new PopStateEvent("popstate"));
    setEcosMessage(t("importExport.ecos.handoff"));
  };

  return (
    <>
      <section className="page-head">
        <p className="eyebrow">{t("importExport.eyebrow")}</p>
        <h1>{t("importExport.title")}</h1>
        <p>{t("importExport.subtitle")}</p>
      </section>

      {message && <p className="form-message">{message}</p>}

      <section className="import-export-grid">
        <article className="panel transfer-panel">
          <div className="panel-head">
            <div>
              <h2 className="panel-title-inline"><FileInput size={20} aria-hidden="true" />{t("importExport.import.title")}</h2>
              <p>{t("importExport.import.subtitle")}</p>
            </div>
          </div>
          <label className="file-drop compact-drop">
            <Upload size={18} aria-hidden="true" />
            {t("importExport.file.choose")}
            <input type="file" accept=".csv,.tsv,.xml,.json" onChange={handleFile} />
          </label>
          <div className="import-summary">
            <span>{t("importExport.summary.rows", { count: importSummary.total })}</span>
            <span>{previewLoading ? t("importExport.summary.reading") : t("importExport.summary.ready", { count: importSummary.selected })}</span>
            <span>{t("importExport.summary.updates", { count: importSummary.updates })}</span>
            <span className={importSummary.errors ? "danger" : ""}>{t("importExport.summary.notes", { count: importSummary.errors })}</span>
            <span>{t("importExport.summary.saved", { count: importSummary.saved })}</span>
            {importTable && <span>{t("importExport.summary.mapped", { count: mappingSummary.mapped })}</span>}
            {importTable && <span className={mappingSummary.unmapped ? "danger" : ""}>{t("importExport.summary.open", { count: mappingSummary.unmapped })}</span>}
          </div>
        </article>

        <article className="panel transfer-panel">
          <div className="panel-head">
            <div>
              <h2 className="panel-title-inline"><Download size={20} aria-hidden="true" />{t("importExport.export.title")}</h2>
              <p>{t("importExport.export.subtitle")}</p>
            </div>
          </div>
          <div className="export-actions">
            <button type="button" className="secondary-button" disabled={loading || vehicles.length === 0} onClick={() => downloadText("railkeeper-bestand.csv", `\uFEFF${vehiclesToCSV(vehicles, fieldLabel, t("common.yes"), t("common.no"))}`, "text/csv;charset=utf-8")}>
              <Download size={15} aria-hidden="true" />
              {t("importExport.export.csv")}
            </button>
            <button type="button" className="secondary-button" disabled={loading || vehicles.length === 0} onClick={() => downloadText("railkeeper-bestand.json", JSON.stringify({ format: "railkeeper-vehicles", version: 1, vehicles }, null, 2), "application/json;charset=utf-8")}>
              <Download size={15} aria-hidden="true" />
              {t("importExport.export.json")}
            </button>
            <button type="button" className="secondary-button" disabled={loading || vehicles.length === 0} onClick={() => printInventory(vehicles, fieldLabel, language, t)}>
              <Printer size={15} aria-hidden="true" />
              {t("importExport.export.print")}
            </button>
          </div>
        </article>
      </section>

      <section className="panel transfer-panel ecos-panel">
        <div className="panel-head">
          <div>
            <h2 className="panel-title-inline"><Database size={20} aria-hidden="true" />{t("importExport.ecos.title")}</h2>
            <p>{t("importExport.ecos.subtitle")}</p>
          </div>
        </div>
        <div className="ecos-connection-grid">
          <label>
            {t("importExport.ecos.host")}
            <input value={ecosHost} onChange={(event) => setEcosHost(event.target.value)} placeholder={t("importExport.ecos.hostPlaceholder")} />
          </label>
          <label>
            {t("importExport.ecos.port")}
            <input value={ecosPort} onChange={(event) => setEcosPort(event.target.value)} inputMode="numeric" placeholder="15471" />
          </label>
          <div className="ecos-actions">
            <button type="button" className="secondary-button" onClick={testECoSConnection} disabled={ecosBusy !== "idle" || !ecosHost.trim()}>
              <Check size={15} aria-hidden="true" />
              {t("importExport.ecos.connect")}
            </button>
            {ecosResult?.connected && (
              <button type="button" className="primary-button" onClick={probeECoSRawData} disabled={ecosBusy !== "idle" || !ecosHost.trim()}>
                <Download size={15} aria-hidden="true" />
                {t("importExport.ecos.fetchData")}
              </button>
            )}
          </div>
        </div>
        <div className="ecos-status-strip">
          <span className={ecosResult?.connected ? "status-ok" : ecosResult ? "status-error" : ""}>
            {ecosResult ? (ecosResult.connected ? t("importExport.ecos.connected") : t("importExport.ecos.notConnected")) : t("importExport.ecos.idle")}
          </span>
          {ecosResult?.status && <span>{t("importExport.ecos.status", { status: ecosResult.status })}</span>}
          {ecosResult?.protocolVersion && <span>{t("importExport.ecos.protocol", { version: ecosResult.protocolVersion })}</span>}
          {ecosResult?.applicationVersion && <span>{t("importExport.ecos.application", { version: ecosResult.applicationVersion })}</span>}
        </div>
        {ecosBusy !== "idle" && (
          <p className="ecos-busy-indicator">
            {ecosBusy === "connecting" ? t("importExport.ecos.connecting") : t("importExport.ecos.fetching")}
          </p>
        )}
        {ecosMessage && <p className="form-message">{ecosMessage}</p>}
        <p className="source-note backup-note">{t("importExport.ecos.note")}</p>
        <p className="source-note backup-note">{t("importExport.ecos.mappingNote")}</p>
        {renderECoSRawProbe(ecosRawProbe, vehicles, t, cv8Manufacturers, openECoSVehicleDraft)}
      </section>


      {importTable && (
        <section className="panel column-mapping-panel">
          <div className="panel-head">
            <div>
              <h2>{t("importExport.mapping.title")}</h2>
              <p>{t("importExport.mapping.subtitle", { file: importTable.fileName })}</p>
            </div>
            <Database size={20} aria-hidden="true" />
          </div>
          <div className="column-mapping-grid">
            {importTable.mappings.map((mapping) => (
              <label key={mapping.index} className={mapping.key ? "" : "unmapped"}>
                <span>
                  <strong title={mapping.header}>{mapping.header || t("importExport.mapping.column", { number: mapping.index + 1 })}</strong>
                  <small>{mapping.key ? t("importExport.mapping.mapped") : t("importExport.mapping.unmapped")}</small>
                </span>
                <select value={mapping.key} onChange={(event) => setColumnMapping(mapping.index, event.target.value as VehicleImportField | "")}>
                  <option value="">{t("importExport.mapping.ignore")}</option>
                  {vehicleImportFields.map((field) => (
                    <option key={field.key} value={field.key}>{fieldLabel(field.key)}</option>
                  ))}
                </select>
              </label>
            ))}
          </div>
          <p className="source-note backup-note">{t("importExport.mapping.note")}</p>
        </section>
      )}

      {rows.length > 0 && (
      <section className="panel import-review-panel">
        <div className="panel-head">
          <div>
            <h2 className="panel-title-inline"><ClipboardCheck size={20} aria-hidden="true" />{t("importExport.review.title")}</h2>
            <p>{t("importExport.review.subtitle")}</p>
          </div>
          <button type="button" className="primary-button" disabled={saving || importSummary.selected === 0} onClick={saveSelected}>
            <Save size={15} aria-hidden="true" />
            {t("importExport.review.saveSelection")}
          </button>
        </div>

        {rows.length === 0 ? (
          <p className="empty-state">{t("importExport.review.empty")}</p>
        ) : (
          <div className="table-wrap import-table">
            <table>
              <thead>
                <tr>
                  <th>{t("importExport.review.apply")}</th>
                  <th>{t("importExport.review.action")}</th>
                  <th>{t("importExport.review.inventory")}</th>
                  <th>{fieldLabel("manufacturer")}</th>
                  <th>{t("importExport.review.article")}</th>
                  <th>{fieldLabel("name")}</th>
                  <th>{fieldLabel("gauge")}</th>
                  <th>{t("exhibition.status")}</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((row) => {
                  const existing = row.duplicateVehicleId ? vehicles.find((vehicle) => vehicle.id === row.duplicateVehicleId) : undefined;
                  const changes = getImportChanges(row, existing, fieldLabel, t("common.yes"), t("common.no"));
                  return (
                    <Fragment key={row.id}>
                      <tr className={row.status === "error" ? "import-row-error" : row.status === "warning" ? "import-row-warning" : row.status === "saved" ? "import-row-saved" : ""}>
                        <td><input type="checkbox" checked={row.selected} disabled={row.status === "saved" || row.status === "error"} onChange={(event) => setRowSelected(row.id, event.target.checked)} /></td>
                        <td>
                          <select value={row.mode} disabled={row.status === "saved"} onChange={(event) => setRowMode(row.id, event.target.value as ImportRow["mode"])}>
                            <option value="create">{t("importExport.review.create")}</option>
                            <option value="update" disabled={!row.duplicateVehicleId}>{t("importExport.review.update")}</option>
                          </select>
                        </td>
                        <td><input value={row.vehicle.inventoryNumber || ""} onChange={(event) => updateRow(row.id, { inventoryNumber: event.target.value })} /></td>
                        <td><input value={row.vehicle.manufacturer} onChange={(event) => updateRow(row.id, { manufacturer: event.target.value })} /></td>
                        <td><input value={row.vehicle.articleNumber || ""} onChange={(event) => updateRow(row.id, { articleNumber: event.target.value })} /></td>
                        <td><input value={row.vehicle.name} onChange={(event) => updateRow(row.id, { name: event.target.value })} /></td>
                        <td><input value={row.vehicle.gauge} onChange={(event) => updateRow(row.id, { gauge: event.target.value })} /></td>
                        <td>
                          <span className={`import-status ${row.status}`}>
                            {row.status === "saved" ? <Check size={14} /> : row.status === "error" || row.status === "warning" ? <AlertTriangle size={14} /> : <Check size={14} />}
                            {row.status === "saved" ? t("common.saved") : row.issues[0] || t("common.ready")}
                          </span>
                        </td>
                      </tr>
                      {((existing && row.mode === "update") || (row.functionSuggestions && row.functionSuggestions.length > 0) || (row.cvSuggestions && row.cvSuggestions.length > 0)) && (
                        <tr className="import-change-row">
                          <td colSpan={8}>
                            <div className="import-change-panel">
                              {existing && row.mode === "update" && (
                                <>
                                  <div>
                                    <strong>{t("importExport.review.updatePreview")}</strong>
                                    <span>{t("importExport.review.overwrites", { count: changes.filter((change) => change.status === "overwrite").length })}, {t("importExport.review.fills", { count: changes.filter((change) => change.status === "fill").length })}</span>
                                  </div>
                                  {changes.length === 0 ? (
                                    <p>{t("importExport.review.noValues")}</p>
                                  ) : (
                                    <table>
                                      <thead>
                                        <tr>
                                          <th>{t("importExport.review.field")}</th>
                                          <th>{t("importExport.review.current")}</th>
                                          <th>{t("importExport.review.import")}</th>
                                          <th>{t("exhibition.status")}</th>
                                        </tr>
                                      </thead>
                                      <tbody>
                                        {changes.map((change) => (
                                          <tr key={change.key} className={`change-${change.status}`}>
                                            <td>{change.label}</td>
                                            <td>{change.current}</td>
                                            <td>{change.incoming}</td>
                                            <td>{t(`importExport.review.status.${change.status}`)}</td>
                                          </tr>
                                        ))}
                                      </tbody>
                                    </table>
                                  )}
                                </>
                              )}
                              {row.functionSuggestions && row.functionSuggestions.length > 0 && (
                                <div className="import-function-panel">
                                  <div>
                                    <strong>ECoS-Funktionstasten</strong>
                                    <span>{row.functionSuggestions.length} Werte werden übernommen</span>
                                  </div>
                                  <div className="import-function-grid">
                                    {row.functionSuggestions.map((fn) => (
                                      <span key={fn.functionKey} title={fn.notes || undefined}>
                                        <strong>{fn.functionKey}</strong>
                                        {fn.name || (typeof fn.ecosDescription === "number" ? `ECoS ${fn.ecosDescription}` : "ECoS")}
                                        {fn.active && <em>aktiv</em>}
                                      </span>
                                    ))}
                                  </div>
                                </div>
                              )}
                              {row.cvSuggestions && row.cvSuggestions.length > 0 && (
                                <div className="import-function-panel">
                                  <div>
                                    <strong>ECoS-CV-Werte</strong>
                                    <span>{row.cvSuggestions.length} Werte werden übernommen</span>
                                  </div>
                                  <div className="import-function-grid">
                                    {row.cvSuggestions.slice(0, 16).map((cv) => (
                                      <span key={`${cv.cvNumber}-${cv.decoderProfile || ""}`} title={cv.description || undefined}>
                                        <strong>CV {cv.cvNumber}</strong>
                                        {cv.value}
                                      </span>
                                    ))}
                                    {row.cvSuggestions.length > 16 && <span>{t("importExport.ecos.draft.moreCVs", { count: row.cvSuggestions.length - 16 })}</span>}
                                  </div>
                                </div>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
      )}
    </>
  );
}
