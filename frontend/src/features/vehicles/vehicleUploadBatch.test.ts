import { describe, expect, it } from "vitest";
import { ApiError } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { createVehicleUploadBatch, maxVehicleUploadFiles, retryableUploadItem, vehicleUploadCounts,
  vehicleUploadFailure } from "./vehicleUploadBatch";

describe("confirmed vehicle upload batches", () => {
  it("captures the target, files and metadata without following later form changes", () => {
    const vehicle = vehicleFixture(), file = new File(["manual"], "manual.txt");
    const metadata = { description: "Confirmed note" };
    const batch = createVehicleUploadBatch("attachment", vehicle, [file], metadata);
    metadata.description = "Later typing"; vehicle.name = "Another view";
    expect(batch.vehicleLabel).toContain("BR 106");
    expect(batch.metadata.description).toBe("Confirmed note");
    expect(batch.items[0].file).toBe(file);
    expect(batch.phase).toBe("confirm");
  });

  it("bounds one confirmation and distinguishes a newly started batch", () => {
    const file = new File(["x"], "file.txt"), vehicle = vehicleFixture();
    expect(() => createVehicleUploadBatch("attachment", vehicle, [])).toThrow(RangeError);
    expect(() => createVehicleUploadBatch("attachment", vehicle, Array.from({ length: maxVehicleUploadFiles + 1 }, () => file))).toThrow(RangeError);
    const first = createVehicleUploadBatch("attachment", vehicle, [file]);
    expect(createVehicleUploadBatch("attachment", vehicle, [file]).id).not.toBe(first.id);
  });

  it("never retries acknowledged, ambiguous or explicitly skipped files", () => {
    const batch = createVehicleUploadBatch("attachment", vehicleFixture(), Array.from({ length: 5 }, (_, index) => new File(["x"], `${index}.txt`)));
    batch.items[0] = { ...batch.items[0], status: "saved", storedId: "stored" };
    batch.items[1].status = "unknown";
    batch.items[2].status = "failed";
    batch.items[3].status = "skipped";
    expect(batch.items.filter(retryableUploadItem).map(item => item.file.name)).toEqual(["2.txt", "4.txt"]);
    expect(vehicleUploadCounts(batch)).toEqual({ total: 5, pending: 1, uploading: 0, saved: 1, failed: 1, unknown: 1, skipped: 1 });
  });

  it("treats network loss, incomplete responses and server timeouts as ambiguous", () => {
    expect(vehicleUploadFailure(new TypeError("Connection lost"))).toBe("unknown");
    for (const status of [408, 499, 500, 502, 504]) expect(vehicleUploadFailure(new ApiError("No response", "failed", status))).toBe("unknown");
    for (const status of [400, 403, 404, 409, 413, 415, 429]) expect(vehicleUploadFailure(new ApiError("Rejected", "failed", status))).toBe("failed");
  });
});
