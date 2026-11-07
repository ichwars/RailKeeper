import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError } from "../../shared/api";
import { createdImageFixture, cvFileFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleUploads } from "./useVehicleUploads";

const files = () => ["first", "second", "third"].map(name => new File([name], `${name}.xml`, { type: "text/xml" }));
const images = () => ["first", "second", "third"].map(name => new File([name], `${name}.png`, { type: "image/png" }));
function setup() {
  const options = { selected: vehicleFixture(), readonly: false, saving: false, setSaving: vi.fn(),
    onMessage: vi.fn(), onStatus: vi.fn(), onImageStored: vi.fn(), refreshSaved: vi.fn().mockResolvedValue(undefined),
    reload: vi.fn().mockResolvedValue(undefined) };
  return { options, ...renderHook(({ selected }) => useVehicleUploads({ ...options, selected }),
    { initialProps: { selected: options.selected } }) };
}

describe("confirmed vehicle uploads", () => {
  afterEach(() => vi.restoreAllMocks());

  it("requires confirmation and retries only a rejected or unattempted file", async () => {
    const uploads = vi.spyOn(api, "uploadVehicleCVFile").mockResolvedValueOnce(cvFileFixture({ id: "first" }))
      .mockRejectedValueOnce(new ApiError("Too large", "too_large", 413)).mockResolvedValueOnce(cvFileFixture({ id: "third" }));
    const { result, options } = setup(), chosen = files();
    act(() => { result.current.commands.prepare("decoder", chosen, { decoderProfile: "Original", description: "Captured" }); });
    expect(uploads).not.toHaveBeenCalled();
    await act(() => result.current.commands.confirm());
    expect(result.current.batch?.items.map(item => item.status)).toEqual(["saved", "failed", "pending"]);
    expect(options.refreshSaved).toHaveBeenCalledExactlyOnceWith("vehicle-1", "decoder");
    act(() => result.current.commands.skip(result.current.batch!.items[1].key));
    await act(() => result.current.commands.confirm());
    expect(uploads.mock.calls.map(call => call[1].name)).toEqual(["first.xml", "second.xml", "third.xml"]);
    expect(uploads).toHaveBeenLastCalledWith("vehicle-1", chosen[2], "Original", "Captured");
    expect(result.current.batch?.items.map(item => item.status)).toEqual(["saved", "skipped", "saved"]);
    act(() => result.current.commands.close());
    expect(options.onStatus).toHaveBeenCalledWith(expect.stringContaining("2 gespeichert"));
  });

  it("never retries a lost response and refreshes the list even without acknowledged success", async () => {
    const uploads = vi.spyOn(api, "uploadVehicleCVFile").mockRejectedValueOnce(new TypeError("Connection lost"))
      .mockResolvedValue(cvFileFixture());
    const { result, options } = setup(), chosen = files().slice(0, 2);
    act(() => { result.current.commands.prepare("decoder", chosen, {}); });
    await act(() => result.current.commands.confirm());
    expect(options.reload).toHaveBeenCalledExactlyOnceWith("vehicle-1");
    expect(result.current.batch?.items.map(item => item.status)).toEqual(["unknown", "pending"]);
    await act(() => result.current.commands.confirm());
    expect(uploads.mock.calls.map(call => call[1].name)).toEqual(["first.xml", "second.xml"]);
    act(() => result.current.commands.close());
    expect(options.onMessage).toHaveBeenLastCalledWith(expect.stringContaining("1 unklar"));
  });

  it("chains only its own image receipts and freezes a stale confirmation after another writer", async () => {
    const uploads = vi.spyOn(api, "uploadVehicleImage").mockResolvedValueOnce(createdImageFixture({ id: "first", imagesRevision: "a".repeat(64) }))
      .mockRejectedValueOnce(new ApiError("Changed", "vehicle_image_conflict", 409));
    const { result, options, rerender } = setup(), chosen = images();
    act(() => { result.current.commands.prepare("image", chosen, { isPrimary: true }); });
    rerender({ selected: { ...options.selected, imagesRevision: "c".repeat(64) } });
    await act(() => result.current.commands.confirm());
    expect(uploads.mock.calls.map(call => call[5])).toEqual(["f".repeat(64), "a".repeat(64)]);
    expect(options.onImageStored).toHaveBeenCalledWith(expect.objectContaining({ id: "first" }));
    expect(result.current.batch?.needsReview).toBe(true);
    expect(result.current.batch?.items.map(item => item.status)).toEqual(["saved", "failed", "pending"]);
    await act(() => result.current.commands.confirm());
    expect(uploads).toHaveBeenCalledTimes(2);
  });

  it("retains an acknowledged image when its receipt or subsequent refresh is incomplete", async () => {
    vi.spyOn(api, "uploadVehicleImage").mockResolvedValue(createdImageFixture({ imagesRevision: "" }));
    const { result, options } = setup();
    options.refreshSaved.mockRejectedValue(new Error("Refresh unavailable"));
    act(() => { result.current.commands.prepare("image", images(), {}); });
    await act(() => result.current.commands.confirm());
    expect(result.current.batch?.items.map(item => item.status)).toEqual(["saved", "pending", "pending"]);
    expect(result.current.batch?.needsReview).toBe(true);
    expect(result.current.batch?.notice).toContain("Bestätigte Dateien bleiben gespeichert");
    await act(() => result.current.commands.confirm());
    expect(api.uploadVehicleImage).toHaveBeenCalledOnce();
  });

  it("releases selected files on cancellation and ignores an old response after changing scope", async () => {
    let complete!: (value: ReturnType<typeof cvFileFixture>) => void;
    vi.spyOn(api, "uploadVehicleCVFile").mockReturnValue(new Promise(resolve => { complete = resolve; }));
    const { result, options, rerender } = setup();
    act(() => { result.current.commands.prepare("decoder", files(), {}); result.current.commands.close(); });
    expect(result.current.batch).toBeNull();
    act(() => { result.current.commands.prepare("decoder", files(), {}, true); });
    expect(result.current.batch?.phase).toBe("uploading");
    act(() => result.current.commands.close());
    expect(result.current.batch).not.toBeNull();
    act(() => result.current.commands.reset());
    rerender({ selected: vehicleFixture({ id: "other" }) });
    await act(async () => { complete(cvFileFixture()); });
    expect(result.current.batch).toBeNull();
    expect(options.refreshSaved).not.toHaveBeenCalled();
    expect(options.setSaving).toHaveBeenLastCalledWith(false);
  });

  it("rejects unsupported or oversized selections and missing comparison versions before writes", () => {
    const upload = vi.spyOn(api, "uploadVehicleImage"), { result, options, rerender } = setup();
    act(() => { result.current.commands.prepare("image", [new File(["x"], "script.exe")], {}); });
    expect(options.onMessage).toHaveBeenLastCalledWith(expect.stringContaining("script.exe"));
    act(() => { result.current.commands.prepare("image", Array.from({ length: 21 }, () => files()[0]), {}); });
    expect(result.current.batch).toBeNull();
    rerender({ selected: vehicleFixture({ imagesRevision: undefined }) });
    act(() => { result.current.commands.prepare("image", images(), {}); });
    expect(result.current.batch).toBeNull(); expect(upload).not.toHaveBeenCalled();
  });
});
