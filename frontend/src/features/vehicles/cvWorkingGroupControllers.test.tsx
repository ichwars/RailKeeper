import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, type DecoderConfigurationContext, type DecoderCVWorkspace } from "../../shared/api";
import { cvFileFixture, cvValueFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { cvWorkingTarget, type CVWorkingGroupOptions } from "./cvWorkingGroup";
import { useVehicleCVController } from "./useVehicleCVController";
import { useVehicleCVImportController } from "./useVehicleCVImportController";
import { useVehicleDecoderFilesController } from "./useVehicleDecoderFilesController";
import { useVehicleUploads } from "./useVehicleUploads";
import { useVehicleMediaDeletion } from "./useVehicleMediaDeletion";

const group = (id: string, profile: string, revision = "a"): DecoderCVWorkspace => ({ id, profile,
  vehicleId: "vehicle-1", origin: "legacy", installationId: null, createdAt: "", updatedAt: "",
  valueCount: 1, sourceCount: 1, contentRevision: "c".repeat(64), bindingRevision: revision.repeat(64) });
const a = group("raw", " Legacy "), b = group("normal", "Legacy", "b");
const context = (groups = [a, b]): DecoderConfigurationContext => ({ vehicleId: "vehicle-1", vehicleName: "Test",
  inventoryNumber: "1", instances: [], installations: [], workspaces: groups, snapshots: [], revision: "d".repeat(64) });
const working = (target = a, groups = [a, b]): CVWorkingGroupOptions => ({ ownerKey: "editor:1", context: context(groups),
  groups, target: cvWorkingTarget(target), ready: true, reload: vi.fn().mockResolvedValue(context(groups)) });
const fileList = (...files: File[]) => ({ ...files, length: files.length, item: (index: number) => files[index] || null }) as unknown as FileList;
const callbacks = () => ({ setSaving: vi.fn(), onMessage: vi.fn(), onStatus: vi.fn(),
  readonly: false, saving: false, refreshSelectedVehicle: vi.fn().mockResolvedValue(undefined),
  reloadSelectedVehicle: vi.fn().mockResolvedValue(undefined) });

describe("confirmed working groups across independent CV actions", () => {
  afterEach(() => vi.restoreAllMocks());

  it("retains the open nondefault target when a committed temporary draft ID is cleared", async () => {
    const selected = vehicleFixture({ cvValues: [] }), cb = callbacks();
    vi.spyOn(api, "applyVehicleCVChanges").mockResolvedValue({ changed: 1 });
    const { result } = renderHook(() => useVehicleCVController({ ...cb, selected,
      decoderNumber: "", workingGroup: working() }));
    act(() => result.current.commands.updateForm({ cvNumber: 2, value: 7, decoderProfile: b.profile }));
    act(() => result.current.commands.save());
    act(() => result.current.commands.edit(result.current.state.manualDraft.values[0]));
    expect(result.current.state.form).toMatchObject({ decoderProfile: b.profile, workspaceId: b.id });
    await act(() => result.current.commands.applyManualDraft());
    expect(result.current.state.form).toMatchObject({ decoderProfile: b.profile, workspaceId: b.id,
      expectedWorkspaceRevision: b.bindingRevision, cvNumber: 2, value: 7 });
    act(() => result.current.commands.save());
    expect(result.current.state.manualDraft.changes[0].value).toMatchObject({
      decoderProfile: b.profile, workspaceId: b.id, expectedWorkspaceRevision: b.bindingRevision
    });
  });

  it("retains staged and open original targets when the new-input default changes", async () => {
    const stored = cvValueFixture({ decoderProfile: a.profile });
    const selected = vehicleFixture({ cvValues: [stored] }), cb = callbacks();
    const request = vi.spyOn(api, "applyVehicleCVChanges").mockResolvedValue({ changed: 1 });
    const { result, rerender } = renderHook(({ workingGroup }) => useVehicleCVController({ ...cb,
      selected, decoderNumber: "", workingGroup }), { initialProps: { workingGroup: working() } });
    act(() => { result.current.commands.edit(stored); });
    act(() => result.current.commands.updateForm({ value: 9 }));
    act(() => result.current.commands.save());
    act(() => result.current.commands.edit(stored));
    act(() => result.current.commands.updateForm({ description: "Keep open" }));
    rerender({ workingGroup: working(b) });
    expect(result.current.state.form).toMatchObject({ decoderProfile: a.profile, workspaceId: a.id, description: "Keep open" });
    expect(result.current.state.manualDraft.changes[0]).toMatchObject({ scope: { workspaceId: a.id }, value: { workspaceId: a.id, value: 9 } });
    await act(() => result.current.commands.applyManualDraft());
    expect(request).toHaveBeenCalledWith(selected.id, [expect.objectContaining({ workspaceId: a.id,
      expectedWorkspaceRevision: a.bindingRevision, value: expect.objectContaining({ decoderProfile: a.profile, workspaceId: a.id }) })]);
    expect(result.current.state.form.description).toBe("Keep open");
  });

  it("compares both sides of a move and retains a scoped deletion in the same batch", async () => {
    const stored = cvValueFixture({ decoderProfile: a.profile });
    const removed = cvValueFixture({ id: "remove", cvNumber: 2, decoderProfile: b.profile });
    const selected = vehicleFixture({ cvValues: [stored, removed] }), cb = callbacks();
    const request = vi.spyOn(api, "applyVehicleCVChanges").mockResolvedValue({ changed: 2 });
    const { result } = renderHook(() => useVehicleCVController({ ...cb, selected, decoderNumber: "", workingGroup: working() }));
    act(() => result.current.commands.edit(stored));
    act(() => result.current.commands.updateForm({ decoderProfile: b.profile, value: 8 }));
    act(() => result.current.commands.save());
    act(() => result.current.commands.remove(removed));
    await act(() => result.current.commands.applyManualDraft());
    expect(request).toHaveBeenCalledWith(selected.id, expect.arrayContaining([
      expect.objectContaining({ id: stored.id, workspaceId: a.id, expectedWorkspaceRevision: a.bindingRevision,
        value: expect.objectContaining({ workspaceId: b.id, expectedWorkspaceRevision: b.bindingRevision }) }),
      { id: removed.id, expectedRevision: removed.revision, delete: true, workspaceId: b.id, expectedWorkspaceRevision: b.bindingRevision }
    ]));
  });

  it("blocks a changed binding and explicitly refreshes even a new unsaved row", async () => {
    const cb = callbacks(), selected = vehicleFixture({ cvValues: [] });
    const request = vi.spyOn(api, "applyVehicleCVChanges").mockResolvedValue({ changed: 1 });
    const fresh = group(a.id, a.profile, "e"), next = working(b, [fresh, b]);
    vi.spyOn(api, "vehicleCVValues").mockResolvedValue([]);
    const { result, rerender } = renderHook(({ workingGroup }) => useVehicleCVController({ ...cb, selected,
      decoderNumber: "", workingGroup }), { initialProps: { workingGroup: working() } });
    act(() => result.current.commands.updateForm({ cvNumber: 19, value: 9 }));
    act(() => result.current.commands.save());
    const id = result.current.state.manualDraft.changes[0].id;
    rerender({ workingGroup: next });
    expect(result.current.state.manualDraft).toMatchObject({ bindingChanged: true, conflicts: [id] });
    await act(() => result.current.commands.applyManualDraft()); expect(request).not.toHaveBeenCalled();
    await act(() => result.current.commands.rebaseManualChange(id));
    expect(result.current.state.manualDraft.conflicts).toEqual([]);
    expect(result.current.state.manualDraft.changes[0].value).toMatchObject({ cvNumber: 19, value: 9,
      decoderProfile: a.profile, workspaceId: a.id, expectedWorkspaceRevision: fresh.bindingRevision });
    expect(request).not.toHaveBeenCalled();
  });

  it("strips file comparisons and keeps the prepared target through a different default and explicit comparison", async () => {
    const cb = callbacks(), selected = vehicleFixture({ cvValues: [] });
    const request = vi.spyOn(api, "importVehicleCVValues").mockRejectedValueOnce(new ApiError("changed", "cv_value_conflict", 409))
      .mockResolvedValue({ imported: 1 });
    const fresh = group(a.id, a.profile, "e"), next = working(b, [fresh, b]);
    vi.spyOn(api, "vehicleCVValues").mockResolvedValue([]);
    const { result, rerender } = renderHook(({ workingGroup }) => useVehicleCVImportController({ ...cb, selected,
      workingGroup, refreshSaved: cb.refreshSelectedVehicle }), { initialProps: { workingGroup: working() } });
    const file = new File([JSON.stringify([{ cvNumber: 21, value: 7, decoderProfile: "Source",
      workspaceId: b.id, expectedWorkspaceRevision: b.bindingRevision, expectedRevision: "forged" }])], "source.json");
    await act(() => result.current.commands.importValues(fileList(file)));
    rerender({ workingGroup: next });
    expect(result.current.state.preview?.rows[0]).toMatchObject({ sourceProfile: "Source", input: { decoderProfile: a.profile,
      workspaceId: a.id, expectedWorkspaceRevision: a.bindingRevision } });
    await act(() => result.current.commands.applyImportPreview());
    await act(() => result.current.commands.applyImportPreview()); expect(request).toHaveBeenCalledOnce();
    await act(() => result.current.commands.recompareImportPreview());
    expect(result.current.state.preview?.workingTarget?.workspaceId).toBe(a.id);
    expect(result.current.state.preview?.rows[0].input.expectedWorkspaceRevision).toBe(fresh.bindingRevision);
    expect(result.current.state.preview?.rows[0].sourceProfile).toBe("Source");
    expect(request).toHaveBeenCalledOnce();
    await act(() => result.current.commands.applyImportPreview());
    expect(request).toHaveBeenLastCalledWith(selected.id, [expect.objectContaining({ workspaceId: a.id,
      expectedWorkspaceRevision: fresh.bindingRevision, expectedRevision: "" })]);
  });

  it("freezes file-upload scope before preview and compares deletion against its own original group", async () => {
    vi.spyOn(api, "previewVehicleCVFile").mockResolvedValue({ fileName: "source.csv", sizeBytes: 32,
      mimeType: "text/csv", hasMetadata: true, suggestedDecoderProfile: "Source", suggestedDescription: "Keep note",
      suggestedCvValues: [{ cvNumber: 1, value: 7 }] });
    const cb = callbacks(), selected = vehicleFixture(), prepareUpload = vi.fn().mockReturnValue(true), prepareDelete = vi.fn(), onImportPreview = vi.fn();
    const { result, rerender } = renderHook(({ workingGroup }) => useVehicleDecoderFilesController({ ...cb, selected,
      workingGroup, prepareUpload, prepareDelete, onImportPreview, onFunctionSuggestions: vi.fn() }),
      { initialProps: { workingGroup: working() } });
    const file = new File(["CV;Value;decoderProfile\n1;7;Source"], "source.csv");
    await act(() => result.current.commands.uploadFiles(fileList(file)));
    rerender({ workingGroup: working(b) });
    act(() => result.current.commands.applyFirstSuggestion());
    await act(() => result.current.commands.previewValuesForImport());
    expect(onImportPreview).toHaveBeenCalledWith(expect.objectContaining({ rows: [expect.objectContaining({ sourceProfile: "Source",
      input: expect.objectContaining({ decoderProfile: a.profile, workspaceId: a.id }) })] }));
    act(() => result.current.commands.confirmUpload());
    expect(prepareUpload).toHaveBeenCalledWith("decoder", [file], expect.objectContaining({ decoderProfile: a.profile,
      description: "Keep note", workspaceId: a.id, expectedWorkspaceRevision: a.bindingRevision }), true);
    act(() => result.current.commands.remove(cvFileFixture({ decoderProfile: a.profile })));
    expect(prepareDelete).toHaveBeenCalledWith(expect.objectContaining({ scope: { workspaceId: a.id,
      expectedWorkspaceRevision: a.bindingRevision } }));
  });

  it("requires a separate upload confirmation after a stale binding, retaining files and notes", async () => {
    const cb = callbacks(), selected = vehicleFixture(), fresh = group(a.id, a.profile, "e");
    const workingGroup = working(b, [fresh, b]), request = vi.spyOn(api, "uploadVehicleCVFile")
      .mockRejectedValueOnce(new ApiError("changed", "cv_file_conflict", 409)).mockResolvedValue(cvFileFixture());
    const { result } = renderHook(() => useVehicleUploads({ ...cb, selected, workingGroup, onImageStored: vi.fn(),
      refreshSaved: vi.fn(), reload: vi.fn() }));
    const files = [new File(["original bytes"], "source.csv")];
    act(() => { result.current.commands.prepare("decoder", files, { decoderProfile: a.profile, description: "Keep note",
      workspaceId: a.id, expectedWorkspaceRevision: a.bindingRevision, workingGroupLabel: "Old hardware" }); });
    await act(() => result.current.commands.confirm());
    expect(result.current.batch?.needsReview).toBe(true);
    await act(() => result.current.commands.confirm()); expect(request).toHaveBeenCalledOnce();
    await act(() => result.current.commands.reviewWorkingGroup());
    expect(request).toHaveBeenCalledOnce();
    expect(result.current.batch).toMatchObject({ needsReview: false, phase: "result", metadata: { decoderProfile: a.profile,
      description: "Keep note", workspaceId: a.id, expectedWorkspaceRevision: fresh.bindingRevision, previousWorkingGroupLabel: "Old hardware" } });
    expect(result.current.batch?.items[0].file).toBe(files[0]);
    await act(() => result.current.commands.confirm());
    expect(request).toHaveBeenLastCalledWith(selected.id, files[0], a.profile, "Keep note", a.id, fresh.bindingRevision);
  });

  it("carries the frozen file deletion comparison without following the new selection", async () => {
    const cb = callbacks(), selected = vehicleFixture(), request = vi.spyOn(api, "deleteVehicleCVFile").mockResolvedValue(undefined);
    const { result } = renderHook(() => useVehicleMediaDeletion({ ...cb, selected, onImageDeleted: vi.fn(), refreshSaved: vi.fn(), reload: vi.fn() }));
    const file = cvFileFixture({ decoderProfile: a.profile });
    act(() => result.current.commands.prepare({ kind: "decoder", record: file,
      scope: { workspaceId: a.id, expectedWorkspaceRevision: a.bindingRevision }, scopeLabel: "Original hardware" }));
    await act(() => result.current.commands.confirm());
    expect(request).toHaveBeenCalledWith(selected.id, file.id, file.revision, { workspaceId: a.id, expectedWorkspaceRevision: a.bindingRevision });
  });
});
