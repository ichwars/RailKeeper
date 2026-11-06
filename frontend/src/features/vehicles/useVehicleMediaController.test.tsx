import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { api, ApiError, type VehicleAttachment } from "../../shared/api";
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
  const prepareUpload = vi.fn().mockReturnValue(true), prepareDelete = vi.fn();
  const hook = renderHook(() => useVehicleMediaController({
    selected: vehicleFixture(),
    readonly: false,
    saving: false,
    setSaving,
    onMessage,
    refreshSelectedVehicle,
    prepareUpload, prepareDelete
  }));
  return {
    ...hook,
    setSaving,
    onMessage,
    refreshSelectedVehicle,
    prepareUpload, prepareDelete
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
    expect(result.current.state.pendingImages).toContainEqual(expect.objectContaining({ id: "stored", persisted: true }));
    expect(result.current.state.pendingImages.some(image => image.id === pending.id)).toBe(false);
  });

  it("captures image and attachment selections for confirmation without writing", () => {
    const uploadImage = vi.spyOn(api, "uploadVehicleImage"), uploadAttachment = vi.spyOn(api, "uploadVehicleAttachment");
    const { result, prepareUpload } = renderController();
    const image = new File(["x"], "lok.png"), attachment = new File(["x"], "manual.pdf");
    act(() => result.current.setters.setImageUploadMaintenanceId("maintenance-1"));
    act(() => result.current.commands.uploadImages(fileList(image)));
    expect(prepareUpload).toHaveBeenCalledWith("image", [image], { isPrimary: true, maintenanceId: "maintenance-1" });
    act(() => result.current.setters.setAttachmentUploadDescription("Captured note"));
    act(() => result.current.commands.uploadAttachment(fileList(attachment)));
    expect(prepareUpload).toHaveBeenLastCalledWith("attachment", [attachment], { category: "", description: "Captured note" });
    expect(uploadImage).not.toHaveBeenCalled(); expect(uploadAttachment).not.toHaveBeenCalled();
  });

  it("acknowledges an upload before refreshing without treating its old title as a local edit", () => {
    const before = vehicleFixture(), uploaded = imageFixture({ id: "uploaded", title: "Original", isPrimary: false });
    const { result } = renderController();
    act(() => result.current.commands.loadDetail(before));
    act(() => result.current.commands.updatePendingImageTitle("image-1", "Own title"));
    act(() => result.current.commands.acknowledgeImage(uploaded));
    const refreshed = { ...before, images: [...before.images!, { ...uploaded, title: "External newer title" }] };
    act(() => result.current.commands.refreshDetail(refreshed, before));
    expect(result.current.state.pendingImages.find(image => image.id === "uploaded")?.title).toBe("External newer title");
    expect(result.current.state.pendingImages.find(image => image.id === "image-1")?.title).toBe("Own title");
    expect(result.current.state.mediaConflicts).toEqual([]);
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
    refreshSelectedVehicle: vi.fn().mockResolvedValue(undefined), prepareUpload: vi.fn().mockReturnValue(true), prepareDelete: vi.fn()
  };
  return { ...renderHook(() => useVehicleMediaController(options)), options };
}

describe("vehicle media draft and persistence boundary", () => {
  it("requires saving a stored maintenance unlink before preparing image deletion", () => {
    const vehicle = vehicleFixture({ images: [imageFixture({ maintenanceId: "maintenance-1" })] });
    const options = { selected: vehicle, readonly: false, saving: false, setSaving: vi.fn(), onMessage: vi.fn(),
      refreshSelectedVehicle: vi.fn(), prepareUpload: vi.fn(), prepareDelete: vi.fn() };
    const { result } = renderHook(() => useVehicleMediaController(options));
    act(() => result.current.commands.loadDetail(vehicle));
    act(() => result.current.commands.updatePendingImageMaintenance("image-1", ""));
    act(() => result.current.commands.removePendingImage(result.current.state.pendingImages[0]));
    expect(options.prepareDelete).not.toHaveBeenCalled();
    expect(result.current.state.pendingImages).toHaveLength(1);
    expect(options.onMessage).toHaveBeenCalledWith(expect.stringContaining("Wartung"));
  });

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
    expect(options.prepareDelete).toHaveBeenCalledWith({ kind: "image", record: options.selected.images?.[0] });
    expect(api.deleteVehicleImage).not.toHaveBeenCalled();
    expect(result.current.state.pendingImages).toHaveLength(1);
    act(() => result.current.commands.removeDeletedImage(image.id));
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
      prepareUpload: vi.fn().mockReturnValue(true), prepareDelete: vi.fn()
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
