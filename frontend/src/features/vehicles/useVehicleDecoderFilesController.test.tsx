import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type VehicleCVFilePreview } from "../../shared/api";
import { cvFileFixture, functionFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleDecoderFilesController } from "./useVehicleDecoderFilesController";

function fileList(...files: File[]) {
  return {
    ...files,
    length: files.length,
    item: (index: number) => files[index] || null
  } as unknown as FileList;
}

const preview: VehicleCVFilePreview = {
  fileName: "decoder.esux",
  sizeBytes: 1024,
  mimeType: "application/octet-stream",
  hasMetadata: true,
  suggestedDecoderProfile: "ESU LokPilot 5",
  suggestedDescription: "BR 106 Decoderprojekt",
  suggestedCvValues: [{ cvNumber: 1, value: 3, description: "Adresse" }],
  suggestedFunctions: [{ functionKey: "F0", name: "Licht", functionType: "licht" }]
};

function renderController() {
  const setSaving = vi.fn();
  const onMessage = vi.fn();
  const onImportPreview = vi.fn();
  const refreshSelectedVehicle = vi.fn().mockResolvedValue(undefined);
  const hook = renderHook(() => useVehicleDecoderFilesController({
    selected: vehicleFixture(),
    setSaving,
    onMessage,
    onImportPreview,
    refreshSelectedVehicle
  }));
  return { ...hook, setSaving, onMessage, onImportPreview, refreshSelectedVehicle };
}

describe("useVehicleDecoderFilesController", () => {
  beforeEach(() => vi.restoreAllMocks());

  it("previews files and applies detected metadata", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue(preview);
    const { result } = renderController();
    const file = new File(["decoder"], "decoder.esux", { type: "application/octet-stream" });

    await act(() => result.current.commands.uploadFiles(fileList(file)));
    expect(result.current.state.uploadPreview?.previews).toEqual([preview]);

    act(() => result.current.commands.applyFirstSuggestion());
    expect(result.current.state.fileProfile).toBe("ESU LokPilot 5");
    expect(result.current.state.fileDescription).toBe("BR 106 Decoderprojekt");
  });

  it("prepares CV values and imports valid function suggestions", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue(preview);
    vi.spyOn(api, "updateVehicleFunction").mockResolvedValue(functionFixture());
    const { result, onImportPreview, refreshSelectedVehicle } = renderController();
    const file = new File(["decoder"], "decoder.esux", { type: "application/octet-stream" });

    await act(() => result.current.commands.uploadFiles(fileList(file)));
    act(() => result.current.commands.previewValuesForImport());
    expect(onImportPreview).toHaveBeenCalledWith(expect.objectContaining({
      fileName: "Decoder-Datei-Vorschau",
      rows: [expect.objectContaining({ input: expect.objectContaining({ cvNumber: 1, value: 3 }) })]
    }));

    act(() => result.current.commands.applyFunctionSuggestions());
    await waitFor(() => expect(api.updateVehicleFunction).toHaveBeenCalledWith(
      "vehicle-1",
      "F0",
      expect.objectContaining({ name: "Licht", functionType: "licht" })
    ));
    expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1");
  });

  it("uploads and removes stored decoder files", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue(preview);
    vi.spyOn(api, "uploadVehicleCVFile").mockResolvedValue(cvFileFixture());
    vi.spyOn(api, "deleteVehicleCVFile").mockResolvedValue(undefined);
    const { result, refreshSelectedVehicle } = renderController();
    const file = new File(["decoder"], "decoder.esux", { type: "application/octet-stream" });

    await act(() => result.current.commands.uploadFiles(fileList(file)));
    act(() => result.current.commands.updateFileProfile("ESU LokPilot 5"));
    act(() => result.current.commands.confirmUpload());
    await waitFor(() => expect(api.uploadVehicleCVFile).toHaveBeenCalledWith(
      "vehicle-1",
      file,
      "ESU LokPilot 5",
      ""
    ));

    act(() => result.current.commands.remove(cvFileFixture()));
    await waitFor(() => expect(api.deleteVehicleCVFile).toHaveBeenCalledWith("vehicle-1", "cv-file-1"));
    expect(refreshSelectedVehicle).toHaveBeenCalledTimes(2);
  });
});
