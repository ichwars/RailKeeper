import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, type VehicleCVValue } from "../../shared/api";
import { cvValueFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { emptyVehicle, type ECoSVehicleDraftPayload } from "./vehicleViewModel";
import { adoptVehicleImportDraft } from "./vehicleImportAdoption";

const mapping = { id: "mapping", vehicleId: "vehicle-1", provider: "ecos", externalId: "77", syncStatus: "linked",
  createdAt: "before", updatedAt: "before" };
const vehicle = () => vehicleFixture({ externalMappings: [mapping] });
const draft = (): ECoSVehicleDraftPayload => ({ source: "ecos", mode: "update", sourceSummary: { objectId: 77,
  name: "Legacy confirmed source", address: "3", protocol: "DCC", profile: "" }, vehicle: emptyVehicle,
  importedKeys: [], externalMapping: { provider: "ecos", externalId: "77" }, cvValues: [
    { cvNumber: 1, value: 4, decoderProfile: "A" }, { cvNumber: 2, value: 5, decoderProfile: "A" }
  ], functionValues: [], unclearFields: [] });
const callbacks = () => ({ markSaved: vi.fn(), clear: vi.fn(), returnToSession: vi.fn(), reloadVehicles: vi.fn(), closeEditor: vi.fn() });

describe("legacy source CV adoption", () => {
  afterEach(() => vi.restoreAllMocks());
  it("does not write or acknowledge a source when its pending comparison belongs to a closed editor", async () => {
    let finish!: (values: VehicleCVValue[]) => void, current = true;
    vi.spyOn(api, "vehicleCVValues").mockReturnValue(new Promise(resolve => { finish = resolve; }));
    const batch = vi.spyOn(api, "importVehicleCVValues"), cb = callbacks();
    const pending = adoptVehicleImportDraft(vehicle(), draft(), { ...cb, isCurrent: () => current });
    current = false; finish([]); await pending;
    expect(batch).not.toHaveBeenCalled(); expect(cb.markSaved).not.toHaveBeenCalled(); expect(cb.clear).not.toHaveBeenCalled();
  });
  it("submits one atomic guarded batch, skips identical identities and preserves all source metadata", async () => {
    const source = draft(); source.cvValues[1] = { ...source.cvValues[1], cvNumber: 257, indexCV31: 0, indexCV32: 0,
      sourceFileId: "source", sourceLine: 27, sourceExportedAt: "2026-08-20T08:00:00" };
    const existing = cvValueFixture({ decoderProfile: "A", value: 4, description: "" });
    vi.spyOn(api, "vehicleCVValues").mockResolvedValue([existing]);
    const batch = vi.spyOn(api, "importVehicleCVValues").mockResolvedValue({ imported: 1 });
    const perRowCreate = vi.spyOn(api, "createVehicleCVValue"), perRowUpdate = vi.spyOn(api, "updateVehicleCVValue");
    vi.spyOn(api, "vehicle").mockResolvedValue(vehicle()); const cb = callbacks();
    await adoptVehicleImportDraft(vehicle(), source, cb);
    expect(batch).toHaveBeenCalledExactlyOnceWith("vehicle-1", [{ ...source.cvValues[1], expectedRevision: "" }]);
    expect(perRowCreate).not.toHaveBeenCalled(); expect(perRowUpdate).not.toHaveBeenCalled();
    expect(cb.markSaved).toHaveBeenCalledOnce(); expect(cb.clear).toHaveBeenCalledOnce();
  });
  it("pins the observed revision and does not acknowledge a rejected batch", async () => {
    const source = draft(), existing = cvValueFixture({ decoderProfile: "A", value: 3 });
    vi.spyOn(api, "vehicleCVValues").mockResolvedValue([existing]);
    const batch = vi.spyOn(api, "importVehicleCVValues").mockRejectedValue(new ApiError("Changed", "cv_value_conflict", 409));
    const cb = callbacks();
    await expect(adoptVehicleImportDraft(vehicle(), source, cb)).rejects.toMatchObject({ code: "cv_value_conflict" });
    expect(batch).toHaveBeenCalledWith("vehicle-1", [expect.objectContaining({ cvNumber: 1, expectedRevision: existing.revision }),
      expect.objectContaining({ cvNumber: 2, expectedRevision: "" })]);
    expect(cb.markSaved).not.toHaveBeenCalled(); expect(cb.clear).not.toHaveBeenCalled();
  });
  it("does not resubmit identical results after a lost response", async () => {
    const source = draft();
    const current = vi.spyOn(api, "vehicleCVValues").mockResolvedValueOnce([])
      .mockResolvedValueOnce(source.cvValues.map((value, index) => cvValueFixture({ ...value, id: `cv-${index}`, description: "" })));
    const batch = vi.spyOn(api, "importVehicleCVValues").mockRejectedValueOnce(new TypeError("Lost reply"));
    vi.spyOn(api, "vehicle").mockResolvedValue(vehicle()); const cb = callbacks();
    await expect(adoptVehicleImportDraft(vehicle(), source, cb)).rejects.toMatchObject({ code: "cv_import_unknown" });
    await adoptVehicleImportDraft(vehicle(), source, cb);
    expect(current).toHaveBeenCalledTimes(2); expect(batch).toHaveBeenCalledOnce(); expect(cb.markSaved).toHaveBeenCalledOnce();
  });
  it("rejects malformed identities before writing any CV and keeps the source draft", async () => {
    const source = draft(); source.cvValues[1].cvNumber = 0;
    vi.spyOn(api, "vehicleCVValues").mockResolvedValue([]);
    const batch = vi.spyOn(api, "importVehicleCVValues"), cb = callbacks();
    await expect(adoptVehicleImportDraft(vehicle(), source, cb)).rejects.toMatchObject({ code: "cv_import_invalid" });
    expect(batch).not.toHaveBeenCalled(); expect(cb.clear).not.toHaveBeenCalled();
  });
});
