import { describe, expect, it } from "vitest";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { isValidCVValueInput } from "./cvImport";
import {
  cvValuesFromImport, maxCVImportValues, maxCVJSONImportBytes, maxCVTextImportBytes, readCVImportFile
} from "./cvFileImport";

describe("CV import file limits", () => {
  it("reimports a full-size pretty-printed RailKeeper export with record metadata", async () => {
    const rows = Array.from({ length: maxCVImportValues }, (_, index) => cvValueFixture({
      id: `00000000-0000-4000-8000-${String(index).padStart(12, "0")}`,
      vehicleId: "11111111-1111-4111-8111-111111111111", cvNumber: 257 + index % 256,
      indexCV31: 16, indexCV32: Math.floor(index / 256), value: index % 256,
      sourceLine: index + 4, sourceExportedAt: "2026-10-17T17:41:59"
    }));
    const file = new File([JSON.stringify({ cvValues: rows }, null, 2)], "vehicle-cv.json");
    expect(file.size).toBeGreaterThan(maxCVTextImportBytes);
    expect(file.size).toBeLessThan(maxCVJSONImportBytes);
    const imported = await readCVImportFile(file);
    expect(imported).toHaveLength(maxCVImportValues);
    expect(imported.every(isValidCVValueInput)).toBe(true);
    expect(imported[9999]).toMatchObject({ cvNumber: 272, indexCV31: 16, indexCV32: 39,
      value: 15, sourceLine: 10003, sourceExportedAt: "2026-10-17T17:41:59", decoderProfile: "ESU LokPilot 5" });
    expect(new Blob([JSON.stringify({ values: imported })]).size).toBeLessThan(8 * 1024 * 1024);
  });

  it.each([
    ["large.json", maxCVJSONImportBytes, /32 MiB/],
    ["large.txt", maxCVTextImportBytes, /2 MiB/],
    ["large.csv", maxCVTextImportBytes, /2 MiB/]
  ] as const)("rejects an oversized %s before reading it", async (name, limit, message) => {
    const file = new File(["{}"], name);
    Object.defineProperty(file, "size", { value: limit + 1 });
    await expect(readCVImportFile(file)).rejects.toThrow(message);
  });

  it("rejects excessive row counts before converting individual records", () => {
    expect(() => cvValuesFromImport(JSON.stringify({ cvValues: Array(maxCVImportValues + 1).fill(null) })))
      .toThrow(/10[.,]000/);
    expect(() => cvValuesFromImport("CV;Wert\n" + "1;3\n".repeat(maxCVImportValues + 1)))
      .toThrow(/10[.,]000/);
  });
});
