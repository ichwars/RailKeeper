import type { FormEvent } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api, ApiError, type Vehicle, type VehicleFunctionChange, type VehicleImage } from "../../shared/api";
import { createdImageFixture, imageFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { uploadedImageToPending, type PendingArticleImage } from "./vehicleTransforms";
import { emptyVehicle, type ECoSVehicleDraftPayload } from "./vehicleViewModel";
import { createVehicleMutationCommands } from "./vehicleMutationCommands";

function commandOptions() {
  const selected: Vehicle | null = null;
  return {
    draftOwner: "tester",
    editor: {
      form: { ...emptyVehicle, manufacturer: "Piko", name: "BR 106" },
      selected: selected as Vehicle | null,
      mode: "create" as const,
      setSaving: vi.fn(),
      setSelectedDetail: vi.fn(),
      setMode: vi.fn(),
      setSaveAttempted: vi.fn(),
      setActiveTab: vi.fn(),
      openModelSection: vi.fn(),
      close: vi.fn()
    },
    validation: { missingRequiredLabels: [] as string[] },
    media: { pendingImages: [] },
    spareParts: { selectedInputs: vi.fn(() => []), clearSelected: vi.fn() },
    functions: { changes: vi.fn((): VehicleFunctionChange[] => []), acknowledgeSaved: vi.fn() },
    ecos: {
      draft: null as ECoSVehicleDraftPayload | null,
      unclearFieldCount: 0,
      markSaved: vi.fn(),
      clear: vi.fn(),
      returnToSession: vi.fn()
    },
    deletion: { candidate: null as Vehicle | null, setCandidate: vi.fn() },
    reloadVehicles: vi.fn(),
    onMessage: vi.fn(),
    t: (key: string) => key
  };
}

const submitEvent = () => ({ preventDefault: vi.fn() }) as unknown as FormEvent;

describe("createVehicleMutationCommands", () => {
  afterEach(() => vi.restoreAllMocks());

  it("preserves untouched stored ordering numbers and renumbers only an explicit reordering", async () => {
    const selected = vehicleFixture({ images: [imageFixture({ id: "primary", sortOrder: 8 }),
      imageFixture({ id: "other", sortOrder: 2, isPrimary: false })] });
    const pending = selected.images!.map(uploadedImageToPending);
    const options = commandOptions();
    const update = vi.spyOn(api, "updateVehicle").mockResolvedValue(selected);
    const editor = { ...options.editor, mode: "edit" as const, selected };
    await createVehicleMutationCommands({ ...options, editor, media: { pendingImages: pending } }).submit(submitEvent());
    expect(update.mock.calls[0][1].images?.map(image => image.sortOrder)).toEqual([8, 2]);
    await createVehicleMutationCommands({ ...options, editor, media: { pendingImages: [...pending].reverse() } }).submit(submitEvent());
    expect(update.mock.calls[1][1].images?.map(image => image.sortOrder)).toEqual([0, 1]);
  });

  it("submits the observed core and image revisions and retains a rejected draft during reload", async () => {
    const options = commandOptions(), selected = vehicleFixture();
    const reloadSelectedVehicle = vi.fn().mockResolvedValue(undefined);
    const update = vi.spyOn(api, "updateVehicle").mockRejectedValue(new ApiError("Changed", "vehicle_image_conflict", 409));
    const form = { ...options.editor.form, name: "Own retained draft" };
    await createVehicleMutationCommands({ ...options,
      editor: { ...options.editor, mode: "edit", selected, form, reloadSelectedVehicle }
    }).submit(submitEvent());
    expect(update).toHaveBeenCalledWith(selected.id, expect.objectContaining({
      name: "Own retained draft", expectedRevision: selected.revision, expectedImagesRevision: selected.imagesRevision
    }));
    expect(reloadSelectedVehicle).toHaveBeenCalledExactlyOnceWith(selected.id);
    expect(options.editor.setSelectedDetail).not.toHaveBeenCalled();
    expect(options.functions.acknowledgeSaved).not.toHaveBeenCalled();
    expect(form.name).toBe("Own retained draft");
    expect(options.onMessage).toHaveBeenLastCalledWith("vehicles.mediaConflict.changed");
  });

  it("blocks unreviewed image conflicts and records missing their comparison version", async () => {
    const options = commandOptions();
    const update = vi.spyOn(api, "updateVehicle");
    await createVehicleMutationCommands({ ...options, media: { ...options.media, hasConflicts: true } }).submit(submitEvent());
    expect(options.onMessage).toHaveBeenLastCalledWith("vehicles.mediaConflict.resolve");
    await createVehicleMutationCommands({ ...options,
      editor: { ...options.editor, mode: "edit", selected: vehicleFixture({ revision: undefined }) }
    }).submit(submitEvent());
    expect(update).not.toHaveBeenCalled();
    expect(options.onMessage).toHaveBeenLastCalledWith("vehicles.mediaConflict.missingRevision");
  });

  it("blocks unresolved refreshed fields before any write", async () => {
    const options = commandOptions();
    const create = vi.spyOn(api, "createVehicle");
    await createVehicleMutationCommands({ ...options, editor: { ...options.editor, hasRefreshConflicts: true } })
      .submit(submitEvent());
    expect(create).not.toHaveBeenCalled();
    expect(options.onMessage).toHaveBeenCalledWith("vehicles.saveState.resolveConflicts");
  });

  it("retries only the remaining remote image after a partial import and retains the stored image ID", async () => {
    const options = commandOptions(), created = vehicleFixture({ id: "created", images: [] });
    const first = createdImageFixture({ id: "stored-first", vehicleId: created.id, url: "/first.png" });
    const second = createdImageFixture({ id: "stored-second", vehicleId: created.id, url: "/second.png", isPrimary: false });
    const firstPending = { id: "remote-first", url: "https://example.test/first.png", title: "First", source: "" };
    const secondPending = { id: "remote-second", url: "https://example.test/second.png", title: "Second", source: "" };
    let pending: PendingArticleImage[] = [firstPending, secondPending];
    const acknowledgeImportedImage = (original: PendingArticleImage, stored: VehicleImage) => {
      pending = pending.map(image => image.id === original.id ? uploadedImageToPending(stored) : image);
    };
    vi.spyOn(api, "createVehicle").mockResolvedValue(created);
    const update = vi.spyOn(api, "updateVehicle").mockResolvedValue({ ...created, images: [first] });
    const imports = vi.spyOn(api, "importVehicleImageFromUrl")
      .mockResolvedValueOnce(first).mockRejectedValueOnce(new Error("Second download failed")).mockResolvedValueOnce(second);
    vi.spyOn(api, "vehicle").mockResolvedValue({ ...created, images: [first, second] });
    await createVehicleMutationCommands({ ...options, media: { pendingImages: pending, acknowledgeImportedImage } })
      .submit(submitEvent());
    expect(pending[0]).toMatchObject({ id: "stored-first", persisted: true });
    await createVehicleMutationCommands({ ...options,
      editor: { ...options.editor, mode: "edit", selected: created },
      media: { pendingImages: pending, acknowledgeImportedImage }
    }).submit(submitEvent());
    expect(api.createVehicle).toHaveBeenCalledOnce();
    expect(update).toHaveBeenCalledWith(created.id, expect.objectContaining({
      images: [expect.objectContaining({ id: "stored-first", url: "/first.png" })]
    }));
    expect(imports.mock.calls.map(call => call[1].url)).toEqual([
      "https://example.test/first.png", "https://example.test/second.png", "https://example.test/second.png"
    ]);
  });

  it("stops at required model fields before writing", async () => {
    const options = commandOptions();
    options.validation.missingRequiredLabels = ["Spurweite"];
    const createVehicle = vi.spyOn(api, "createVehicle");

    await createVehicleMutationCommands(options).submit(submitEvent());

    expect(createVehicle).not.toHaveBeenCalled();
    expect(options.editor.setActiveTab).toHaveBeenCalledWith("model");
    expect(options.editor.openModelSection).toHaveBeenCalledOnce();
    expect(options.onMessage).toHaveBeenCalledWith("vehicles.requiredMissing");
  });

  it("creates and reloads a vehicle", async () => {
    const options = commandOptions();
    const created = vehicleFixture({ id: "created" });
    vi.spyOn(api, "createVehicle").mockResolvedValue(created);
    vi.spyOn(api, "vehicle").mockResolvedValue(created);

    await createVehicleMutationCommands(options).submit(submitEvent());

    expect(api.createVehicle).toHaveBeenCalledWith(expect.objectContaining({ name: "BR 106", images: [] }));
    expect(options.editor.setSelectedDetail).toHaveBeenCalledWith(created);
    expect(options.editor.setMode).toHaveBeenCalledWith("edit");
    expect(options.reloadVehicles).toHaveBeenCalledOnce();
    expect(options.onMessage).toHaveBeenCalledWith("vehicles.createdContinue");
    expect(options.editor.setSaving).toHaveBeenLastCalledWith(false);
  });

  it("keeps the created vehicle selected when a later image import fails and retries as an update", async () => {
    const created = vehicleFixture({ id: "created-once" });
    const options = {
      ...commandOptions(),
      media: { pendingImages: [{ id: "remote", url: "https://example.test/image.png", source: "Reference", title: "Reference" }] }
    };
    vi.spyOn(api, "createVehicle").mockResolvedValue(created);
    vi.spyOn(api, "updateVehicle").mockResolvedValue(created);
    vi.spyOn(api, "vehicle").mockResolvedValue(created);
    vi.spyOn(api, "importVehicleImageFromUrl").mockRejectedValueOnce(new Error("Import unavailable"))
      .mockResolvedValueOnce({ imagesRevision: "b".repeat(64), id: "image", vehicleId: created.id, url: "/local.png", isPrimary: false, sortOrder: 0, createdAt: "", updatedAt: "" });

    await createVehicleMutationCommands(options).submit(submitEvent());

    expect(options.editor.setSelectedDetail).toHaveBeenCalledWith(created);
    expect(options.editor.setMode).toHaveBeenCalledWith("edit");
    expect(options.editor.setSelectedDetail.mock.invocationCallOrder[0])
      .toBeLessThan(vi.mocked(api.importVehicleImageFromUrl).mock.invocationCallOrder[0]);
    expect(options.onMessage).toHaveBeenLastCalledWith("Import unavailable");
    expect(options.editor.close).not.toHaveBeenCalled();
    await createVehicleMutationCommands({
      ...options, editor: { ...options.editor, selected: created, mode: "edit" }
    }).submit(submitEvent());

    expect(api.createVehicle).toHaveBeenCalledOnce();
    expect(api.updateVehicle).toHaveBeenCalledOnce();
    expect(options.reloadVehicles).toHaveBeenCalledOnce();
  });

  it("keeps persisted image identity and source in the core request while importing only new remote images", async () => {
    const created = vehicleFixture({ id: "images" });
    const options = {
      ...commandOptions(),
      media: { pendingImages: [
        { id: "existing", url: "https://example.test/existing.png", source: "Original source", title: "Original", persisted: true },
        { id: "remote", url: "https://example.test/new.png", source: "New source", title: "New", persisted: false }
      ] }
    };
    vi.spyOn(api, "createVehicle").mockResolvedValue(created);
    vi.spyOn(api, "vehicle").mockResolvedValue(created);
    vi.spyOn(api, "importVehicleImageFromUrl").mockResolvedValue({
      imagesRevision: "b".repeat(64), id: "new", vehicleId: created.id, url: "/new.png", isPrimary: false, sortOrder: 1, createdAt: "", updatedAt: ""
    });
    await createVehicleMutationCommands(options).submit(submitEvent());
    expect(api.createVehicle).toHaveBeenCalledWith(expect.objectContaining({ images: [
      expect.objectContaining({ id: "existing", url: "https://example.test/existing.png", sourceUrl: "Original source" })
    ] }));
    expect(api.importVehicleImageFromUrl).toHaveBeenCalledExactlyOnceWith(created.id, expect.objectContaining({
      url: "https://example.test/new.png", sourceUrl: "New source", sortOrder: 1
    }));
  });

  it("stores an ECoS mapping and returns to Digital Centers", async () => {
    const options = commandOptions();
    const created = vehicleFixture({ id: "created" });
    const draft: ECoSVehicleDraftPayload = {
      source: "ecos",
      mode: "create",
      sourceWorkItemId: "item-77",
      sourceSummary: { objectId: 77, name: "BR 106", address: "3", protocol: "DCC", profile: "" },
      vehicle: { ...emptyVehicle, name: "BR 106", category: "Lokomotive", digital: true },
      importedKeys: ["name", "category", "digital"],
      externalMapping: {
        provider: "ecos", externalId: "77", externalName: "BR 106",
        externalAddress: "3", externalProtocol: "DCC", syncStatus: "linked"
      },
      cvValues: [{ cvNumber: 1024, value: 0, decoderProfile: "ECoS" }],
      functionValues: [],
      unclearFields: [],
      returnToDigitalCenters: { sessionId: "session-1", objectId: "77" }
    };
    options.ecos.draft = draft;
    vi.spyOn(api, "importDigitalCenterCVs").mockResolvedValue({ imported: 1, preserved: 0 });
    const createCV = vi.spyOn(api, "createVehicleCVValue");
    vi.spyOn(api, "createVehicle").mockResolvedValue(created);
    vi.spyOn(api, "upsertVehicleExternalMapping").mockResolvedValue({
      id: "mapping-1", vehicleId: created.id, provider: "ecos", externalId: "77",
      syncStatus: "linked", createdAt: "2026-08-23T08:00:00Z", updatedAt: "2026-08-23T08:00:00Z"
    });
    vi.spyOn(api, "vehicle").mockResolvedValue(created);

    await createVehicleMutationCommands(options).submit(submitEvent());

    expect(api.upsertVehicleExternalMapping).toHaveBeenCalledWith(created.id, draft.externalMapping);
    expect(api.createVehicle).toHaveBeenCalledWith(expect.objectContaining({ externalMapping: draft.externalMapping }));
    expect(api.importDigitalCenterCVs).toHaveBeenCalledWith("session-1", "item-77", created.id);
    expect(createCV).not.toHaveBeenCalled();
    expect(options.ecos.markSaved).toHaveBeenCalledWith(draft, created.id);
    expect(options.editor.close).toHaveBeenCalledOnce();
    expect(options.ecos.returnToSession).toHaveBeenCalledWith(draft);
  });

  it("localizes an ECoS mapping ownership conflict", async () => {
    const options = commandOptions();
    const created = vehicleFixture({ id: "created" });
    options.ecos.draft = {
      source: "ecos", mode: "create",
      sourceSummary: { objectId: 77, name: "BR 106", address: "3", protocol: "DCC", profile: "" },
      vehicle: { ...emptyVehicle, name: "BR 106" }, importedKeys: ["name"],
      externalMapping: { provider: "ecos", externalId: "77" },
      cvValues: [], functionValues: [], unclearFields: []
    };
    vi.spyOn(api, "createVehicle").mockResolvedValue(created);
    vi.spyOn(api, "upsertVehicleExternalMapping").mockRejectedValue(
      new ApiError("conflict", "external_mapping_conflict", 409)
    );
    vi.spyOn(api, "vehicle").mockResolvedValue(created);

    await createVehicleMutationCommands(options).submit(submitEvent());

    expect(options.onMessage).toHaveBeenCalledWith("vehicles.ecosDraft.mappingConflict");
    expect(options.reloadVehicles).toHaveBeenCalledOnce();
  });

  it("does not perform a second mapping write after an atomic mapped creation", async () => {
    const options = commandOptions();
    const mapping = { id: "m", vehicleId: "created", provider: "cs3", externalId: "77", syncStatus: "linked",
      createdAt: "before", updatedAt: "before" };
    const created = vehicleFixture({ id: "created", externalMappings: [mapping] });
    options.ecos.draft = { source: "cs3", mode: "create", sourceSummary: { objectId: 77, name: "Source", address: "3", protocol: "DCC", profile: "" },
      vehicle: options.editor.form, importedKeys: ["name"], externalMapping: { provider: "cs3", externalId: "77" },
      cvValues: [], functionValues: [], unclearFields: [] };
    vi.spyOn(api, "createVehicle").mockResolvedValue(created); vi.spyOn(api, "vehicle").mockResolvedValue(created);
    const extraWrite = vi.spyOn(api, "upsertVehicleExternalMapping");
    await createVehicleMutationCommands(options).submit(submitEvent());
    expect(extraWrite).not.toHaveBeenCalled();
    expect(options.ecos.markSaved).toHaveBeenCalledWith(options.ecos.draft, created.id);
  });

  it("retains a rejected source draft without acknowledging another vehicle creation", async () => {
    const options = commandOptions();
    vi.spyOn(api, "createVehicle").mockRejectedValue(new ApiError("Source already linked", "external_mapping_conflict", 409));
    await createVehicleMutationCommands(options).submit(submitEvent());
    expect(options.editor.setSelectedDetail).not.toHaveBeenCalled();
    expect(options.functions.acknowledgeSaved).not.toHaveBeenCalled();
    expect(options.editor.close).not.toHaveBeenCalled();
    expect(options.reloadVehicles).toHaveBeenCalledOnce();
    expect(options.onMessage).toHaveBeenLastCalledWith("vehicles.ecosDraft.mappingConflict");
  });

  it("deletes the selected candidate and closes its editor", async () => {
    const candidate = vehicleFixture({ id: "delete-me" });
    const options = commandOptions();
    options.editor.selected = candidate;
    options.deletion.candidate = candidate;
    vi.spyOn(api, "deleteVehicle").mockResolvedValue(undefined);

    createVehicleMutationCommands(options).confirmDelete();
    await vi.waitFor(() => expect(options.reloadVehicles).toHaveBeenCalledOnce());

    expect(api.deleteVehicle).toHaveBeenCalledWith(candidate.id);
    expect(options.editor.close).toHaveBeenCalledOnce();
    expect(options.deletion.setCandidate).toHaveBeenCalledWith(null);
  });

  it("submits function edits/deletions in the vehicle request without per-row writes", async () => {
    const options = commandOptions(), saved = vehicleFixture({ id: "saved" });
    const changes: VehicleFunctionChange[] = [
      { functionKey: "F0", expectedRevision: "a".repeat(64), value: { name: "Manual", mode: "moment", flagsVersion: 2 } },
      { functionKey: "F3", expectedRevision: "b".repeat(64), delete: true }
    ];
    options.functions.changes.mockReturnValue(changes);
    vi.spyOn(api, "createVehicle").mockResolvedValue(saved);
    vi.spyOn(api, "vehicle").mockResolvedValue(saved);
    const perRow = vi.spyOn(api, "updateVehicleFunction"), deleteRow = vi.spyOn(api, "deleteVehicleFunction");
    await createVehicleMutationCommands(options).submit(submitEvent());
    expect(api.createVehicle).toHaveBeenCalledWith(expect.objectContaining({ functionChanges: changes }));
    expect(options.functions.acknowledgeSaved).toHaveBeenCalledWith(saved.functions, saved.id);
    expect(perRow).not.toHaveBeenCalled(); expect(deleteRow).not.toHaveBeenCalled();
  });

  it("preserves the draft and selection when the atomic request conflicts", async () => {
    const options = commandOptions();
    options.functions.changes.mockReturnValue([{ functionKey: "F1", expectedRevision: "a".repeat(64), value: { name: "Draft" } }]);
    vi.spyOn(api, "createVehicle").mockRejectedValue(new ApiError("Conflict", "vehicle_function_conflict", 409));
    await createVehicleMutationCommands(options).submit(submitEvent());
    expect(options.onMessage).toHaveBeenLastCalledWith("vehicles.functions.conflict");
    expect(options.functions.acknowledgeSaved).not.toHaveBeenCalled();
    expect(options.editor.setSelectedDetail).not.toHaveBeenCalled();
    expect(options.editor.close).not.toHaveBeenCalled();
    expect(options.editor.setSaving).toHaveBeenLastCalledWith(false);
  });

  it("rejects overlong UTF-8 notes before any vehicle write", async () => {
    const options = commandOptions();
    options.functions.changes.mockReturnValue([{ functionKey: "F1", expectedRevision: "", value: { notes: "ä".repeat(501) } }]);
    const create = vi.spyOn(api, "createVehicle");
    await createVehicleMutationCommands(options).submit(submitEvent());
    expect(create).not.toHaveBeenCalled();
    expect(options.editor.setActiveTab).toHaveBeenCalledWith("control");
    expect(options.onMessage).toHaveBeenCalledWith("vehicles.functions.invalidFields");
  });
});
