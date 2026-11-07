import { invalidatePendingRequests } from "../../shared/requestGeneration";
import { useCallback, useEffect, useRef, useState, type SetStateAction } from "react";
import { api, ApiError, type Vehicle, type VehicleImage } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { attachmentCategoryForFile, isAllowedImageFile, isBlockedAttachmentFile, isBlockedCVFile } from "./vehicleFiles";
import { createVehicleUploadBatch, maxVehicleUploadFiles, retryableUploadItem, validImagesRevision,
  vehicleUploadCounts, vehicleUploadFailure, type VehicleUploadBatch, type VehicleUploadItem,
  type VehicleUploadKind, type VehicleUploadMetadata } from "./vehicleUploadBatch";
import { CVWorkingGroupUnavailable, cvWorkingGroupLabel, refreshCVInputScope, type CVWorkingGroupOptions } from "./cvWorkingGroup";

type Options = {
  selected: Vehicle | null;
  readonly: boolean;
  saving: boolean;
  setSaving: (value: boolean) => void;
  onMessage: (message: string) => void;
  onStatus: (message: string) => void;
  onImageStored: (image: VehicleImage) => void;
  refreshSaved: (vehicleId: string, kind: VehicleUploadKind) => Promise<void>;
  reload: (vehicleId: string) => Promise<void>;
  workingGroup?: CVWorkingGroupOptions;
};

export function useVehicleUploads(options: Options) {
  const { t } = useI18n();
  const optionsRef = useRef(options); optionsRef.current = options;
  const [batch, setBatchState] = useState<VehicleUploadBatch | null>(null);
  const batchRef = useRef(batch), busy = useRef(false), generation = useRef(0), alive = useRef(true);
  useEffect(() => { alive.current = true; return () => { alive.current = false; invalidatePendingRequests(generation); }; }, []);
  const setBatch = useCallback((action: SetStateAction<VehicleUploadBatch | null>) => {
    const next = typeof action === "function" ? action(batchRef.current) : action;
    batchRef.current = next;
    setBatchState(next);
  }, []);
  const patchItem = (key: string, patch: Partial<VehicleUploadItem>) => setBatch(current => current ? {
    ...current, items: current.items.map(item => item.key === key ? { ...item, ...patch } : item)
  } : current);

  const reset = () => {
    generation.current++;
    if (busy.current) optionsRef.current.setSaving(false);
    busy.current = false;
    setBatch(null);
  };

  const store = async (item: VehicleUploadItem, value: VehicleUploadBatch) => {
    const { vehicleId, metadata } = value;
    if (value.kind === "image") {
      const image = await api.uploadVehicleImage(vehicleId, item.file, item.file.name,
        Boolean(metadata.isPrimary), metadata.maintenanceId || "", metadata.imagesRevision);
      return { id: image.id, image, imagesRevision: image.imagesRevision };
    }
    const record = value.kind === "decoder"
      ? metadata.workspaceId
        ? await api.uploadVehicleCVFile(vehicleId, item.file, metadata.decoderProfile || "", metadata.description || "", metadata.workspaceId, metadata.expectedWorkspaceRevision)
        : await api.uploadVehicleCVFile(vehicleId, item.file, metadata.decoderProfile || "", metadata.description || "")
      : await api.uploadVehicleAttachment(vehicleId, item.file,
        metadata.category || attachmentCategoryForFile(item.file), metadata.description || "", metadata.maintenanceId || "");
    return { id: record.id, image: undefined, imagesRevision: undefined };
  };

  const run = async () => {
    const original = batchRef.current, opts = optionsRef.current;
    if (!original || busy.current || opts.saving || opts.readonly || original.needsReview ||
      opts.selected?.id !== original.vehicleId || !original.items.some(retryableUploadItem)) return;
    const version = ++generation.current;
    const active = () => alive.current && generation.current === version &&
      optionsRef.current.selected?.id === original.vehicleId && batchRef.current?.id === original.id &&
      (original.ownerKey === undefined || optionsRef.current.workingGroup?.ownerKey === original.ownerKey);
    busy.current = true;
    opts.setSaving(true); opts.onMessage("");
    setBatch(current => current ? { ...current, phase: "uploading", notice: undefined } : current);
    let saved = 0;
    try {
      for (const originalItem of original.items.filter(retryableUploadItem)) {
        if (!active() || optionsRef.current.readonly) break;
        const current = batchRef.current;
        if (!current) break;
        const item = current.items.find(value => value.key === originalItem.key);
        if (!item || !retryableUploadItem(item)) continue;
        patchItem(item.key, { status: "uploading", error: undefined });
        try {
          const stored = await store(item, current);
          if (!active()) break;
          if (!stored.id) throw new Error(t("vehicles.mediaUpload.incomplete"));
          patchItem(item.key, { status: "saved", storedId: stored.id }); saved++;
          if (stored.image) {
            optionsRef.current.onImageStored(stored.image);
            if (!validImagesRevision(stored.imagesRevision)) {
              setBatch(value => value ? { ...value, needsReview: true, notice: t("vehicles.mediaUpload.missingRevision") } : value);
              break;
            }
            setBatch(value => value ? { ...value, metadata: { ...value.metadata, imagesRevision: stored.imagesRevision } } : value);
          }
        } catch (error) {
          if (!active()) break;
          if (batchRef.current?.items.find(value => value.key === item.key)?.status === "saved") {
            setBatch(value => value ? { ...value, needsReview: true, notice: t("vehicles.mediaUpload.refreshFailed") } : value);
            break;
          }
          const groupChanged = original.kind === "decoder" && Boolean(original.metadata.workspaceId) && error instanceof ApiError && error.code === "cv_file_conflict";
          const changed = groupChanged || error instanceof ApiError && error.code === "vehicle_image_conflict";
          patchItem(item.key, { status: vehicleUploadFailure(error),
            error: changed ? t("vehicles.mediaConflict.changed") : error instanceof Error ? error.message : String(error) });
          if (changed) setBatch(value => value ? { ...value, needsReview: true, notice: t(groupChanged ? "cvScope.fileBindingChanged" : "vehicles.mediaUpload.reviewAgain") } : value);
          break;
        }
      }
      if (active()) {
        const current = batchRef.current;
        try {
          if (saved) await optionsRef.current.refreshSaved(original.vehicleId, original.kind);
          else if (current && (current.needsReview || vehicleUploadCounts(current).unknown)) await optionsRef.current.reload(original.vehicleId);
        } catch {
          if (active()) setBatch(value => value ? { ...value, notice: t("vehicles.mediaUpload.refreshFailed") } : value);
        }
      }
    } finally {
      if (alive.current && generation.current === version) {
        busy.current = false;
        setBatch(value => value ? { ...value, phase: "result" } : value);
        optionsRef.current.setSaving(false);
      }
    }
  };

  const prepare = (kind: VehicleUploadKind, files: File[], metadata: VehicleUploadMetadata, confirmed = false) => {
    const opts = optionsRef.current;
    if (!opts.selected || opts.readonly || opts.saving || busy.current || batchRef.current || !files.length) return false;
    if (files.length > maxVehicleUploadFiles) {
      opts.onMessage(t("vehicles.mediaUpload.limit", { count: maxVehicleUploadFiles })); return false;
    }
    const invalid = files.find(file => kind === "image" ? !isAllowedImageFile(file)
      : kind === "decoder" ? isBlockedCVFile(file) : isBlockedAttachmentFile(file));
    if (invalid) { opts.onMessage(t(`vehicles.mediaUpload.invalid.${kind}`, { name: invalid.name })); return false; }
    if (kind === "image" && !validImagesRevision(opts.selected.imagesRevision)) {
      opts.onMessage(t("vehicles.mediaConflict.missingRevision")); return false;
    }
    opts.onMessage("");
    const next = createVehicleUploadBatch(kind, opts.selected, files, {
      ...metadata, ...(kind === "image" ? { imagesRevision: opts.selected.imagesRevision } : {})
    });
    setBatch({ ...next, ...(opts.workingGroup ? { ownerKey: opts.workingGroup.ownerKey } : {}) });
    if (confirmed) void run();
    return true;
  };

  const reviewWorkingGroup = async () => {
    const original = batchRef.current, opts = optionsRef.current;
    if (!original || original.kind !== "decoder" || !original.metadata.workspaceId || !original.needsReview ||
      busy.current || opts.saving || opts.readonly || !opts.workingGroup || opts.selected?.id !== original.vehicleId) return;
    const version = ++generation.current;
    const active = () => alive.current && version === generation.current && batchRef.current?.id === original.id &&
      optionsRef.current.workingGroup?.ownerKey === original.ownerKey;
    busy.current = true; opts.setSaving(true); setBatch(current => current ? { ...current, phase: "reviewing" } : current);
    try {
      const context = await opts.workingGroup.reload();
      if (!active()) return;
      if (!context) throw new CVWorkingGroupUnavailable();
      const scope = refreshCVInputScope({ cvNumber: 1, value: 0, decoderProfile: original.metadata.decoderProfile || "",
        workspaceId: original.metadata.workspaceId, expectedWorkspaceRevision: original.metadata.expectedWorkspaceRevision }, context.workspaces);
      const label = cvWorkingGroupLabel(context, scope.workspaceId, scope.decoderProfile, t);
      setBatch(current => current ? { ...current, metadata: { ...current.metadata, expectedWorkspaceRevision: scope.expectedWorkspaceRevision,
        workingGroupLabel: label, ...(original.metadata.workingGroupLabel !== label ? { previousWorkingGroupLabel: original.metadata.workingGroupLabel } : {}) },
        needsReview: false, notice: t("cvScope.fileReviewed") } : current);
    } catch {
      if (active()) setBatch(current => current ? { ...current, notice: t("cvScope.unavailable") } : current);
    } finally {
      if (version === generation.current) {
        busy.current = false; optionsRef.current.setSaving(false);
        setBatch(current => current ? { ...current, phase: "result" } : current);
      }
    }
  };

  const skip = (key: string) => {
    if (busy.current) return;
    const item = batchRef.current?.items.find(value => value.key === key);
    if (item && retryableUploadItem(item)) patchItem(key, { status: "skipped" });
  };
  const close = () => {
    if (busy.current) return;
    const current = batchRef.current;
    if (current?.phase === "result") {
      const counts = vehicleUploadCounts(current);
      if (counts.failed || counts.unknown || counts.pending) optionsRef.current.onMessage(t("vehicles.mediaUpload.partial", counts));
      else if (counts.saved) optionsRef.current.onStatus(t("vehicles.mediaUpload.completed", counts));
    }
    setBatch(null);
  };
  return { batch, commands: { prepare, confirm: run, reviewWorkingGroup, skip, close, reset } };
}
