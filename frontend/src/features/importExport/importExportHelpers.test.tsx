import { describe, expect, it } from "vitest";

import { analogVehicleFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import {
  defaultColumnMappings,
  detectDelimiter,
  ecosFunctionSuggestions,
  findECoSMatch,
  getImportChanges,
  importRowsFromTable,
  mergeImportedVehicle,
  normalizeHeader,
  parseBoolean,
  parseDelimited,
  parseXMLImport,
  rawECoSFunctions,
  rawECoSUnknownAttributes,
  vehicleImportFields,
  vehiclesToCSV
} from "./importExportHelpers";

describe("import/export helpers", () => {
  it("normalizes headers and booleans", () => {
    expect(normalizeHeader(" Fahrzeug-Nr. ")).toBe("fahrzeug-nr");
    expect(normalizeHeader("Spurweite (mm)")).toBe("spurweite-mm");
    expect(parseBoolean("Vorhanden")).toBe(true);
    expect(parseBoolean("nein")).toBe(false);
  });

  it("detects delimiters and parses quoted rows", () => {
    const text = 'Hersteller;Bezeichnung;Beschreibung\nESU;BR 106;"Diesel; DR"';
    expect(detectDelimiter(text)).toBe(";");
    expect(parseDelimited(text, ";")).toEqual([
      ["Hersteller", "Bezeichnung", "Beschreibung"],
      ["ESU", "BR 106", "Diesel; DR"]
    ]);
  });

  it("parses XML records into a table", () => {
    const table = parseXMLImport(
      "<vehicles><vehicle><manufacturer>ESU</manufacturer><name>BR 106</name></vehicle></vehicles>"
    );
    expect(table[0]).toEqual(["manufacturer", "name"]);
    expect(table[1]).toEqual(["ESU", "BR 106"]);
  });

  it("maps columns and flags existing inventory numbers", () => {
    const table = [
      ["Inventarnummer", "Hersteller", "Bezeichnung", "Spurweite", "Kategorie", "Gattung", "Digital"],
      ["RK-LOK-000001", "ESU", "BR 106", "H0", "Lokomotive", "Diesellokomotive", "ja"],
      ["RK-LOK-000003", "Roco", "BR 80", "H0", "Lokomotive", "Dampflok", "nein"]
    ];
    const mappings = defaultColumnMappings(table);
    const rows = importRowsFromTable(table, [vehicleFixture()], mappings);

    expect(rows[0]).toMatchObject({ mode: "update", status: "warning", selected: false });
    expect(rows[1]).toMatchObject({ mode: "create", status: "ok", selected: true });
    expect(rows[1].vehicle.digital).toBe(false);
  });

  it("matches ECoS locomotives by mapping, decoder and normalized name", () => {
    const mapped = vehicleFixture({
      externalMappings: [{
        id: "mapping-1",
        vehicleId: "vehicle-1",
        provider: "ecos",
        externalId: "77",
        syncStatus: "linked",
        createdAt: "2026-07-23T08:00:00Z",
        updatedAt: "2026-07-23T08:00:00Z"
      }]
    });
    expect(findECoSMatch({ objectId: 77, name: "Unknown", address: 999 }, [mapped])?.source).toBe("mapping");
    expect(findECoSMatch({ objectId: 12, name: "Unknown", address: 1001 }, [vehicleFixture()])?.source).toBe("decoder");
    expect(findECoSMatch({ objectId: 13, name: "BR106", address: 0 }, [vehicleFixture()])?.source).toBe("name");
  });

  it("creates ECoS function suggestions only from static descriptions", () => {
    const suggestions = ecosFunctionSuggestions({
      objectId: 1001,
      functions: [{ index: 0, description: 3 }]
    }, []);

    expect(suggestions).toHaveLength(1);
    expect(suggestions[0]).toMatchObject({ functionKey: "F0" });
    expect(suggestions[0]).not.toHaveProperty("active");
    expect(suggestions[0].notes).toBe("Aus ECoS funcdesc 3.");
  });

  it("keeps runtime and image attributes out of raw ECoS review", () => {
    const locomotive = {
      objectId: 1001,
      functions: [{ index: 2, description: 6 }],
      attributes: {
        speed: ["22"],
        dir: ["1"],
        funcset: ["101"],
        image: ["12"],
        custom: ["retained"]
      }
    };

    expect(rawECoSFunctions(locomotive)).toEqual([{ index: 2, description: 6 }]);
    expect(rawECoSUnknownAttributes(locomotive)).toEqual([["custom", ["retained"]]]);
  });

  it("merges selected fields and reports import changes", () => {
    const existing = analogVehicleFixture({ manufacturer: "Roco", name: "BR 86" });
    const merged = mergeImportedVehicle(existing, { manufacturer: "Fleischmann", name: "", gauge: "H0" }, ["manufacturer", "name"]);
    expect(merged.manufacturer).toBe("Fleischmann");
    expect(merged.name).toBe("BR 86");

    const row = importRowsFromTable(
      [["Inventarnummer", "Hersteller"], [existing.inventoryNumber, "Fleischmann"]],
      [existing]
    )[0];
    const changes = getImportChanges(row, existing, (key) => key, "ja", "nein");
    expect(changes).toContainEqual({
      key: "manufacturer",
      label: "manufacturer",
      current: "Roco",
      incoming: "Fleischmann",
      status: "overwrite"
    });
  });

  it("exports escaped CSV values", () => {
    const csv = vehiclesToCSV(
      [vehicleFixture({
        description: "Diesel; DR",
        maximumSpeedKmh: 120,
        homeBase: "Bw Leipzig-West"
      })],
      (key) => key,
      "ja",
      "nein"
    );
    expect(csv).toContain("inventoryNumber;manufacturer");
    expect(csv).toContain('"Diesel; DR"');
    expect(csv).toContain("maximumSpeedKmh;homeBase");
    expect(csv).toContain("120;Bw Leipzig-West");
  });

  it("round-trips every scalar model field added to the CSV exchange", () => {
    const csvLabels: Record<string, string> = {
      decoderType: "Decoder-Typ",
      acquisitionType: "Erwerbsart",
      acquiredFrom: "Erworben von/bei",
      purchasePrice: "Kaufpreis",
      purchaseDate: "Kaufdatum",
      storageLocation: "Lagerort",
      storageDetails: "Lagerdetails",
      condition: "Zustand",
      conditionDetails: "Zustandsdetails",
      packaging: "Verpackung"
    };
    const source = vehicleFixture({
      decoderType: "ESU LokPilot 5",
      acquisitionType: "Gebrauchtkauf",
      acquiredFrom: "Modellbahn Börse",
      purchasePrice: "149.90",
      purchaseDate: "2026-08-15",
      storageLocation: "Vitrine",
      storageDetails: "Fach 3",
      condition: "Sehr gut",
      conditionDetails: "Leichte Laufspuren",
      packaging: "Originalverpackung"
    });

    const csv = vehiclesToCSV([source], (key) => csvLabels[key] || key, "ja", "nein");
    const imported = importRowsFromTable(parseDelimited(csv, ";"), []);

    expect(vehicleImportFields.map((field) => field.key)).toEqual(expect.arrayContaining(Object.keys(csvLabels)));
    expect(imported[0].vehicle).toMatchObject({
      decoderType: source.decoderType,
      acquisitionType: source.acquisitionType,
      acquiredFrom: source.acquiredFrom,
      purchasePrice: source.purchasePrice,
      purchaseDate: source.purchaseDate,
      storageLocation: source.storageLocation,
      storageDetails: source.storageDetails,
      condition: source.condition,
      conditionDetails: source.conditionDetails,
      packaging: source.packaging
    });
    expect(csv).not.toContain("images");
  });

  it("recognizes common English headers for the complete scalar model data", () => {
    const headers = [
      "Decoder type",
      "Acquisition type",
      "Acquired from",
      "Purchase price",
      "Purchase date",
      "Storage location",
      "Storage details",
      "Condition",
      "Condition details",
      "Packaging"
    ];

    expect(defaultColumnMappings([headers]).map((mapping) => mapping.key)).toEqual([
      "decoderType",
      "acquisitionType",
      "acquiredFrom",
      "purchasePrice",
      "purchaseDate",
      "storageLocation",
      "storageDetails",
      "condition",
      "conditionDetails",
      "packaging"
    ]);
  });

  it("imports and validates the optional maximum speed", () => {
    const valid = importRowsFromTable([
      ["Hersteller", "Bezeichnung", "Spurweite", "Kategorie", "Gattung", "Höchstgeschwindigkeit", "Heimat-Bw"],
      ["Piko", "BR 118", "H0", "Lokomotive", "Diesellok", "120", "Bw Leipzig-West"]
    ], []);
    expect(valid[0].vehicle).toMatchObject({ maximumSpeedKmh: 120, homeBase: "Bw Leipzig-West" });
    expect(valid[0].issues).toEqual([]);

    const invalid = importRowsFromTable([
      ["Hersteller", "Bezeichnung", "Spurweite", "Kategorie", "Gattung", "Höchstgeschwindigkeit"],
      ["Piko", "BR 118", "H0", "Lokomotive", "Diesellok", "1001"]
    ], []);
    expect(invalid[0].issues).toContain("Höchstgeschwindigkeit muss zwischen 1 und 1000 km/h liegen");
  });

  it("keeps duplicate rows with an invalid maximum speed blocked", () => {
    const existing = vehicleFixture();
    const rows = importRowsFromTable([
      ["Inventarnummer", "Höchstgeschwindigkeit"],
      [existing.inventoryNumber, "1001"]
    ], [existing]);

    expect(rows[0]).toMatchObject({ mode: "update", status: "error", selected: false });
    expect(rows[0].issues).toContain("Höchstgeschwindigkeit muss zwischen 1 und 1000 km/h liegen");
  });

  it("rejects a home base longer than 200 characters before saving", () => {
    const rows = importRowsFromTable([
      ["Hersteller", "Bezeichnung", "Spurweite", "Kategorie", "Gattung", "Heimat-Bw"],
      ["Piko", "BR 118", "H0", "Lokomotive", "Diesellok", "B".repeat(201)]
    ], []);

    expect(rows[0]).toMatchObject({ status: "error", selected: false });
    expect(rows[0].issues).toContain("Heimat-Bw / Einsatzstelle darf höchstens 200 Zeichen enthalten");
  });
});
