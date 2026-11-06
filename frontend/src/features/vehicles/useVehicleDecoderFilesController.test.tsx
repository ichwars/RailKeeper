import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type VehicleCVFilePreview } from "../../shared/api";
import { cvFileFixture, vehicleFixture } from "../../test/fixtures/vehicles";
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
  const onFunctionSuggestions = vi.fn().mockReturnValue({ applied: 1, blocked: [] });
  const prepareUpload = vi.fn().mockReturnValue(true), prepareDelete = vi.fn(), onStatus = vi.fn();
  const hook = renderHook(() => useVehicleDecoderFilesController({
    selected: vehicleFixture(), readonly: false, saving: false, onStatus,
    setSaving,
    onMessage,
    onImportPreview,
    onFunctionSuggestions,
    prepareUpload, prepareDelete
  }));
  return { ...hook, setSaving, onMessage, onImportPreview, onFunctionSuggestions, prepareUpload, prepareDelete, onStatus };
}

describe("useVehicleDecoderFilesController", () => {
  beforeEach(() => vi.restoreAllMocks());

  it("ignores late preview results and confirmed suggestions from another vehicle", async () => {
    let complete!: (value: VehicleCVFilePreview) => void;
    vi.spyOn(api, "previewVehicleCVFile").mockReturnValue(new Promise(resolve => { complete = resolve; }));
    const prepareUpload = vi.fn(), onFunctionSuggestions = vi.fn(), onImportPreview = vi.fn();
    const { result, rerender } = renderHook(({ selected }) => useVehicleDecoderFilesController({
      selected, readonly: false, saving: false, setSaving: vi.fn(), onMessage: vi.fn(), onStatus: vi.fn(),
      prepareUpload, prepareDelete: vi.fn(), onFunctionSuggestions, onImportPreview
    }), { initialProps: { selected: vehicleFixture() } });
    let pending!: Promise<void>;
    act(() => { pending = result.current.commands.uploadFiles(fileList(new File(["x"], "decoder.esux"))); });
    rerender({ selected: vehicleFixture({ id: "other" }) });
    await act(async () => { complete(preview); await pending; });
    expect(result.current.state.uploadPreview).toBeNull();
    act(() => result.current.commands.confirmUpload());
    act(() => result.current.commands.applyFunctionSuggestions());
    expect(prepareUpload).not.toHaveBeenCalled(); expect(onFunctionSuggestions).not.toHaveBeenCalled();
    expect(onImportPreview).not.toHaveBeenCalled();
  });

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

  it("prepares CV values and sends function suggestions to the draft without writes", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue(preview);
    const updateFunction = vi.spyOn(api, "updateVehicleFunction");
    const { result, onImportPreview, onFunctionSuggestions, prepareUpload, prepareDelete, onStatus, onMessage } = renderController();
    const file = new File(["decoder"], "decoder.esux", { type: "application/octet-stream" });

    await act(() => result.current.commands.uploadFiles(fileList(file)));
    await act(() => result.current.commands.previewValuesForImport());
    expect(onImportPreview).toHaveBeenCalledWith(expect.objectContaining({
      fileName: "Decoder-Datei-Vorschau",
      rows: [expect.objectContaining({ input: expect.objectContaining({ cvNumber: 1, value: 3 }) })]
    }));

    act(() => result.current.commands.applyFunctionSuggestions());
    expect(onFunctionSuggestions).toHaveBeenCalledWith([preview]);
    expect(updateFunction).not.toHaveBeenCalled();
    expect(prepareUpload).not.toHaveBeenCalled();
    expect(onStatus).toHaveBeenLastCalledWith(expect.stringContaining("Entwurf"));
  });

  it("uploads and removes stored decoder files", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue(preview);
    vi.spyOn(api, "uploadVehicleCVFile").mockResolvedValue(cvFileFixture());
    vi.spyOn(api, "deleteVehicleCVFile").mockResolvedValue(undefined);
    const { result, prepareUpload, prepareDelete } = renderController();
    const file = new File(["decoder"], "decoder.esux", { type: "application/octet-stream" });

    await act(() => result.current.commands.uploadFiles(fileList(file)));
    act(() => result.current.commands.updateFileProfile("ESU LokPilot 5"));
    act(() => result.current.commands.confirmUpload());
    expect(prepareUpload).toHaveBeenCalledWith("decoder", [file], { decoderProfile: "ESU LokPilot 5", description: "" }, true);
    expect(api.uploadVehicleCVFile).not.toHaveBeenCalled();
    expect(result.current.state.uploadPreview).toBeNull();
    act(() => result.current.commands.remove(cvFileFixture()));
    expect(prepareDelete).toHaveBeenCalledWith({ kind: "decoder", record: cvFileFixture() });
    expect(api.deleteVehicleCVFile).not.toHaveBeenCalled();
  });

  it("keeps detected row and preview profiles separate from the file-storage profile", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue({ ...preview,
      suggestedCvValues: [
        { cvNumber: 1, value: 17, decoderProfile: "LokPilot 5 micro DCC" },
        { cvNumber: 29, value: 14 }
      ]
    });
    const { result, onImportPreview } = renderController();
    await act(() => result.current.commands.uploadFiles(fileList(new File(["decoder"], "decoder.esux"))));
    act(() => result.current.commands.updateFileProfile("Eigenes Dateiprofil"));
    await act(() => result.current.commands.previewValuesForImport());
    expect(onImportPreview).toHaveBeenCalledWith(expect.objectContaining({ rows: [
      expect.objectContaining({ input: expect.objectContaining({ decoderProfile: "LokPilot 5 micro DCC" }) }),
      expect.objectContaining({ input: expect.objectContaining({ decoderProfile: "ESU LokPilot 5" }) })
    ] }));
    expect(result.current.state.fileProfile).toBe("Eigenes Dateiprofil");
  });

  it("uses the file profile only when no row or decoder profile was detected", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue({ ...preview, suggestedDecoderProfile: "" });
    const { result, onImportPreview } = renderController();
    await act(() => result.current.commands.uploadFiles(fileList(new File(["decoder"], "decoder.esux"))));
    act(() => result.current.commands.updateFileProfile("Eigenes Dateiprofil"));
    await act(() => result.current.commands.previewValuesForImport());
    expect(onImportPreview).toHaveBeenCalledWith(expect.objectContaining({ rows: [
      expect.objectContaining({ input: expect.objectContaining({ decoderProfile: "Eigenes Dateiprofil" }) })
    ] }));
  });
});
