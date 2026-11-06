import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError } from "../../shared/api";
import { cvFileFixture, imageFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleMediaDeletion } from "./useVehicleMediaDeletion";

function setup() {
  const options = { selected: vehicleFixture(), readonly: false, saving: false, setSaving: vi.fn(),
    onMessage: vi.fn(), onStatus: vi.fn(), onImageDeleted: vi.fn(), refreshSaved: vi.fn().mockResolvedValue(undefined),
    reload: vi.fn().mockResolvedValue(undefined) };
  return { options, ...renderHook(({ selected }) => useVehicleMediaDeletion({ ...options, selected }),
    { initialProps: { selected: options.selected } }) };
}
describe("confirmed media deletion", () => {
  afterEach(() => vi.restoreAllMocks());
  it("captures the reviewed decoder-file revision and reference count until explicit confirmation", async () => {
    const remove = vi.spyOn(api, "deleteVehicleCVFile").mockResolvedValue(undefined);
    const { result, options, rerender } = setup(), record = cvFileFixture({ sourceReferenceCount: 7 });
    act(() => result.current.commands.prepare({ kind: "decoder", record }));
    expect(remove).not.toHaveBeenCalled(); expect(result.current.candidate?.references).toBe(7);
    act(() => result.current.commands.close()); expect(result.current.candidate).toBeNull();
    act(() => result.current.commands.prepare({ kind: "decoder", record }));
    rerender({ selected: vehicleFixture({ cvFiles: [{ ...record, revision: "b".repeat(64), sourceReferenceCount: 8 }] }) });
    await act(() => result.current.commands.confirm());
    expect(remove).toHaveBeenCalledExactlyOnceWith("vehicle-1", record.id, record.revision);
    expect(options.refreshSaved).toHaveBeenCalledExactlyOnceWith("vehicle-1", "decoder");
  });
  it("reloads a changed file and blocks reuse of the old confirmation", async () => {
    const remove = vi.spyOn(api, "deleteVehicleImage").mockRejectedValue(new ApiError("Changed", "image_conflict", 409));
    const { result, options } = setup();
    act(() => result.current.commands.prepare({ kind: "image", record: imageFixture() }));
    await act(() => result.current.commands.confirm());
    expect(options.reload).toHaveBeenCalledExactlyOnceWith("vehicle-1");
    expect(options.refreshSaved).not.toHaveBeenCalled();
    expect(options.onImageDeleted).not.toHaveBeenCalled(); expect(result.current.candidate?.stale).toBe(true);
    await act(() => result.current.commands.confirm()); expect(remove).toHaveBeenCalledOnce();
  });
  it("does not repeat a successful deletion when only the following refresh fails", async () => {
    const remove = vi.spyOn(api, "deleteVehicleImage").mockResolvedValue(undefined);
    const { result, options } = setup(); options.refreshSaved.mockRejectedValue(new Error("Refresh unavailable"));
    act(() => result.current.commands.prepare({ kind: "image", record: imageFixture() }));
    await act(() => result.current.commands.confirm());
    expect(options.onImageDeleted).toHaveBeenCalledExactlyOnceWith("image-1");
    expect(result.current.candidate).toMatchObject({ stale: true, busy: false });
    expect(result.current.candidate?.error).toContain("wurde gelöscht");
    await act(() => result.current.commands.confirm()); expect(remove).toHaveBeenCalledOnce();
  });
  it("rejects linked images, foreign records, missing revisions and changed vehicle scopes", async () => {
    const remove = vi.spyOn(api, "deleteVehicleImage"), { result, rerender } = setup();
    act(() => result.current.commands.prepare({ kind: "image", record: imageFixture({ maintenanceId: "maintenance" }) }));
    expect(result.current.candidate).toBeNull();
    act(() => result.current.commands.prepare({ kind: "image", record: imageFixture({ revision: undefined }) }));
    expect(result.current.candidate).toBeNull();
    act(() => result.current.commands.prepare({ kind: "image", record: imageFixture({ vehicleId: "other" }) }));
    expect(result.current.candidate).toBeNull();
    act(() => result.current.commands.prepare({ kind: "image", record: imageFixture() }));
    rerender({ selected: vehicleFixture({ id: "other" }) });
    await act(() => result.current.commands.confirm()); expect(remove).not.toHaveBeenCalled();
  });
});
