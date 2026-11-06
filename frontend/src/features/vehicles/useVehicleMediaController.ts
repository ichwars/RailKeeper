import { useVehicleImageDraftController } from "./useVehicleImageDraftController";
import { type DragEvent, useRef, useState } from "react";

import { api, type Vehicle, type VehicleAttachment, type VehicleImage } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { reconcileDraft, reconcileImageDraft } from "./vehicleDraftRefresh";
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
  onImageUploadComplete: () => void;
};

export function useVehicleMediaController({
  selected,
  readonly,
  saving,
  setSaving,
  onMessage,
  refreshSelectedVehicle,
  onImageUploadComplete
}: UseVehicleMediaControllerOptions) {
  const { t } = useI18n();
  const imageDraft = useVehicleImageDraftController();
  const { pendingImages, previewImage } = imageDraft.state;
  const { setPendingImages, setPreviewImage } = imageDraft.setters;
  const { addImages, setPrimaryPendingImage, updatePendingImageTitle,
    updatePendingImageMaintenance, movePendingImage } = imageDraft.commands;
  const [attachmentDeleteCandidate, setAttachmentDeleteCandidate] = useState<VehicleAttachment | null>(null);
  const [attachmentEdits, setAttachmentEdits] = useState<AttachmentEditState>({});
  const [imageUploadMaintenanceId, setImageUploadMaintenanceId] = useState("");
  const [attachmentUploadCategory, setAttachmentUploadCategory] = useState("");
  const [attachmentUploadDescription, setAttachmentUploadDescription] = useState("");
  const [attachmentDragActive, setAttachmentDragActive] = useState(false);
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const attachmentInputRef = useRef<HTMLInputElement | null>(null);

  const reset = (clearPreview = false) => {
    imageDraft.commands.reset(clearPreview);
    setAttachmentEdits({});
    setImageUploadMaintenanceId("");
    setAttachmentUploadCategory("");
    setAttachmentUploadDescription("");
    setAttachmentDragActive(false);
    setAttachmentDeleteCandidate(null);
  };

  const loadDetail = (detail: Vehicle) => {
    imageDraft.commands.loadDetail(detail);
    setAttachmentEdits(attachmentsToEditState(detail.attachments));
  };

  const refreshAttachments = (detail: Vehicle, previous: Vehicle) => {
    const base = attachmentsToEditState(previous.attachments), incoming = attachmentsToEditState(detail.attachments);
    setAttachmentEdits(current => ({ ...current, ...Object.fromEntries(Object.entries(incoming).map(([id, value]) => [
      id, current[id] ? reconcileDraft(current[id], base[id] || current[id], value).value : value
    ])) }));
  };

  const refreshDetail = (detail: Vehicle, previous: Vehicle) => {
    setPendingImages(current => reconcileImageDraft(current, previous, detail));
    refreshAttachments(detail, previous);
  };

  const acknowledgeCoreSaved = (detail: Vehicle, previous: Vehicle, phase: "core" | "complete") => {
    setPendingImages(current => [
      ...vehicleImagesToPending(detail),
      ...(phase === "core" ? current.filter(image => !image.persisted && /^https?:\/\//i.test(image.url)) : [])
    ]);
    refreshAttachments(detail, previous);
  };

  const acknowledgeImportedImage = (pending: PendingArticleImage, image: VehicleImage) => {
    setPendingImages(current => current.map(value => value.id === pending.id ? uploadedImageToPending(image)
      : image.isPrimary && value.persisted ? { ...value, isPrimary: false } : value));
  };

  const removePendingImage = (image: PendingArticleImage) => {
    if (image.maintenanceId) {
      onMessage(t("vehicles.uploads.maintenanceLinkedCannotDelete"));
      return;
    }

    const removeFromState = () => imageDraft.commands.removeFromState(image.id);

    if (selected && image.persisted) {
      setSaving(true);
      api.deleteVehicleImage(selected.id, image.id)
        .then(() => {
          removeFromState();
          return refreshSelectedVehicle(selected.id);
        })
        .catch((error: Error) => onMessage(error.message))
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
      .catch((error: Error) => onMessage(error.message))
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
      .catch((error: Error) => onMessage(error.message))
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
        description: current[attachmentId]?.description || "",
        category: current[attachmentId]?.category || "",
        maintenanceId: current[attachmentId]?.maintenanceId || "",
        ...patch
      }
    }));
  };

  const saveAttachment = (attachment: VehicleAttachment) => {
    if (!selected) return;
    const edit = attachmentEdits[attachment.id] || { description: "", category: "", maintenanceId: "" };
    setSaving(true);
    api.updateVehicleAttachment(selected.id, attachment.id, edit)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  const deleteAttachment = (attachment: VehicleAttachment) => {
    if (!selected) return;
    setSaving(true);
    setAttachmentDeleteCandidate(null);
    api.deleteVehicleAttachment(selected.id, attachment.id)
      .then(() => refreshSelectedVehicle(selected.id))
      .catch((error: Error) => onMessage(error.message))
      .finally(() => setSaving(false));
  };

  return {
    state: {
      attachmentDeleteCandidate,
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
