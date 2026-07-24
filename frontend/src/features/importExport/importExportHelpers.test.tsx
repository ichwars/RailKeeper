import { describe, expect, it } from "vitest";

import { analogVehicleFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import {
  defaultColumnMappings,
  detectDelimiter,
  findECoSMatch,
  getImportChanges,
  importRowsFromTable,
  mergeImportedVehicle,
  normalizeHeader,
  parseBoolean,
  parseDelimited,
  parseXMLImport,
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
      [vehicleFixture({ description: "Diesel; DR" })],
      (key) => key,
      "ja",
      "nein"
    );
    expect(csv).toContain("inventoryNumber;manufacturer");
    expect(csv).toContain('"Diesel; DR"');
  });
});
