import { describe, expect, it } from "vitest";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { buildCVImportPreview } from "./cvImport";
import { guardedCVImportValues, recompareCVImport } from "./cvImportGuard";

describe("confirmed CV import comparison", () => {
  it("pins observed revisions and absence while ignoring comparison tokens supplied by a file", () => {
    const before = cvValueFixture({ decoderProfile: "A" });
    const preview = buildCVImportPreview("source.json", [
      { cvNumber: 1, value: 4, decoderProfile: "A", expectedRevision: "untrusted", sourceLine: 7 },
      { cvNumber: 2, value: 5, decoderProfile: "A", expectedRevision: "untrusted" },
      { cvNumber: 3, value: 9, decoderProfile: "A" }
    ], [before]);
    preview.rows[2].selected = false;
    expect(guardedCVImportValues(preview)).toEqual([
      expect.objectContaining({ cvNumber: 1, expectedRevision: before.revision, sourceLine: 7 }),
      expect.objectContaining({ cvNumber: 2, expectedRevision: "" })
    ]);
    expect(preview.rows[0].input.expectedRevision).toBe("untrusted");
    expect(() => guardedCVImportValues({ ...preview, rows: [{ ...preview.rows[0], existing: { ...before, revision: undefined } }] }))
      .toThrow("missingRevision");
  });
  it("retains only the original selection on explicit comparison and deselects already identical results", () => {
    const preview = { ...buildCVImportPreview("source.csv", [
      { cvNumber: 1, value: 4, decoderProfile: "A" }, { cvNumber: 2, value: 5, decoderProfile: "A" },
      { cvNumber: 3, value: 6, decoderProfile: "A" }
    ], []), vehicleId: "vehicle-1" };
    preview.rows[2].selected = false;
    const refreshed = recompareCVImport(preview, [
      cvValueFixture({ cvNumber: 1, value: 4, decoderProfile: "A", description: "" }),
      cvValueFixture({ id: "other", cvNumber: 2, value: 7, decoderProfile: "A", description: "" })
    ]);
    expect(refreshed.rows.map(row => row.selected)).toEqual([false, true, false]);
    expect(refreshed.vehicleId).toBe(preview.vehicleId);
    expect(refreshed.rows.map(row => row.input)).toEqual(preview.rows.map(row => row.input));
  });
});
