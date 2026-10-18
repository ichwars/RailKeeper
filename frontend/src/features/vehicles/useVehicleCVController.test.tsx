import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
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
  const hook = renderHook(() => useVehicleCVController({
    selected: vehicleFixture({ cvValues: [cvValueFixture()] }),
    decoderNumber: "1001",
    setSaving,
    onMessage,
    refreshSelectedVehicle
  }));
  return { ...hook, setSaving, onMessage, refreshSelectedVehicle };
}

describe("useVehicleCVController", () => {
  beforeEach(() => vi.restoreAllMocks());

  it("validates and updates an existing CV value with the same key", async () => {
    vi.spyOn(api, "updateVehicleCVValue").mockResolvedValue(cvValueFixture({ value: 5 }));
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

    await waitFor(() => expect(api.updateVehicleCVValue).toHaveBeenCalledWith(
      "vehicle-1",
      "cv-1",
      expect.objectContaining({ cvNumber: 1, value: 5 })
    ));
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

});
