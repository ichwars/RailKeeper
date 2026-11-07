import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError } from "../../shared/api";
import { cvValueFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { buildCVImportPreview } from "./cvImport";
import { useVehicleCVImportController } from "./useVehicleCVImportController";

function setup() {
  const options = { selected: vehicleFixture(), readonly: false, saving: false, setSaving: vi.fn(),
    onMessage: vi.fn(), refreshSaved: vi.fn().mockResolvedValue(undefined) };
  return { options, ...renderHook(({ selected }) => useVehicleCVImportController({ ...options, selected }),
    { initialProps: { selected: options.selected } }) };
}
const preview = () => ({ ...buildCVImportPreview("source.csv", [{ cvNumber: 1, value: 4, decoderProfile: "A" }],
  [cvValueFixture({ decoderProfile: "A", value: 3 })]), vehicleId: "vehicle-1" });

describe("CV import guards and acknowledgments", () => {
  afterEach(() => vi.restoreAllMocks());
  it("retains a stale selection and requires an explicit fresh comparison before another write", async () => {
    const request = vi.spyOn(api, "importVehicleCVValues").mockRejectedValueOnce(new ApiError("Changed", "cv_value_conflict", 409))
      .mockResolvedValueOnce({ imported: 1 });
    const latest = cvValueFixture({ decoderProfile: "A", value: 7, revision: "d".repeat(64) });
    vi.spyOn(api, "vehicleCVValues").mockResolvedValue([latest]);
    const { result } = setup();
    act(() => result.current.commands.setImportPreview(preview()));
    await act(() => result.current.commands.applyImportPreview());
    expect(result.current.state.needsReview).toBe(true);
    expect(result.current.state.preview?.rows[0].input.value).toBe(4);
    await act(() => result.current.commands.applyImportPreview()); expect(request).toHaveBeenCalledOnce();
    await act(() => result.current.commands.recompareImportPreview());
    expect(result.current.state.needsReview).toBe(false);
    expect(result.current.state.preview?.rows[0].existing?.value).toBe(7);
    expect(request).toHaveBeenCalledOnce();
    await act(() => result.current.commands.applyImportPreview());
    expect(request).toHaveBeenLastCalledWith("vehicle-1", [expect.objectContaining({ value: 4, expectedRevision: latest.revision })]);
  });
  it("does not repeat a lost response and deselects a result already found on the server", async () => {
    const request = vi.spyOn(api, "importVehicleCVValues").mockRejectedValueOnce(new TypeError("Response lost"));
    vi.spyOn(api, "vehicleCVValues").mockResolvedValue([cvValueFixture({ decoderProfile: "A", value: 4, description: "" })]);
    const { result } = setup();
    act(() => result.current.commands.setImportPreview(preview()));
    await act(() => result.current.commands.applyImportPreview());
    expect(result.current.state.needsReview).toBe(true);
    await act(() => result.current.commands.applyImportPreview()); expect(request).toHaveBeenCalledOnce();
    await act(() => result.current.commands.recompareImportPreview());
    expect(result.current.state.preview?.rows[0]).toMatchObject({ status: "same", selected: false });
    await act(() => result.current.commands.applyImportPreview()); expect(request).toHaveBeenCalledOnce();
  });
  it("acknowledges storage before a failed refresh and prevents repeating the confirmed selection", async () => {
    const request = vi.spyOn(api, "importVehicleCVValues").mockResolvedValue({ imported: 1 });
    const { result, options } = setup(); options.refreshSaved.mockRejectedValue(new Error("Refresh failed"));
    act(() => result.current.commands.setImportPreview(preview()));
    await act(() => result.current.commands.applyImportPreview());
    expect(result.current.state).toMatchObject({ preview: null, imported: 1 });
    expect(options.onMessage).toHaveBeenLastCalledWith(expect.stringContaining("bestätigt gespeichert"));
    await act(() => result.current.commands.applyImportPreview()); expect(request).toHaveBeenCalledOnce();
  });
  it("ignores an acknowledged response after resetting into another vehicle", async () => {
    let complete!: (value: { imported: number }) => void;
    vi.spyOn(api, "importVehicleCVValues").mockReturnValue(new Promise(resolve => { complete = resolve; }));
    const { result, options, rerender } = setup();
    act(() => result.current.commands.setImportPreview(preview()));
    let pending!: Promise<void>;
    act(() => { pending = result.current.commands.applyImportPreview(); });
    act(() => result.current.commands.reset()); rerender({ selected: vehicleFixture({ id: "other" }) });
    await act(async () => { complete({ imported: 1 }); await pending; });
    expect(result.current.state).toMatchObject({ preview: null, imported: 0 });
    expect(options.refreshSaved).not.toHaveBeenCalled();
    expect(options.setSaving).toHaveBeenLastCalledWith(false);
  });
});
