import { useVehicleImageDraftController } from "./useVehicleImageDraftController";
import { type DragEvent, type SetStateAction, useCallback, useRef, useState } from "react";

import { api, ApiError, type Vehicle, type VehicleAttachment, type VehicleImage } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { activeMediaConflicts, refreshMediaDraft, resolveMediaImage, type MediaDraftConflict, type MediaMetadataDraft } from "./mediaDraftConflicts";
import { attachmentCategoryForFile, isAllowedImageFile, isBlockedAttachmentFile } from "./vehicleFiles";
import {
  attachmentsToEditState,
  type AttachmentEditState,
  type PendingArticleImage,
  uploadedImageToPending,
  vehicleImagesToPending
} from "./vehicleTransforms";

type UseVehicleMediaControllerOptions = {
  selected: Vehicle | null;
  readonly: boolean;
  saving: boolean;
  setSaving: (saving: boolean) => void;
  onMessage: (message: string) => void;
  refreshSelectedVehicle: (vehicleId?: string) => Promise<void>;
  reloadSelectedVehicle?: (vehicleId?: string) => Promise<void>;
  onImageUploadComplete: () => void;
};

export function useVehicleMediaController({
  selected,
  readonly,
  saving,
  setSaving,
  onMessage,
  refreshSelectedVehicle,
  reloadSelectedVehicle,
  onImageUploadComplete
}: UseVehicleMediaControllerOptions) {
  const { t } = useI18n();
  const imageDraft = useVehicleImageDraftController();
  const { pendingImages, previewImage } = imageDraft.state;
  const { setPendingImages, setPreviewImage } = imageDraft.setters;
  const { addImages, setPrimaryPendingImage, updatePendingImageTitle,
    updatePendingImageMaintenance, movePendingImage } = imageDraft.commands;
  const [attachmentDeleteCandidate, setAttachmentDeleteCandidate] = useState<VehicleAttachment | null>(null);
  const [metadata, setMetadataState] = useState<MediaMetadataDraft>({ edits: {}, conflicts: [] });
  const metadataRef = useRef(metadata);
  const setMetadata = useCallback((action: SetStateAction<MediaMetadataDraft>) => {
    const next = typeof action === "function" ? action(metadataRef.current) : action;
    metadataRef.current = next;
    setMetadataState(next);
  }, []);
  const setAttachmentEdits = useCallback((action: SetStateAction<AttachmentEditState>) => setMetadata(current => ({
    ...current, edits: typeof action === "function" ? action(current.edits) : action
  })), [setMetadata]);
  const attachmentEdits = metadata.edits;
  const [imageUploadMaintenanceId, setImageUploadMaintenanceId] = useState("");
  const [attachmentUploadCategory, setAttachmentUploadCategory] = useState("");
  const [attachmentUploadDescription, setAttachmentUploadDescription] = useState("");
  const [attachmentDragActive, setAttachmentDragActive] = useState(false);
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const attachmentInputRef = useRef<HTMLInputElement | null>(null);

  const reset = (clearPreview = false) => {
    imageDraft.commands.reset(clearPreview);
    setMetadata({ edits: {}, conflicts: [] });
    setImageUploadMaintenanceId("");
    setAttachmentUploadCategory("");
    setAttachmentUploadDescription("");
    setAttachmentDragActive(false);
    setAttachmentDeleteCandidate(null);
  };

  const loadDetail = (detail: Vehicle) => {
    imageDraft.commands.loadDetail(detail);
    setMetadata({ edits: attachmentsToEditState(detail.attachments), conflicts: [] });
  };

  const refreshAttachments = (detail: Vehicle, previous: Vehicle) => {
    const refreshed = refreshMediaDraft(imageDraft.refs.pendingImagesRef.current, metadataRef.current, previous, detail);
    setMetadata({ edits: refreshed.state.edits, conflicts: [
      ...metadataRef.current.conflicts.filter(value => value.kind === "image"),
      ...refreshed.state.conflicts.filter(value => value.kind === "attachment")
    ] });
  };

  const refreshDetail = (detail: Vehicle, previous: Vehicle) => {
    const refreshed = refreshMediaDraft(imageDraft.refs.pendingImagesRef.current, metadataRef.current, previous, detail);
    setPendingImages(refreshed.images);
    setMetadata(refreshed.state);
  };

  const resolveConflict = (conflict: MediaDraftConflict, choice: "draft" | "saved") => {
    if (!selected || readonly || saving) return;
    if (conflict.kind === "image") {
      setPendingImages(current => resolveMediaImage(current, conflict, selected, choice));
    } else if (choice === "saved" || conflict.deleted) {
      setAttachmentEdits(current => {
        const next = { ...current }, saved = attachmentsToEditState(selected.attachments)[conflict.id];
        if (!saved) delete next[conflict.id];
        else if (next[conflict.id]) next[conflict.id] = { ...next[conflict.id],
          ...Object.fromEntries(conflict.fields.map(field => [field.field, Reflect.get(saved, field.field)])) };
        return next;
      });
    }
    setMetadata(current => ({ ...current, conflicts: current.conflicts.filter(value =>
      value.kind !== conflict.kind || value.id !== conflict.id) }));
    onMessage("");
  };

  const handleMediaError = async (error: unknown) => {
    const conflict = error instanceof ApiError && ["image_conflict", "vehicle_image_conflict", "attachment_conflict"].includes(error.code);
    if (conflict && selected) await reloadSelectedVehicle?.(selected.id);
    onMessage(conflict ? t("vehicles.mediaConflict.changed") : error instanceof Error ? error.message : String(error));
  };

  const acknowledgeCoreSaved = (detail: Vehicle, previous: Vehicle, phase: "core" | "complete") => {
    setPendingImages(current => [
      ...vehicleImagesToPending(detail),
      ...(phase === "core" ? current.filter(image => !image.persisted && /^https?:\/\//i.test(image.url)) : [])
    ]);
    refreshAttachments(detail, previous);
    setMetadata(current => ({ ...current, conflicts: current.conflicts.filter(value => value.kind !== "image") }));
  };

  const acknowledgeImportedImage = (pending: PendingArticleImage, image: VehicleImage) => {
    setPendingImages(current => current.map(value => value.id === pending.id ? uploadedImageToPending(image)
      : image.isPrimary && value.persisted ? { ...value, isPrimary: false } : value));
  };

  const removePendingImage = (image: PendingArticleImage) => {
    if (readonly || saving) return;
    if (image.maintenanceId) {
      onMessage(t("vehicles.uploads.maintenanceLinkedCannotDelete"));
      return;
    }

    const removeFromState = () => imageDraft.commands.removeFromState(image.id);

    if (selected && image.persisted) {
      const revision = selected.images?.find(value => value.id === image.id)?.revision;
      if (!revision) { onMessage(t("vehicles.mediaConflict.missingRevision")); return; }
      setSaving(true);
      api.deleteVehicleImage(selected.id, image.id, revision)
        .then(() => {
          removeFromState();
          return refreshSelectedVehicle(selected.id);
        })
        .catch(handleMediaError)
        .finally(() => setSaving(false));
      return;
    }
    removeFromState();
  };

  const uploadImages = (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const uploadFiles = Array.from(files);
    const invalid = uploadFiles.find((file) => !isAllowedImageFile(file));
    if (invalid) {
      onMessage(`${invalid.name} ist kein erlaubtes Bildformat.`);
      if (imageInputRef.current) imageInputRef.current.value = "";
      return;
    }

    setSaving(true);
    onMessage("");
    (async () => {
      for (const file of uploadFiles) {
        const image = await api.uploadVehicleImage(
          selected.id,
          file,
          file.name,
          pendingImages.length === 0,
          imageUploadMaintenanceId
        );
        addImages([uploadedImageToPending(image)]);
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .then(onImageUploadComplete)
      .catch(handleMediaError)
      .finally(() => {
        setSaving(false);
        if (imageInputRef.current) imageInputRef.current.value = "";
      });
  };

  const uploadAttachment = (files: FileList | null) => {
    if (!selected || !files || files.length === 0) return;
    const uploadFiles = Array.from(files);
    const blocked = uploadFiles.find(isBlockedAttachmentFile);
    if (blocked) {
      onMessage(
        `${blocked.name} ist als Beilage nicht erlaubt. ` +
        "Erlaubt sind PDF, TXT, CSV, JSON, XML, ZIP sowie JPG, PNG und WebP."
      );
      if (attachmentInputRef.current) attachmentInputRef.current.value = "";
      return;
    }

    setSaving(true);
    onMessage("");
    (async () => {
      for (const file of uploadFiles) {
        await api.uploadVehicleAttachment(
          selected.id,
          file,
          attachmentUploadCategory || attachmentCategoryForFile(file),
          attachmentUploadDescription,
          ""
        );
      }
    })()
      .then(() => refreshSelectedVehicle(selected.id))
      .catch(handleMediaError)
      .finally(() => {
        setSaving(false);
        if (attachmentInputRef.current) attachmentInputRef.current.value = "";
      });
  };

  const onAttachmentDrag = (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
    event.stopPropagation();
    if (readonly || !selected || saving) return;
    setAttachmentDragActive(event.type === "dragenter" || event.type === "dragover");
  };

  const onAttachmentDrop = (event: DragEvent<HTMLElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setAttachmentDragActive(false);
    if (readonly || !selected || saving) return;
    uploadAttachment(event.dataTransfer.files);
  };

  const updateAttachmentEdit = (
    attachmentId: string,
    patch: Partial<{ description: string; category: string; maintenanceId: string }>
  ) => {
    setAttachmentEdits((current) => ({
      ...current,
      [attachmentId]: {
        ...current[attachmentId],
        description: current[attachmentId]?.description || "",
        category: current[attachmentId]?.category || "",
        maintenanceId: current[attachmentId]?.maintenanceId || "",
        ...patch
      }
    }));
  };

  const saveAttachment = (attachment: VehicleAttachment) => {
    if (!selected || readonly || saving) return;
    if (activeMediaConflicts(metadataRef.current, imageDraft.refs.pendingImagesRef.current, selected)
      .some(value => value.kind === "attachment" && value.id === attachment.id)) {
      onMessage(t("vehicles.mediaConflict.resolve")); return;
    }
    const edit = metadataRef.current.edits[attachment.id] || { description: "", category: "", maintenanceId: "" };
    if (!edit.expectedRevision) { onMessage(t("vehicles.mediaConflict.missingRevision")); return; }
    setSaving(true);
    api.updateVehicleAttachment(selected.id, attachment.id, edit)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch(handleMediaError)
      .finally(() => setSaving(false));
  };

  const deleteAttachment = (attachment: VehicleAttachment) => {
    if (!selected || readonly || saving) return;
    if (!attachment.revision) { onMessage(t("vehicles.mediaConflict.missingRevision")); return; }
    setSaving(true);
    setAttachmentDeleteCandidate(null);
    api.deleteVehicleAttachment(selected.id, attachment.id, attachment.revision)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch(handleMediaError)
      .finally(() => setSaving(false));
  };

  return {
    state: {
      attachmentDeleteCandidate,
      mediaConflicts: activeMediaConflicts(metadata, pendingImages, selected),
      pendingImages,
      previewImage,
      attachmentEdits,
      imageUploadMaintenanceId,
      attachmentUploadCategory,
      attachmentUploadDescription,
      attachmentDragActive
    },
    refs: { imageInputRef, attachmentInputRef },
    setters: {
      setAttachmentDeleteCandidate,
      setPendingImages,
      setPreviewImage,
      setAttachmentEdits,
      setImageUploadMaintenanceId,
      setAttachmentUploadCategory,
      setAttachmentUploadDescription,
      setAttachmentDragActive
    },
    commands: {
      reset,
      loadDetail,
      refreshDetail,
      acknowledgeCoreSaved,
      acknowledgeImportedImage,
      resolveConflict,
      addImages,
      setPrimaryPendingImage,
      updatePendingImageTitle,
      updatePendingImageMaintenance,
      movePendingImage,
      removePendingImage,
      uploadImages,
      uploadAttachment,
      onAttachmentDrag,
      onAttachmentDrop,
      updateAttachmentEdit,
      saveAttachment,
      deleteAttachment
    }
  };
}
