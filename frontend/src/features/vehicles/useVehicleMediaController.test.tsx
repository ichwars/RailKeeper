import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { api, ApiError, type VehicleAttachment } from "../../shared/api";
import { createdImageFixture, imageFixture, vehicleFixture } from "../../test/fixtures/vehicles";
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

  it("keeps local image metadata and pending remote images while refreshing a stored upload", () => {
    const previous = vehicleFixture();
    const { result } = renderController();
    act(() => result.current.commands.loadDetail(previous));
    act(() => result.current.commands.updatePendingImageTitle("image-1", "Unsaved title"));
    act(() => result.current.commands.addImages([{
      id: "remote", url: "https://example.test/new.png", title: "Pending", source: "", isPrimary: false
    }]));
    const uploaded = imageFixture({ id: "uploaded", isPrimary: false });
    const incoming = vehicleFixture({ images: [...(previous.images || []), uploaded] });
    act(() => result.current.commands.refreshDetail(incoming, previous));
    expect(result.current.state.pendingImages.map(image => image.id)).toEqual(["image-1", "uploaded", "remote"]);
    expect(result.current.state.pendingImages[0].title).toBe("Unsaved title");
    expect(result.current.state.pendingImages[2].persisted).not.toBe(true);
  });

  it("replaces each confirmed remote image by its real stored ID before a later import can fail", () => {
    const { result } = renderController();
    const pending = { id: "remote", url: "https://example.test/new.png", title: "Pending", source: "" };
    act(() => result.current.commands.addImages([pending]));
    act(() => result.current.commands.acknowledgeImportedImage(pending, imageFixture({ id: "stored" })));
    expect(result.current.state.pendingImages).toEqual([expect.objectContaining({ id: "stored", persisted: true })]);
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
    vi.spyOn(api, "uploadVehicleImage").mockResolvedValue(createdImageFixture({ id: "image-uploaded" }));
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
      revision: "a".repeat(64),
      vehicleId: "vehicle-1",
      fileName: "manual.pdf",
      originalName: "manual.pdf",
      sizeBytes: 100,
      createdAt: "2026-07-23T08:00:00Z",
      updatedAt: "2026-07-23T08:00:00Z"
    };
    vi.spyOn(api, "updateVehicleAttachment").mockResolvedValue(attachment);

    act(() => result.current.commands.loadDetail(vehicleFixture({ attachments: [attachment] })));
    act(() => result.current.commands.updateAttachmentEdit(attachment.id, { category: "Anleitung" }));
    act(() => result.current.commands.saveAttachment(attachment));
    await waitFor(() => expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1"));
    expect(api.updateVehicleAttachment).toHaveBeenCalledWith("vehicle-1", attachment.id, {
      expectedRevision: attachment.revision,
      description: "",
      category: "Anleitung",
      maintenanceId: ""
    });

    act(() => result.current.commands.reset(true));
    expect(result.current.state.pendingImages).toEqual([]);
    expect(result.current.state.attachmentEdits).toEqual({});
  });
});

function renderDraftBoundaryController() {
  const options = {
    selected: vehicleFixture(), readonly: false, saving: false, setSaving: vi.fn(), onMessage: vi.fn(),
    refreshSelectedVehicle: vi.fn().mockResolvedValue(undefined), onImageUploadComplete: vi.fn()
  };
  return { ...renderHook(() => useVehicleMediaController(options)), options };
}

describe("vehicle media draft and persistence boundary", () => {
  it("edits and removes a new image only in the draft", () => {
    const remove = vi.spyOn(api, "deleteVehicleImage");
    const { result } = renderDraftBoundaryController();
    const image = { id: "new", url: "https://example.test/new.png", source: "Reference", title: "New" };
    act(() => result.current.commands.addImages([image]));
    act(() => result.current.commands.updatePendingImageTitle(image.id, "Draft title"));
    expect(result.current.state.pendingImages[0]).toMatchObject({ title: "Draft title", isPrimary: true });
    act(() => result.current.commands.removePendingImage(result.current.state.pendingImages[0]));
    expect(result.current.state.pendingImages).toEqual([]);
    expect(remove).not.toHaveBeenCalled();
  });

  it("keeps maintenance-linked drafts and deletes an unlinked stored image with its original revision", async () => {
    vi.spyOn(api, "deleteVehicleImage").mockResolvedValue(undefined);
    const { result, options } = renderDraftBoundaryController();
    const image = {
      id: "image-1", url: "/local.png", title: "Stored", source: "", persisted: true, maintenanceId: "maintenance"
    };
    act(() => result.current.commands.addImages([image]));
    act(() => result.current.commands.removePendingImage(image));
    expect(api.deleteVehicleImage).not.toHaveBeenCalled();
    expect(result.current.state.pendingImages).toHaveLength(1);
    act(() => result.current.commands.updatePendingImageMaintenance(image.id, ""));
    act(() => result.current.commands.removePendingImage(result.current.state.pendingImages[0]));
    await waitFor(() => expect(options.refreshSelectedVehicle).toHaveBeenCalledWith(options.selected.id));
    expect(api.deleteVehicleImage).toHaveBeenCalledExactlyOnceWith(options.selected.id, image.id, options.selected.images?.[0].revision);
    expect(result.current.state.pendingImages).toEqual([]);
  });
});


describe("media revision guards", () => {
  it("retains attachment drafts through a conflict and requires a reviewed choice before writing", async () => {
    const attachment: VehicleAttachment = {
      id: "a", revision: "a".repeat(64), vehicleId: "vehicle-1", fileName: "manual.txt", originalName: "manual.txt",
      sizeBytes: 4, createdAt: "original", updatedAt: "original", description: "Original", category: "manual"
    };
    const before = vehicleFixture({ attachments: [attachment] });
    const incoming = vehicleFixture({ attachments: [{ ...attachment, description: "Newer", category: "invoice", revision: "b".repeat(64) }] });
    const options = {
      readonly: false, saving: false, setSaving: vi.fn(), onMessage: vi.fn(),
      refreshSelectedVehicle: vi.fn().mockResolvedValue(undefined), reloadSelectedVehicle: vi.fn().mockResolvedValue(undefined),
      onImageUploadComplete: vi.fn()
    };
    const update = vi.spyOn(api, "updateVehicleAttachment").mockRejectedValueOnce(new ApiError("Changed", "attachment_conflict", 409))
      .mockResolvedValueOnce(incoming.attachments![0]);
    const { result, rerender } = renderHook(({ selected }) => useVehicleMediaController({ ...options, selected }),
      { initialProps: { selected: before } });
    act(() => result.current.commands.loadDetail(before));
    act(() => result.current.commands.updateAttachmentEdit(attachment.id, { description: "Own" }));
    act(() => result.current.commands.saveAttachment(attachment));
    await waitFor(() => expect(options.reloadSelectedVehicle).toHaveBeenCalledWith(before.id));
    expect(options.refreshSelectedVehicle).not.toHaveBeenCalled();
    expect(result.current.state.attachmentEdits.a.description).toBe("Own");
    act(() => result.current.commands.refreshDetail(incoming, before));
    rerender({ selected: incoming });
    expect(result.current.state.mediaConflicts).toHaveLength(1);
    act(() => result.current.commands.saveAttachment(incoming.attachments![0]));
    expect(update).toHaveBeenCalledOnce();
    act(() => result.current.commands.updateAttachmentEdit(attachment.id, { description: "Own final" }));
    expect(result.current.state.mediaConflicts[0].fields[0].local).toBe("Own final");
    act(() => result.current.commands.resolveConflict(result.current.state.mediaConflicts[0], "draft"));
    act(() => result.current.commands.saveAttachment(incoming.attachments![0]));
    await waitFor(() => expect(options.refreshSelectedVehicle).toHaveBeenCalledWith(before.id));
    expect(update).toHaveBeenLastCalledWith(before.id, attachment.id, {
      description: "Own final", category: "invoice", maintenanceId: "", expectedRevision: "b".repeat(64)
    });
  });
});
