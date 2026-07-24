import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { imageFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleMediaController } from "./useVehicleMediaController";

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
  const onImageUploadComplete = vi.fn();
  const hook = renderHook(() => useVehicleMediaController({
    selected: vehicleFixture(),
    readonly: false,
    saving: false,
    setSaving,
    onMessage,
    refreshSelectedVehicle,
    onImageUploadComplete
  }));
  return {
    ...hook,
    setSaving,
    onMessage,
    refreshSelectedVehicle,
    onImageUploadComplete
  };
}

describe("useVehicleMediaController", () => {
  it("loads, reorders and updates pending images", () => {
    const vehicle = vehicleFixture();
    const { result } = renderController();

    act(() => result.current.commands.loadDetail(vehicle));
    act(() => result.current.commands.addImages([{
      id: "image-2",
      url: "/image-2.jpg",
      title: "Seite",
      source: "",
      isPrimary: false,
      maintenanceId: ""
    }]));
    act(() => result.current.commands.movePendingImage("image-2", -1));
    act(() => result.current.commands.setPrimaryPendingImage("image-2"));
    act(() => result.current.commands.updatePendingImageTitle("image-2", "Neue Seite"));

    expect(result.current.state.pendingImages.map((image) => image.id)).toEqual(["image-2", "image-1"]);
    expect(result.current.state.pendingImages[0]).toMatchObject({ title: "Neue Seite", isPrimary: true });
    expect(result.current.state.pendingImages[1].isPrimary).toBe(false);
  });

  it("rejects blocked files before an API call", () => {
    const uploadImage = vi.spyOn(api, "uploadVehicleImage");
    const uploadAttachment = vi.spyOn(api, "uploadVehicleAttachment");
    const { result, onMessage } = renderController();

    act(() => result.current.commands.uploadImages(fileList(new File(["x"], "script.exe"))));
    expect(uploadImage).not.toHaveBeenCalled();
    expect(onMessage).toHaveBeenCalledWith("script.exe ist kein erlaubtes Bildformat.");

    act(() => result.current.commands.uploadAttachment(fileList(new File(["x"], "script.exe"))));
    expect(uploadAttachment).not.toHaveBeenCalled();
    expect(onMessage).toHaveBeenLastCalledWith(expect.stringContaining("script.exe ist als Beilage nicht erlaubt"));
  });

  it("uploads an image and refreshes the selected vehicle", async () => {
    vi.spyOn(api, "uploadVehicleImage").mockResolvedValue(imageFixture({ id: "image-uploaded" }));
    const { result, setSaving, refreshSelectedVehicle, onImageUploadComplete } = renderController();
    const file = new File(["image"], "lok.png", { type: "image/png" });

    act(() => result.current.commands.uploadImages(fileList(file)));

    await waitFor(() => expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1"));
    expect(api.uploadVehicleImage).toHaveBeenCalledWith("vehicle-1", file, "lok.png", true, "");
    expect(onImageUploadComplete).toHaveBeenCalledOnce();
    expect(setSaving).toHaveBeenNthCalledWith(1, true);
    expect(setSaving).toHaveBeenLastCalledWith(false);
  });

  it("saves attachment metadata and resets media state", async () => {
    const { result, refreshSelectedVehicle } = renderController();
    const attachment = {
      id: "attachment-1",
      vehicleId: "vehicle-1",
      fileName: "manual.pdf",
      originalName: "manual.pdf",
      sizeBytes: 100,
      createdAt: "2026-07-23T08:00:00Z",
      updatedAt: "2026-07-23T08:00:00Z"
    };
    vi.spyOn(api, "updateVehicleAttachment").mockResolvedValue(attachment);

    act(() => result.current.commands.updateAttachmentEdit(attachment.id, { category: "Anleitung" }));
    act(() => result.current.commands.saveAttachment(attachment));
    await waitFor(() => expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1"));
    expect(api.updateVehicleAttachment).toHaveBeenCalledWith("vehicle-1", attachment.id, {
      description: "",
      category: "Anleitung",
      maintenanceId: ""
    });

    act(() => result.current.commands.reset(true));
    expect(result.current.state.pendingImages).toEqual([]);
    expect(result.current.state.attachmentEdits).toEqual({});
  });
});
