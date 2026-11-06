import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, ApiError } from "../../shared/api";
import { cvValueFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleCVController } from "./useVehicleCVController";

function fileList(...files: File[]) {
  return {
    ...files,
    length: files.length,
    item: (index: number) => files[index] || null
  } as unknown as FileList;
}

function renderController() {
  const setSaving = vi.fn();
  const onMessage = vi.fn();
  const refreshSelectedVehicle = vi.fn().mockResolvedValue(undefined);
  const reloadSelectedVehicle = vi.fn().mockResolvedValue(undefined);
  const initialVehicle = vehicleFixture({ cvValues: [cvValueFixture()] });
  const hook = renderHook(({ vehicle }) => useVehicleCVController({
    selected: vehicle,
    decoderNumber: "1001",
    setSaving,
    onMessage,
    refreshSelectedVehicle,
    reloadSelectedVehicle
  }), { initialProps: { vehicle: initialVehicle } });
  return { ...hook, setSaving, onMessage, refreshSelectedVehicle, reloadSelectedVehicle, initialVehicle };
}

describe("useVehicleCVController", () => {
  beforeEach(() => vi.restoreAllMocks());

  it("validates and updates an existing CV value with the same key", async () => {
    const update = vi.spyOn(api, "updateVehicleCVValue");
    vi.spyOn(api, "applyVehicleCVChanges").mockResolvedValue({ changed: 1 });
    const { result, onMessage, refreshSelectedVehicle } = renderController();

    act(() => result.current.commands.updateForm({ cvNumber: 0, value: 300 }));
    act(() => result.current.commands.save());
    expect(onMessage).toHaveBeenLastCalledWith("CV-Nummer muss 1-1024 und Wert 0-255 sein.");

    act(() => result.current.commands.updateForm({
      cvNumber: 1,
      value: 5,
      decoderProfile: "ESU LokPilot 5"
    }));
    act(() => result.current.commands.save());

    expect(result.current.state.manualDraft.changes).toHaveLength(1);
    expect(result.current.state.manualDraft.values[0]).toMatchObject({ cvNumber: 1, value: 5, description: "Adresse" });
    expect(update).not.toHaveBeenCalled();
    expect(refreshSelectedVehicle).not.toHaveBeenCalled();
    await act(() => result.current.commands.applyManualDraft());
    expect(api.applyVehicleCVChanges).toHaveBeenCalledWith("vehicle-1", [expect.objectContaining({
      id: "cv-1", expectedRevision: "c".repeat(64), value: expect.objectContaining({ cvNumber: 1, value: 5 })
    })]);
    expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1");
  });

  it("previews an import and applies selected new rows", async () => {
    vi.spyOn(api, "importVehicleCVValues").mockResolvedValue({ imported: 1 });
    const { result, refreshSelectedVehicle } = renderController();
    const file = new File(["CV;Wert;Beschreibung\n2;7;Startspannung"], "decoder.csv", { type: "text/csv" });

    await act(() => result.current.commands.importValues(fileList(file)));
    expect(result.current.state.importPreview).toMatchObject({ fileName: "decoder.csv" });
    expect(result.current.state.importStats).toMatchObject({ selected: 1, new: 1 });

    act(() => result.current.commands.applyImportPreview());
    await waitFor(() => expect(api.importVehicleCVValues).toHaveBeenCalledWith(
      "vehicle-1",
      [expect.objectContaining({ cvNumber: 2, value: 7 })]
    ));
    expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1");
    await waitFor(() => expect(result.current.state.importPreview).toBeNull());
  });
  it("previews a full LokProgrammer export and sends one atomic request", async () => {
    const values = Array.from({ length: 1367 }, (_, i) => ({ cvNumber: 257 + i % 256,
      value: i % 256, indexCV31: 16, indexCV32: Math.floor(i / 256) }));
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue({ fileName: "full.txt", sizeBytes: 20000,
      mimeType: "text/plain", hasMetadata: false, suggestedDecoderProfile: "LokPilot 5 micro DCC",
      suggestedCvValues: values });
    vi.spyOn(api, "importVehicleCVValues").mockResolvedValue({ imported: 1367 });
    const { result } = renderController();
    await act(() => result.current.commands.importValues(fileList(new File(["Decoder: LokPilot 5 micro DCC\nCV 1 = 3"], "full.txt"))));
    expect(result.current.state.importStats).toMatchObject({ new: 1367, selected: 1367, invalid: 0 });
    act(() => result.current.commands.applyImportPreview());
    await waitFor(() => expect(api.importVehicleCVValues).toHaveBeenCalledOnce());
    expect(vi.mocked(api.importVehicleCVValues).mock.calls[0][1]).toHaveLength(1367);
    await waitFor(() => expect(result.current.state.importPreview).toBeNull());
  });

  it("does not inherit the edited CV profile or attachment when importing a list", async () => {
    const { result } = renderController();
    act(() => result.current.commands.edit(cvValueFixture({ decoderProfile: "Anderer Decoder", sourceFileId: "old-source" })));
    const file = new File([JSON.stringify({ cvValues: [{ cvNumber: 1, value: 17,
      decoderProfile: "LokPilot 5 micro DCC", sourceFileId: "exported-source",
      sourceLine: 4, sourceExportedAt: "2026-10-17T17:41:59" }] })], "import.json");
    await act(() => result.current.commands.importValues(fileList(file)));
    expect(result.current.state.importPreview?.rows[0].input).toMatchObject({
      cvNumber: 1, value: 17, decoderProfile: "LokPilot 5 micro DCC", sourceFileId: "",
      sourceLine: 4, sourceExportedAt: "2026-10-17T17:41:59"
    });
    expect(result.current.state.form).toMatchObject({ decoderProfile: "Anderer Decoder", sourceFileId: "old-source" });
  });

  it("retains all staged edits after a revision conflict and reloads without claiming a save", async () => {
    vi.spyOn(api, "applyVehicleCVChanges").mockRejectedValue(new ApiError("stale", "cv_value_conflict", 409));
    const { result, reloadSelectedVehicle, refreshSelectedVehicle, rerender, initialVehicle } = renderController();
    act(() => result.current.commands.updateForm({ cvNumber: 1, value: 5, decoderProfile: "ESU LokPilot 5" }));
    act(() => result.current.commands.save());
    const pending = result.current.state.manualDraft.changes;
    await act(() => result.current.commands.applyManualDraft());
    expect(result.current.state.manualDraft.changes).toEqual(pending);
    expect(reloadSelectedVehicle).toHaveBeenCalledWith("vehicle-1");
    expect(refreshSelectedVehicle).not.toHaveBeenCalled();
    rerender({ vehicle: { ...initialVehicle, cvValues: [cvValueFixture({ value: 7, revision: "d".repeat(64) })] } });
    expect(result.current.state.manualDraft.conflicts).toEqual(["cv-1"]);
    act(() => result.current.commands.rebaseManualChange("cv-1"));
    expect(result.current.state.manualDraft.conflicts).toEqual([]);
    expect(result.current.state.manualDraft.changes[0]).toMatchObject({ original: { revision: "d".repeat(64) }, value: { value: 5 } });
  });

  it("saves only the staged selection and retains further unsubmitted CV form input", async () => {
    vi.spyOn(api, "applyVehicleCVChanges").mockResolvedValue({ changed: 1 });
    const { result } = renderController();
    act(() => result.current.commands.updateForm({ cvNumber: 2, value: 7, decoderProfile: "New" }));
    act(() => result.current.commands.save());
    act(() => result.current.commands.updateForm({ cvNumber: 3, value: 9, description: "Still typing" }));
    await act(() => result.current.commands.applyManualDraft());
    expect(api.applyVehicleCVChanges).toHaveBeenCalledWith("vehicle-1", [expect.objectContaining({ value: expect.objectContaining({ cvNumber: 2 }) })]);
    expect(result.current.state.manualDraft.changes).toEqual([]);
    expect(result.current.state.manualDraft.lastSavedCount).toBe(1);
    expect(result.current.state.form).toMatchObject({ cvNumber: 3, value: 9, description: "Still typing" });
  });

  it("stages deletion without a request and restores all stored rows when the draft is discarded", () => {
    const remove = vi.spyOn(api, "deleteVehicleCVValue");
    const { result } = renderController();
    act(() => result.current.commands.remove(cvValueFixture()));
    expect(result.current.state.manualDraft.values).toEqual([]);
    expect(result.current.state.manualDraft.changes[0].value).toBeNull();
    expect(remove).not.toHaveBeenCalled();
    act(() => result.current.commands.discardManualDraft());
    expect(result.current.state.manualDraft.changes).toEqual([]);
    expect(result.current.state.manualDraft.values).toHaveLength(1);
  });

});
