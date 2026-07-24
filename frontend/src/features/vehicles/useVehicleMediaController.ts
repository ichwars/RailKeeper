import { type DragEvent, useRef, useState } from "react";

import { api, type Vehicle, type VehicleAttachment } from "../../shared/api";
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
  const [attachmentDeleteCandidate, setAttachmentDeleteCandidate] = useState<VehicleAttachment | null>(null);
  const [pendingImages, setPendingImages] = useState<PendingArticleImage[]>([]);
  const [previewImage, setPreviewImage] = useState<PendingArticleImage | null>(null);
  const [attachmentEdits, setAttachmentEdits] = useState<AttachmentEditState>({});
  const [imageUploadMaintenanceId, setImageUploadMaintenanceId] = useState("");
  const [attachmentUploadCategory, setAttachmentUploadCategory] = useState("");
  const [attachmentUploadDescription, setAttachmentUploadDescription] = useState("");
  const [attachmentDragActive, setAttachmentDragActive] = useState(false);
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const attachmentInputRef = useRef<HTMLInputElement | null>(null);

  const reset = (clearPreview = false) => {
    setPendingImages([]);
    setAttachmentEdits({});
    setImageUploadMaintenanceId("");
    setAttachmentUploadCategory("");
    setAttachmentUploadDescription("");
    setAttachmentDragActive(false);
    setAttachmentDeleteCandidate(null);
    if (clearPreview) setPreviewImage(null);
  };

  const loadDetail = (detail: Vehicle) => {
    setPendingImages(vehicleImagesToPending(detail));
    setAttachmentEdits(attachmentsToEditState(detail.attachments));
  };

  const addImages = (images: PendingArticleImage[]) => {
    setPendingImages((current) => {
      const existing = new Set(current.map((image) => image.url));
      const next = [...current, ...images.filter((image) => !existing.has(image.url))];
      if (!next.some((image) => image.isPrimary) && next.length > 0) {
        next[0] = { ...next[0], isPrimary: true };
      }
      return next;
    });
  };

  const setPrimaryPendingImage = (id: string) => {
    setPendingImages((current) => current.map((image) => ({ ...image, isPrimary: image.id === id })));
  };

  const updatePendingImageTitle = (id: string, title: string) => {
    setPendingImages((current) => current.map((image) => (image.id === id ? { ...image, title } : image)));
  };

  const updatePendingImageMaintenance = (id: string, maintenanceId: string) => {
    setPendingImages((current) => current.map((image) => (
      image.id === id ? { ...image, maintenanceId } : image
    )));
  };

  const movePendingImage = (id: string, direction: -1 | 1) => {
    setPendingImages((current) => {
      const index = current.findIndex((image) => image.id === id);
      const target = index + direction;
      if (index < 0 || target < 0 || target >= current.length) return current;
      const next = [...current];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  };

  const removePendingImage = (image: PendingArticleImage) => {
    if (image.maintenanceId) {
      onMessage("Bild ist mit einer Wartung verkn?pft. Bitte zuerst die Verkn?pfung entfernen und speichern.");
      return;
    }

    const removeFromState = () => {
      setPendingImages((current) => {
        const next = current.filter((entry) => entry.id !== image.id);
        if (next.length > 0 && !next.some((entry) => entry.isPrimary)) {
          next[0] = { ...next[0], isPrimary: true };
        }
        return next;
      });
    };

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
